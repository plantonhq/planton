package tofumodule

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/failure"
	"github.com/plantonhq/planton/pkg/iac/tofu/generators"
	"github.com/plantonhq/planton/pkg/iac/tofu/tfbackend"
	"github.com/plantonhq/planton/shared/iac/terraform"
	"google.golang.org/protobuf/proto"
)

// Init initializes an HCL module (tofu or terraform) with optional JSON streaming.
// The binaryName parameter specifies which CLI binary to use ("tofu" or "terraform").
//
// ctx controls the lifetime of the child process (see newReapableCommand): cancelling it
// terminates the whole tofu process group so a cancelled init is never orphaned.
//
// backendBody lines, when given, are rendered inside the backend.tf declaration (see
// tfbackend.WriteBackendFile) -- the remote backend's `workspaces { name = ... }` block
// travels this way because HCL blocks cannot ride -backend-config flags.
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
	backendBody ...string,
) (err error) {
	if err := tfbackend.WriteBackendFile(modulePath, backendType, backendBody...); err != nil {
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
	var diagnostics bytes.Buffer
	cmd.Stderr = io.MultiWriter(os.Stderr, &diagnostics)

	// If a channel is provided, stream stdout line-by-line (see
	// streamCommandJSONOutput for the read-before-Wait ordering that avoids a
	// "file already closed" race). An init that fails before its JSON stream
	// starts (a module it cannot parse, a provider it cannot fetch) says why
	// only on stderr, so a failed init carries that text.
	if jsonLogEventsChan != nil {
		err := streamCommandJSONOutput(binaryName, cmd, jsonLogEventsChan)
		return failure.Annotate(withStderr(err, diagnostics.String()), diagnostics.String())
	}

	// Otherwise stream stdout directly to the console.
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to execute %s command %s", binaryName, cmd.String())
	}

	return nil
}
