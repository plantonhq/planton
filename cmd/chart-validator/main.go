// chart-validator is the CI chart-validation binary: the SAME `chart
// validate` command the planton CLI mounts, linked WITHOUT the CLI's
// pulumi/cloud SDK world.
//
// Why it exists: the charts release lane and the charts PR lint lane validate
// every chart against this ref's own contract before packaging (a release
// once shipped charts its own tree rejected and bricked every installed
// desktop). They used to get the validator by compiling the ENTIRE CLI --
// 6-9 minutes of build for a 1-second packaging step, measured across
// v0.4.1-v0.4.4, almost all of it linking pulumi provider SDKs that chart
// validation never touches. This main links only pkg/infrachart/validatecmd
// (chart engine + generated kind registry + protovalidate), cutting the
// compile to a fraction while keeping the verdict engine byte-identical.
//
// Two designs were considered and rejected (see the 2026-08 platform ADRs
// and changelogs for the fuller record):
//   - reimplementing validation in an interpreted language: a second engine
//     is a silent-drift surface -- the protovalidate-java conformance gate
//     exists because engines disagreeing already shipped a broken release;
//   - committing a prebuilt binary: its bulk IS the generated contract
//     (kind schemas + CEL rules), which changes every release even when no
//     validator source changes, so a committed binary validates against a
//     stale contract -- the exact silent failure the lanes exist to prevent.
package main

import (
	"fmt"
	"os"

	"github.com/plantonhq/planton/pkg/infrachart/validatecmd"
)

func main() {
	cmd := validatecmd.NewChartValidateCommand(validatecmd.Options{})
	cmd.Use = "chart-validator [chart-dir ...]"
	if err := cmd.Execute(); err != nil {
		// The command runs with SilenceErrors (the handler prints the full
		// per-chart report itself); the host prints the one-line summary,
		// exactly as the CLI roots do.
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
