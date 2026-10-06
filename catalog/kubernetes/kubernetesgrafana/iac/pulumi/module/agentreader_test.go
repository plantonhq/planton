package module

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	kubernetesprovider "github.com/plantonhq/planton/catalog/kubernetes"
	kubernetesgrafanav1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesgrafana/v1alpha1"
	"github.com/plantonhq/planton/shared"
)

// These tests pin what spec.agent_reader resolves to: the defaults, the
// admin keys the Job reads, the image override, a Job name that moves with
// exactly what the Job reconciles, a Helm rendering the block never
// touches (so turning it on never restarts Grafana), and an environment
// that carries the script's whole contract. The live agent-reader
// scenarios prove the rest through both engines' installs: the token, the
// Viewer role, the refused write, the replacement and the refusal.

func localsWithReader(name string, spec *kubernetesgrafanav1alpha1.KubernetesGrafanaSpec) *Locals {
	spec.Namespace = literal("observability")
	return initializeLocals(nil, &kubernetesgrafanav1alpha1.KubernetesGrafanaIacInput{
		Target: &kubernetesgrafanav1alpha1.KubernetesGrafana{
			Metadata: &shared.CatalogObjectMetadata{Name: name, Org: "acme", Env: "prod"},
			Spec:     spec,
		},
	})
}

func readerOf(t *testing.T, reader *kubernetesgrafanav1alpha1.KubernetesGrafanaAgentReader) *agentReader {
	t.Helper()
	locals := localsWithReader("hub", &kubernetesgrafanav1alpha1.KubernetesGrafanaSpec{AgentReader: reader})
	if locals.AgentReader == nil {
		t.Fatal("a declared agent_reader resolved to nothing")
	}
	return locals.AgentReader
}

func int32Ptr(i int32) *int32    { return &i }
func stringPtr(s string) *string { return &s }

func TestAnEmptyAgentReaderResolvesToTheDefaults(t *testing.T) {
	reader := readerOf(t, &kubernetesgrafanav1alpha1.KubernetesGrafanaAgentReader{})
	checks := map[string][2]string{
		"name":             {reader.Name, "hub-agent-reader"},
		"service account":  {reader.ServiceAccountName, "agent-reader"},
		"script":           {reader.ScriptName, "hub-agent-reader-script"},
		"image":            {reader.Image, "docker.io/alpine/k8s:1.35.8"},
		"admin secret":     {reader.AdminSecretName, "hub"},
		"admin user key":   {reader.AdminUserKey, "admin-user"},
		"admin password":   {reader.AdminPasswordKey, "admin-password"},
		"pull secret name": {reader.PullSecretName, ""},
	}
	for what, c := range checks {
		if c[0] != c[1] {
			t.Errorf("%s = %q, want %q", what, c[0], c[1])
		}
	}
	if reader.TokenGeneration != 1 || reader.Disabled {
		t.Errorf("generation %d disabled %v, want 1 and false", reader.TokenGeneration, reader.Disabled)
	}
	if !strings.HasPrefix(reader.JobName, "hub-agent-reader-") || len(reader.JobName) != len("hub")+22 {
		t.Errorf("job name %q is not <name>-agent-reader-<8 hex>", reader.JobName)
	}
}

func TestTheJobReadsTheDeclaredAdminSecretsKeys(t *testing.T) {
	locals := localsWithReader("hub", &kubernetesgrafanav1alpha1.KubernetesGrafanaSpec{
		AdminSecret: &kubernetesgrafanav1alpha1.KubernetesGrafanaAdminSecret{
			Name: "grafana-admin", UserKey: stringPtr("user"), PasswordKey: stringPtr("pass"),
		},
		AgentReader: &kubernetesgrafanav1alpha1.KubernetesGrafanaAgentReader{},
	})
	r := locals.AgentReader
	if r.AdminSecretName != "grafana-admin" || r.AdminUserKey != "user" || r.AdminPasswordKey != "pass" {
		t.Errorf("admin = %s/%s/%s, want grafana-admin/user/pass", r.AdminSecretName, r.AdminUserKey, r.AdminPasswordKey)
	}
}

func TestAnImageOverrideKeepsTheDefaultForWhatItLeavesEmpty(t *testing.T) {
	cases := map[string]struct {
		image *kubernetesprovider.ContainerImage
		want  string
	}{
		"repo only": {&kubernetesprovider.ContainerImage{Repo: "mirror.example.com/alpine/k8s"}, "mirror.example.com/alpine/k8s:1.35.8"},
		"tag only":  {&kubernetesprovider.ContainerImage{Tag: "1.34.2"}, "docker.io/alpine/k8s:1.34.2"},
		"both":      {&kubernetesprovider.ContainerImage{Repo: "mirror.example.com/k8s", Tag: "2"}, "mirror.example.com/k8s:2"},
	}
	for name, c := range cases {
		reader := readerOf(t, &kubernetesgrafanav1alpha1.KubernetesGrafanaAgentReader{Image: c.image})
		if reader.Image != c.want {
			t.Errorf("%s: image = %q, want %q", name, reader.Image, c.want)
		}
	}
}

