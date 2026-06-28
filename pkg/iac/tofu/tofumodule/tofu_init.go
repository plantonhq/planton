package tofumodule

import (
	"context"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/iac/terraform"
	"github.com/plantonhq/openmcf/pkg/iac/tofu/generators"
	"github.com/plantonhq/openmcf/pkg/iac/tofu/tfbackend"
	"google.golang.org/protobuf/proto"
)

// Init initializes an HCL module (tofu or terraform) with optional JSON streaming.
// The binaryName parameter specifies which CLI binary to use ("tofu" or "terraform").
//
// ctx controls the lifetime of the child process (see newReapableCommand): cancelling it
// terminates the whole tofu process group so a cancelled init is never orphaned.
func Init(
	ctx context.Context,
	binaryName string,
	modulePath string,
	manifestObject proto.Message,
	backendType terraform.TerraformBackendType,
	backendConfigInput []string,
	providerConfigEnvVars []string,
	isReconfigure bool,
	isJsonOutput bool,
	jsonLogEventsChan chan string,
) (err error) {
	if err := tfbackend.WriteBackendFile(modulePath, backendType); err != nil {
		return errors.Wrapf(err, "failed to write backend file")
	}

	tfVarsFile := filepath.Join(modulePath, ".terraform", "terraform.tfvars")
	if err := generators.WriteVarFile(manifestObject, tfVarsFile); err != nil {
		return errors.Wrapf(err, "failed to write %s file", tfVarsFile)
	}

	// Build the init command
	cmdArgs := []string{
		terraform.TerraformOperationType_init.String(),
		"--var-file", tfVarsFile,
	}
	if isReconfigure {
		cmdArgs = append(cmdArgs, "-reconfigure")
	}
	if isJsonOutput {
		cmdArgs = append(cmdArgs, "-json")
	}
	for _, backendConfig := range backendConfigInput {
		cmdArgs = append(cmdArgs, "--backend-config", backendConfig)
	}

	cmd := newReapableCommand(ctx, binaryName, cmdArgs...)
	cmd.Dir = modulePath
	// https://stackoverflow.com/a/41133244
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, providerConfigEnvVars...)

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	// If a channel is provided, stream stdout line-by-line (see
	// streamCommandJSONOutput for the read-before-Wait ordering that avoids a
	// "file already closed" race).
	if jsonLogEventsChan != nil {
		return streamCommandJSONOutput(binaryName, cmd, jsonLogEventsChan)
	}

	// Otherwise stream stdout directly to the console.
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to execute %s command %s", binaryName, cmd.String())
	}

	return nil
}
