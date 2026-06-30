package tofumodule

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
	"github.com/plantonhq/planton/internal/cli/workspace"
	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/pkg/fileutil"
	"github.com/plantonhq/planton/pkg/iac/tofu/tofuzip"
)

// GetModulePathResult contains the module path and a cleanup function
type GetModulePathResult struct {
	ModulePath    string
	RepoPath      string
	CleanupFunc   func() error
	ShouldCleanup bool
}

// GetModulePath returns the path to the Terraform/OpenTofu module directory.
// If moduleDir is provided and is a valid Terraform module directory, it returns that.
// Otherwise, it tries to download the module zip from GitHub releases (fast path).
// If zip download fails or is unavailable, it falls back to the staging area (clone repo).
// If moduleVersion is provided, it uses that version for the download/checkout.
// The returned GetModulePathResult includes a cleanup function that should be called after execution
// unless noCleanup is true.
func GetModulePath(moduleDir, kindName, moduleVersion string, noCleanup bool) (*GetModulePathResult, error) {

	// If the module directory is provided, check if it is a valid terraform module directory
	if moduleDir != "" {
		isTerraformModuleDir, err := isTerraformModuleDirectory(moduleDir)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to check if %s is a valid terraform module directory", moduleDir)
		}

		// If the module directory is a valid terraform module directory, return the module directory
		if isTerraformModuleDir {
			return &GetModulePathResult{
				ModulePath:    moduleDir,
				RepoPath:      moduleDir,
				CleanupFunc:   func() error { return nil },
				ShouldCleanup: false,
			}, nil
		}
	}

	// Determine target version - use moduleVersion if provided, otherwise CLI version
	targetVersion := moduleVersion
	if targetVersion == "" && version.Version != "" && version.Version != version.DefaultVersion {
		targetVersion = version.Version
	}

	// Try zip download mode first (fast path)
	// Only available for released versions, not dev builds
	if targetVersion != "" && tofuzip.CanUseZipMode() {
		result, err := tryZipDownload(kindName, targetVersion)
		if err != nil {
			// Log warning and fall back to staging
			cliprint.PrintWarning(fmt.Sprintf("Zip download failed, falling back to staging: %v", err))
		} else if result != nil {
			return result, nil
		}
	}

	// Fall back to staging mode (clone repo)
	return getModulePathFromStaging(kindName, moduleVersion, noCleanup)
}

// tryZipDownload attempts to download the Terraform module zip from GitHub releases.
// Returns nil, nil if zip mode is not available (indicating fallback should be used).
// Returns result, nil on success.
// Returns nil, error on failure (indicating fallback should be used with a warning).
func tryZipDownload(kindName, releaseVersion string) (*GetModulePathResult, error) {
	cliprint.PrintStep(fmt.Sprintf("Downloading Terraform module for %s...", kindName))

	// Ensure module is downloaded/cached
	modulePath, err := tofuzip.EnsureModule(kindName, releaseVersion)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to download module zip for %s", kindName)
	}

	// Verify the module path has .tf files
	isTerraformModuleDir, err := isTerraformModuleDirectory(modulePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to verify downloaded module at %s", modulePath)
	}

	if !isTerraformModuleDir {
		return nil, errors.Errorf("downloaded module at %s does not contain .tf files", modulePath)
	}

	// Zip mode doesn't need cleanup (cached for future use)
	return &GetModulePathResult{
		ModulePath:    modulePath,
		RepoPath:      modulePath,
		CleanupFunc:   func() error { return nil },
		ShouldCleanup: false,
	}, nil
}

