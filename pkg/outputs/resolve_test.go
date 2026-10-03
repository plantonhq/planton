//go:build !codegen
// +build !codegen

package outputs

import (
	"testing"

	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/shared/catalogkind"
)

func TestResolve_ReturnsConcreteType(t *testing.T) {
	kind := catalogkind.CatalogKind_Auth0ResourceServer
	msg, err := resolveOutputsMessage(kind)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg == nil {
		t.Fatal("expected non-nil message")
	}

	fullName := string(msg.ProtoReflect().Descriptor().FullName())
	expected := "dev.planton.auth0.auth0resourceserver.v1alpha1.Auth0ResourceServerOutputs"
	if fullName != expected {
		t.Errorf("expected message type %s, got %s", expected, fullName)
	}
}

// TestResolve_AllRegisteredKinds iterates every CatalogKind that has a
// registered top-level message and verifies that Outputs resolution
// succeeds. This catches kinds that don't follow the status.outputs pattern.
func TestResolve_AllRegisteredKinds(t *testing.T) {
	var resolved, skipped, failed int

	for _, kind := range catalogkindreflect.KindsList() {
		_, instErr := catalogkindreflect.NewInstance(kind)
		if instErr != nil {
			skipped++
			continue
		}

		_, err := resolveOutputsMessage(kind)
		if err != nil {
			t.Errorf("kind %s: resolve failed: %v", kind.String(), err)
			failed++
			continue
		}
		resolved++
	}

	t.Logf("resolved=%d  skipped=%d  failed=%d", resolved, skipped, failed)

	if failed > 0 {
		t.Errorf("%d kinds failed Outputs resolution", failed)
	}
}

func TestResolve_UnknownKind(t *testing.T) {
	_, err := resolveOutputsMessage(catalogkind.CatalogKind_unspecified)
	if err == nil {
		t.Fatal("expected error for unspecified kind, got nil")
	}
}
