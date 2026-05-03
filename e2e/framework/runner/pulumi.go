package runner

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// PulumiResult captures the outcome of a Pulumi CLI invocation.
type PulumiResult struct {
	Stdout   string
	Stderr   string
	Duration time.Duration
	ExitCode int
}

// PulumiDeploy runs `pulumi up` for the given module directory and stack.
func PulumiDeploy(moduleDir, stackName, backendURL, stackInputFilePath string) (*PulumiResult, error) {
	if err := pulumiEnsureStack(moduleDir, stackName, backendURL); err != nil {
		return nil, errors.Wrap(err, "failed to ensure pulumi stack exists")
	}

	args := []string{"up", "--stack", stackName, "--yes", "--skip-preview", "--non-interactive"}
	return runPulumi(moduleDir, backendURL, stackInputFilePath, "", args)
}

// PulumiDestroy runs `pulumi destroy` for the given module directory and stack.
func PulumiDestroy(moduleDir, stackName, backendURL, stackInputFilePath string) (*PulumiResult, error) {
	args := []string{"destroy", "--stack", stackName, "--yes", "--non-interactive"}
	return runPulumi(moduleDir, backendURL, stackInputFilePath, "", args)
}

// PulumiRemoveStack removes the stack entirely after destroy.
func PulumiRemoveStack(moduleDir, stackName, backendURL string) error {
	args := []string{"stack", "rm", stackName, "--yes", "--non-interactive"}
	_, err := runPulumi(moduleDir, backendURL, "", "", args)
	return err
}

// PulumiStackOutputs retrieves stack outputs as a raw string.
func PulumiStackOutputs(moduleDir, stackName, backendURL string) (string, error) {
	args := []string{"stack", "output", "--stack", stackName, "--json", "--non-interactive"}
	result, err := runPulumi(moduleDir, backendURL, "", "", args)
	if err != nil {
		return "", err
	}
	return result.Stdout, nil
}

func pulumiEnsureStack(moduleDir, stackName, backendURL string) error {
	// Try selecting the stack first
	selectArgs := []string{"stack", "select", stackName, "--non-interactive"}
	_, err := runPulumi(moduleDir, backendURL, "", "", selectArgs)
	if err == nil {
		return nil
	}

	// Stack doesn't exist -- create it
	initArgs := []string{"stack", "init", stackName, "--non-interactive"}
	_, err = runPulumi(moduleDir, backendURL, "", "", initArgs)
	if err != nil {
		return errors.Wrapf(err, "failed to create stack %s", stackName)
	}
	return nil
}

func runPulumi(moduleDir, backendURL, stackInputFilePath, kubeContext string, args []string) (*PulumiResult, error) {
	cmd := exec.Command("pulumi", args...)
	cmd.Dir = moduleDir

	env := os.Environ()
	if backendURL != "" {
		env = append(env, "PULUMI_BACKEND_URL="+backendURL)
	}
	// Empty passphrase for local file backend (no encryption needed for E2E)
	env = append(env, "PULUMI_CONFIG_PASSPHRASE=")
	if stackInputFilePath != "" {
		env = append(env, "STACK_INPUT_YAML_FILE="+stackInputFilePath)
	}
	if kubeContext != "" {
		env = append(env, "KUBE_CTX="+kubeContext)
	}
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	result := &PulumiResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
	}

	if err != nil {
		combined := strings.TrimSpace(stdout.String() + "\n" + stderr.String())
		return result, errors.Wrapf(err, "pulumi %s failed:\n%s", strings.Join(args, " "), combined)
	}

	return result, nil
}

// PulumiLogin logs into the specified backend. For file backends, this is a no-op
// if PULUMI_BACKEND_URL is already set, but we call it to ensure the backend is initialized.
func PulumiLogin(backendURL string) error {
	cmd := exec.Command("pulumi", "login", backendURL, "--non-interactive")
	cmd.Env = append(os.Environ(), "PULUMI_CONFIG_PASSPHRASE=")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "pulumi login to %s failed: %s", backendURL, stderr.String())
	}
	return nil
}

// GenerateStackName creates a unique, deterministic stack name for an E2E test run.
// Format: e2e-{component}-{shortID}
func GenerateStackName(component string, runID string) string {
	short := runID
	if len(short) > 8 {
		short = short[:8]
	}
	return fmt.Sprintf("e2e-%s-%s", component, short)
}
