// Package validatecmd is the single home of the `chart validate` cobra
// command. It lives as a leaf package -- NOT in cmd/planton/root -- because
// the command mounts in THREE binaries and one of them must stay small:
//
//  1. the standalone planton CLI's `chart` tree (via cmd/planton/root, which
//     re-exports the constructor for compatibility),
//  2. the Planton Platform CLI's `chart` tree (cross-repo, through the
//     cmd/planton/root re-export -- an import-path contract; do not move it),
//  3. cmd/chart-validator, the CI validation binary: the charts release and
//     PR lanes used to `go build .` the ENTIRE CLI -- pulumi provider SDKs,
//     cloud SDKs, the whole command forest -- spending 6-9 minutes of compile
//     for a 1-second packaging job (measured across v0.4.1-v0.4.4). This
//     package's dependency set is only the chart engine, the generated kind
//     registry, and protovalidate, so a binary linking just this compiles in
//     a fraction of the time.
//
// One implementation, three mounts: validation must run the exact consumer
// engine at the exact ref (the protovalidate-java conformance gate exists
// because engines drift), so the CI binary is a smaller LINK of the same
// command, never a second implementation.
package validatecmd

import (
	"errors"
	"fmt"
	"runtime"
	"sync"

	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/pkg/infrachart"
	"github.com/spf13/cobra"
)

// NewChartValidateCommand builds the `chart validate` command. It is a
// constructor rather than a package variable because cobra commands are
// stateful (flag values live on the instance) and this command mounts in
// multiple binaries -- each host mounts its own instance.
//
// The command is fully offline -- it dials nothing and needs no configuration
// -- so hosts that guard backend-requiring commands must exempt it.
//
// checks are the host's own rules over each rendered document
// (infrachart.Options.DocumentChecks): the open-source CLI passes none, and the
// Planton Platform CLI passes the platform's rule for where secrets may go.
func NewChartValidateCommand(checks ...infrachart.DocumentCheck) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [chart-dir ...]",
		Short: "render and validate infra-charts against the compiled-in kind registry",
		Long: `Render each chart's templates with the defaults declared in values.yaml and validate
every produced manifest offline: the kind must exist, every field must exist on the
kind's spec (unknown or renamed fields fail), the whole document (metadata as well as
spec) must pass its validation rules, every valueFrom reference must resolve to a real field on the referenced kind (and
references to a field's default kind must use the annotated composition key), and the
chart's internal references must form a dependency graph without cycles.

Each bool param is additionally flipped once so conditional manifests are exercised in
both branches. A reference whose target another variant defines but the current one
does not is a warning that says what to verify (either a toggle broke the composition,
or the parameter points that arm at a resource owned outside the chart); a reference no
variant defines is a warning too (the resource must already exist in the target
environment).

The reserved org and env template variables render as "acme" and "dev" unless set:
--set org=<slug> --set env=<slug> works everywhere, and a host CLI whose global --org and
--env flags name the organization and environment (the Planton Platform CLI) binds them
when they are typed; --set wins over both.

This requires no backend and no network: everything validates against the schemas
compiled into this binary, rendered the way the control plane renders them. The control
plane validates each document again when a chart is published or installed; a host CLI
may add its own document checks (the Planton Platform CLI adds where secrets may go:
a secret field takes only a $secret/ reference, and a generated secret feeds only a
secret field).`,
		Example: `
	# Validate one chart
	planton chart validate charts/gcp/cloud-run-service

	# Validate several charts
	planton chart validate charts/gcp/postgres-production charts/gcp/gke-environment

	# Validate every chart under a directory tree
	planton chart validate --all charts/

	# Exercise a specific parameter combination beyond the automatic toggle flips
	planton chart validate charts/gcp/static-website-cdn --set dnsEnabled=false

	# Render for a specific environment (catches names that break on a hyphenated slug)
	planton chart validate charts/gcp/cloud-run-service --set env=production-eu`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return chartValidateHandler(cmd, args, checks)
		},
		// The handler prints the full per-chart report itself; the returned error is a
		// one-line summary for the host root's error path, so neither cobra's usage
		// dump nor its error echo should repeat it.
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.Flags().Bool("all", false, "treat the given directories as roots and validate every chart found under them")
	cmd.Flags().Bool("verbose", false, "also list charts and variants that validated cleanly, and print warnings for passing charts")
	cmd.Flags().StringArray("set", nil, "override a param value (key=value, repeatable); org and env may also be overridden")
	return cmd
}

