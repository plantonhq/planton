package pulumistack

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/failure"
	"github.com/plantonhq/planton/pkg/iac/iacinput"
	"github.com/plantonhq/planton/pkg/iac/iacinput/iacinputproviderconfig"
	"github.com/plantonhq/planton/pkg/iac/pulumi/backendconfig"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule"
	pulumimoduleiacinput "github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
	"github.com/plantonhq/planton/pkg/kubernetes/execcredential"
	"github.com/plantonhq/planton/pkg/kubernetes/kubeconfig"
	"github.com/plantonhq/planton/shared/iac/pulumi"
	log "github.com/sirupsen/logrus"
	"golang.org/x/term"
)

func Run(moduleDir, stackFqdn, targetManifestPath string, pulumiOperation pulumi.PulumiOperationType,
	isUpdatePreview bool, isAutoApprove bool, valueOverrides map[string]string, showDiff bool, moduleVersion string, noCleanup bool,
	kubeContext string, iacInputFilePath string, providerConfig *iacinputproviderconfig.ProviderConfig,
	opts ...RunOption) error {
	var cfg runConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	manifestObject, err := manifest.LoadWithOverrides(targetManifestPath, valueOverrides)
	if err != nil {
		return errors.Wrapf(err, "failed to override values in target manifest file")
	}

	// Stack selection follows the tofu backend's precedence direction: the
	// --stack flag wins, the manifest annotation fills in when the flag is
	// absent. (The annotation silently overriding an explicit flag was the
	// one inverted precedence in the IaC surface — normalized deliberately.)
	finalStackFqdn := stackFqdn
	if finalStackFqdn == "" {
		if manifestBackendConfig, err := backendconfig.ExtractFromManifest(manifestObject); err == nil && manifestBackendConfig != nil {
			if manifestBackendConfig.StackFqdn != "" {
				cyan := color.New(color.FgCyan).SprintFunc()
				fmt.Printf("\nDetected Stack from Annotations: %s\n\n", cyan(manifestBackendConfig.StackFqdn))
				finalStackFqdn = manifestBackendConfig.StackFqdn
			}
		}
	}

	// Validate that we have a stack FQDN
	if finalStackFqdn == "" {
		return errors.New("Pulumi stack FQDN is required. Provide it via --stack flag or set pulumi.planton.dev/stack.fqdn annotation in manifest")
	}

	// Resolve the state backend URL (flag > annotation > env). When resolved,
	// PULUMI_BACKEND_URL pins the backend for this run and its output reads;
	// when not, pulumi keeps today's behavior — the machine's ambient login.
	backendUrl, backendUrlSource := backendconfig.ResolveBackendURL(manifestObject, cfg.backendUrl)
	if backendUrl != "" {
		cyan := color.New(color.FgCyan).SprintFunc()
		fmt.Printf("Backend URL (%s): %s\n", backendUrlSource, cyan(backendUrl))
	}

	kindName, err := catalogkindreflect.ExtractKindFromProto(manifestObject)
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

	// Determine IaC input file path:
	// - If user provided --iac-input flag, use that file directly
	// - Otherwise, build IaC input from manifest and write to temp file
	var finalIacInputFilePath string
	if iacInputFilePath != "" {
		// User provided a pre-built IaC input file
		finalIacInputFilePath = iacInputFilePath
	} else {
		// Build IaC input from manifest
		iacInputYamlContent, err := iacinput.BuildIacInputYaml(manifestObject, providerConfig)
		if err != nil {
			return errors.Wrap(err, "failed to build IaC input yaml")
		}

		// Write IaC input to file (avoids env var size limits for large manifests)
		finalIacInputFilePath = filepath.Join(pulumiModuleRepoPath, "iac-input.yaml")
		if err := os.WriteFile(finalIacInputFilePath, []byte(iacInputYamlContent), 0600); err != nil {
			return errors.Wrap(err, "failed to write IaC input file")
		}
	}

	// Update project name in Pulumi.yaml
	// For binary mode, we regenerate the Pulumi.yaml with the correct project name
	if err := UpdateProjectNameInPulumiYaml(pulumiModuleRepoPath, pulumiProjectName); err != nil {
		return errors.Wrapf(err, "failed to update project name in %s/Pulumi.yaml", pulumiModuleRepoPath)
	}

	// Map to Pulumi CLI verbs
	op := pulumiOperation.String()
	switch pulumiOperation {
	case pulumi.PulumiOperationType_update:
		op = "up"
	case pulumi.PulumiOperationType_refresh:
		op = "refresh"
	case pulumi.PulumiOperationType_destroy:
		op = "destroy"
	}
	if isUpdatePreview {
		op = "preview"
	}

	// Build pulumi command with optional flags
	args := []string{op, "--stack", finalStackFqdn}
	if isAutoApprove {
		args = append(args, "--yes")
		// For 'pulumi up', skip preview to avoid TTY prompts in CI/non-interactive shells
		if op == "up" {
			args = append(args, "--skip-preview")
		}
	}
	// Destroy-time resource hooks (BeforeDelete/AfterDelete) require the
	// program that registered them; without --run-program Pulumi skips
	// those hooks silently. Kyverno's webhook-GC sentinel depends on this.
	if op == "destroy" {
		args = append(args, "--run-program")
	}
	if showDiff {
		args = append(args, "--diff")
	}

	pulumiCmd := exec.Command("pulumi", args...)

	// extraEnv is shared between the operation itself and the post-update
	// output reads, so capture sees exactly the backend the update used.
	extraEnv := []string{pulumimoduleiacinput.FilePathEnvVar + "=" + finalIacInputFilePath}
	if op == "destroy" {
		// The program runs during destroy only for its delete hooks; it must
		// know that, so steps that can fail for reasons unrelated to what is
		// being deleted stand aside (see OperationEnvVar).
		extraEnv = append(extraEnv, pulumimoduleiacinput.OperationEnvVar+"="+pulumimoduleiacinput.OperationDestroy)
	}
	if backendUrl != "" {
		extraEnv = append(extraEnv, "PULUMI_BACKEND_URL="+backendUrl)
	}

	// Set environment variables
	pulumiCmd.Env = append(os.Environ(), extraEnv...)
	if kubeContext != "" {
		pulumiCmd.Env = append(pulumiCmd.Env, kubeconfig.KubeContextEnvVar+"="+kubeContext)
	}

	// Advertise this binary as the kubeconfig exec-credential command: the module
	// process renders kubeconfigs for managed clusters (EKS/GKE) but cannot know the
	// engine-spawning binary's path on its own. Failure to resolve our own path is
	// only fatal for those providers, so it degrades to a warning here and surfaces
	// as a clear builder error if a kubernetes module actually needs the contract.
	if executable, err := os.Executable(); err == nil {
		pulumiCmd.Env = append(pulumiCmd.Env, execcredential.CommandPathEnvVar+"="+executable)
	} else {
		log.Warnf("could not resolve own executable path for %s: %v", execcredential.CommandPathEnvVar, err)
	}

	// Set the working directory to the repository path
	pulumiCmd.Dir = pulumiModuleRepoPath

	// Stdin stays the terminal so Pulumi can prompt. In a terminal, stdout and
	// stderr are handed straight through: Pulumi checks whether ITS stdout is
	// a TTY to choose the interactive tree view, and a pipe in between would
	// silently downgrade every laptop run to the plain log. Anywhere else (a
	// runner, CI, the e2e harness) the output is streamed AND kept, because a
	// provider's raw text (the API server answering Forbidden, a chart
	// repository that does not resolve) can be all that reaches the reader,
	// and the layer that runs the engine adds what it means and what to do.
	// Pulumi writes diagnostics to stdout, so both streams are kept.
	var output bytes.Buffer
	pulumiCmd.Stdin = os.Stdin
	pulumiCmd.Stdout = os.Stdout
	pulumiCmd.Stderr = os.Stderr
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		pulumiCmd.Stdout = io.MultiWriter(os.Stdout, &output)
		pulumiCmd.Stderr = io.MultiWriter(os.Stderr, &output)
	}

	// Log execution mode and directory info (debug level only)
	if pathResult.UseBinary {
		log.Debugf("execution mode: binary (no compilation)")
		log.Debugf("binary path: %s", pathResult.BinaryPath)
	} else {
		log.Debugf("execution mode: source (compilation required)")
	}
	log.Debugf("workspace directory: %s", pulumiModuleRepoPath)
	fmt.Println()

	// Print handoff message after all setup is complete
	cliprint.PrintHandoff("Pulumi")

	if err := pulumiCmd.Run(); err != nil {
		return failure.Annotate(errors.Wrapf(err, "failed to execute pulumi command %s", op), output.String())
	}

	// Capture must run here, before the deferred workspace cleanup fires:
	// the output reads run in the module workspace with the same backend the
	// update used. Only a real update captures — previews change nothing, and
	// refresh/destroy have no fresh outputs to read.
	if cfg.captureSink != nil && pulumiOperation == pulumi.PulumiOperationType_update && !isUpdatePreview {
		if captureErr := captureOutputs(finalStackFqdn, pulumiModuleRepoPath, kindName,
			extraEnv, cfg.captureSink); captureErr != nil {
			// The update already succeeded; a capture failure must not turn a
			// deployed stack into a failed command. Report and move on.
			log.Warnf("outputs could not be captured after update: %v", captureErr)
		}
	}

	return nil
}
