package verify

import (
	"testing"
)

// recordingChecker captures the path a verifier asks for, so the tests pin
// the exact Management API path each id turns into; body is what a read of
// that path returns.
type recordingChecker struct {
	path   string
	exists bool
	body   map[string]interface{}
}

func (c *recordingChecker) ResourceExists(path string) (bool, error) {
	c.path = path
	return c.exists, nil
}

func (c *recordingChecker) ReadResource(path string) (map[string]interface{}, bool, error) {
	c.path = path
	return c.body, c.exists, nil
}

// A user's id carries "|", a reserved path character; the verifier must
// escape it so the Management API sees one path segment. Every other kind's
// plain id passes through unchanged under the same escaping.
func TestFormatPathEscapesReservedCharacters(t *testing.T) {
	cases := []struct {
		kind     string
		id       string
		wantPath string
	}{
		{"auth0user", "auth0|66f1c2d3e4a5b6c7d8e9f0a1", "users/auth0%7C66f1c2d3e4a5b6c7d8e9f0a1"},
		{"auth0role", "rol_abc123", "roles/rol_abc123"},
		{"auth0client", "AbCdEf123", "clients/AbCdEf123"},
		{"auth0customdomain", "cd_0123456789abcdef", "custom-domains/cd_0123456789abcdef"},
		{"auth0clientfrommetadatadocument", "tpc_AbCdEf123", "clients/tpc_AbCdEf123"},
	}
	for _, tc := range cases {
		v, err := GetVerifier(tc.kind)
		if err != nil {
			t.Fatalf("%s: %v", tc.kind, err)
		}
		checker := &recordingChecker{exists: true}
		if err := v.VerifyExists(checker, tc.id); err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.kind, err)
		}
		if checker.path != tc.wantPath {
			t.Errorf("%s: path = %q, want %q", tc.kind, checker.path, tc.wantPath)
		}
	}
}

// The id output defaults to "id" for every kind that does not name its own;
// the user kind reads its identifier from user_id, the API's own name for it.
func TestIDOutputDefaultsAndOverrides(t *testing.T) {
	for kind, want := range map[string]string{
		"auth0client":                     "id",
		"auth0role":                       "id",
		"auth0eventstream":                "id",
		"auth0user":                       "user_id",
		"auth0clientfrommetadatadocument": "client_id",
	} {
		v, err := GetVerifier(kind)
		if err != nil {
			t.Fatalf("%s: %v", kind, err)
		}
		if got := v.IDOutput(); got != want {
			t.Errorf("%s: IDOutput() = %q, want %q", kind, got, want)
		}
	}
}

func TestGetVerifierUnknownComponent(t *testing.T) {
	if _, err := GetVerifier("auth0nothing"); err == nil {
		t.Fatal("expected an error for an unregistered component")
	}
}

// A verification succeeded only if Auth0 reports the domain ready, and its
// destroy must leave the domain verified: absent-after-destroy is the domain
// still ready, never the domain gone.
func TestCustomDomainVerificationRequiresAReadyDomain(t *testing.T) {
	v, err := GetVerifier("auth0customdomainverification")
	if err != nil {
		t.Fatal(err)
	}
	if v.IDOutput() != "custom_domain_id" {
		t.Fatalf("IDOutput = %q, want custom_domain_id", v.IDOutput())
	}

	ready := &recordingChecker{exists: true, body: map[string]interface{}{"status": "ready"}}
	if err := v.VerifyExists(ready, "cd_0123456789abcdef"); err != nil {
		t.Fatalf("ready domain: %v", err)
	}
	if ready.path != "custom-domains/cd_0123456789abcdef" {
		t.Errorf("path = %q, want custom-domains/cd_0123456789abcdef", ready.path)
	}
	if err := v.VerifyAbsent(ready, "cd_0123456789abcdef"); err != nil {
		t.Fatalf("destroy leaving the domain verified: %v", err)
	}

	pending := &recordingChecker{exists: true, body: map[string]interface{}{"status": "pending_verification"}}
	if err := v.VerifyExists(pending, "cd_0123456789abcdef"); err == nil {
		t.Error("a pending_verification domain passed verification")
	}

	gone := &recordingChecker{exists: false}
	if err := v.VerifyAbsent(gone, "cd_0123456789abcdef"); err == nil {
		t.Error("a destroy that removed the custom domain passed")
	}
}

// A tenant's settings have no identifier and always exist: the verifier names
// no id output, and "absent" after destroy is the settings still answering.
func TestTenantSettingsVerifierReadsTheTenantsSettings(t *testing.T) {
	v, err := GetVerifier("auth0tenantsettings")
	if err != nil {
		t.Fatal(err)
	}
	if v.IDOutput() != "" {
		t.Fatalf("IDOutput = %q, want none", v.IDOutput())
	}
	checker := &recordingChecker{exists: true}
	if err := v.VerifyExists(checker, ""); err != nil {
		t.Fatalf("deployed settings: %v", err)
	}
	if checker.path != "tenants/settings" {
		t.Errorf("path = %q, want tenants/settings", checker.path)
	}
	if err := v.VerifyAbsent(checker, ""); err != nil {
		t.Fatalf("settings left in place by destroy: %v", err)
	}
}
