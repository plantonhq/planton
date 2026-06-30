package tofumodule

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/apis/dev/planton/shared/iac/terraform"
	"github.com/plantonhq/planton/pkg/iac/tofu/generators"
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

	// Keep stdin/stderr for interactive prompt or error streaming
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	fmt.Printf("%s module directory: %s\n", binaryName, modulePath)
	fmt.Printf("running command: %s\n", cmd.String())

	// If JSON output, stream stdout line-by-line (see streamCommandJSONOutput for
	// the read-before-Wait ordering that avoids a "file already closed" race).
	if isJsonOutput {
		return streamCommandJSONOutput(binaryName, cmd, jsonLogEventsChan)
	}

	// Otherwise stream stdout directly to the console.
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to execute %s command %s", binaryName, cmd.String())
	}

	return nil
}
