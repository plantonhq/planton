package iacrunner

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/internal/cli/flag"
	"github.com/plantonhq/planton/internal/cli/prompt"
	"github.com/plantonhq/planton/internal/cli/ui"
	"github.com/plantonhq/planton/internal/cli/ui/preflightreport"
	"github.com/plantonhq/planton/pkg/outputs"
	"github.com/plantonhq/planton/pkg/setdeploy"
	"github.com/spf13/cobra"
)

// Exit codes for the set lane, distinct so CI can tell WHY a run stopped:
// a preflight refusal means nothing was handed to an IaC engine and the fix
// is in the inputs; a deploy failure means state may have advanced and the
// recovery is re-running the same command.
const (
	ExitDeployFailure    = 1
	ExitPreflightRefusal = 2
)

// setScopedFlagRefusals names the apply flags that address exactly ONE
// resource and therefore refuse set input, each with the sentence that says
// what to do instead. Set-wide flags (backend type/bucket/region/endpoint,
// backend-url, module-version, kube-context, auto-approve, yes) stay legal.
var setScopedFlagRefusals = []struct {
	flagName flag.Flag
	isString bool
	sentence string
}{
	{flag.Set, false, "--set addresses one resource's fields and this input is a set — put the value in the manifest itself"},
	{flag.ModuleDir, true, "--module-dir supplies one kind's module and this input is a set — the module catalog resolves each node's module"},
	{flag.LocalModule, false, "--local-module supplies one kind's module and this input is a set — the module catalog resolves each node's module"},
	{flag.Stack, true, "--stack names one pulumi stack and this input is a set — set annotation pulumi.planton.dev/stack.fqdn on each pulumi manifest"},
	{flag.BackendKey, true, "--backend-key names one state file and this input is a set — set annotation tofu.planton.dev/backend.key on each manifest (sharing a key would make the resources overwrite each other's state)"},
}

// RunSet is the multi-manifest apply lane: the preflight wall, one approval,
// then sequential dependency-ordered deploys with output-fed references.
// It returns the process exit code; the cobra handler exits with it. Split
// from the handler so the whole lane — refusals, report, exit codes — is
// drivable from tests without a process boundary.
func RunSet(cmd *cobra.Command, docs []setdeploy.Doc) int {
	// Single-resource-scoped flags refuse set input with the fix named.
	for _, refusal := range setScopedFlagRefusals {
		var isSet bool
		if refusal.isString {
			v, _ := cmd.Flags().GetString(string(refusal.flagName))
			isSet = v != ""
		} else {
			switch refusal.flagName {
			case flag.Set:
				v, _ := cmd.Flags().GetStringToString(string(flag.Set))
				isSet = len(v) > 0
			default:
				v, _ := cmd.Flags().GetBool(string(refusal.flagName))
				isSet = v
			}
		}
		if isSet {
			cliprint.PrintError(refusal.sentence)
			return ExitPreflightRefusal
		}
	}

	flags := setFlagsFromCommand(cmd)

	cliprint.PrintStep(fmt.Sprintf("Preflighting %d manifests...", len(docs)))
	plan := setdeploy.Preflight(docs, flags, setdeploy.LiveProbes{})
	preflightreport.Print(plan.Report)

	if plan.Report.Refused() {
		return ExitPreflightRefusal
	}

	// One approval for the whole set — the preflight report and the deploy
	// order ARE the plan being approved. CI passes --auto-approve (or --yes);
	// interactively the question is asked once, never per node.
	if !approvedSetDeploy(cmd, len(plan.Order)) {
		cliprint.PrintError("deploy not approved — pass --auto-approve (or answer yes) to deploy this set")
		return ExitPreflightRefusal
	}

	deployer := &setdeploy.EngineDeployer{Flags: flags}
	defer deployer.Close()

	result, err := setdeploy.Execute(plan, deployer, setEventRenderer{})
	if err != nil {
		cliprint.PrintError(err.Error())
		return ExitDeployFailure
	}
	if !result.Succeeded() {
		printSetFailureSummary(plan, result)
		return ExitDeployFailure
	}

	cliprint.PrintSuccess(fmt.Sprintf("All %d resources deployed", len(plan.Order)))
	return 0
}

