package root

import (
	"github.com/plantonhq/planton/cmd/planton/root/pulumi"
	"github.com/plantonhq/planton/internal/cli/flag"
	"github.com/plantonhq/planton/internal/cli/iacflags"
	"github.com/spf13/cobra"
)

var Pulumi = &cobra.Command{
	Use:   "pulumi",
	Short: "run a pulumi stack",
}

func init() {
	Pulumi.PersistentFlags().String(string(flag.Manifest), "", "path of the manifest file")

	Pulumi.PersistentFlags().String(string(flag.InputDir), "", "directory containing target.yaml and credential yaml files")
	Pulumi.PersistentFlags().String(string(flag.KustomizeDir), "", "directory containing kustomize configuration")
	Pulumi.PersistentFlags().String(string(flag.Overlay), "", "kustomize overlay to use (e.g., prod, dev, staging)")
	// The flag default is EMPTY, not the current directory: an empty value
	// means "no explicit choice", which lets the module resolver distinguish
	// a user-chosen directory (validated loudly) from the implicit
	// current-directory probe (falls through quietly to download/staging).
	Pulumi.PersistentFlags().String(string(flag.ModuleDir), "", "directory containing the pulumi module (defaults to the current directory when it contains a valid module)")
	Pulumi.PersistentFlags().StringToString(string(flag.Set), map[string]string{}, "override resource manifest values using key=value pairs")

	Pulumi.PersistentFlags().String(string(flag.Stack), "", "pulumi stack fqdn in the format of <org>/<project>/<stack>")
	Pulumi.PersistentFlags().Bool(string(flag.Yes), false, "Automatically approve and perform the update after previewing it")
	Pulumi.PersistentFlags().Bool(string(flag.Force), false, "Force removal of stack even if resources exist (use with delete/rm command)")
	Pulumi.PersistentFlags().Bool(string(flag.Diff), false, "Show detailed resource diffs")

	// Staging/cleanup flags
	Pulumi.PersistentFlags().Bool(string(flag.NoCleanup), false, "Do not cleanup the workspace copy after execution (keeps cloned modules)")
	Pulumi.PersistentFlags().String(string(flag.ModuleVersion), "",
		"Checkout a specific version (tag, branch, or commit SHA) of the IaC modules in the workspace copy.\n"+
			"This allows using a different module version than what's in the staging area without affecting it.")

	iacflags.AddKubeContextFlag(Pulumi)

	// IaC input file flag
	Pulumi.PersistentFlags().StringP(string(flag.IacInput), "i", "", "path to a YAML file containing the IaC input (bypasses building IaC input from manifest)")

	// Provider config flag (unified)
	Pulumi.PersistentFlags().StringP(string(flag.ProviderConfig), "p", "", "path to provider credentials file")

	Pulumi.AddCommand(
		pulumi.Init,
		pulumi.Refresh,
		pulumi.Preview,
		pulumi.Update,
		pulumi.Destroy,
		pulumi.Delete,
		pulumi.Cancel,
	)
}
