package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A second act must name the same object the first act created, so its
// tokens expand to the FIRST act's run id and scenario slug -- never its own
// file name -- and the file is found beside the authored scenario, not beside
// the first act's expanded temp copy.
func TestPrepareSecondAct_ExpandsWithTheFirstActsValues(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "minimal.yaml")
	second := filepath.Join(dir, "minimal.to.yaml")
	write := func(path, body string) {
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(first, `apiVersion: stripe.planton.dev/v1
kind: StripeCoupon
metadata:
  name: e2e-coupon
  annotations:
    planton.dev/e2e-upgrade-manifest: minimal.to.yaml
spec:
  name: e2e-${E2E_RUN_ID}-${E2E_SCENARIO}
`)
	write(second, `apiVersion: stripe.planton.dev/v1
kind: StripeCoupon
metadata:
  name: e2e-coupon
  annotations:
    planton.dev/e2e-second-act: minimal.yaml
spec:
  name: e2e-${E2E_RUN_ID}-${E2E_SCENARIO}-renamed
`)

	la, err := readLifecycleAnnotations(first)
	if err != nil {
		t.Fatalf("reading the first act: %v", err)
	}
	if la.upgradeManifest != second {
		t.Fatalf("upgrade manifest = %q, want the file beside the authored scenario %q", la.upgradeManifest, second)
	}

	prepared, err := prepareSecondAct(la.upgradeManifest, "ab12cd34-t", ScenarioSlug(first), LaneClock(), nil)
	if err != nil {
		t.Fatalf("preparing the second act: %v", err)
	}
	data, err := os.ReadFile(prepared)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "name: e2e-ab12cd34-t-minimal-renamed") {
		t.Fatalf("the second act must expand to the first act's run id and slug; got:\n%s", data)
	}
	if prepared == second {
		t.Fatal("the authored second act must never be rewritten in place")
	}
}

// Annotations are metadata: they must read from a scenario as authored even while a run token
// stands in a typed number field, which cannot parse into the kind's message before expansion.
func TestManifestAnnotation_ReadsPastATokenInANumberField(t *testing.T) {
	path := filepath.Join(t.TempDir(), "restricted.yaml")
	body := `apiVersion: stripe.planton.dev/v1alpha1
kind: StripePromotionCode
metadata:
  name: e2e-promotion-code
  annotations:
    planton.dev/e2e-required-env: PLANTON_E2E_SOMETHING
spec:
  code: E2E
  expiresAt: ${E2E_UNIX_TIME_PLUS:30d}
`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := ManifestAnnotation(path, ScenarioRequiredEnvAnnotation)
	if err != nil || got != "PLANTON_E2E_SOMETHING" {
		t.Fatalf("annotation = %q, %v; want it read from the YAML", got, err)
	}
}

// A reinstall names either "true" (the scenario again) or a second act beside
// it, which must declare itself the scenario's second act so discovery never
// runs it as a lane of its own.
func TestReadLifecycleAnnotations_ReinstallSecondAct(t *testing.T) {
	scenario := func(t *testing.T, reinstall string, files map[string]string) string {
		t.Helper()
		dir := t.TempDir()
		first := filepath.Join(dir, "restore.yaml")
		body := "apiVersion: kubernetes.planton.dev/v1alpha1\nkind: KubernetesStigmer\nmetadata:\n  name: stigmer\n  annotations:\n    planton.dev/e2e-reinstall: \"" + reinstall + "\"\nspec: {}\n"
		if err := os.WriteFile(first, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		for name, content := range files {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return first
	}
	secondAct := func(firstAct string) string {
		return "apiVersion: kubernetes.planton.dev/v1alpha1\nkind: KubernetesStigmer\nmetadata:\n  name: stigmer\n  annotations:\n    planton.dev/e2e-second-act: " + firstAct + "\nspec: {}\n"
	}

	t.Run("true reinstalls the scenario itself", func(t *testing.T) {
		la, err := readLifecycleAnnotations(scenario(t, "true", nil))
		if err != nil || !la.reinstall || la.reinstallManifest != "" {
			t.Fatalf("got %+v, %v; want a reinstall of the scenario itself", la, err)
		}
	})
	t.Run("a file reinstalls that second act", func(t *testing.T) {
		first := scenario(t, "restore.again.yaml", map[string]string{"restore.again.yaml": secondAct("restore.yaml")})
		la, err := readLifecycleAnnotations(first)
		if err != nil {
			t.Fatal(err)
		}
		if !la.reinstall || la.reinstallManifest != filepath.Join(filepath.Dir(first), "restore.again.yaml") {
			t.Fatalf("got %+v; want the second act beside the scenario", la)
		}
	})
	t.Run("a missing file is refused", func(t *testing.T) {
		_, err := readLifecycleAnnotations(scenario(t, "restore.again.yaml", nil))
		if err == nil || !strings.Contains(err.Error(), "does not exist beside the scenario") {
			t.Fatalf("err %v, want the missing file named", err)
		}
	})
	t.Run("a file not marked as this scenario's second act is refused", func(t *testing.T) {
		first := scenario(t, "restore.again.yaml", map[string]string{"restore.again.yaml": secondAct("other.yaml")})
		_, err := readLifecycleAnnotations(first)
		if err == nil || !strings.Contains(err.Error(), "planton.dev/e2e-second-act: restore.yaml") {
			t.Fatalf("err %v, want the missing second-act mark named", err)
		}
	})
}