func TestTheJobNameMovesWithExactlyWhatTheJobReconciles(t *testing.T) {
	base := readerOf(t, &kubernetesgrafanav1alpha1.KubernetesGrafanaAgentReader{}).JobName
	if again := readerOf(t, &kubernetesgrafanav1alpha1.KubernetesGrafanaAgentReader{}).JobName; again != base {
		t.Fatalf("an unchanged declaration renamed the Job: %s then %s", base, again)
	}
	moved := map[string]*kubernetesgrafanav1alpha1.KubernetesGrafanaAgentReader{
		"account":    {ServiceAccountName: stringPtr("agent-teammates")},
		"generation": {TokenGeneration: int32Ptr(2)},
		"disabled":   {Disabled: true},
	}
	for what, declared := range moved {
		if readerOf(t, declared).JobName == base {
			t.Errorf("a changed %s kept the Job's name, so the change would never run", what)
		}
	}
	// The image is how the Job runs, not what it reconciles: a new image
	// replaces the Job under its name (ReplaceOnChanges) without a new run
	// identity.
	image := &kubernetesprovider.ContainerImage{Repo: "mirror.example.com/alpine/k8s"}
	if readerOf(t, &kubernetesgrafanav1alpha1.KubernetesGrafanaAgentReader{Image: image}).JobName != base {
		t.Error("an image override renamed the Job")
	}
}

func TestTheJobNameHashesTheCanonicalStringBothEnginesBuild(t *testing.T) {
	// OpenTofu's twin: substr(sha256(join("|", [name, tostring(generation),
	// disabled ? "true" : "false", sha256(local.agent_reader_script)])), 0, 8).
	script := sha256.Sum256([]byte(agentReaderScript))
	sum := sha256.Sum256([]byte("agent-teammates|3|false|" + hex.EncodeToString(script[:])))
	want := "hub-agent-reader-" + hex.EncodeToString(sum[:])[:8]
	if got := agentReaderJobName("hub", "agent-teammates", 3, false); got != want {
		t.Errorf("job name = %s, want %s", got, want)
	}
}

func TestTheAgentReaderNeverTouchesTheHelmValues(t *testing.T) {
	without := renderedValues(t, localsWithReader("hub", &kubernetesgrafanav1alpha1.KubernetesGrafanaSpec{}))
	with := renderedValues(t, localsWithReader("hub", &kubernetesgrafanav1alpha1.KubernetesGrafanaSpec{
		AgentReader: &kubernetesgrafanav1alpha1.KubernetesGrafanaAgentReader{TokenGeneration: int32Ptr(4)},
	}))
	if !reflect.DeepEqual(without, with) {
		t.Error("declaring agent_reader changed the chart values, which would restart Grafana")
	}
}

func TestNoAgentReaderRendersNothing(t *testing.T) {
	if localsWithReader("hub", &kubernetesgrafanav1alpha1.KubernetesGrafanaSpec{}).AgentReader != nil {
		t.Error("an agent reader resolved without the block")
	}
}

func TestTheEnvironmentCarriesTheScriptsWholeContract(t *testing.T) {
	locals := localsWithReader("hub", &kubernetesgrafanav1alpha1.KubernetesGrafanaSpec{
		AgentReader: &kubernetesgrafanav1alpha1.KubernetesGrafanaAgentReader{},
	})
	provided := map[string]string{"GRAFANA_ADMIN_USER": "", "GRAFANA_ADMIN_PASSWORD": ""}
	for _, kv := range agentReaderEnv(locals) {
		provided[kv[0]] = kv[1]
	}
	for name := range provided {
		if name != "HOME" && !strings.Contains(agentReaderScript, "$"+name) {
			t.Errorf("the module sets %s, which the script never reads", name)
		}
	}
	for _, read := range []string{"GRAFANA_URL", "ADMIN_SECRET", "NAMESPACE", "RELEASE_NAME", "SERVICE_ACCOUNT",
		"TOKEN_GENERATION", "DISABLED", "TOKEN_SECRET", "OWNER_SERVICE_ACCOUNT", "SECRET_LABELS"} {
		if _, ok := provided[read]; !ok {
			t.Errorf("the script reads $%s, which the module never sets", read)
		}
	}
	labels, _ := json.Marshal(locals.Labels)
	if got := provided["SECRET_LABELS"]; got != string(labels) || !strings.HasPrefix(got, `{"planton.ai/environment":"prod",`) {
		t.Errorf("SECRET_LABELS = %s, want the resource's labels %s as key-sorted JSON (OpenTofu's jsonencode)", got, labels)
	}
	if provided["GRAFANA_URL"] != "http://hub.observability.svc.cluster.local" {
		t.Errorf("GRAFANA_URL = %s, want the in-cluster endpoint", provided["GRAFANA_URL"])
	}
}
