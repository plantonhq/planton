package manifestgraph

import (
	"fmt"
	"strings"

	"github.com/plantonhq/planton/pkg/refannotations"
	"github.com/plantonhq/planton/pkg/refcheck"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// Target is a checked reference's resolved destination: which kind and name
// it points at, the env it names (empty means "the consumer's own env"), and
// the EFFECTIVE field path after annotation defaults applied.
type Target struct {
	Kind      catalogkind.CatalogKind
	Name      string
	Env       string
	FieldPath string
}

// Identity derives the target's graph identity, falling back to the
// consumer's env when the reference names none — the same fallback the
// platform's edge selectors apply. The slug derives through the ONE slug
// function (see the phantom-node warning in the package doc).
func (t Target) Identity(consumerEnv string) Identity {
	env := t.Env
	if env == "" {
		env = consumerEnv
	}
	return Identity{Kind: t.Kind, Slug: GenerateSlug(t.Name), Env: env}
}

// EffectiveKind resolves which kind a reference points at: the reference's
// explicit kind wins, else the field's default_kind annotation, else
// unspecified. This is the kind-half of CheckRef, exported separately because
// resolution lookups need it without the full rule evaluation.
func EffectiveKind(use RefUse) catalogkind.CatalogKind {
	return refannotations.Of(use.Field).EffectiveKind(use.Ref.GetKind())
}

// CheckRef validates one valueFrom reference against the foreign-key
// annotations on its declaring field and the referenced kind's proto surface.
// It returns the resolved target (for dependency-graph construction and
// resolution) and any problems found.
//
// The field's composition keys are its candidate entries plus
// (default_kind, default_kind_field_path) — see pkg/refannotations. The rules,
// in order:
//
//  1. A reference must have a target kind: an explicit valueFrom.kind, or the
//     field's default_kind annotation.
//  2. A reference must have a field path: an explicit valueFrom.fieldPath, or
//     — only when the target kind has exactly one key on this field — that
//     key. A kind the field composes from through more than one output must
//     name which.
//  3. An explicit path on a keyed kind must be one of that kind's keys: the
//     key is the composition the modules are proven to accept, and choosing
//     another output is the id/name/self-link mismatch class that otherwise
//     only surfaces at deploy time. A path that EXTENDS a key is not an
//     override: list elements and map entries are addressed by index or key
//     (`status.outputs.backend_pool_ids.web`), data the annotation cannot
//     name.
//  4. The effective field path must resolve against the target kind's actual
//     proto surface (outputs, spec, or metadata).
//
// The platform's Java reader applies the same rules; the case table in
// rules_test.go is the contract both copy.
func CheckRef(use RefUse) (Target, []string) {
	var problems []string
	annotations := refannotations.Of(use.Field)

	targetKind := annotations.EffectiveKind(use.Ref.GetKind())
	if targetKind == catalogkind.CatalogKind_unspecified {
		problem := fmt.Sprintf("%s: valueFrom does not name a kind and the field declares no default kind — add an explicit `kind:`", use.FieldPath)
		if kinds := annotations.Kinds(); len(kinds) > 0 {
			problem += fmt.Sprintf(" (the field accepts %s)", joinKinds(kinds))
		}
		return Target{}, append(problems, problem)
	}

	keys := annotations.KeysFor(targetKind)
	effectivePath := use.Ref.GetFieldPath()
	switch {
	case effectivePath == "" && len(keys) == 1:
		effectivePath = keys[0]
	case effectivePath == "" && len(keys) > 1:
		problems = append(problems,
			fmt.Sprintf("%s: valueFrom targets %s, which this field composes from more than one output (%s) — add an explicit `fieldPath:` naming one",
				use.FieldPath, targetKind, strings.Join(keys, ", ")))
		return Target{Kind: targetKind, Name: use.Ref.GetName(), Env: use.Ref.GetEnv()}, problems
	case effectivePath == "":
		problems = append(problems,
			fmt.Sprintf("%s: valueFrom targets %s but has no fieldPath, and no annotated default applies — add an explicit `fieldPath:`", use.FieldPath, targetKind))
		return Target{Kind: targetKind, Name: use.Ref.GetName(), Env: use.Ref.GetEnv()}, problems
	case !annotations.AcceptsPath(targetKind, effectivePath):
		contract := fmt.Sprintf("%q", keys[0])
		if len(keys) > 1 {
			contract = "one of " + strings.Join(quoteAll(keys), ", ")
		}
		problems = append(problems,
			fmt.Sprintf("%s: valueFrom overrides the annotated composition key for %s — the field's contract is %s but the reference names %q (id/name/self-link format mismatches only surface at deploy time; use the annotated path)",
				use.FieldPath, targetKind, contract, effectivePath))
	}

	if reason := refcheck.ResolveValueFromPath(targetKind, effectivePath); reason != "" {
		problems = append(problems,
			fmt.Sprintf("%s: valueFrom fieldPath %q does not resolve on %s: %s", use.FieldPath, effectivePath, targetKind, reason))
	}

	return Target{Kind: targetKind, Name: use.Ref.GetName(), Env: use.Ref.GetEnv(), FieldPath: effectivePath}, problems
}

// joinKinds renders kinds for a sentence: "AwsS3Bucket, AwsCloudwatchLogGroup".
func joinKinds(kinds []catalogkind.CatalogKind) string {
	names := make([]string, len(kinds))
	for i, k := range kinds {
		names[i] = k.String()
	}
	return strings.Join(names, ", ")
}

func quoteAll(values []string) []string {
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = fmt.Sprintf("%q", v)
	}
	return quoted
}
