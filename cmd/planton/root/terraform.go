package root

import (
	"os"

	"github.com/plantonhq/planton/cmd/planton/root/terraform"
	"github.com/plantonhq/planton/internal/cli/flag"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var Terraform = &cobra.Command{
	Use:   "terraform",
	Short: "run terraform commands",
}

func init() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal("failed to get current working directory")
	}

	Terraform.PersistentFlags().String(string(flag.Manifest), "", "path of the deployment-component manifest file")

	Terraform.PersistentFlags().String(string(flag.InputDir), "", "directory containing target.yaml and credential yaml files")
	Terraform.PersistentFlags().String(string(flag.KustomizeDir), "", "directory containing kustomize configuration")
	Terraform.PersistentFlags().String(string(flag.Overlay), "", "kustomize overlay to use (e.g., prod, dev, staging)")
	Terraform.PersistentFlags().String(string(flag.ModuleDir), pwd, "directory containing the terraform module")
	Terraform.PersistentFlags().StringToString(string(flag.Set), map[string]string{}, "override resource manifest values using key=value pairs")

	// Provider config flag (unified)
	Terraform.PersistentFlags().StringP(string(flag.ProviderConfig), "p", "", "path to provider credentials file")

	Terraform.AddCommand(
		terraform.Apply,
		terraform.Destroy,
		terraform.Init,
		terraform.Plan,
		terraform.Refresh,
	)
}