// getModulePathFromStaging is the fallback path that clones the repo to staging.
// This is used when zip download is not available or fails.
func getModulePathFromStaging(kindName, moduleVersion string, noCleanup bool) (*GetModulePathResult, error) {
	// Get the tofu workspace directory
	tofuModuleWorkspaceDir, err := getWorkspaceDir()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get tofu module workspace directory")
	}

	// Determine target version - use CLI version if set and not default
	targetVersion := ""
	if version.Version != "" && version.Version != version.DefaultVersion {
		targetVersion = version.Version
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

	// Copy from staging to tofu workspace
	cliprint.PrintStep("Copying modules to workspace...")
	terraformModuleRepoPath, err := staging.CopyToWorkspace(tofuModuleWorkspaceDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to copy staging to workspace")
	}
	cliprint.PrintSuccess("Modules copied to workspace")

	// If moduleVersion is specified, checkout that version in the workspace copy
	if moduleVersion != "" {
		if err := staging.CheckoutVersionInWorkspace(terraformModuleRepoPath, moduleVersion); err != nil {
			// Clean up on error
			_ = staging.CleanupWorkspaceCopy(terraformModuleRepoPath)
			return nil, errors.Wrapf(err, "failed to checkout module version %s", moduleVersion)
		}
	}

	terraformModulePath, err := getTerraformModulePath(terraformModuleRepoPath, kindName)
	if err != nil {
		// Clean up on error
		_ = staging.CleanupWorkspaceCopy(terraformModuleRepoPath)
		return nil, errors.Wrapf(err, "failed to get terraform module path for %s", kindName)
	}

	// Create cleanup function
	cleanupFunc := func() error {
		return staging.CleanupWorkspaceCopy(terraformModuleRepoPath)
	}

	return &GetModulePathResult{
		ModulePath:    terraformModulePath,
		RepoPath:      terraformModuleRepoPath,
		CleanupFunc:   cleanupFunc,
		ShouldCleanup: !noCleanup,
	}, nil
}

// GetModulePathLegacy is the legacy function signature for backward compatibility.
// It calls GetModulePath with noCleanup=false and no moduleVersion, returns just the module path.
// Note: This does not perform cleanup - callers should migrate to GetModulePath for proper cleanup handling.
func GetModulePathLegacy(moduleDir, kindName string) (string, error) {
	result, err := GetModulePath(moduleDir, kindName, "", false)
	if err != nil {
		return "", err
	}
	return result.ModulePath, nil
}

// GetLocalModulePath constructs the Terraform module path from a local planton git repository.
// It takes the local repo path and kindName, and returns the path to the Terraform module.
// This is used when --local-module flag is specified.
func GetLocalModulePath(localRepoPath, kindName string) (string, error) {
	return getTerraformModulePath(localRepoPath, kindName)
}

// IsTerraformModuleDirectory checks if the given directory contains any files with .tf extension.
// It returns true if any .tf file exists, false otherwise. If an error occurs during the check, it returns an error.
func isTerraformModuleDirectory(moduleDir string) (bool, error) {
	entries, err := os.ReadDir(moduleDir)
	if err != nil {
		return false, errors.Wrapf(err, "failed to read directory %s", moduleDir)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tf") {
			return true, nil
		}
	}
	return false, nil
}

func getTerraformModulePath(moduleRepoDir, kindName string) (string, error) {
	kind := crkreflect.KindFromString(kindName)
	kindProvider := crkreflect.GetProvider(kind)
	if kindProvider == cloudresourcekind.CloudResourceProvider_cloud_resource_provider_unspecified {
		return "", errors.New("failed to get kind provider")
	}

	kindDirPath := filepath.Join(
		moduleRepoDir,
		"apis/dev/planton/provider",
		strings.ReplaceAll(kindProvider.String(), "_", ""))

	terraformModulePath := filepath.Join(
		kindDirPath,
		strings.ToLower(kindName),
		"v1/iac/tf",
	)

	if _, err := os.Stat(terraformModulePath); os.IsNotExist(err) {
		return "", errors.Wrapf(err, "failed to get %s module directory", kindName)
	}

	return terraformModulePath, nil
}

// getWorkspaceDir returns the path of the workspace directory to which terraform module repo can be cloned.
func getWorkspaceDir() (string, error) {
	cliWorkspaceDir, err := workspace.GetWorkspaceDir()
	if err != nil {
		return "", errors.Wrapf(err, "failed to get cli workspace directory")
	}
	//base directory will always be ${HOME}/.planton/tofu
	tofuModuleWorkspaceDir := filepath.Join(cliWorkspaceDir, "tofu")
	if !fileutil.IsDirExists(tofuModuleWorkspaceDir) {
		if err := os.MkdirAll(tofuModuleWorkspaceDir, os.ModePerm); err != nil {
			return "", errors.Wrapf(err, "failed to ensure %s dir", tofuModuleWorkspaceDir)
		}
	}
	return tofuModuleWorkspaceDir, nil
}
