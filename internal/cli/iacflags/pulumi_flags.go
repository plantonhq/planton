package iacflags

import (
	"github.com/plantonhq/planton/internal/cli/flag"
	"github.com/spf13/cobra"
)

// AddPulumiFlags adds Pulumi-specific flags to the command.
func AddPulumiFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String(string(flag.Stack), "",
		"pulumi stack fqdn in the format of <org>/<project>/<stack>")

	cmd.PersistentFlags().Bool(string(flag.Yes), false,
		"Automatically approve and perform the update after previewing it (Pulumi)")

	cmd.PersistentFlags().Bool(string(flag.Diff), false,
		"Show detailed resource diffs (Pulumi)")
}
