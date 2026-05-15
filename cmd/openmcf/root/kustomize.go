package root

import (
	"github.com/plantonhq/openmcf/cmd/openmcf/root/kustomize"
	"github.com/spf13/cobra"
)

var Kustomize = &cobra.Command{
	Use:   "kustomize",
	Short: "Generate and manage kustomize OpenAPI schemas for OpenMCF resources",
}

func init() {
	Kustomize.AddCommand(
		kustomize.Schema,
		kustomize.Init,
	)
}
