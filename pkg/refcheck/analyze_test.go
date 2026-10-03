//go:build !codegen
// +build !codegen

package refcheck

import (
	"testing"

	"github.com/plantonhq/planton/shared/catalogkind"
)

// TestForeignKeyReferencesAllResolve enforces the registry-wide invariant: every
// (default_kind_field_path) annotation must resolve against the referenced kind's
// status.outputs (or spec) message. A dangling reference here means a composition
// that silently fails to resolve at deploy time.
func TestForeignKeyReferencesAllResolve(t *testing.T) {
	findings := Analyze()
	for _, f := range findings {
		t.Errorf("dangling foreign-key reference %s:%s -> %s %q: %s",
			f.Kind, f.FieldPath, f.TargetKind, f.RefPath, f.Reason)
	}
	if n := len(findings); n > 0 {
		t.Errorf("%d dangling foreign-key reference(s)", n)
	}
}

// A reference into a kind that names its object with spec.name must not read the Planton
// resource's metadata.name: the path resolves on every kind, so only this rule catches a
// consumer handed a Secret, ConfigMap or ServiceAccount name that does not exist.
func TestOwnNameReason(t *testing.T) {
	cases := []struct {
		name    string
		kind    catalogkind.CatalogKind
		refPath string
		wantBad bool
	}{
		{"metadata.name into a kind with spec.name is refused", catalogkind.CatalogKind_KubernetesSecret, "metadata.name", true},
		{"metadata.name into a ServiceAccount is refused", catalogkind.CatalogKind_KubernetesServiceAccount, "metadata.name", true},
		{"the published name output is sound", catalogkind.CatalogKind_KubernetesSecret, "status.outputs.secret_name", false},
		{"spec.name itself is sound", catalogkind.CatalogKind_KubernetesSecret, "spec.name", false},
		{"metadata.name into a kind without spec.name is sound", catalogkind.CatalogKind_KubernetesClusterIssuer, "metadata.name", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ownNameReason(c.kind, c.refPath) != ""; got != c.wantBad {
				t.Errorf("ownNameReason(%s, %q) refused=%v, want %v", c.kind, c.refPath, got, c.wantBad)
			}
		})
	}
}
