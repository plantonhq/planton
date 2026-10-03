package runner

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/iacinput"
)

// PulumiResult captures the outcome of a Pulumi CLI invocation.
type PulumiResult struct {
	Stdout   string
	Stderr   string
	Duration time.Duration
	ExitCode int
}

// PulumiDeploy runs `pulumi up` for the given module directory and stack.
func PulumiDeploy(moduleDir, stackName, backendURL, iacInputFilePath string) (*PulumiResult, error) {
	if err := pulumiEnsureStack(moduleDir, stackName, backendURL); err != nil {
		return nil, errors.Wrap(err, "failed to ensure pulumi stack exists")
	}

	// --refresh reconciles Pulumi state with the live cloud before applying.
	// Dependency stacks are keyed by run id, so every scenario in a run reuses
	// the same stack name; if an earlier scenario's teardown half-completed,
	// the stale state would otherwise make this up a silent no-op while the
	// actual infra component is gone.
	args := []string{"up", "--stack", stackName, "--yes", "--skip-preview", "--non-interactive", "--refresh"}
	return runPulumi(moduleDir, backendURL, iacInputFilePath, "", args)
}

// PulumiPreviewExpectNoChanges re-plans the just-applied stack and fails if
// any change is still pending — the Pulumi arm of the IDEMPOTENCY phase. No
// --refresh here: the point is whether the program's desired state matches
// what the apply recorded, not whether the cloud drifted in the seconds since.
func PulumiPreviewExpectNoChanges(moduleDir, stackName, backendURL, iacInputFilePath string) (*PulumiResult, error) {
	args := []string{"preview", "--stack", stackName, "--expect-no-changes", "--non-interactive"}
	result, err := runPulumi(moduleDir, backendURL, iacInputFilePath, "", args)
	if err != nil {
		return result, errors.Wrap(err, "pulumi preview reported pending changes after apply (idempotency violation)")
	}
	return result, nil
}

// PulumiDestroy runs `pulumi destroy` for the given module directory and stack.
// --run-program keeps the program available so BeforeDelete/AfterDelete
// resource hooks fire (e.g. Kyverno's webhook-GC sentinel). Without it,
// delete hooks are silently skipped and destroy can leave cluster-scoped
// residue the module intended to clean up.
func PulumiDestroy(moduleDir, stackName, backendURL, iacInputFilePath string) (*PulumiResult, error) {
	args := []string{"destroy", "--stack", stackName, "--yes", "--non-interactive", "--run-program"}
	return runPulumi(moduleDir, backendURL, iacInputFilePath, "", args)
}

// PulumiRemoveStack removes the stack entirely after destroy.
func PulumiRemoveStack(moduleDir, stackName, backendURL string) error {
	args := []string{"stack", "rm", stackName, "--yes", "--non-interactive"}
	_, err := runPulumi(moduleDir, backendURL, "", "", args)
	return err
}

// PulumiOutputs retrieves outputs as a raw string.
//
// --show-secrets is load-bearing: without it Pulumi masks every secret
// output as the literal string "[secret]", which corrupts a sensitive
// SCALAR output silently (the sentinel populates the proto field and
// counts as verified) and fails a sensitive MAP output's JSON parse
// (first live hit: a name-keyed authorization_keys map). The Terraform
// path reads real values (`terraform output -json` includes sensitive
// outputs), so the flag keeps both engines' verification bar identical
// -- and dependency-output capture rides this same helper, where a
// consumer's value_from reference to a fixture's sensitive output would
// otherwise resolve to the sentinel and deploy silently wrong. No
// caller prints raw output values (counts and names only), so the
// unmasked values never reach logs.
func PulumiOutputs(moduleDir, stackName, backendURL string) (string, error) {
	args := []string{"stack", "output", "--stack", stackName, "--json", "--non-interactive", "--show-secrets"}
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

func runPulumi(moduleDir, backendURL, iacInputFilePath, kubeContext string, args []string) (*PulumiResult, error) {
	cmd := exec.Command("pulumi", args...)
	cmd.Dir = moduleDir

	env := os.Environ()
	if backendURL != "" {
		env = append(env, "PULUMI_BACKEND_URL="+backendURL)
	}
	// Empty passphrase for local file backend (no encryption needed for E2E)
	env = append(env, "PULUMI_CONFIG_PASSPHRASE=")
	if iacInputFilePath != "" {
		env = append(env, "IAC_INPUT_YAML_FILE="+iacInputFilePath)
	}
	if kubeContext != "" {
		env = append(env, "KUBE_CTX="+kubeContext)
	}
	// The same operation signal the platform's runner gives module programs:
	// a destroy re-runs the program for its delete hooks only, and steps that
	// can fail for unrelated reasons stand aside (iacinput.OperationEnvVar).
	if len(args) > 0 && args[0] == "destroy" {
		env = append(env, iacinput.OperationEnvVar+"="+iacinput.OperationDestroy)
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

// maxStackNameLen caps generated stack names. The cap itself is a
// conservative budget (Pulumi accepts up to 100 characters); what matters is
// HOW the cap is enforced -- see GenerateStackName.
const maxStackNameLen = 50

// GenerateStackName creates a unique, deterministic stack name for an E2E
// test run. Format: e2e-{label}-{shortID}, capped at maxStackNameLen.
//
// The cap is enforced by replacing the tail with a short hash of the FULL
// composed name -- never by blind truncation. Labels carry uniquifying
// suffixes at the END (an install profile's per-document index, the scenario
// name, the run id), so a blind prefix cut collapses distinct stacks into
// one shared name, and every `pulumi up` after the first silently REPLACES
// the previous document's resources in that shared stack (live-caught: a
// four-document subnet install profile deployed as one stack, each subnet
// deleting its predecessor -- the gateway's subnet was gone by deploy time).
// The hash keeps distinct full names distinct, and the derivation stays
// deterministic, so per-run stack reuse across scenarios is unaffected.
func GenerateStackName(label string, runID string) string {
	short := runID
	if len(short) > 8 {
		short = short[:8]
	}
	name := fmt.Sprintf("e2e-%s-%s", label, short)
	if len(name) <= maxStackNameLen {
		return name
	}
	digest := sha256.Sum256([]byte(name))
	return name[:maxStackNameLen-9] + "-" + hex.EncodeToString(digest[:4])
}

// boundStackName truncates a stack name to maxStackNameLen while preserving
// uniqueness: the head stays readable and the tail is replaced by a short
// stable hash of the FULL name. A plain head-truncate is a correctness bug,
// not a cosmetic one -- it cuts off exactly the disambiguating parts (a
// dependency's manifest-name tail, the run-id suffix), collapsing distinct
// stacks onto one name.
func boundStackName(name string) string {
	if len(name) <= maxStackNameLen {
		return name
	}
	sum := sha256.Sum256([]byte(name))
	return name[:maxStackNameLen-9] + "-" + hex.EncodeToString(sum[:4])
}
