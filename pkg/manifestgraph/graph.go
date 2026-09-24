package manifestgraph

import (
	"fmt"

	"github.com/plantonhq/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// EdgeSource names the composition fact an edge was derived from. Two
// sources are the author's own words; the rest are inferences the graph
// draws from metadata and matching names. The distinction is load-bearing
// once: an inferred edge that closes a cycle with an authored edge yields
// (see BuildGraph), because the author's words always win over an inference.
type EdgeSource string

const (
	// EdgeSourceValueFrom: a valueFrom reference (authored).
	EdgeSourceValueFrom EdgeSource = "value-from"
	// EdgeSourceRelationship: a metadata.relationships entry (authored).
	EdgeSourceRelationship EdgeSource = "relationship"
	// EdgeSourceNamespacePlacement: a literal namespace on a placement field
	// whose namespace the set deploys (inferred).
	EdgeSourceNamespacePlacement EdgeSource = "namespace-placement"
	// EdgeSourceConnectionPlacement: a workload's planton.dev/connection
	// naming the connection a sibling cluster publishes (inferred).
	EdgeSourceConnectionPlacement EdgeSource = "connection-placement"
	// EdgeSourceLiteralSibling: a literal in a default_kind field naming a
	// sibling of that kind by its slug (inferred).
	EdgeSourceLiteralSibling EdgeSource = "literal-sibling"
	// EdgeSourceOperatorPrerequisite: the kind metadata's operator
	// prerequisite, present exactly once in the set (inferred).
	EdgeSourceOperatorPrerequisite EdgeSource = "operator-prerequisite"
)

// Authored reports whether the source is the author's own statement rather
// than an inference of the graph.
func (s EdgeSource) Authored() bool {
	return s == EdgeSourceValueFrom || s == EdgeSourceRelationship
}

// Dependency is one edge: the producer the consumer depends on, the source
// that derived it, and the consumer field that carried the fact ("" when the
// fact lives in metadata rather than in a field).
type Dependency struct {
	Producer  int
	Source    EdgeSource
	FieldPath string
}

// Graph is the set's dependency graph: for each node, the nodes it depends
// on (its producers) with the provenance of each edge, plus everything the
// edge derivation learned that is NOT an in-set edge — derived targets,
// external targets, dropped inferences — as structured findings and
// assumption records.
type Graph struct {
	Set *Set

	// DependsOn[i] lists the dependencies of node i. Deduplicated by
	// producer; when two sources derive the same edge the first one to run
	// keeps it, and the sources run authored-first so an authored edge is
	// never mistaken for an inference.
	DependsOn [][]Dependency

	// Derived are targets implied by literal namespace placement that are
	// not in the set: their existence is the deploy target's own concern
	// (the platform materializes them as skip-only nodes; a backendless
	// deploy simply proceeds — the module fails honestly if the namespace
	// is genuinely absent).
	Derived []Identity

	// Findings carry per-reference rule violations, the external-target
	// classes, and the inferred edges dropped to keep the set orderable.
	// Severity is the consumer's policy (see FindingClass).
	Findings []Finding
}

// Producers returns the plain adjacency list (consumer -> producer indexes)
// the ordering and cycle routines read.
func (g *Graph) Producers() [][]int {
	edges := make([][]int, len(g.DependsOn))
	for consumer, deps := range g.DependsOn {
		for _, dep := range deps {
			edges[consumer] = append(edges[consumer], dep.Producer)
		}
	}
	return edges
}

// BuildGraph derives the set's dependency graph from the manifests' own
// composition facts — the six edge sources the platform's orchestrator
// uses, in the same semantics:
//
//   - valueFrom references, by their EFFECTIVE kind (annotation defaults
//     materialize BEFORE ordering, so an annotation-riding reference orders
//     exactly like an explicit one);
//   - explicit metadata.relationships (every relationship type is an
//     ordering fact: the related resource comes first);
//   - literal namespace placement: a namespace-annotated field holding a
//     LITERAL value implies the namespace, an edge when the set deploys it
//     and a derived record when it does not;
//   - connection placement: a Kubernetes workload whose planton.dev/connection
//     annotation names the connection a sibling cluster will publish runs on
//     that cluster, so the cluster comes first (see connection.go);
//   - literal siblings: a LITERAL in any field annotated with a default kind
//     names a sibling when a node of that kind with that slug is in the set
//     (a route naming its gateway, a certificate naming its issuer) — the
//     shape the console's wizard writes and kubectl users think in. A
//     Kubernetes name is an identity and matches; a cloud id ("sg-0abc")
//     slugs to nothing in the set and never does; no node is ever minted
//     for an unmatched literal;
//   - operator prerequisites: the kind metadata's operator prerequisite,
//     when exactly one instance is in the set (see prerequisite.go).
//
// The first two are the author's words; the other four are inferences. An
// inferred edge that closes a cycle with the author's edges is dropped and
// reported (FindingDerivedEdgeDropped) so the set stays orderable; a cycle
// among authored edges stands and TopoOrder reports it.
//
// A reference that names an env explicitly only forms an edge when that
// identity is in the set; otherwise it is the env-external finding class.
func BuildGraph(set *Set) *Graph {
	g := &Graph{Set: set, DependsOn: make([][]Dependency, len(set.Nodes))}
	publishers := publishedConnectionIndex(set)
	instances := kindInstanceIndex(set)

	seen := make([]map[int]bool, len(set.Nodes))
	addEdge := func(consumer, producer int, source EdgeSource, fieldPath string) {
		if consumer == producer {
			return
		}
		if seen[consumer] == nil {
			seen[consumer] = map[int]bool{}
		}
		if seen[consumer][producer] {
			return
		}
		seen[consumer][producer] = true
		g.DependsOn[consumer] = append(g.DependsOn[consumer], Dependency{Producer: producer, Source: source, FieldPath: fieldPath})
	}

	derivedSeen := map[Identity]bool{}

	for i := range set.Nodes {
		node := &set.Nodes[i]
		nodeID := node.Identity

		// Source 1: valueFrom references.
		for _, use := range node.refUses {
			target, problems := CheckRef(use)
			for _, p := range problems {
				g.Findings = append(g.Findings, Finding{
					Class: FindingRefRule, Source: node.Source, Node: &nodeID,
					FieldPath: use.FieldPath, Message: p,
				})
			}
			if target.Kind == cloudresourcekind.CloudResourceKind_unspecified || target.Name == "" {
				continue
			}
			targetID := target.Identity(nodeID.Env)
			if producer, ok := set.Lookup(targetID); ok {
				addEdge(i, producer, EdgeSourceValueFrom, use.FieldPath)
				continue
			}
			t := target
			class := FindingExternalValueFrom
			reason := "the set does not deploy it — the value must come from a resource that already exists"
			if target.Env != "" && target.Env != nodeID.Env {
				class = FindingEnvExternalValueFrom
				reason = fmt.Sprintf("it names env %q explicitly — a reference into another environment's resources by design", target.Env)
			}
			g.Findings = append(g.Findings, Finding{
				Class: class, Source: node.Source, Node: &nodeID, FieldPath: use.FieldPath, Target: &t,
				Message: fmt.Sprintf("%s: references %s %q outside this set; %s", use.FieldPath, target.Kind, target.Name, reason),
			})
		}

		// Source 2: explicit relationships.
		meta := node.Metadata()
		for ri, rel := range meta.GetRelationships() {
			relTarget := Target{Kind: rel.GetKind(), Name: rel.GetName(), Env: rel.GetEnv()}
			targetID := relTarget.Identity(nodeID.Env)
			fieldPath := fmt.Sprintf("metadata.relationships[%d]", ri)
			if producer, ok := set.Lookup(targetID); ok {
				addEdge(i, producer, EdgeSourceRelationship, fieldPath)
				continue
			}
			t := relTarget
			g.Findings = append(g.Findings, Finding{
				Class: FindingExternalRelationship, Source: node.Source, Node: &nodeID,
				FieldPath: fieldPath, Target: &t,
				Message: fmt.Sprintf("relationship %s %s %q is outside this set — its existence is assumed and verified by the module at apply",
					rel.GetType(), rel.GetKind(), rel.GetName()),
			})
		}

		// Source 3: literal namespace placement.
		for _, nsName := range literalNamespacePlacements(node.Msg) {
			nsID := Identity{Kind: cloudresourcekind.CloudResourceKind_KubernetesNamespace, Slug: GenerateSlug(nsName), Env: nodeID.Env}
			if producer, ok := set.Lookup(nsID); ok {
				addEdge(i, producer, EdgeSourceNamespacePlacement, "")
				continue
			}
			if !derivedSeen[nsID] {
				derivedSeen[nsID] = true
				g.Derived = append(g.Derived, nsID)
			}
		}

		// Source 4: connection placement. A consumed connection no sibling
		// publishes is not a finding: it may be user-authored, published by a
		// cluster deployed elsewhere, or resolved from a default — all backend
		// facts the platform checks at admission, none this lane can see.
		if slug := ConsumedConnectionSlug(node); slug != "" {
			if producer, ok := publishers[slug]; ok {
				addEdge(i, producer, EdgeSourceConnectionPlacement, "")
			}
		}

		// Source 5: literal siblings. The namespace is source 3's (it alone
		// mints a derived target when absent); every other annotated kind
		// matches a sibling or is nothing — never a finding, because a
		// literal that names no sibling is the common, legitimate case (an
		// externally created gateway, a cloud id).
		for _, use := range node.literalUses {
			kind := annotatedKind(use.Field)
			if kind == cloudresourcekind.CloudResourceKind_unspecified || kind == cloudresourcekind.CloudResourceKind_KubernetesNamespace {
				continue
			}
			siblingID := Identity{Kind: kind, Slug: GenerateSlug(use.Value), Env: nodeID.Env}
			if producer, ok := set.Lookup(siblingID); ok {
				addEdge(i, producer, EdgeSourceLiteralSibling, use.FieldPath)
			}
		}

		// Source 6: operator prerequisites, when the set holds exactly one.
		for _, operator := range operatorPrerequisites(nodeID.Kind) {
			if producer := soleInstance(instances, operator, nodeID.Env); producer >= 0 {
				addEdge(i, producer, EdgeSourceOperatorPrerequisite, "")
			}
		}
	}

	g.dropInferredEdgesOnCycles()
	return g
}

// dropInferredEdgesOnCycles keeps the set orderable when an inference closes
// a cycle with the author's own edges: while a cycle exists and carries at
// least one inferred edge, the first inferred edge along the cycle (in
// reference order, so the choice is deterministic) is removed and reported.
// A cycle made only of authored edges is left standing for TopoOrder to
// report — the author wrote it, and the author must break it.
func (g *Graph) dropInferredEdgesOnCycles() {
	for {
		cycle := FindCycle(g.Producers())
		if cycle == nil {
			return
		}
		dropped := false
		for k := range cycle {
			consumer := cycle[k]
			producer := cycle[(k+1)%len(cycle)]
			if g.dropInferredEdge(consumer, producer) {
				dropped = true
				break
			}
		}
		if !dropped {
			return
		}
	}
}

// dropInferredEdge removes the consumer -> producer edge when it is an
// inference, records the finding, and reports whether it did.
func (g *Graph) dropInferredEdge(consumer, producer int) bool {
	deps := g.DependsOn[consumer]
	for k, dep := range deps {
		if dep.Producer != producer || dep.Source.Authored() {
			continue
		}
		g.DependsOn[consumer] = append(deps[:k:k], deps[k+1:]...)
		consumerID := g.Set.Nodes[consumer].Identity
		producerID := g.Set.Nodes[producer].Identity
		t := Target{Kind: producerID.Kind, Name: g.Set.Nodes[producer].Name, Env: producerID.Env}
		g.Findings = append(g.Findings, Finding{
			Class: FindingDerivedEdgeDropped, Source: g.Set.Nodes[consumer].Source, Node: &consumerID,
			FieldPath: dep.FieldPath, Target: &t,
			Message: fmt.Sprintf("%s would order after %s by %s, but that closes a cycle with the authored dependencies — the inferred edge is dropped and the authored order stands",
				consumerID, producerID, dep.Source),
		})
		return true
	}
	return false
}

// annotatedKind reads the default_kind annotation off a StringValueOrRef
// field's descriptor (unspecified when the field carries none).
func annotatedKind(fd protoreflect.FieldDescriptor) cloudresourcekind.CloudResourceKind {
	if fd == nil || fd.Options() == nil {
		return cloudresourcekind.CloudResourceKind_unspecified
	}
	kind, _ := proto.GetExtension(fd.Options(), foreignkeyv1.E_DefaultKind).(cloudresourcekind.CloudResourceKind)
	return kind
}

// literalNamespacePlacements finds top-level spec fields that name a
// kubernetes namespace as a LITERAL: a StringValueOrRef field annotated
// default_kind=KubernetesNamespace and not containment_exempt, holding the
// literal arm. A literal namespace is a placement fact — the namespace must
// exist first — even though no valueFrom reference spells the dependency out.
func literalNamespacePlacements(msg proto.Message) []string {
	var names []string
	top := msg.ProtoReflect()
	specFd := top.Descriptor().Fields().ByName("spec")
	if specFd == nil || specFd.Kind() != protoreflect.MessageKind || !top.Has(specFd) {
		return nil
	}
	spec := top.Get(specFd).Message()
	spec.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		if fd.Kind() != protoreflect.MessageKind || fd.IsMap() || fd.IsList() {
			return true
		}
		if string(fd.Message().FullName()) != stringValueOrRefFullName {
			return true
		}
		opts := fd.Options()
		if opts == nil {
			return true
		}
		if annotatedKind(fd) != cloudresourcekind.CloudResourceKind_KubernetesNamespace {
			return true
		}
		if exempt, _ := proto.GetExtension(opts, foreignkeyv1.E_ContainmentExempt).(bool); exempt {
			return true
		}
		svor, ok := v.Message().Interface().(*foreignkeyv1.StringValueOrRef)
		if !ok || svor.GetValue() == "" {
			return true
		}
		names = append(names, svor.GetValue())
		return true
	})
	return names
}
