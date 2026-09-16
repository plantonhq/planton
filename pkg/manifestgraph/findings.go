package manifestgraph

import "fmt"

// FindingClass names one class of graph finding. Severity is deliberately
// NOT part of the class: it is consumer policy. Chart validation treats an
// external valueFrom target as a warning (charts compose onto resources
// owned elsewhere by design); a backendless deployment refuses the same
// class (no backend exists to discover the target); the E2E harness treats
// an unresolved reference as legal by design. One vocabulary, three policies.
type FindingClass string

const (
	// FindingDuplicateIdentity: two manifests in the set claim the same
	// (kind, slug, env) — a graph cannot tell them apart.
	FindingDuplicateIdentity FindingClass = "duplicate-identity"

	// FindingRefRule: a reference violates the strict reference rules
	// (missing kind, missing field path, composition-key override, or a
	// field path that does not resolve on the target kind).
	FindingRefRule FindingClass = "ref-rule"

	// FindingExternalValueFrom: a valueFrom reference whose target is not in
	// the set. The reference NEEDS the target's value, so whether this is a
	// warning or a refusal is the consumer's policy.
	FindingExternalValueFrom FindingClass = "external-value-from"

	// FindingEnvExternalValueFrom: a valueFrom reference that names an env
	// explicitly and lands outside the set — it points at another
	// environment's resource by design.
	FindingEnvExternalValueFrom FindingClass = "env-external-value-from"

	// FindingExternalRelationship: a metadata.relationships entry whose
	// target is not in the set. Relationships carry no value need — the
	// target's EXISTENCE is assumed and verified by the module at apply — so
	// this is a stated assumption, not a broken reference.
	FindingExternalRelationship FindingClass = "external-relationship"

	// FindingCycle: the set's dependencies form a cycle; no deploy order
	// exists.
	FindingCycle FindingClass = "cycle"

	// FindingDerivedEdgeDropped: an edge the graph INFERRED (a literal
	// naming a sibling, an operator prerequisite, a namespace or connection
	// placement) closed a cycle with the author's own edges and was dropped
	// so the set stays orderable. The author's words always win over an
	// inference; the drop is reported so the order's reason is visible.
	FindingDerivedEdgeDropped FindingClass = "derived-edge-dropped"

	// FindingUnresolvedRef: resolution had no outputs for the reference's
	// target — the target was not deployed by this run. Whether that is an
	// error is the consumer's policy.
	FindingUnresolvedRef FindingClass = "unresolved-ref"

	// FindingMissingOutput: the target's outputs WERE available but lack the
	// referenced field — the resource deployed and does not export what the
	// reference names. Sharper than unresolved: the composition itself is
	// wrong, not the deployment order.
	FindingMissingOutput FindingClass = "missing-output"
)

// Finding is one structured fact about the set: which node (by source label
// and identity) carries it, which field, which target it concerns, and the
// human sentence. Consumers map findings to their own severity vocabulary.
type Finding struct {
	Class     FindingClass
	Source    string
	Node      *Identity
	FieldPath string
	Target    *Target
	Message   string
}

func (f Finding) String() string {
	if f.Source == "" {
		return fmt.Sprintf("[%s] %s", f.Class, f.Message)
	}
	return fmt.Sprintf("[%s] %s: %s", f.Class, f.Source, f.Message)
}
