package module

import (
	"strings"
	"testing"

	v1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesstigmer/v1alpha1"
)

// The minimal install renders only what it declares, so the operator's own
// defaults decide everything else.
func TestMinimalStigmerRendersOnlyWhatItDeclares(t *testing.T) {
	spec := paritySpec()
	got := normalized(t, stigmerSpec(spec))
	want := `{"owner":{"email":"owner@acme.example"},"runner":{"llm":{"secretName":"stigmer-model-key"}}}`
	if got != want {
		t.Errorf("minimal Stigmer spec:\n got %s\nwant %s", got, want)
	}
	if values := helmValues(spec); len(values) != 0 {
		t.Errorf("a minimal install sets no chart values, got %v", values)
	}
}

// The model Secret holds only the keys declared: the runner maps every key
// as optional, so a present but empty key would reach it as an empty setting.
func TestModelSecretHoldsOnlyDeclaredKeys(t *testing.T) {
	spec := paritySpec()
	spec.Runner.Model.GetKeys().AnthropicBaseUrl = "http://models.ai.svc.cluster.local:8080"
	secrets := materializedSecrets(spec)
	if len(secrets) != 1 || secrets[0].Name != vars.ModelSecret {
		t.Fatalf("want only the model Secret, got %+v", secrets)
	}
	if got := strings.Join(sortedKeys(secrets[0].Data), ","); got != "ANTHROPIC_API_KEY,ANTHROPIC_BASE_URL" {
		t.Errorf("model Secret keys: %s", got)
	}
}

// A key kept in the user's own Secret is read from there, and the module
// writes no Secret for it.
func TestExistingSecretsAreReferencedNotWritten(t *testing.T) {
	spec := externalSecrets()
	if secrets := materializedSecrets(spec); len(secrets) != 0 {
		t.Errorf("every credential is the user's own Secret, yet the module writes %+v", secrets)
	}
	rendered := stigmerSpec(spec)
	runner := rendered["runner"].(map[string]interface{})
	if runner["llm"].(map[string]interface{})["secretName"] != "model-keys" {
		t.Errorf("the model key's Secret: %v", runner["llm"])
	}
	database := rendered["database"].(map[string]interface{})["postgresql"].(map[string]interface{})["external"].(map[string]interface{})
	if got := normalized(t, database["passwordSecret"]); got != `{"key":"password","name":"stigmer-db-app"}` {
		t.Errorf("the database password's Secret: %s", got)
	}
}

// The chart's image tag is never set: the chart derives it from its version,
// and a second tag would run a different Stigmer from the one declared.
func TestChartValuesNeverCarryAnImageTag(t *testing.T) {
	values := helmValues(bundledEverything())
	image, _ := values["image"].(map[string]interface{})
	if _, set := image["tag"]; set {
		t.Errorf("the chart's image tag is set: %v", image)
	}
	if image["repository"] != "mirror.example.com/stigmer/stigmer-operator" {
		t.Errorf("the mirror is not carried: %v", image)
	}
}

// Secrets in extra settings are written to the module's Secret and read by
// reference; the user's own Secrets are read where they are.
func TestExtraEnvKeepsSecretsOutOfTheSpec(t *testing.T) {
	env := extraEnv(sampleExtraEnv(), vars.ServerExtraEnvSecret)
	got := normalized(t, env)
	for _, want := range []string{
		`{"name":"LOG_FORMAT","value":"json"}`,
		`{"name":"MCP_TOKEN","valueFrom":{"secretKeyRef":{"key":"MCP_TOKEN","name":"stigmer-server-extra-env"}}}`,
		`{"name":"SEARCH_KEY","valueFrom":{"secretKeyRef":{"key":"key","name":"search"}}}`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("extraEnv lacks %s:\n%s", want, got)
		}
	}
	if strings.Contains(got, `"token"`) {
		t.Errorf("a secret value reached the Stigmer's spec: %s", got)
	}
}

// The upgrade annotation uses the operator's own spelling of the version.
func TestUpgradeWithoutBackupAnnotation(t *testing.T) {
	spec := paritySpec()
	spec.UpgradeWithoutBackupTo = "0.3.2"
	if got := stigmerAnnotations(spec)[vars.UpgradeWithoutBackupAnnotation]; got != "v0.3.2" {
		t.Errorf("annotation value: %q", got)
	}
}

func TestVersionBelow(t *testing.T) {
	for _, c := range []struct {
		requested, running string
		below              bool
	}{
		{"0.3.1", "v0.3.2", true},
		{"0.3.2", "v0.3.2", false},
		{"0.3.3", "v0.3.2", false},
		{"0.2.9", "v0.3.0", true},
		{"1.0.0", "v0.9.9", false},
		{"0.3.1", "", false},
		{"0.3.1", "dev", false},
	} {
		if got := versionBelow(c.requested, c.running); got != c.below {
			t.Errorf("versionBelow(%q, %q) = %v, want %v", c.requested, c.running, got, c.below)
		}
	}
}

// The owner's password command appears only when the bundled sign-in runs,
// by Stigmer's own rule.
func TestOwnerPasswordCommandFollowsTheBundledSignIn(t *testing.T) {
	internal := &Locals{Spec: paritySpec(), Namespace: "stigmer"}
	if outputs(internal)[OpOwnerPasswordCommand] != "" {
		t.Error("an internal install has no sign-in, so no owner password")
	}
	exposed := paritySpec()
	exposed.Ingress = bundledEverything().Ingress
	if !strings.Contains(outputs(&Locals{Spec: exposed, Namespace: "stigmer"})[OpOwnerPasswordCommand], "stigmer-owner-setup") {
		t.Error("an exposed install with no identity runs the bundled sign-in")
	}
	exposed.Server = &v1.KubernetesStigmerServer{AllowUnauthenticatedExposure: true}
	if outputs(&Locals{Spec: exposed, Namespace: "stigmer"})[OpOwnerPasswordCommand] != "" {
		t.Error("allow_unauthenticated_exposure turns the sign-in off")
	}
}
