package root

import (
	"github.com/plantonhq/planton/cmd/planton/root/tofu"
	"github.com/plantonhq/planton/internal/cli/flag"
	"github.com/plantonhq/planton/internal/cli/iacflags"
	"github.com/spf13/cobra"
)

var Tofu = &cobra.Command{
	Use:   "tofu",
	Short: "run open-tofu commands",
}

func init() {
	Tofu.PersistentFlags().String(string(flag.Manifest), "", "path of the kind manifest file")

	// The shared manifest resolver reads this flag on every source-resolution
	// path, so every IaC command group must register it (the pulumi group does
	// the same); without it, `tofu <cmd> --manifest ...` fails before running.
	Tofu.PersistentFlags().StringP(string(flag.IacInput), "i", "", "path to a YAML file containing the IaC input (extracts manifest from target field)")

	Tofu.PersistentFlags().String(string(flag.InputDir), "", "directory containing target.yaml and credential yaml files")
	Tofu.PersistentFlags().String(string(flag.KustomizeDir), "", "directory containing kustomize configuration")
	Tofu.PersistentFlags().String(string(flag.Overlay), "", "kustomize overlay to use (e.g., prod, dev, staging)")
	// The flag default is EMPTY, not the current directory: an empty value
	// means "no explicit choice", which lets the module resolver distinguish
	// a user-chosen directory (validated loudly) from the implicit
	// current-directory probe (falls through quietly to download/staging).
	Tofu.PersistentFlags().String(string(flag.ModuleDir), "", "directory containing the terraform module (defaults to the current directory when it contains a valid module)")
	Tofu.PersistentFlags().StringToString(string(flag.Set), map[string]string{}, "override resource manifest values using key=value pairs")

	// Provider config flag (unified)
	Tofu.PersistentFlags().StringP(string(flag.ProviderConfig), "p", "", "path to provider credentials file")

	iacflags.AddKubeContextFlag(Tofu)

	Tofu.AddCommand(
		tofu.Apply,
		tofu.Destroy,
		tofu.GenerateHelmCrds,
		tofu.GenerateModule,
		tofu.GenerateVariables,
		tofu.Init,
		tofu.LoadTfVars,
		tofu.Plan,
		tofu.Refresh,
	)
}
