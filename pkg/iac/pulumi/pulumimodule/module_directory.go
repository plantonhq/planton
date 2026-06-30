package pulumimodule

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/internal/cli/staging"
	"github.com/plantonhq/planton/internal/cli/version"
	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/pkg/fileutil"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumibinary"
)

// GetPathResult contains the module path and a cleanup function
type GetPathResult struct {
	ModulePath    string
	RepoPath      string
	BinaryPath    string // Path to the pre-built binary (if using binary mode)
	CleanupFunc   func() error
	ShouldCleanup bool
	UseBinary     bool // Indicates if binary mode is being used
}

// GetPath returns the path to the Pulumi module directory.
// If moduleDir is provided and is a valid Pulumi module directory, it returns that.
// Otherwise, it tries to use pre-built binaries (faster, no compilation needed).
// If binary download fails, it falls back to the staging approach (clone + compile).
// If moduleVersion is provided, it checks out that version (tag, branch, or commit SHA) in the workspace copy.
// The returned GetPathResult includes a cleanup function that should be called after execution
// unless noCleanup is true.
func GetPath(moduleDir string, stackFqdn, kindName string, moduleVersion string, noCleanup bool) (*GetPathResult, error) {
	isPulumiModuleDir, err := IsPulumiModuleDirectory(moduleDir)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to check if %s is a valid pulumi module directory", moduleDir)
	}
	if isPulumiModuleDir {
		// User provided a valid module directory, use it directly
		return &GetPathResult{
			ModulePath:    moduleDir,
			RepoPath:      moduleDir,
			CleanupFunc:   func() error { return nil },
			ShouldCleanup: false,
			UseBinary:     false,
		}, nil
	}

	// Determine target version for binary download:
	// 1. If moduleVersion is specified, use that (e.g., "v0.3.9" or "v0.3.1-pulumi-awsecsservice-20260107.01")
	// 2. Otherwise, use CLI version if set and not default
	cliVersion := ""
	if version.Version != "" && version.Version != version.DefaultVersion {
		cliVersion = version.Version
	}

	// Determine which version to use for binary download
	binaryVersion := moduleVersion
	if binaryVersion == "" {
		binaryVersion = cliVersion
	}

	// Try binary approach first (if we have a version)
	// Binary approach is faster and doesn't require Go toolchain
	if binaryVersion != "" {
		result, err := tryBinaryApproach(stackFqdn, kindName, binaryVersion, noCleanup)
		if err == nil {
			return result, nil
		}
		// Log warning and fall through to legacy approach
		cliprint.PrintWarning(fmt.Sprintf("Binary download not available, falling back to source: %v", err))
	}

	// Fall back to staging/clone approach
	return getPathFromStaging(stackFqdn, kindName, cliVersion, moduleVersion, noCleanup)
}

// tryBinaryApproach attempts to use pre-built binary for execution.
// releaseVersion can be:
// - CLI version like "v0.3.2" (downloads from main planton release)
// - Module version like "v0.3.1-pulumi-awsecsservice-20260107.01" (downloads from component-specific release)
// Returns an error if binary is not available or download fails.
func tryBinaryApproach(stackFqdn, kindName, releaseVersion string, noCleanup bool) (*GetPathResult, error) {
	cliprint.PrintStep(fmt.Sprintf("Checking for Pulumi module: %s (%s)...", kindName, releaseVersion))

	// Ensure binary is downloaded and cached
	binaryPath, err := pulumibinary.EnsureBinary(kindName, releaseVersion)
	if err != nil {
		return nil, errors.Wrap(err, "failed to ensure binary")
	}

	// Setup workspace with Pulumi.yaml pointing to the binary
	cliprint.PrintStep("Setting up module workspace...")
	workspaceResult, err := pulumibinary.SetupBinaryWorkspace(binaryPath, stackFqdn, kindName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to setup binary workspace")
	}

	cliprint.PrintSuccess("Module workspace ready")

	return &GetPathResult{
		ModulePath:    workspaceResult.WorkspacePath,
		BinaryPath:    workspaceResult.BinaryPath,
		CleanupFunc:   workspaceResult.CleanupFunc,
		ShouldCleanup: !noCleanup && workspaceResult.ShouldCleanup,
		UseBinary:     true,
	}, nil
}

