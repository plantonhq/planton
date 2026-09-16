// Package manifestgraph is the one home for treating a SET of cloud-resource
// manifests as a dependency graph: node identity, reference collection, the
// strict reference rules, edge derivation from the manifests' own composition
// facts, topological ordering, classification of what resolves inside the set
// versus what points out of it, and rewriting resolved references to literals.
//
// Three consumers share it so their semantics can never diverge: infra-chart
// validation (the authoring gate), the E2E harness's reference resolution,
// and multi-manifest deployment. The package is deliberately a library in the
// strictest sense — it never prints, never exits, and returns structured
// findings whose severity is the CONSUMER's policy: chart validation warns
// about cross-chart references by design (charts compose onto resources owned
// elsewhere), while a backendless deployment must refuse the same class
// (there is no backend to discover the target).
//
// Identity is (kind, slug, env), matching the platform's dependency graph:
// an explicit metadata.slug passes through, otherwise the slug derives from
// metadata.name (see GenerateSlug). Edge-TARGET identity derives through the
// SAME slug function — a selector that slugs a referenced name differently
// would mint a phantom node instead of joining the real one, and the failure
// is silent by design (an unmatched identity is a no-op, never an error).
//
// Edges come from four sources, mirroring the platform's orchestrator:
// valueFrom references (by their EFFECTIVE kind — the field's default_kind
// annotation applies before ordering, so an annotation-riding reference
// orders exactly like an explicit one), explicit metadata.relationships,
// literal namespace placement (a literal value in a namespace-annotated field
// implies the namespace must exist — a derived target, never deployed here),
// and connection placement (a Kubernetes workload whose planton.dev/connection
// annotation names the connection a sibling cluster publishes runs on that
// cluster — see connection.go for the two-ended naming contract).
package manifestgraph
