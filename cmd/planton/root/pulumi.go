package root

import (
	"github.com/plantonhq/planton/cmd/planton/root/pulumi"
	"github.com/plantonhq/planton/internal/cli/flag"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var Pulumi = &cobra.Command{
	Use:   "pulumi",
	Short: "run a pulumi stack",
}

func init() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal("failed to get current working directory")
	}

	Pulumi.PersistentFlags().String(string(flag.Manifest), "", "path of the deployment-component manifest file")

	Pulumi.PersistentFlags().String(string(flag.InputDir), "", "directory containing target.yaml and credential yaml files")
	Pulumi.PersistentFlags().String(string(flag.KustomizeDir), "", "directory containing kustomize configuration")
	Pulumi.PersistentFlags().String(string(flag.Overlay), "", "kustomize overlay to use (e.g., prod, dev, staging)")
	Pulumi.PersistentFlags().String(string(flag.ModuleDir), pwd, "directory containing the pulumi module")
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

	// Kubernetes context flag
	Pulumi.PersistentFlags().String(string(flag.KubeContext), "", "kubectl context to use for Kubernetes deployments (overrides manifest label)")

	// Stack input file flag
	Pulumi.PersistentFlags().StringP(string(flag.StackInput), "i", "", "path to a YAML file containing the stack input (bypasses building stack input from manifest)")

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
