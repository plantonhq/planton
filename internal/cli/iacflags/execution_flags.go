package iacflags

import (
	"github.com/plantonhq/planton/internal/cli/flag"
	"github.com/spf13/cobra"
)

// AddExecutionFlags adds common execution-related flags for IaC commands.
func AddExecutionFlags(cmd *cobra.Command) {
	// The flag default is EMPTY, not the current directory: an empty value
	// means "no explicit choice", which lets the module resolver distinguish
	// a user-chosen directory (validated loudly) from the implicit
	// current-directory probe (falls through quietly to download/staging).
	cmd.PersistentFlags().String(string(flag.ModuleDir), "",
		"directory containing the provisioner module (defaults to the current directory when it contains a valid module)")

	cmd.PersistentFlags().String(string(flag.ModuleVersion), "",
		"Checkout a specific version (tag, branch, or commit SHA) of the IaC modules in the workspace copy.\n"+
			"This allows using a different module version than what's in the staging area without affecting it.")

	cmd.PersistentFlags().Bool(string(flag.NoCleanup), false,
		"Do not cleanup the workspace copy after execution (keeps cloned modules)")

	AddKubeContextFlag(cmd)

	cmd.PersistentFlags().StringToString(string(flag.Set), map[string]string{},
		"override resource manifest values using key=value pairs")

	// --local-module is not declared here: it is a root persistent flag
	// (root.RegisterPersistentFlags), and a second declaration on the command
	// would hide the root's -- the person types one, the handler reads the other.
}

// AddKubeContextFlag registers --kube-context, the one flag every engine
// group's handlers read to pick the kubeconfig context for a Kubernetes
// deploy. Declared once so the root lifecycle commands and the pulumi, tofu,
// and terraform groups cannot drift: a handler that reads a flag its group
// never registered silently sees "" and deploys to the current context.
func AddKubeContextFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().String(string(flag.KubeContext), "",
		"kubectl context to use for Kubernetes deployments (overrides manifest label)")
}
