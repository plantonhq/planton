//go:build !codegen
// +build !codegen

package refcheck

import "testing"

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
