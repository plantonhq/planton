package runner

import (
	"testing"

	auth0resourceserverv1 "github.com/plantonhq/planton/apis/dev/planton/provider/auth0/auth0resourceserver/v1"
)

func TestVerifyOutputTransformation_Auth0ResourceServer(t *testing.T) {
	rawOutputs := map[string]interface{}{
		"id":                     "auth0|abc123",
		"identifier":             "https://api.example.com",
		"name":                   "Example API",
		"signing_alg":            "RS256",
		"signing_secret":         nil,
		"token_lifetime":         float64(3600),
		"token_lifetime_for_web": float64(7200),
		"allow_offline_access":   true,
		"skip_consent_for_verifiable_first_party_clients": false,
		"enforce_policies": true,
		"token_dialect":    "access_token_authz",
		"is_system":        false,
		"client_id":        nil,
	}

	msg, flatOutputs, err := VerifyOutputTransformation("auth0resourceserver", rawOutputs, "")
	if err != nil {
		t.Fatalf("VerifyOutputTransformation failed: %v", err)
	}

	if msg == nil {
		t.Fatal("expected non-nil proto message")
	}

	if flatOutputs == nil {
		t.Fatal("expected non-nil flat outputs")
	}

	typed, ok := msg.(*auth0resourceserverv1.Auth0ResourceServerStackOutputs)
	if !ok {
		t.Fatalf("expected *Auth0ResourceServerStackOutputs, got %T", msg)
	}

	assertField(t, "id", typed.GetId(), "auth0|abc123")
	assertField(t, "identifier", typed.GetIdentifier(), "https://api.example.com")
	assertField(t, "name", typed.GetName(), "Example API")
	assertField(t, "signing_alg", typed.GetSigningAlg(), "RS256")
	assertField(t, "token_lifetime", typed.GetTokenLifetime(), "3600")
	assertField(t, "token_lifetime_for_web", typed.GetTokenLifetimeForWeb(), "7200")
	assertField(t, "allow_offline_access", typed.GetAllowOfflineAccess(), "true")
	assertField(t, "enforce_policies", typed.GetEnforcePolicies(), "true")
	assertField(t, "token_dialect", typed.GetTokenDialect(), "access_token_authz")
	assertField(t, "is_system", typed.GetIsSystem(), "false")
}

func TestVerifyOutputTransformation_UnknownComponent(t *testing.T) {
	rawOutputs := map[string]interface{}{"id": "test-123"}

	_, _, err := VerifyOutputTransformation("nonexistent_component_xyz", rawOutputs, "")
	if err == nil {
		t.Fatal("expected error for unknown component, got nil")
	}
}

func TestVerifyOutputTransformation_EmptyOutputs(t *testing.T) {
	rawOutputs := map[string]interface{}{}

	msg, flatOutputs, err := VerifyOutputTransformation("auth0resourceserver", rawOutputs, "")
	if err != nil {
		t.Fatalf("VerifyOutputTransformation failed: %v", err)
	}

	if msg == nil {
		t.Fatal("expected non-nil proto message for empty outputs")
	}

	if len(flatOutputs) != 0 {
		t.Errorf("expected empty flat outputs, got %d entries", len(flatOutputs))
	}
}

func TestVerifyOutputTransformation_FlatOutputsCorrect(t *testing.T) {
	rawOutputs := map[string]interface{}{
		"id":             "test-id",
		"token_lifetime": float64(3600),
		"is_system":      false,
	}

	_, flatOutputs, err := VerifyOutputTransformation("auth0resourceserver", rawOutputs, "")
	if err != nil {
		t.Fatalf("VerifyOutputTransformation failed: %v", err)
	}

	assertFlatEntry(t, flatOutputs, "id", "test-id")
	assertFlatEntry(t, flatOutputs, "token_lifetime", "3600")
	assertFlatEntry(t, flatOutputs, "is_system", "false")
}

func assertField(t *testing.T, fieldName, got, expected string) {
	t.Helper()
	if got != expected {
		t.Errorf("field %s: expected %q, got %q", fieldName, expected, got)
	}
}

func assertFlatEntry(t *testing.T, m map[string]string, key, expected string) {
	t.Helper()
	val, ok := m[key]
	if !ok {
		t.Errorf("expected key %q not found in flat outputs", key)
		return
	}
	if val != expected {
		t.Errorf("flat output %q: expected %q, got %q", key, expected, val)
	}
}
