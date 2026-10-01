package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeTempManifest is shared with refresolve_test.go.

func TestExpandManifestTokens_ReplacesAllOccurrences(t *testing.T) {
	src := writeTempManifest(t, strings.Join([]string{
		"apiVersion: gcp.planton.dev/v1alpha1",
		"kind: GcpWorkloadIdentityPool",
		"metadata:",
		"  name: fixed-name",
		"spec:",
		"  workloadIdentityPoolId: e2e-pool-${E2E_RUN_ID}",
		"  displayName: pool ${E2E_RUN_ID}",
	}, "\n"))

	out, err := ExpandManifestTokens(src, "ab12cd34", "minimal", LaneClock())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == src {
		t.Fatal("expected a new expanded file, got the original path")
	}
	t.Cleanup(func() { os.Remove(out) })

	expanded, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("failed to read expanded manifest: %v", err)
	}
	got := string(expanded)
	if strings.Contains(got, RunIDToken) {
		t.Errorf("expanded manifest still contains %s:\n%s", RunIDToken, got)
	}
	if !strings.Contains(got, "workloadIdentityPoolId: e2e-pool-ab12cd34") {
		t.Errorf("identifier was not expanded:\n%s", got)
	}
	if !strings.Contains(got, "displayName: pool ab12cd34") {
		t.Errorf("second occurrence was not expanded:\n%s", got)
	}

	// The source manifest must be untouched.
	original, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("failed to re-read source manifest: %v", err)
	}
	if !strings.Contains(string(original), RunIDToken) {
		t.Error("source manifest was modified by expansion")
	}
}

func TestExpandManifestTokens_PassthroughWithoutToken(t *testing.T) {
	src := writeTempManifest(t, "apiVersion: gcp.planton.dev/v1alpha1\nkind: GcpServiceAccount\n")

	out, err := ExpandManifestTokens(src, "ab12cd34", "minimal", LaneClock())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != src {
		t.Errorf("expected the original path for a token-free manifest, got %s", out)
	}
}

func TestExpandManifestTokens_ErrorsOnTokenWithoutRunID(t *testing.T) {
	src := writeTempManifest(t, "spec:\n  id: e2e-${E2E_RUN_ID}\n")

	if _, err := ExpandManifestTokens(src, "", "minimal", LaneClock()); err == nil {
		t.Fatal("expected an error when the manifest uses the token but no run id is provided")
	}
}

func TestExpandManifestTokens_KeepsScenarioBasename(t *testing.T) {
	src := writeTempManifest(t, "spec:\n  id: e2e-${E2E_RUN_ID}\n")

	out, err := ExpandManifestTokens(src, "ab12cd34", "minimal", LaneClock())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() { os.Remove(out) })

	// Verifier dispatch keys behavioral variants off the scenario name in
	// the manifest path — the expanded copy must preserve the basename.
	if filepath.Base(out) != filepath.Base(src) {
		t.Errorf("expanded copy basename = %q, want %q", filepath.Base(out), filepath.Base(src))
	}
}

func TestExpandManifestTokens_ExpandsEnvTokens(t *testing.T) {
	t.Setenv("PLANTON_E2E_TEST_ROLE_ARN", "arn:aws:iam::123456789012:role/e2e-test")
	src := writeTempManifest(t, strings.Join([]string{
		"spec:",
		"  irsaRoleArn: ${E2E_ENV:PLANTON_E2E_TEST_ROLE_ARN}",
		"  bucket: e2e-${E2E_RUN_ID}",
	}, "\n"))

	out, err := ExpandManifestTokens(src, "ab12cd34", "minimal", LaneClock())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() { os.Remove(out) })

	expanded, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("failed to read expanded manifest: %v", err)
	}
	got := string(expanded)
	if !strings.Contains(got, "irsaRoleArn: arn:aws:iam::123456789012:role/e2e-test") {
		t.Errorf("env token was not expanded:\n%s", got)
	}
	if !strings.Contains(got, "bucket: e2e-ab12cd34") {
		t.Errorf("run-id token must still expand alongside env tokens:\n%s", got)
	}
}

