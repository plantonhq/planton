package component

import (
	"testing"

	v1 "github.com/plantonhq/planton/operator/api/v1"
)

// The bundled secrets manager is deployed by default: it backs secret storage,
// the credential store, and the OIDC signing key. Opting out is
// the explicit act. The three predicates that answer "is the vault on?" must
// agree, so all three are pinned here against the same arms.
func TestOpenBAO_IsEnabledByDefault(t *testing.T) {
	o := &OpenBAO{}

	p := ingressPlatform(false)
	if !o.IsEnabled(p) || !isVaultEnabled(p) {
		t.Error("vault must be enabled with no spec.vault at all")
	}

	p.Spec.Vault = &v1.OpenBAOSpec{}
	if !o.IsEnabled(p) || !isVaultEnabled(p) {
		t.Error("vault must be enabled with an empty spec.vault")
	}

	off := false
	p.Spec.Vault = &v1.OpenBAOSpec{Enabled: &off}
	if o.IsEnabled(p) || isVaultEnabled(p) {
		t.Error("explicit enabled=false must disable the vault")
	}

	on := true
	p.Spec.Vault = &v1.OpenBAOSpec{Enabled: &on}
	if !o.IsEnabled(p) || !isVaultEnabled(p) {
		t.Error("explicit enabled=true must enable the vault")
	}
}

// The operator passes on only a DECLARED default secret backend: with nothing
// declared, the control plane seeds the default it can serve (the vault when
// it runs, the local backend under the secrets key when it does not), so the
// operator never re-derives that rule -- whichever way the vault is set.
func TestDeclaredSecretBackend_PassesOnOnlyADeclaration(t *testing.T) {
	p := ingressPlatform(false)
	if binding := declaredSecretBackend(p); binding != nil {
		t.Fatalf("a default install declares nothing, got %+v", binding)
	}

	off := false
	p.Spec.Vault = &v1.OpenBAOSpec{Enabled: &off}
	if binding := declaredSecretBackend(p); binding != nil {
		t.Errorf("a vault opt-out declares nothing either, got %+v", binding)
	}

	p.Spec.Bootstrap = &v1.BootstrapSpec{
		SecretBackend: &v1.BootstrapSecretBackendSpec{
			Type: "awsSecretsManager",
			AwsSecretsManager: &v1.BootstrapAwsSecretsManagerSpec{
				Region: "ap-south-1",
			},
		},
	}
	binding := declaredSecretBackend(p)
	if binding == nil || binding.Type != "aws-secrets-manager" || binding.AwsRegion != "ap-south-1" {
		t.Errorf("a declared backend passes on whatever the vault is set to, got %+v", binding)
	}

	p.Spec.Vault = nil
	p.Spec.Bootstrap.SecretBackend = &v1.BootstrapSecretBackendSpec{Type: "platform"}
	if binding := declaredSecretBackend(p); binding == nil || binding.Type != "platform" {
		t.Errorf("a declared platform backend passes on, got %+v", binding)
	}
}
