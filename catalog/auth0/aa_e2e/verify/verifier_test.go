package verify

import (
	"testing"
)

// recordingChecker captures the path a verifier asks for, so the tests pin
// the exact Management API path each id turns into.
type recordingChecker struct {
	path   string
	exists bool
}

func (c *recordingChecker) ResourceExists(path string) (bool, error) {
	c.path = path
	return c.exists, nil
}

// A user's id carries "|", a reserved path character; the verifier must
// escape it so the Management API sees one path segment. Every other kind's
// plain id passes through unchanged under the same escaping.
func TestFormatPathEscapesReservedCharacters(t *testing.T) {
	cases := []struct {
		component string
		id        string
		wantPath  string
	}{
		{"auth0user", "auth0|66f1c2d3e4a5b6c7d8e9f0a1", "users/auth0%7C66f1c2d3e4a5b6c7d8e9f0a1"},
		{"auth0role", "rol_abc123", "roles/rol_abc123"},
		{"auth0client", "AbCdEf123", "clients/AbCdEf123"},
	}
	for _, tc := range cases {
		v, err := GetVerifier(tc.component)
		if err != nil {
			t.Fatalf("%s: %v", tc.component, err)
		}
		checker := &recordingChecker{exists: true}
		if err := v.VerifyExists(checker, tc.id); err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.component, err)
		}
		if checker.path != tc.wantPath {
			t.Errorf("%s: path = %q, want %q", tc.component, checker.path, tc.wantPath)
		}
	}
}

// The id output defaults to "id" for every kind that does not name its own;
// the user kind reads its identifier from user_id, the API's own name for it.
func TestIDOutputDefaultsAndOverrides(t *testing.T) {
	for component, want := range map[string]string{
		"auth0client":      "id",
		"auth0role":        "id",
		"auth0eventstream": "id",
		"auth0user":        "user_id",
	} {
		v, err := GetVerifier(component)
		if err != nil {
			t.Fatalf("%s: %v", component, err)
		}
		if got := v.IDOutput(); got != want {
			t.Errorf("%s: IDOutput() = %q, want %q", component, got, want)
		}
	}
}

func TestGetVerifierUnknownComponent(t *testing.T) {
	if _, err := GetVerifier("auth0nothing"); err == nil {
		t.Fatal("expected an error for an unregistered component")
	}
}
