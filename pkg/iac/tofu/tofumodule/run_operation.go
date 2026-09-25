package tofumodule

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/failure"
	"github.com/plantonhq/planton/pkg/iac/tofu/generators"
	"github.com/plantonhq/planton/shared/iac/terraform"
	"google.golang.org/protobuf/proto"
)

// RunOperation runs an HCL-based IaC command (tofu or terraform), optionally adding -json flag
// and streaming output lines. It recovers from any panic in the stdout-reading goroutine.
// The binaryName parameter specifies which CLI binary to use ("tofu" or "terraform").
//
// ctx controls the lifetime of the child process: cancelling it terminates the entire
// tofu process group (see newReapableCommand), so a cancelled/superseded stack job never
// orphans a tofu that would keep holding the state lock.
func RunOperation(
	ctx context.Context,
	binaryName string,
	modulePath string,
	terraformOperation terraform.TerraformOperationType,
	isAutoApprove bool,
	isDestroyPlan bool,
	manifestObject proto.Message,
	providerConfigEnvVars []string,
	isJsonOutput bool,
	jsonLogEventsChan chan string,
) (err error) {
	// Write or update terraform.tfvars
	tfVarsFile := filepath.Join(modulePath, ".terraform", "terraform.tfvars")
	if err := generators.WriteVarFile(manifestObject, tfVarsFile); err != nil {
		return errors.Wrapf(err, "failed to write %s file", tfVarsFile)
	}

	// Determine command and arguments
	op := terraformOperation.String()
	args := []string{op, "--var-file", tfVarsFile}

	if terraformOperation == terraform.TerraformOperationType_plan {
		args = append(args, "--out", "terraform.tfplan")
		if isDestroyPlan {
			args = append(args, "--destroy")
		}
	}

	// Add --auto-approve if needed
	if (terraformOperation == terraform.TerraformOperationType_apply ||
		terraformOperation == terraform.TerraformOperationType_destroy) && isAutoApprove {
		args = append(args, "--auto-approve")
	}

	// If the caller wants JSON output, add the -json flag
	if isJsonOutput {
		args = append(args, "-json")
	}

	cmd := newReapableCommand(ctx, binaryName, args...)
	cmd.Dir = modulePath
	// https://stackoverflow.com/a/41133244
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, providerConfigEnvVars...)

	// Keep stdin for interactive prompts. Diagnostics stream to the terminal
	// as the engine prints them AND are kept, because some failures reach the
	// user only as a provider's raw text (a repository host that does not
	// resolve, an API server answering Forbidden at apply): nothing inside a
	// module can rephrase those, so the layer that runs the engine reads
	// them afterwards and adds what they mean and what to do.
	cmd.Stdin = os.Stdin
	var diagnostics bytes.Buffer
	cmd.Stderr = io.MultiWriter(os.Stderr, &diagnostics)

	fmt.Printf("%s module directory: %s\n", binaryName, modulePath)
	fmt.Printf("running command: %s\n", cmd.String())

	// If JSON output, stream stdout line-by-line (see streamCommandJSONOutput for
	// the read-before-Wait ordering that avoids a "file already closed" race).
	// The JSON consumer receives the engine's diagnostics as events; a failure
	// the engine raises before its JSON stream starts exists only on stderr,
	// so a failed command carries that text too.
	if isJsonOutput {
		err := streamCommandJSONOutput(binaryName, cmd, jsonLogEventsChan)
		return failure.Annotate(withStderr(err, diagnostics.String()), diagnostics.String())
	}

	// Otherwise stream stdout directly to the console.
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return failure.Annotate(
			errors.Wrapf(err, "failed to execute %s command %s", binaryName, cmd.String()),
			diagnostics.String())
	}

	return nil
}
