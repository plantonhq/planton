package iacflags

import (
	"os"

	"github.com/plantonhq/planton/internal/cli/flag"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// AddExecutionFlags adds common execution-related flags for IaC commands.
func AddExecutionFlags(cmd *cobra.Command) {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal("failed to get current working directory")
	}

	cmd.PersistentFlags().String(string(flag.ModuleDir), pwd,
		"directory containing the provisioner module")

	cmd.PersistentFlags().String(string(flag.ModuleVersion), "",
		"Checkout a specific version (tag, branch, or commit SHA) of the IaC modules in the workspace copy.\n"+
			"This allows using a different module version than what's in the staging area without affecting it.")

	cmd.PersistentFlags().Bool(string(flag.NoCleanup), false,
		"Do not cleanup the workspace copy after execution (keeps cloned modules)")

	cmd.PersistentFlags().String(string(flag.KubeContext), "",
		"kubectl context to use for Kubernetes deployments (overrides manifest label)")

	cmd.PersistentFlags().StringToString(string(flag.Set), map[string]string{},
		"override resource manifest values using key=value pairs")

	cmd.PersistentFlags().Bool(string(flag.LocalModule), false,
		"Use the local planton repository to derive the module directory")
}
