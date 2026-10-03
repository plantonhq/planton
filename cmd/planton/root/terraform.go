package root

import (
	"github.com/plantonhq/planton/cmd/planton/root/terraform"
	"github.com/plantonhq/planton/internal/cli/flag"
	"github.com/plantonhq/planton/internal/cli/iacflags"
	"github.com/spf13/cobra"
)

var Terraform = &cobra.Command{
	Use:   "terraform",
	Short: "run terraform commands",
}

func init() {
	Terraform.PersistentFlags().String(string(flag.Manifest), "", "path of the kind manifest file")

	Terraform.PersistentFlags().String(string(flag.InputDir), "", "directory containing target.yaml and credential yaml files")
	Terraform.PersistentFlags().String(string(flag.KustomizeDir), "", "directory containing kustomize configuration")
	Terraform.PersistentFlags().String(string(flag.Overlay), "", "kustomize overlay to use (e.g., prod, dev, staging)")
	// The flag default is EMPTY, not the current directory: an empty value
	// means "no explicit choice", which lets the module resolver distinguish
	// a user-chosen directory (validated loudly) from the implicit
	// current-directory probe (falls through quietly to download/staging).
	Terraform.PersistentFlags().String(string(flag.ModuleDir), "", "directory containing the terraform module (defaults to the current directory when it contains a valid module)")
	Terraform.PersistentFlags().StringToString(string(flag.Set), map[string]string{}, "override resource manifest values using key=value pairs")

	// IaC input file flag: the shared manifest resolver reads this on every
	// command, so it must be registered here (as on the pulumi command group)
	// or resolution fails before --manifest is even considered.
	Terraform.PersistentFlags().StringP(string(flag.IacInput), "i", "", "path to a YAML file containing the IaC input (bypasses building IaC input from manifest)")

	// Provider config flag (unified)
	Terraform.PersistentFlags().StringP(string(flag.ProviderConfig), "p", "", "path to provider credentials file")

	iacflags.AddKubeContextFlag(Terraform)

	Terraform.AddCommand(
		terraform.Apply,
		terraform.Destroy,
		terraform.Init,
		terraform.Plan,
		terraform.Refresh,
	)
}
