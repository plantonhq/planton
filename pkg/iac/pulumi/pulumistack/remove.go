package pulumistack

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/pkg/iac/pulumi/backendconfig"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule"
)

// Remove deletes a Pulumi stack and all its configuration/state from the backend.
// This is a destructive operation that removes the stack metadata.
// Note: This does NOT destroy cloud resources - run 'pulumi destroy' first if needed.
func Remove(moduleDir, stackFqdn, targetManifestPath string, valueOverrides map[string]string, force bool, moduleVersion string, noCleanup bool) error {
	manifestObject, err := manifest.LoadWithOverrides(targetManifestPath, valueOverrides)
	if err != nil {
		return errors.Wrapf(err, "failed to override values in target manifest file")
	}

	// Try to extract backend configuration from manifest labels
	// If found, use it instead of the provided stackFqdn
	finalStackFqdn := stackFqdn
	if manifestBackendConfig, err := backendconfig.ExtractFromManifest(manifestObject); err == nil && manifestBackendConfig != nil {
		if manifestBackendConfig.StackFqdn != "" {
			cyan := color.New(color.FgCyan).SprintFunc()
			fmt.Printf("\nDetected Stack from Labels: %s\n\n", cyan(manifestBackendConfig.StackFqdn))
			finalStackFqdn = manifestBackendConfig.StackFqdn
		}
	}

	// Validate that we have a stack FQDN
	if finalStackFqdn == "" {
		return errors.New("Pulumi stack FQDN is required. Provide it via --stack flag or set pulumi.planton.dev/stack.fqdn label in manifest")
	}

	kindName, err := crkreflect.ExtractKindFromProto(manifestObject)
	if err != nil {
		return errors.Wrapf(err, "failed to extract kind name from manifest proto")
	}

	pathResult, err := pulumimodule.GetPath(moduleDir, finalStackFqdn, kindName, moduleVersion, noCleanup)
	if err != nil {
		return errors.Wrapf(err, "failed to get pulumi-module directory")
	}

	// Setup cleanup to run after execution
	if pathResult.ShouldCleanup {
		defer func() {
			if cleanupErr := pathResult.CleanupFunc(); cleanupErr != nil {
				fmt.Printf("Warning: failed to cleanup workspace copy: %v\n", cleanupErr)
			}
		}()
	}

	pulumiModuleRepoPath := pathResult.ModulePath

	pulumiProjectName, err := ExtractProjectName(finalStackFqdn)
	if err != nil {
		return errors.Wrapf(err, "failed to extract project name from stack fqdn")
	}

	if err := UpdateProjectNameInPulumiYaml(pulumiModuleRepoPath, pulumiProjectName); err != nil {
		return errors.Wrapf(err, "failed to update project name in %s/Pulumi.yaml", pulumiModuleRepoPath)
	}

	// Build pulumi stack rm command
	args := []string{"stack", "rm", "--stack", finalStackFqdn, "--yes"}
	if force {
		// Force removal even if resources exist
		args = append(args, "--force")
	}

	pulumiCmd := exec.Command("pulumi", args...)

	// Set the working directory to the repository path
	pulumiCmd.Dir = pulumiModuleRepoPath

	// Stream to terminal and also capture output for error classification
	buf := &bytes.Buffer{}
	mwOut := io.MultiWriter(os.Stdout, buf)
	mwErr := io.MultiWriter(os.Stderr, buf)

	// Set stdin, stdout, and stderr to the current terminal
	pulumiCmd.Stdin = os.Stdin
	pulumiCmd.Stdout = mwOut
	pulumiCmd.Stderr = mwErr

	fmt.Printf("\npulumi module directory: %s\n", pulumiModuleRepoPath)

	// Print handoff message after all setup is complete
	cliprint.PrintHandoff("Pulumi")

	fmt.Printf("Removing stack: %s\n\n", finalStackFqdn)

	if err := pulumiCmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to remove pulumi stack")
	}

	fmt.Printf("\n✓ Successfully removed stack: %s\n", finalStackFqdn)
	return nil
}
