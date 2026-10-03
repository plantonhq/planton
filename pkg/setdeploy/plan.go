package setdeploy

import (
	"github.com/plantonhq/planton/pkg/iac/provisioner"
	"github.com/plantonhq/planton/pkg/iac/tofu/backendconfig"
	"github.com/plantonhq/planton/pkg/manifestgraph"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// Flags carries the set-wide inputs a caller may legally supply for a whole
// set. Anything per-node by nature (state key, stack FQDN, field overrides,
// module directory) is deliberately absent: those ride each manifest's own
// annotations, and the CLI refuses their single-resource flags on set input.
type Flags struct {
	// Tofu/terraform state backend, set-wide. Merged per node with the
	// manifest's annotations at the same precedence the single-manifest lane
	// uses (flags override annotations). The state KEY is never set-wide.
	BackendType     string
	BackendBucket   string
	BackendRegion   string
	BackendEndpoint string

	// PulumiBackendURL pins pulumi state for the whole set (flag >
	// annotation > env, per node).
	PulumiBackendURL string

	// ModuleVersion pins the catalog module version for every node.
	ModuleVersion string

	// KubeContext overrides each node's kube-context annotation, matching the
	// single-manifest lane's flag-beats-annotation rule.
	KubeContext string
}

// NodePlan is everything preflight resolved about one node that execution
// needs: the routing facts (provisioner, kind, provider) and the state
// configuration, all settled BEFORE the first handoff so execution never
// discovers configuration mid-run.
type NodePlan struct {
	// Index is the node's position in the set (Plan.Set.Nodes[Index]).
	Index int

	Identity manifestgraph.Identity
	Source   string

	// KindName is the manifest's kind as the module resolvers address it.
	KindName string

	// Provisioner routes the node to its engine. When the manifest carries no
	// provisioner label the set lane defaults to tofu — stated in the report,
	// never prompted (a set deploy is one decision, not N interviews).
	Provisioner        provisioner.ProvisionerType
	ProvisionerDefault bool

	// Provider is the cloud provider the kind belongs to, driving the
	// credential check.
	Provider catalogkind.CatalogProvider

	// TofuBackend is the node's merged state backend configuration
	// (tofu/terraform nodes only).
	TofuBackend *backendconfig.TofuBackendConfig

	// PulumiStackFqdn and PulumiBackendURL are the node's pulumi state
	// identity and location (pulumi nodes only).
	PulumiStackFqdn  string
	PulumiBackendURL string

	// KubeContext is the node's resolved kubectl context ("" when none).
	KubeContext string
}

// Plan is a preflighted set: the graph, the deploy order, and every node's
// resolved execution facts. A Plan exists even when the report refuses — the
// renderer needs the graph to show what WOULD deploy — but Execute must never
// run a refused plan.
type Plan struct {
	Set   *manifestgraph.Set
	Graph *manifestgraph.Graph

	// Order is the deployment order as indexes into Set.Nodes; nil when the
	// graph has a cycle.
	Order []int

	// Nodes holds per-node execution facts, indexed like Set.Nodes.
	Nodes []NodePlan

	// Report is the preflight wall's outcome.
	Report *Report
}
