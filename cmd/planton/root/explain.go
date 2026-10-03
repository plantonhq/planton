package root

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/explain"
	"github.com/spf13/cobra"
)

// Explain is the standalone binary's offline API reference (kubectl-explain
// for the kind catalog). It is deliberately NOT part of the engine
// command set (RegisterCommands): the Planton Platform CLI registers its own
// `explain` covering BOTH its platform APIs and these catalog kinds
// (through the same pkg/explain engine), so mounting this one there would
// collide with a strictly richer command.
var Explain = &cobra.Command{
	Use:   "explain <kind>[.<field.path>]",
	Short: "explain a catalog kind's schema (offline)",
	Long: `Explain any catalog kind the way kubectl explain teaches Kubernetes
resources: every spec field as it is written in YAML manifests, what it
means, its validation constraints and defaults, and the outputs other
resources can reference via valueFrom.

Drill into any field with a dotted path of the exact keys written in YAML.

Fully offline -- the schema and its documentation are compiled into this
binary; no control plane, no login, no network.`,
	Example: `
	# The full schema of a kind
	planton explain aws-vpc

	# One field, with its allowed values documented
	planton explain aws-vpc.spec.instanceTenancy

	# Every kind this binary knows
	planton explain --list`,
	Args: cobra.MaximumNArgs(1),
	RunE: explainHandler,
	// The handler prints the report itself; a returned error is a one-line
	// summary that must not be buried under usage output.
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	Explain.Flags().Bool("list", false, "list every catalog kind this binary knows")
	Explain.Flags().StringP("output", "o", "", "output format: json for the machine-readable report")
}

func explainHandler(cmd *cobra.Command, args []string) error {
	listFlag, _ := cmd.Flags().GetBool("list")
	output, _ := cmd.Flags().GetString("output")

	if listFlag || len(args) == 0 {
		if !listFlag {
			return errors.New("provide a kind (e.g. aws-vpc) or --list to see all kinds")
		}
		for _, name := range explain.KindNames() {
			fmt.Fprintln(cmd.OutOrStdout(), name)
		}
		return nil
	}

	// The first dotted segment is the kind; the rest is the field path.
	segments := strings.Split(args[0], ".")
	resource, err := explain.ResolveKindName(segments[0])
	if err != nil {
		return err
	}

	report, err := explain.DefaultEngine().Explain(resource, segments[1:])
	if err != nil {
		return err
	}

	if output == "json" {
		body, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return errors.Wrap(err, "failed to marshal explain report")
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(body))
		return nil
	}
	fmt.Fprint(cmd.OutOrStdout(), explain.Render(report))
	return nil
}
