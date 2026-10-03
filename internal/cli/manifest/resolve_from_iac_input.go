package manifest

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/plantonhq/planton/internal/cli/flag"
	"github.com/plantonhq/planton/pkg/clipboard"
	"github.com/plantonhq/planton/pkg/iac/iacinput"
)

// clipboardFlagValues contains all flag names that indicate clipboard input.
var clipboardFlagValues = []string{"--clipboard", "--clip", "--cb", "-c"}

// resolveFromIacInput checks for --iac-input flag and extracts manifest from it.
// Returns empty string if flag not provided.
// When an IaC input file is provided, the manifest is extracted from the "target" field
// and written to a temporary file.
//
// Special case: If the flag value is a clipboard flag name (e.g., "-i --clip"),
// this indicates the user wants to read IaC input from clipboard.
func resolveFromIacInput(cmd *cobra.Command) (manifestPath string, isTemp bool, err error) {
	// Not every command tree registers --iac-input (it is a pulumi-tree
	// flag); for commands that don't, this source is simply absent rather
	// than an error -- otherwise every tofu/terraform command that resolves
	// a manifest would crash before reading --manifest.
	if cmd.Flags().Lookup(string(flag.IacInput)) == nil {
		return "", false, nil
	}

	iacInputPath, err := cmd.Flags().GetString(string(flag.IacInput))
	if err != nil {
		return "", false, errors.Wrap(err, "failed to get iac-input flag")
	}

	if iacInputPath == "" {
		return "", false, nil
	}

	// Detect clipboard flag captured as -i value (e.g., "-i --clip")
	// User intent is clear: read IaC input from clipboard
	if isClipboardFlagValue(iacInputPath) {
		return resolveIacInputFromClipboard()
	}

	manifestPath, err = iacinput.ExtractManifestFromIacInput(iacInputPath)
	if err != nil {
		return "", false, errors.Wrapf(err, "failed to extract manifest from IaC input %s", iacInputPath)
	}

	return manifestPath, true, nil
}

// isClipboardFlagValue checks if the value matches a clipboard flag name.
func isClipboardFlagValue(value string) bool {
	for _, f := range clipboardFlagValues {
		if value == f {
			return true
		}
	}
	return false
}

// resolveIacInputFromClipboard reads IaC input YAML from clipboard
// and extracts the manifest from the "target" field.
// Returns structured errors for beautiful display by command handlers.
func resolveIacInputFromClipboard() (string, bool, error) {
	raw, err := clipboard.Read()
	if err != nil {
		// Check for empty clipboard and return structured error
		if strings.Contains(err.Error(), "empty") {
			return "", false, &ClipboardEmptyError{}
		}
		return "", false, errors.Wrap(err, "failed to read from clipboard")
	}

	// Check if content is an IaC input (has "target" field)
	if !iacinput.IsIacInput(raw) {
		return "", false, &ClipboardNotIacInputError{Raw: raw}
	}

	manifestPath, err := iacinput.ExtractManifestFromBytes(raw)
	if err != nil {
		return "", false, errors.Wrap(err, "failed to extract manifest from clipboard IaC input")
	}

	return manifestPath, true, nil
}
