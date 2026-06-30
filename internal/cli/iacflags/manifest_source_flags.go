package iacflags

import (
	"github.com/plantonhq/planton/internal/cli/flag"
	"github.com/spf13/cobra"
)

// clipboardFlagNames contains all the flag names that enable clipboard input.
// The primary flag is "clipboard" with aliases "clip" and "cb".
var clipboardFlagNames = []string{string(flag.Clipboard), "clip", "cb"}

// AddManifestSourceFlags adds flags for specifying the manifest source.
// Priority order: --clipboard > --stack-input > --manifest > --input-dir > --kustomize-dir+--overlay
// Clipboard flag supports aliases: --clipboard, --clip, --cb, -c
func AddManifestSourceFlags(cmd *cobra.Command) {
	// Primary clipboard flag with shorthand -c
	cmd.PersistentFlags().BoolP(string(flag.Clipboard), "c", false,
		"read manifest content from system clipboard (aliases: --clip, --cb)")

	// Hidden alias flags for convenience
	cmd.PersistentFlags().Bool("clip", false, "alias for --clipboard")
	cmd.PersistentFlags().Bool("cb", false, "alias for --clipboard")

	// Hide alias flags from help output
	_ = cmd.PersistentFlags().MarkHidden("clip")
	_ = cmd.PersistentFlags().MarkHidden("cb")

	cmd.PersistentFlags().StringP(string(flag.Manifest), "f", "",
		"path of the deployment-component manifest file")

	cmd.PersistentFlags().StringP(string(flag.StackInput), "i", "",
		"path to a YAML file containing the stack input (extracts manifest from target field)")

	cmd.PersistentFlags().String(string(flag.InputDir), "",
		"directory containing target.yaml and credential yaml files")

	cmd.PersistentFlags().String(string(flag.KustomizeDir), "",
		"directory containing kustomize configuration")

	cmd.PersistentFlags().String(string(flag.Overlay), "",
		"kustomize overlay to use (e.g., prod, dev, staging)")
}

// IsClipboardFlagSet checks if any clipboard flag (--clipboard, --clip, or --cb) is set.
// Returns true if clipboard input should be used, false otherwise.
func IsClipboardFlagSet(cmd *cobra.Command) (bool, error) {
	for _, name := range clipboardFlagNames {
		if cmd.Flags().Changed(name) {
			val, err := cmd.Flags().GetBool(name)
			if err != nil {
				return false, err
			}
			if val {
				return true, nil
			}
		}
	}
	return false, nil
}