// Multi-line env values (PEM certificate/key material) must render as a
// double-quoted single-line YAML scalar -- plain splicing would put every
// line after the first outside the field's scalar and corrupt the manifest.
func TestExpandManifestTokens_MultilineEnvValueAsQuotedScalar(t *testing.T) {
	pem := "-----BEGIN CERTIFICATE-----\nMIIBfake\n-----END CERTIFICATE-----"
	t.Setenv("PLANTON_E2E_TEST_CERT_PEM", pem)
	src := writeTempManifest(t, strings.Join([]string{
		"spec:",
		"  certificates: ${E2E_ENV:PLANTON_E2E_TEST_CERT_PEM}",
		"  name: e2e-${E2E_RUN_ID}",
	}, "\n"))

	out, err := ExpandManifestTokens(src, "ab12cd34", "minimal", LaneClock())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() { os.Remove(out) })

	expanded, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("failed to read expanded manifest: %v", err)
	}
	got := string(expanded)
	want := `  certificates: "-----BEGIN CERTIFICATE-----\nMIIBfake\n-----END CERTIFICATE-----"`
	if !strings.Contains(got, want) {
		t.Errorf("multi-line env value not rendered as a quoted scalar:\n%s", got)
	}
	if !strings.Contains(got, "name: e2e-ab12cd34") {
		t.Errorf("run-id token must still expand alongside a multi-line env token:\n%s", got)
	}
}

// A multi-line value has no safe rendering outside the whole-mapping-value
// position -- embedding it mid-string must fail loudly, never write an
// unparseable manifest.
func TestExpandManifestTokens_ErrorsOnEmbeddedMultilineEnvToken(t *testing.T) {
	t.Setenv("PLANTON_E2E_TEST_MULTILINE", "line1\nline2")
	src := writeTempManifest(t, "spec:\n  v: prefix-${E2E_ENV:PLANTON_E2E_TEST_MULTILINE}\n")

	if _, err := ExpandManifestTokens(src, "ab12cd34", "minimal", LaneClock()); err == nil {
		t.Fatal("expected an error for a multi-line env token embedded mid-string")
	}
}

func TestExpandManifestTokens_ErrorsOnUnsetEnvToken(t *testing.T) {
	src := writeTempManifest(t, "spec:\n  arn: ${E2E_ENV:PLANTON_E2E_DEFINITELY_UNSET_VAR}\n")

	if _, err := ExpandManifestTokens(src, "ab12cd34", "minimal", LaneClock()); err == nil {
		t.Fatal("expected an error when an env token's variable is unset")
	}
}

func TestExpandManifestTokens_RejectsNonPrefixedEnvToken(t *testing.T) {
	t.Setenv("AWS_SECRET_ACCESS_KEY", "leak-me")
	src := writeTempManifest(t, "spec:\n  arn: ${E2E_ENV:AWS_SECRET_ACCESS_KEY}\n")

	if _, err := ExpandManifestTokens(src, "ab12cd34", "minimal", LaneClock()); err == nil {
		t.Fatal("expected an error for env tokens outside the allowed prefix")
	}
}

func TestExpandManifestTokens_EnvTokensWithoutRunID(t *testing.T) {
	t.Setenv("PLANTON_E2E_TEST_VALUE", "some-value")
	src := writeTempManifest(t, "spec:\n  v: ${E2E_ENV:PLANTON_E2E_TEST_VALUE}\n")

	out, err := ExpandManifestTokens(src, "", "minimal", LaneClock())
	if err != nil {
		t.Fatalf("env-token-only manifests must not require a run id: %v", err)
	}
	t.Cleanup(func() { os.Remove(out) })
}

