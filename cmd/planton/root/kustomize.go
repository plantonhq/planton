package root

import (
	"github.com/plantonhq/planton/cmd/planton/root/kustomize"
	"github.com/spf13/cobra"
)

var Kustomize = &cobra.Command{
	Use:   "kustomize",
	Short: "Generate and manage kustomize OpenAPI schemas for Planton resources",
}

func init() {
	Kustomize.AddCommand(
		kustomize.Schema,
		kustomize.Init,
	)
}