func setFlagsFromCommand(cmd *cobra.Command) setdeploy.Flags {
	backendType, _ := cmd.Flags().GetString(string(flag.BackendType))
	backendBucket, _ := cmd.Flags().GetString(string(flag.BackendBucket))
	backendRegion, _ := cmd.Flags().GetString(string(flag.BackendRegion))
	backendEndpoint, _ := cmd.Flags().GetString(string(flag.BackendEndpoint))
	pulumiBackendURL, _ := cmd.Flags().GetString(string(flag.BackendUrl))
	moduleVersion, _ := cmd.Flags().GetString(string(flag.ModuleVersion))
	kubeContext, _ := cmd.Flags().GetString(string(flag.KubeContext))
	return setdeploy.Flags{
		BackendType:      backendType,
		BackendBucket:    backendBucket,
		BackendRegion:    backendRegion,
		BackendEndpoint:  backendEndpoint,
		PulumiBackendURL: pulumiBackendURL,
		ModuleVersion:    moduleVersion,
		KubeContext:      kubeContext,
	}
}

// approvedSetDeploy applies the one-approval rule: --auto-approve or --yes
// approves; otherwise an interactive terminal is asked once; a non-
// interactive run without approval refuses (hanging a CI runner on a hidden
// question is the failure mode this refuses to have).
func approvedSetDeploy(cmd *cobra.Command, nodeCount int) bool {
	if autoApprove, _ := cmd.Flags().GetBool(string(flag.AutoApprove)); autoApprove {
		return true
	}
	if yes, _ := cmd.Flags().GetBool(string(flag.Yes)); yes {
		return true
	}
	if !prompt.IsInteractive() {
		return false
	}
	fmt.Printf("Deploy %d resources in the order above? [y/N]: ", nodeCount)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	answer := strings.ToLower(strings.TrimSpace(input))
	return answer == "y" || answer == "yes"
}

// setEventRenderer is the OSS CLI's voice for the execution loop, in the
// house step-log style.
type setEventRenderer struct{}

func (setEventRenderer) NodeStarting(position, total int, node setdeploy.NodePlan) {
	fmt.Println()
	cliprint.PrintStep(fmt.Sprintf("[%d/%d] %s", position, total, node.Identity))
}

func (setEventRenderer) NodeSucceeded(node setdeploy.NodePlan, captured *outputs.CaptureResult) {
	cliprint.PrintSuccess(fmt.Sprintf("%s deployed", node.Identity))
	ui.OutputsSummary(captured)
}

func (setEventRenderer) NodeWarning(node setdeploy.NodePlan, message string) {
	cliprint.PrintWarning(message)
}

func (setEventRenderer) NodeFailed(node setdeploy.NodePlan, err error) {
	// The engine's own error, verbatim — paraphrasing an IaC failure hides
	// the one detail the user needs.
	ui.EngineFailure(fmt.Sprintf("%s failed", node.Identity), err)
}

// printSetFailureSummary tells the whole truth after a node failure: what
// deployed (with outputs captured), what failed, what never started, and the
// honest recovery — state backends hold everything, so re-running the same
// command re-applies completed nodes as no-ops and continues from the
// failure.
func printSetFailureSummary(plan *setdeploy.Plan, result *setdeploy.Result) {
	fmt.Println()
	for _, idx := range plan.Order {
		node := plan.Set.Nodes[idx]
		switch result.Statuses[idx] {
		case setdeploy.NodeStatusSucceeded:
			cliprint.PrintSuccess(fmt.Sprintf("%s deployed (outputs captured)", node.Identity))
		case setdeploy.NodeStatusFailed:
			cliprint.PrintError(fmt.Sprintf("%s FAILED — the engine's error is above, verbatim", node.Identity))
		case setdeploy.NodeStatusNeverStarted:
			cliprint.PrintInfo(fmt.Sprintf("%s never started", node.Identity))
		}
	}
	fmt.Println()
	cliprint.PrintInfo("State is held by the backends: run the same command again — completed resources re-apply as no-ops and the deploy continues from the failure.")
}