// hostScope reads the host root's global --org or --env when the person typed
// it. The command declares neither: a host CLI already owns both names as its
// global flags, and a local copy would hide them (the person types one, the
// handler reads the other). An untyped flag stays empty, so a host's saved
// context never leaks into the synthetic render and Validate's defaults apply.
func hostScope(cmd *cobra.Command, name string) string {
	if f := cmd.Flag(name); f != nil && f.Changed {
		return f.Value.String()
	}
	return ""
}

func chartValidateHandler(cmd *cobra.Command, args []string, checks []infrachart.DocumentCheck) error {
	all, _ := cmd.Flags().GetBool("all")
	verbose, _ := cmd.Flags().GetBool("verbose")
	setFlags, _ := cmd.Flags().GetStringArray("set")
	org, env := hostScope(cmd, "org"), hostScope(cmd, "env")

	if len(args) == 0 {
		return errors.New("provide at least one chart directory (or a root directory with --all)")
	}

	setOverrides, err := infrachart.ParseSetFlags(setFlags)
	if err != nil {
		return err
	}

	dirs, err := infrachart.DiscoverCharts(args, all)
	if err != nil {
		return err
	}
	if len(dirs) == 0 {
		return errors.New("no charts found (a chart directory contains a Chart.yaml)")
	}

	// Charts validate CONCURRENTLY -- each is an independent, CPU-bound unit
	// (template rendering + CEL evaluation), the shared protovalidate
	// validator is documented safe for concurrent use, and rendering shares
	// only a read-only environment (the sandbox bans every stateful template
	// construct). All printing happens AFTER the wait, strictly in discovery
	// order, so output is deterministic and byte-identical to the sequential
	// predecessor's -- workers never print.
	type outcome struct {
		report *infrachart.Report
		err    error
	}
	outcomes := make([]outcome, len(dirs))
	semaphore := make(chan struct{}, runtime.NumCPU())
	var wg sync.WaitGroup
	for i, dir := range dirs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			report, err := infrachart.Validate(dir, infrachart.Options{Org: org, Env: env, Set: setOverrides, DocumentChecks: checks})
			outcomes[i] = outcome{report: report, err: err}
		}()
	}
	wg.Wait()

	passed, failed := 0, 0
	for i, dir := range dirs {
		report, err := outcomes[i].report, outcomes[i].err
		if err != nil {
			failed++
			cliprint.PrintError(fmt.Sprintf("✘ %s: %v", dir, err))
			continue
		}
		if !report.HasErrors() {
			passed++
			if verbose {
				cliprint.PrintSuccess(fmt.Sprintf("✔ %s (%d variants)", dir, len(report.Variants)))
				printIssues(report, severityOnly(infrachart.SeverityWarning))
			}
			continue
		}
		failed++
		cliprint.PrintError(fmt.Sprintf("✘ %s", dir))
		printIssues(report, nil)
	}

	fmt.Printf("\nchart validation: %d passed, %d failed out of %d\n", passed, failed, passed+failed)
	if failed > 0 {
		return fmt.Errorf("%d of %d charts failed validation", failed, passed+failed)
	}
	return nil
}

// severityOnly returns a filter admitting only the given severity.
func severityOnly(s infrachart.Severity) func(infrachart.Issue) bool {
	return func(issue infrachart.Issue) bool { return issue.Severity == s }
}

// printIssues prints a report's issues grouped by variant. A nil filter
// prints everything.
func printIssues(report *infrachart.Report, filter func(infrachart.Issue) bool) {
	for _, variant := range report.Variants {
		for _, issue := range variant.Issues {
			if filter != nil && !filter(issue) {
				continue
			}
			label := issue.File
			if issue.ResourceKind != "" {
				label = fmt.Sprintf("%s (%s/%s)", issue.File, issue.ResourceKind, issue.ResourceName)
			}
			fmt.Printf("  [%s] %s: %s: %s\n", variant.Name, issue.Severity, label, issue.Message)
		}
	}
}