// The scenario token gives reservation-window prerequisite identifiers a
// scenario-unique cloud-side name, so a multi-scenario kind never deletes
// and recreates the same name within one engine run.
func TestExpandManifestTokens_ExpandsScenarioToken(t *testing.T) {
	src := writeTempManifest(t, "spec:\n  databaseName: e2e-db-${E2E_SCENARIO}-${E2E_RUN_ID}\n")

	out, err := ExpandManifestTokens(src, "ab12cd34", "enterprise", LaneClock())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() { os.Remove(out) })
	got, _ := os.ReadFile(out)
	if !strings.Contains(string(got), "databaseName: e2e-db-enterprise-ab12cd34") {
		t.Errorf("scenario token not expanded:\n%s", got)
	}
}

func TestExpandManifestTokens_ErrorsOnScenarioTokenWithoutScenario(t *testing.T) {
	src := writeTempManifest(t, "spec:\n  databaseName: e2e-db-${E2E_SCENARIO}\n")

	if _, err := ExpandManifestTokens(src, "ab12cd34", "", LaneClock()); err == nil {
		t.Fatal("expected an error when the scenario token has no scenario in scope")
	}
}

func TestScenarioSlug(t *testing.T) {
	if got := ScenarioSlug("/tmp/x/enterprise.yaml"); got != "enterprise" {
		t.Errorf("slug = %q, want enterprise", got)
	}
	if got := ScenarioSlug(""); got != "" {
		t.Errorf("empty path should give empty slug, got %q", got)
	}
}

func TestEngineScopedRunID_DistinctPerEngine(t *testing.T) {
	pulumiID := EngineScopedRunID("ab12cd34", "pulumi")
	terraformID := EngineScopedRunID("ab12cd34", "terraform")
	if pulumiID == terraformID {
		t.Errorf("engine-scoped ids must differ per engine, both were %q", pulumiID)
	}
	if got, want := pulumiID, "ab12cd34-p"; got != want {
		t.Errorf("pulumi id = %q, want %q", got, want)
	}
	if got, want := terraformID, "ab12cd34-t"; got != want {
		t.Errorf("terraform id = %q, want %q", got, want)
	}
	if got := EngineScopedRunID("ab12cd34", ""); got != "ab12cd34" {
		t.Errorf("empty engine should pass the run id through, got %q", got)
	}
}

// The run-clock token expands to the lane's clock plus the offset, identically in every manifest
// the lane expands, and a malformed or non-future offset fails loudly instead of deploying text.
func TestExpandManifestTokens_RunClock(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	body := "spec:\n  expiresAt: ${E2E_UNIX_TIME_PLUS:30d}\n  activeFrom: ${E2E_UNIX_TIME_PLUS:365d1m}\n"
	clock := LaneClock()
	first, err := ExpandManifestTokens(write("a.yaml", body), "", "", clock)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ExpandManifestTokens(write("b.yaml", body), "", "", clock)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := os.ReadFile(first)
	b, _ := os.ReadFile(second)
	want := fmt.Sprintf("spec:\n  expiresAt: %d\n  activeFrom: %d\n",
		clock.Add(30*24*time.Hour).Unix(), clock.Add(365*24*time.Hour+time.Minute).Unix())
	if string(a) != want || string(b) != want {
		t.Fatalf("both manifests must name the same instants:\n got %q\n and %q\nwant %q", a, b, want)
	}

	for _, bad := range []string{"${E2E_UNIX_TIME_PLUS:soon}", "${E2E_UNIX_TIME_PLUS:0d}", "${E2E_UNIX_TIME_PLUS:-5m}", "${E2E_UNIX_TIME_PLUS:}"} {
		if _, err := ExpandManifestTokens(write("bad.yaml", "spec:\n  expiresAt: "+bad+"\n"), "", "", clock); err == nil {
			t.Errorf("%s must be refused", bad)
		}
	}
}
