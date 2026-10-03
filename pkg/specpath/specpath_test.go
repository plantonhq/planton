package specpath

import (
	"strings"
	"testing"

	kindv1 "github.com/plantonhq/planton/iac/catalogkindpermissions/v1"
)

// The tests walk a real committed descriptor (CatalogKindPermissionsSpec) so
// they exercise the same protoreflect surfaces production paths do, with no
// synthetic fixture to maintain.
func TestValidate(t *testing.T) {
	desc := (&kindv1.CatalogKindPermissionsSpec{}).ProtoReflect().Descriptor()

	valid := []string{
		"aws",                             // message leaf
		"aws.statements",                  // repeated message leaf
		"aws.statements.actions",          // traversal THROUGH a repeated field to a repeated scalar leaf
		"kubernetes.rules.cluster_scoped", // scalar leaf
	}
	for _, path := range valid {
		if err := Validate(desc, path); err != nil {
			t.Errorf("Validate(%q) = %v, want nil", path, err)
		}
	}

	invalid := []string{
		"",                             // empty
		"nope",                         // unknown root
		"aws.nope",                     // unknown nested
		"aws.statements.sid.deeper",    // continues past a scalar
		"aws.statements.actions.après", // continues past a repeated scalar
	}
	for _, path := range invalid {
		if err := Validate(desc, path); err == nil {
			t.Errorf("Validate(%q) = nil, want error", path)
		}
	}
}

// TestResolve pins the live-message half's contract: explicit-presence
// semantics, the terminal descriptor surviving absent intermediates, and
// -- the load-bearing pin -- the refusal to traverse THROUGH repeated
// fields, which Validate deliberately permits (cost-driver paths MAY
// walk through repeated fields; paths a live evaluator resolves may
// not).
func TestResolve(t *testing.T) {
	spec := &kindv1.CatalogKindPermissionsSpec{
		Aws: &kindv1.AwsPermissions{
			Statements: []*kindv1.AwsStatement{
				{Sid: "ManageTable", Actions: []string{"dynamodb:CreateTable"}},
			},
		},
	}
	msg := spec.ProtoReflect()

	// A present message leaf.
	resolved, err := Resolve(msg, "aws")
	if err != nil || !resolved.Present {
		t.Fatalf("Resolve(aws): err=%v present=%v, want present", err, resolved.Present)
	}

	// A non-empty repeated terminal resolves present; callers count it.
	resolved, err = Resolve(msg, "aws.statements")
	if err != nil || !resolved.Present || !resolved.Field.IsList() {
		t.Fatalf("Resolve(aws.statements): err=%v present=%v list=%v, want a present list",
			err, resolved.Present, resolved.Field.IsList())
	}
	if resolved.Value.List().Len() != 1 {
		t.Errorf("aws.statements: got %d elements, want 1", resolved.Value.List().Len())
	}

	// Traversal THROUGH a repeated field is refused -- which element
	// would the value come from? Validate accepts this same path.
	if _, err := Resolve(msg, "aws.statements.actions"); err == nil {
		t.Fatal("Resolve(aws.statements.actions) = nil, want the through-repeated refusal")
	} else if !strings.Contains(err.Error(), "cannot pick an element") {
		t.Fatalf("Resolve(aws.statements.actions): got %v, want the cannot-pick-an-element refusal", err)
	}
	if err := Validate((&kindv1.CatalogKindPermissionsSpec{}).ProtoReflect().Descriptor(), "aws.statements.actions"); err != nil {
		t.Fatalf("Validate(aws.statements.actions) = %v -- the asymmetry this pin documents just moved", err)
	}

	// An unset INTERMEDIATE resolves absent with the terminal descriptor
	// intact, so callers can reason about kind and field options even
	// for absent values.
	resolved, err = Resolve(msg, "kubernetes.rules")
	if err != nil {
		t.Fatalf("Resolve(kubernetes.rules): %v", err)
	}
	if resolved.Present {
		t.Error("Resolve(kubernetes.rules): present through an unset intermediate, want absent")
	}
	if resolved.Field == nil || string(resolved.Field.Name()) != "rules" {
		t.Errorf("Resolve(kubernetes.rules): terminal descriptor lost on the absent path")
	}

	// Unknown segments error.
	if _, err := Resolve(msg, "aws.nope"); err == nil {
		t.Error("Resolve(aws.nope) = nil, want error")
	}
	if _, err := Resolve(msg, ""); err == nil {
		t.Error("Resolve(\"\") = nil, want error")
	}
}

// TestResolvableTerminal pins the descriptor-level twin of Resolve's
// traversal contract: it refuses exactly the paths Resolve refuses, so a
// gate validating with it can never bless a path that later errors at
// replay.
func TestResolvableTerminal(t *testing.T) {
	desc := (&kindv1.CatalogKindPermissionsSpec{}).ProtoReflect().Descriptor()

	// Terminal repeated fields are legal, same as Resolve.
	terminal, err := ResolvableTerminal(desc, "aws.statements")
	if err != nil || !terminal.IsList() {
		t.Fatalf("ResolvableTerminal(aws.statements): err=%v, want a repeated terminal", err)
	}

	// Through-repeated traversal is refused, same as Resolve -- while
	// the permissive Terminal accepts it (pinned in TestValidate).
	if _, err := ResolvableTerminal(desc, "aws.statements.actions"); err == nil {
		t.Fatal("ResolvableTerminal(aws.statements.actions) = nil, want the through-repeated refusal")
	} else if !strings.Contains(err.Error(), "cannot pick an element") {
		t.Fatalf("ResolvableTerminal(aws.statements.actions): got %v, want the cannot-pick-an-element refusal", err)
	}

	// Scalar leaves and unknown segments behave like Terminal.
	if _, err := ResolvableTerminal(desc, "kubernetes.rules.cluster_scoped"); err == nil {
		t.Error("ResolvableTerminal through kubernetes.rules (repeated): want the refusal")
	}
	if _, err := ResolvableTerminal(desc, "nope"); err == nil {
		t.Error("ResolvableTerminal(nope) = nil, want error")
	}
	if _, err := ResolvableTerminal(desc, ""); err == nil {
		t.Error("ResolvableTerminal(\"\") = nil, want error")
	}
}
