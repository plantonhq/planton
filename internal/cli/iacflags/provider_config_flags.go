package iacflags

import (
	"github.com/plantonhq/planton/internal/cli/flag"
	"github.com/spf13/cobra"
)

// AddProviderConfigFlags adds the unified provider config flag.
// The provider type is auto-detected from the manifest's apiVersion and kind.
func AddProviderConfigFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP(string(flag.ProviderConfig), "p", "",
		"path to provider credentials file (provider type auto-detected from manifest)")
}
