package manifest

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/plantonhq/planton/internal/cli/flag"
	"github.com/plantonhq/planton/pkg/clipboard"
	"github.com/plantonhq/planton/pkg/iac/stackinput"
)

// clipboardFlagValues contains all flag names that indicate clipboard input.
var clipboardFlagValues = []string{"--clipboard", "--clip", "--cb", "-c"}

// resolveFromStackInput checks for --stack-input flag and extracts manifest from it.
// Returns empty string if flag not provided.
// When a stack input file is provided, the manifest is extracted from the "target" field
// and written to a temporary file.
//
// Special case: If the flag value is a clipboard flag name (e.g., "-i --clip"),
// this indicates the user wants to read stack input from clipboard.
func resolveFromStackInput(cmd *cobra.Command) (manifestPath string, isTemp bool, err error) {
	stackInputPath, err := cmd.Flags().GetString(string(flag.StackInput))
	if err != nil {
		return "", false, errors.Wrap(err, "failed to get stack-input flag")
	}

	if stackInputPath == "" {
		return "", false, nil
	}

	// Detect clipboard flag captured as -i value (e.g., "-i --clip")
	// User intent is clear: read stack input from clipboard
	if isClipboardFlagValue(stackInputPath) {
		return resolveStackInputFromClipboard()
	}

	manifestPath, err = stackinput.ExtractManifestFromStackInput(stackInputPath)
	if err != nil {
		return "", false, errors.Wrapf(err, "failed to extract manifest from stack input %s", stackInputPath)
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

// resolveStackInputFromClipboard reads stack input YAML from clipboard
// and extracts the manifest from the "target" field.
// Returns structured errors for beautiful display by command handlers.
func resolveStackInputFromClipboard() (string, bool, error) {
	raw, err := clipboard.Read()
	if err != nil {
		// Check for empty clipboard and return structured error
		if strings.Contains(err.Error(), "empty") {
			return "", false, &ClipboardEmptyError{}
		}
		return "", false, errors.Wrap(err, "failed to read from clipboard")
	}

	// Check if content is a stack input (has "target" field)
	if !stackinput.IsStackInput(raw) {
		return "", false, &ClipboardNotStackInputError{Raw: raw}
	}

	manifestPath, err := stackinput.ExtractManifestFromBytes(raw)
	if err != nil {
		return "", false, errors.Wrap(err, "failed to extract manifest from clipboard stack input")
	}

	return manifestPath, true, nil
}
