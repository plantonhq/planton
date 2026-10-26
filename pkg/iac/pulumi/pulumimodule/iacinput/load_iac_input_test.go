package iacinput

import (
	"testing"

	kubernetesagentsubstratev1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesagentsubstrate/v1alpha1"
)

// A spec written under an older release breaks a rule its kind's current release added (the Envoy
// router needs at least one client): an apply refuses it, a destroy tears it down.
func TestValidateUnlessDestroying(t *testing.T) {
	spec := &kubernetesagentsubstratev1.KubernetesAgentSubstrateSpec{}

	t.Setenv(OperationEnvVar, "")
	if err := validateUnlessDestroying(spec); err == nil {
		t.Fatal("an apply must judge the spec by its rules")
	}

	t.Setenv(OperationEnvVar, OperationDestroy)
	if err := validateUnlessDestroying(spec); err != nil {
		t.Fatalf("a destroy must not be stopped by a rule on what to create: %v", err)
	}
}
