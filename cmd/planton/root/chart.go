//go:build !codegen
// +build !codegen

package root

import (
	"github.com/plantonhq/planton/pkg/infrachart/validatecmd"
	"github.com/spf13/cobra"
)

// Chart is the standalone binary's offline infra-chart tooling. It is deliberately NOT
// part of the engine command set (RegisterCommands): the Planton Platform CLI mounts its
// own `chart` command tree (build/publish against the control plane) and adds the shared
// offline subcommands to it via their exported constructors -- NewChartValidateCommand
// below -- so both binaries carry the exact same validate command with no drift.
var Chart = &cobra.Command{
	Use:   "chart",
	Short: "work with infra-charts offline",
}

func init() {
	Chart.AddCommand(NewChartValidateCommand(ChartValidateOptions{}))
}

// ChartValidateOptions re-exports the rules a host adds to `chart validate`,
// so a host names them through the same import path as the constructor.
type ChartValidateOptions = validatecmd.Options

// NewChartValidateCommand re-exports the shared `chart validate` constructor.
//
// This path (cmd/planton/root) is a CROSS-REPO import contract: the Planton
// Platform CLI mounts the command through it -- keep the re-export even
// though the implementation lives in pkg/infrachart/validatecmd. It moved to
// that leaf package so the CI chart-validator binary can link the identical
// command without this package's full CLI world (pulumi/cloud SDKs) entering
// its build graph; see the validatecmd package comment for the three mounts.
// A host passes its own rules: document checks (the Planton Platform CLI
// passes where secrets may go) and the reference tokens it resolves before
// anything deploys (the Planton Platform CLI passes its $secret/ and $var/
// grammar, so a rule written about a value waits for the token's value). This
// CLI passes neither, because its backendless deploy takes literals.
func NewChartValidateCommand(opts ChartValidateOptions) *cobra.Command {
	return validatecmd.NewChartValidateCommand(opts)
}
