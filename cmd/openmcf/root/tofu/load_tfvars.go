package tofu

import (
	"github.com/plantonhq/openmcf/internal/manifest"
	"github.com/plantonhq/openmcf/pkg/iac/tofu/generators"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var LoadTfVars = &cobra.Command{
	Use:   "load-tfvars",
	Short: "load a openmcf manifest into tfvars format",
	Example: `
	openmcf tofu load-tfvars --manifest manifest.yaml
	`,
	Args: cobra.ExactArgs(1), //path of the manifest to load
	Run:  loadTfVarsHandler,
}

func loadTfVarsHandler(cmd *cobra.Command, args []string) {
	manifestPath := args[0]
	updatedManifest, err := manifest.LoadWithOverrides(manifestPath, map[string]string{})
	if err != nil {
		log.Fatal(err)
	}
	tfvarsString, err := generators.RenderTFVars(updatedManifest)
	if err != nil {
		log.Fatal("failed to generate Terraform variables: ", err)
	}
	println(tfvarsString)
}
