package localmodule

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/pkg/iac/gitrepo"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule"
	"github.com/plantonhq/planton/pkg/iac/tofu/tofumodule"
	"github.com/spf13/cobra"
)

// GetModuleDir resolves the local module directory from the planton repo.
// It reads the manifest file, extracts the kind, and constructs the module path
// based on the provisioner type.
//
// Parameters:
//   - targetManifestPath: Path to the manifest YAML file
//   - cmd: Cobra command (used to get the local repo path from flags/env)
//   - prov: The IaC provisioner type (pulumi or terraform)
//
// Returns the resolved module directory path or an error with helpful guidance.
func GetModuleDir(targetManifestPath string, cmd *cobra.Command, prov shared.IacProvisioner) (string, error) {
	// Get local repo path first (needed for error messages)
	repoPath := gitrepo.GetLocalRepoPath(cmd)

	// Read manifest file
	manifestBytes, err := os.ReadFile(targetManifestPath)
	if err != nil {
		return "", &Error{
			Stage:   "reading manifest",
			Cause:   err,
			Context: fmt.Sprintf("manifest path: %s", targetManifestPath),
			Hint:    "Verify the manifest file exists and is readable.",
		}
	}

	// Extract kind from YAML (no proto loading needed)
	cloudResourceKind, err := crkreflect.ExtractKindFromYaml(manifestBytes)
	if err != nil {
		return "", &Error{
			Stage:   "detecting resource kind",
			Cause:   err,
			Context: fmt.Sprintf("manifest: %s", targetManifestPath),
			Hint:    "Ensure the manifest has valid 'apiVersion' and 'kind' fields.",
		}
	}
	kindName := crkreflect.ExtractKindNameByKind(cloudResourceKind)

	cliprint.PrintStep(fmt.Sprintf("Using local module from: %s", repoPath))

	// Verify repo path exists
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		return "", &Error{
			Stage:   "locating local repository",
			Cause:   err,
			Context: fmt.Sprintf("repo path: %s", repoPath),
			Hint: fmt.Sprintf("The planton repository was not found at '%s'.\n"+
				"  Options:\n"+
				"  1. Clone the repo: git clone https://github.com/plantonhq/planton %s\n"+
				"  2. Set a different path: --planton-git-repo /your/path\n"+
				"  3. Use environment variable: export PLANTON_GIT_REPO=/your/path",
				repoPath, repoPath),
		}
	}

	// Get module path based on provisioner
	var moduleDir string
	var provName string
	switch prov {
	case shared.IacProvisioner_pulumi:
		provName = "pulumi"
		moduleDir, err = pulumimodule.GetLocalModulePath(repoPath, kindName)
	case shared.IacProvisioner_terraform:
		provName = "terraform"
		moduleDir, err = tofumodule.GetLocalModulePath(repoPath, kindName)
	default:
		return "", &Error{
			Stage:   "selecting provisioner",
			Cause:   errors.New("unsupported provisioner"),
			Context: fmt.Sprintf("provisioner: %s", prov.String()),
			Hint:    "Use 'pulumi' or 'terraform' as the provisioner.",
		}
	}

	if err != nil {
		expectedPath := buildExpectedModulePath(repoPath, kindName, provName)
		return "", &Error{
			Stage:   "resolving module path",
			Cause:   err,
			Context: fmt.Sprintf("kind: %s, provisioner: %s", kindName, provName),
			Hint: fmt.Sprintf("The %s module for '%s' was not found.\n"+
				"  Expected location: %s\n"+
				"  Possible fixes:\n"+
				"  1. Verify the kind '%s' is correct in your manifest\n"+
				"  2. Check if the module exists: ls -la %s\n"+
				"  3. Pull latest changes: cd %s && git pull",
				provName, kindName, expectedPath, kindName, expectedPath, repoPath),
		}
	}

	cliprint.PrintSuccess(fmt.Sprintf("Local module path: %s", moduleDir))
	return moduleDir, nil
}

// buildExpectedModulePath constructs the expected module path for error messages
func buildExpectedModulePath(repoPath, kindName, provName string) string {
	// Get provider from kind name (e.g., KubernetesNats -> kubernetes)
	provider := strings.ToLower(kindName)
	for _, p := range []string{"kubernetes", "aws", "gcp", "azure", "civo", "digitalocean", "cloudflare"} {
		if strings.HasPrefix(strings.ToLower(kindName), p) {
			provider = p
			break
		}
	}

	subdir := "pulumi"
	if provName == "terraform" {
		subdir = "tf"
	}

	return filepath.Join(repoPath, "apis/dev/planton/provider", provider,
		strings.ToLower(kindName), "v1/iac", subdir)
}

// Error provides structured error information for better UX
type Error struct {
	Stage   string // What step failed
	Cause   error  // Underlying error
	Context string // Additional context
	Hint    string // Actionable suggestion
}

func (e *Error) Error() string {
	return fmt.Sprintf("local module resolution failed at '%s': %v", e.Stage, e.Cause)
}

// PrintError outputs a user-friendly error message with guidance
func (e *Error) PrintError() {
	cliprint.PrintError(fmt.Sprintf("Failed to resolve local module (%s)", e.Stage))
	if e.Context != "" {
		fmt.Printf("  Context: %s\n", e.Context)
	}
	if e.Hint != "" {
		fmt.Printf("\n  %s\n", e.Hint)
	}
}
