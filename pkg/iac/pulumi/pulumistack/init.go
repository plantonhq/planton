package pulumistack

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/pkg/iac/pulumi/backendconfig"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule"
)

// Init initializes a new Pulumi stack for the given manifest.
// It extracts the stack FQDN from the manifest, prepares the module directory,
// and runs `pulumi stack init` to create the stack in the backend.
func Init(moduleDir, stackFqdn, targetManifestPath string, valueOverrides map[string]string, moduleVersion string, noCleanup bool) error {
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
		return errors.Wrapf(err, "failed to extract project name from %s stack fqdn", finalStackFqdn)
	}

	if err := UpdateProjectNameInPulumiYaml(pulumiModuleRepoPath, pulumiProjectName); err != nil {
		return errors.Wrapf(err, "failed to update project name in %s/Pulumi.yaml", pulumiModuleRepoPath)
	}

	// Build pulumi stack init command
	args := []string{"stack", "init", finalStackFqdn}

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

	fmt.Printf("Initializing stack: %s\n\n", finalStackFqdn)

	if err := pulumiCmd.Run(); err != nil {
		// Check if the error is because stack already exists
		out := buf.String()
		if strings.Contains(out, "already exists") || strings.Contains(out, "stack already exists") {
			fmt.Printf("\n✓ Stack '%s' already exists, skipping initialization\n", finalStackFqdn)
			return nil
		}
		return errors.Wrapf(err, "failed to initialize pulumi stack")
	}

	fmt.Printf("\n✓ Successfully initialized stack: %s\n", finalStackFqdn)
	return nil
}
