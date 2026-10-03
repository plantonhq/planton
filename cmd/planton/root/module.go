//go:build !codegen
// +build !codegen

package root

import (
	"github.com/plantonhq/planton/cmd/planton/root/module"
	"github.com/spf13/cobra"
)

// Module groups the verbs for working with the IaC modules behind catalog
// kinds: ejecting an official module into a user-owned copy to
// customize, and verifying that a customized module still honors its
// kind's contract. Both verbs work standalone — no platform account,
// no control plane.
var Module = &cobra.Command{
	Use:   "module",
	Short: "eject and verify the IaC modules behind catalog kinds",
}

func init() {
	Module.AddCommand(
		module.Eject,
		module.Verify,
	)
}
