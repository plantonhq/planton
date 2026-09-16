package manifestgraph

import (
	"fmt"

	"github.com/plantonhq/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Graph is the set's dependency graph: for each node, the indexes of the
// nodes it depends on (its producers), plus everything the edge derivation
// learned that is NOT an in-set edge — derived targets, external targets —
// as structured findings and assumption records.
type Graph struct {
	Set *Set

	// DependsOn[i] lists the node indexes node i depends on. Deduplicated.
	DependsOn [][]int

	// Derived are targets implied by literal namespace placement that are
	// not in the set: their existence is the deploy target's own concern
	// (the platform materializes them as skip-only nodes; a backendless
	// deploy simply proceeds — the module fails honestly if the namespace
	// is genuinely absent).
	Derived []Identity

	// Findings carry per-reference rule violations and the external-target
	// classes. Severity is the consumer's policy (see FindingClass).
	Findings []Finding
}

// BuildGraph derives the set's dependency graph from the manifests' own
// composition facts — the four edge sources the platform's orchestrator
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
//     that cluster, so the cluster comes first (see connection.go).
//
// A reference that names an env explicitly only forms an edge when that
// identity is in the set; otherwise it is the env-external finding class.
func BuildGraph(set *Set) *Graph {
	g := &Graph{Set: set, DependsOn: make([][]int, len(set.Nodes))}
	publishers := publishedConnectionIndex(set)

	seen := make([]map[int]bool, len(set.Nodes))
	addEdge := func(consumer, producer int) {
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
		g.DependsOn[consumer] = append(g.DependsOn[consumer], producer)
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
				addEdge(i, producer)
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
			if producer, ok := set.Lookup(targetID); ok {
				addEdge(i, producer)
				continue
			}
			t := relTarget
			g.Findings = append(g.Findings, Finding{
				Class: FindingExternalRelationship, Source: node.Source, Node: &nodeID,
				FieldPath: fmt.Sprintf("metadata.relationships[%d]", ri), Target: &t,
				Message: fmt.Sprintf("relationship %s %s %q is outside this set — its existence is assumed and verified by the module at apply",
					rel.GetType(), rel.GetKind(), rel.GetName()),
			})
		}

		// Source 3: literal namespace placement.
		for _, nsName := range literalNamespacePlacements(node.Msg) {
			nsID := Identity{Kind: cloudresourcekind.CloudResourceKind_KubernetesNamespace, Slug: GenerateSlug(nsName), Env: nodeID.Env}
			if producer, ok := set.Lookup(nsID); ok {
				addEdge(i, producer)
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
				addEdge(i, producer)
			}
		}
	}

	return g
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
		annotatedKind, _ := proto.GetExtension(opts, foreignkeyv1.E_DefaultKind).(cloudresourcekind.CloudResourceKind)
		if annotatedKind != cloudresourcekind.CloudResourceKind_KubernetesNamespace {
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
