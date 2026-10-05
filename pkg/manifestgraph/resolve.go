package manifestgraph

import (
	"fmt"
	"strings"

	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

// OutputsLookup supplies a deployed resource's FLATTENED outputs (dotted
// string keys, the platform's flattening) by graph identity. Returning false
// means the identity's outputs are not available — the reference is left
// untouched and reported, and the caller's policy decides what that means
// (a preflighted deployment treats it as a bug; the E2E harness treats it as
// "not deployed by this chain, leave it for the engine's own wall").
type OutputsLookup func(id Identity) (map[string]string, bool)

// ResolveRefs rewrites the manifest's valueFrom references to literal values
// read from deployed resources' outputs. consumerEnv is the manifest's own
// env — the fallback for references that name none.
//
// Semantics, per reference:
//   - the effective kind and field path come from the strict rules (CheckRef);
//     a reference the rules reject is reported and left untouched;
//   - the target's outputs come from the lookup; a lookup miss leaves the
//     reference untouched and reports FindingUnresolvedRef;
//   - a lookup HIT whose outputs lack the referenced field reports
//     FindingMissingOutput — the sharper class, because the resource exists
//     and the composition names a field it does not export;
//   - on success the whole StringValueOrRef is replaced with the literal arm.
//
// Field paths address outputs by their flattened key: the "status.outputs."
// prefix is stripped, the remainder is the dotted key (map-typed composition
// keys address entries by suffix, e.g. "backend_pool_ids.web").
func ResolveRefs(msg proto.Message, consumerEnv string, lookup OutputsLookup) (int, []Finding) {
	resolved := 0
	var findings []Finding

	for _, use := range CollectRefUses(msg) {
		target, problems := CheckRef(use)
		if len(problems) > 0 {
			for _, p := range problems {
				findings = append(findings, Finding{
					Class: FindingRefRule, FieldPath: use.FieldPath, Message: p,
				})
			}
			continue
		}
		if target.Kind == catalogkind.CatalogKind_unspecified || target.Name == "" {
			continue
		}

		targetID := target.Identity(consumerEnv)
		outs, ok := lookup(targetID)
		if !ok {
			t := target
			findings = append(findings, Finding{
				Class: FindingUnresolvedRef, FieldPath: use.FieldPath, Target: &t,
				Message: fmt.Sprintf("%s: no outputs available for %s — the reference stays unresolved", use.FieldPath, targetID),
			})
			continue
		}

		key := strings.TrimPrefix(target.FieldPath, "status.outputs.")
		value, ok := outs[key]
		if !ok {
			t := target
			findings = append(findings, Finding{
				Class: FindingMissingOutput, FieldPath: use.FieldPath, Target: &t,
				Message: fmt.Sprintf("%s: %s has no output %q — the resource deployed but does not export the referenced field", use.FieldPath, targetID, key),
			})
			continue
		}

		use.Replace(&foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
		})
		resolved++
	}

	return resolved, findings
}