// getPathFromStaging uses the legacy staging/clone approach to get the module path.
// This is the fallback when binary approach is not available.
func getPathFromStaging(stackFqdn, kindName, targetVersion, moduleVersion string, noCleanup bool) (*GetPathResult, error) {
	stackWorkspaceDir, err := getWorkspaceDir(stackFqdn)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get %s stack workspace directory", stackFqdn)
	}

	// Ensure staging is set up with the correct version
	cliprint.PrintStep("Ensuring staging area is ready...")
	if err := staging.EnsureStaging(targetVersion); err != nil {
		return nil, errors.Wrap(err, "failed to ensure staging area")
	}
	// Get and display current staging version
	stagingVersion, _ := staging.GetCurrentStagingVersion()
	if stagingVersion != "" {
		cliprint.PrintSuccess(fmt.Sprintf("Staging area ready (modules version: %s)", stagingVersion))
	} else {
		cliprint.PrintSuccess("Staging area ready")
	}

	// Copy from staging to stack workspace
	cliprint.PrintStep("Copying modules to stack workspace...")
	pulumiModuleRepoPath, err := staging.CopyToWorkspace(stackWorkspaceDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to copy staging to workspace")
	}
	cliprint.PrintSuccess("Modules copied to workspace")

	// If moduleVersion is specified, checkout that version in the workspace copy
	if moduleVersion != "" {
		if err := staging.CheckoutVersionInWorkspace(pulumiModuleRepoPath, moduleVersion); err != nil {
			// Clean up on error
			_ = staging.CleanupWorkspaceCopy(pulumiModuleRepoPath)
			return nil, errors.Wrapf(err, "failed to checkout module version %s", moduleVersion)
		}
	}

	pulumiModulePath, err := getPulumiModulePath(pulumiModuleRepoPath, kindName)
	if err != nil {
		// Clean up on error
		_ = staging.CleanupWorkspaceCopy(pulumiModuleRepoPath)
		return nil, errors.Wrapf(err, "failed to get pulumi module path for %s", kindName)
	}

	// Create cleanup function
	cleanupFunc := func() error {
		return staging.CleanupWorkspaceCopy(pulumiModuleRepoPath)
	}

	return &GetPathResult{
		ModulePath:    pulumiModulePath,
		RepoPath:      pulumiModuleRepoPath,
		CleanupFunc:   cleanupFunc,
		ShouldCleanup: !noCleanup,
		UseBinary:     false,
	}, nil
}

// GetPathLegacy is the legacy function signature for backward compatibility.
// It calls GetPath with noCleanup=false and no moduleVersion, returns just the module path.
// Note: This does not perform cleanup - callers should migrate to GetPath for proper cleanup handling.
func GetPathLegacy(moduleDir string, stackFqdn, kindName string) (string, error) {
	result, err := GetPath(moduleDir, stackFqdn, kindName, "", false)
	if err != nil {
		return "", err
	}
	return result.ModulePath, nil
}

// GetLocalModulePath constructs the Pulumi module path from a local planton git repository.
// It takes the local repo path and kindName, and returns the path to the Pulumi module.
// This is used when --local-module flag is specified.
func GetLocalModulePath(localRepoPath, kindName string) (string, error) {
	return getPulumiModulePath(localRepoPath, kindName)
}

// IsPulumiModuleDirectory checks if the given directory contains a Pulumi.yaml file.
// It returns true if the file exists, false otherwise. If an error occurs during the check, it returns an error.
func IsPulumiModuleDirectory(moduleDir string) (bool, error) {
	pulumiYamlPath := moduleDir + "/Pulumi.yaml"
	isExists, err := fileutil.IsExists(pulumiYamlPath)
	if err != nil {
		return false, errors.Wrapf(err, "failed to check if %s exists", pulumiYamlPath)
	}
	return isExists, nil
}

func getPulumiModulePath(moduleRepoDir, kindName string) (string, error) {
	kind := crkreflect.KindFromString(kindName)
	kindProvider := crkreflect.GetProvider(kind)
	if kindProvider == cloudresourcekind.CloudResourceProvider_cloud_resource_provider_unspecified {
		return "", errors.New("failed to get kind provider")
	}

	kindDirPath := filepath.Join(
		moduleRepoDir,
		"apis/dev/planton/provider",
		strings.ReplaceAll(kindProvider.String(), "_", ""))

	pulumiModulePath := filepath.Join(
		kindDirPath,
		strings.ToLower(kindName),
		"v1/iac/pulumi",
	)

	if _, err := os.Stat(pulumiModulePath); os.IsNotExist(err) {
		return "", errors.Wrapf(err, "failed to get %s module directory", kindName)
	}

	return pulumiModulePath, nil
}

// getWorkspaceDir returns the path of the workspace directory to be used while initializing stack using automation api.
// For source mode (staging/clone), workspaces are under: ~/.planton/pulumi/staging-workspaces/{stack-fqdn}
func getWorkspaceDir(stackFqdn string) (string, error) {
	pulumiBaseDir, err := pulumibinary.GetPulumiBaseDir()
	if err != nil {
		return "", errors.Wrapf(err, "failed to get Pulumi base directory")
	}
	// Use "staging-workspaces" to differentiate from binary "workspaces"
	stackWorkspaceDir := filepath.Join(pulumiBaseDir, "staging-workspaces", stackFqdn)
	if !fileutil.IsDirExists(stackWorkspaceDir) {
		if err := os.MkdirAll(stackWorkspaceDir, os.ModePerm); err != nil {
			return "", errors.Wrapf(err, "failed to ensure %s dir", stackWorkspaceDir)
		}
	}
	return stackWorkspaceDir, nil
}
