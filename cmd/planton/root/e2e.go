package root

import (
	"github.com/plantonhq/planton/cmd/planton/root/e2e"
	"github.com/spf13/cobra"
)

var E2E = &cobra.Command{
	Use:   "e2e",
	Short: "E2E test discovery, profiling, and CI integration",
}

func init() {
	E2E.AddCommand(
		e2e.Discover,
	)
}
