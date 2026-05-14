//go:build !codegen
// +build !codegen

package outputs

import (
	"testing"

	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/crkreflect"
)

func TestResolve_ReturnsConcreteType(t *testing.T) {
	kind := cloudresourcekind.CloudResourceKind_Auth0ResourceServer
	msg, err := resolveStackOutputsMessage(kind)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg == nil {
		t.Fatal("expected non-nil message")
	}

	fullName := string(msg.ProtoReflect().Descriptor().FullName())
	expected := "org.openmcf.provider.auth0.auth0resourceserver.v1.Auth0ResourceServerStackOutputs"
	if fullName != expected {
		t.Errorf("expected message type %s, got %s", expected, fullName)
	}
}

// TestResolve_AllRegisteredKinds iterates every CloudResourceKind that has a
// registered top-level message and verifies that StackOutputs resolution
// succeeds. This catches components that don't follow the status.outputs pattern.
func TestResolve_AllRegisteredKinds(t *testing.T) {
	var resolved, skipped, failed int

	for _, kind := range crkreflect.KindsList() {
		_, instErr := crkreflect.NewInstance(kind)
		if instErr != nil {
			skipped++
			continue
		}

		_, err := resolveStackOutputsMessage(kind)
		if err != nil {
			t.Errorf("kind %s: resolve failed: %v", kind.String(), err)
			failed++
			continue
		}
		resolved++
	}

	t.Logf("resolved=%d  skipped=%d  failed=%d", resolved, skipped, failed)

	if failed > 0 {
		t.Errorf("%d kinds failed StackOutputs resolution", failed)
	}
}

func TestResolve_UnknownKind(t *testing.T) {
	_, err := resolveStackOutputsMessage(cloudresourcekind.CloudResourceKind_unspecified)
	if err == nil {
		t.Fatal("expected error for unspecified kind, got nil")
	}
}
