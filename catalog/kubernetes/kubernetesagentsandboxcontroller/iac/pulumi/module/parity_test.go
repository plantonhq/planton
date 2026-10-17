package module

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	v1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesagentsandboxcontroller/v1alpha1"
	"github.com/plantonhq/planton/pkg/failure"
	"github.com/plantonhq/planton/pkg/iac/tofu/generators"
	"github.com/plantonhq/planton/shared"
	"sigs.k8s.io/yaml"
)

// TestEnginesApplyTheSameDocuments is the cross-engine parity proof: for
// each spec, the Terraform module's planned documents (a real `tofu plan`
// of iac/tf, its yaml_body per kubectl_manifest) must equal the documents
// this module composes, object for object, and the planned outputs must equal
// these outputs. The two engines patch upstream's files with separate code,
// so this is the guard that they install the same thing.
//
// It plans against a cluster (the plan-time ownership and definition reads)
// and fetches from GitHub, so it runs when asked: AGENT_SANDBOX_TF_PARITY=1,
// with KUBE_CONFIG_PATH naming a cluster that holds no agent-sandbox install
// and `tofu` on PATH.
func TestEnginesApplyTheSameDocuments(t *testing.T) {
	if os.Getenv("AGENT_SANDBOX_TF_PARITY") != "1" {
		t.Skip("set AGENT_SANDBOX_TF_PARITY=1 (with KUBE_CONFIG_PATH and tofu on PATH) to compare the engines' documents")
	}
	if os.Getenv("KUBE_CONFIG_PATH") == "" {
		t.Fatal("KUBE_CONFIG_PATH must name the cluster the plan reads")
	}
	pluginCache := t.TempDir()

	routerAllowAll := &v1.KubernetesAgentSandboxControllerRouter{
		Authorization: &v1.KubernetesAgentSandboxControllerRouterAuthorization{Mode: &v1.KubernetesAgentSandboxControllerRouterAuthorization_AllowAll{AllowAll: &v1.KubernetesAgentSandboxControllerRouterAllowAll{}}},
	}
	cases := map[string]*v1.KubernetesAgentSandboxControllerSpec{
		"empty": {},
		"extensions with label domains": {
			Extensions: &v1.KubernetesAgentSandboxControllerExtensions{AllowedLabelDomains: []string{"sandbox.users.io", "example.com"}},
		},
		"router with defaults": {Router: routerAllowAll},
		"lean router with token review, both listeners and telemetry": {
			Router: &v1.KubernetesAgentSandboxControllerRouter{
				MinAvailable: i32(0),
				Authorization: &v1.KubernetesAgentSandboxControllerRouterAuthorization{
					Mode: &v1.KubernetesAgentSandboxControllerRouterAuthorization_TokenReview{TokenReview: &v1.KubernetesAgentSandboxControllerRouterTokenReview{
						Ttl: sp("1m"), CacheSize: i32(100), RequireToken: true, Audiences: []string{"sandboxes"},
					}},
				},
				Tls:           &v1.KubernetesAgentSandboxControllerRouterTls{CertificateSecret: literal("router-tls"), Port: i32(9443)},
				Cache:         &v1.KubernetesAgentSandboxControllerRouterCache{Enabled: b(false)},
				NetworkPolicy: &v1.KubernetesAgentSandboxControllerRouterNetworkPolicy{Enabled: b(false)},
				Telemetry:     &v1.KubernetesAgentSandboxControllerRouterTelemetry{OtlpEndpoint: "http://collector:4317", Insecure: true, Headers: secretKeyOf("otel", "headers")},
				Proxy:         &v1.KubernetesAgentSandboxControllerRouterProxy{ShutdownTimeout: sp("1m30s")},
			},
		},
		"every setting": fullSpec(),
	}

	names := make([]string, 0, len(cases))
	for name := range cases {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		spec := cases[name]
		// The test cluster holds no agent-sandbox namespace, so each case
		// owns it; TestTerraformRefusesAMissingNamespace covers the other arm.
		spec.Namespace = literal(vars.Namespace)
		spec.CreateNamespace = true
		t.Run(name, func(t *testing.T) {
			planned, plannedOutputs, err := terraformPlan(t, spec, pluginCache)
			if err != nil {
				t.Fatal(err)
			}

			s := resolveSettings(spec)
			s.identityLabels = identityLabels(parityMetadata)
			files, err := fetchUpstream(s)
			if err != nil {
				t.Fatal(err)
			}
			composed, err := compose(s, files)
			if err != nil {
				t.Fatal(err)
			}
			want := map[string]string{}
			if composed.namespace != nil {
				want[documentID(composed.namespace)] = normalized(t, composed.namespace)
			}
			for _, d := range composed.workloads {
				want[documentID(d)] = normalized(t, d)
			}

			for id, doc := range want {
				if planned[id] == "" {
					t.Errorf("Terraform does not apply %s", id)
					continue
				}
				if planned[id] != doc {
					t.Errorf("%s differs between the engines:\nterraform: %s\npulumi:    %s", id, planned[id], doc)
				}
			}
			for id := range planned {
				if _, ok := want[id]; !ok {
					t.Errorf("Terraform applies %s, which Pulumi does not", id)
				}
			}

			for output, value := range outputValues(s) {
				if normalized(t, plannedOutputs[output]) != normalized(t, value) {
					t.Errorf("output %s: terraform %v, pulumi %v", output, plannedOutputs[output], value)
				}
			}
		})
	}
}

// TestTerraformRefusesAMissingNamespace pins the Terraform twin's refusal of
// a namespace its owner has not created to the Go module's words, so a user
// reads the same three parts on either engine. Same gate as the parity test.
func TestTerraformRefusesAMissingNamespace(t *testing.T) {
	if os.Getenv("AGENT_SANDBOX_TF_PARITY") != "1" {
		t.Skip("set AGENT_SANDBOX_TF_PARITY=1 (with KUBE_CONFIG_PATH and tofu on PATH) to compare the engines' refusals")
	}
	spec := &v1.KubernetesAgentSandboxControllerSpec{Namespace: literal(vars.Namespace)}
	_, _, err := terraformPlan(t, spec, t.TempDir())
	if err == nil {
		t.Fatal("a plan with create_namespace false must be refused while the namespace is missing")
	}
	text := strings.Join(strings.Fields(err.Error()), " ")
	refusal := failure.KubernetesNamespaceMissing(kindName, vars.Namespace)
	for _, part := range []string{refusal.Observed, refusal.Meaning, refusal.NextStep} {
		if !strings.Contains(text, strings.Join(strings.Fields(part), " ")) {
			t.Errorf("the Terraform refusal lacks the Go module's words %q:\n%s", part, text)
		}
	}
}

var parityMetadata = &shared.CatalogObjectMetadata{Name: "parity"}

// terraformPlan plans iac/tf for the spec and returns each kubectl_manifest
// document by its composed identity, and the planned outputs; a refused plan
// returns its output as the error.
func terraformPlan(t *testing.T, spec *v1.KubernetesAgentSandboxControllerSpec, pluginCache string) (map[string]string, map[string]interface{}, error) {
	t.Helper()
	dir := t.TempDir()
	sources, err := filepath.Glob(filepath.Join("..", "..", "tf", "*.tf"))
	if err != nil || len(sources) == 0 {
		t.Fatalf("the Terraform module's files: %v", err)
	}
	for _, source := range sources {
		raw, err := os.ReadFile(source)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, filepath.Base(source)), raw, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	manifest := &v1.KubernetesAgentSandboxController{
		ApiVersion: "kubernetes.planton.dev/v1alpha1",
		Kind:       "KubernetesAgentSandboxController",
		Metadata:   parityMetadata,
		Spec:       spec,
	}
	if err := generators.WriteVarFile(manifest, filepath.Join(dir, "terraform.tfvars")); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) ([]byte, error) {
		t.Helper()
		cmd := exec.Command("tofu", append(args, "-no-color")...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "TF_PLUGIN_CACHE_DIR="+pluginCache)
		out, err := cmd.Output()
		if err != nil {
			stderr := ""
			if exitErr, ok := err.(*exec.ExitError); ok {
				stderr = string(exitErr.Stderr)
			}
			return nil, fmt.Errorf("tofu %v: %v\n%s\n%s", args, err, out, stderr)
		}
		return out, nil
	}
	if _, err := run("init", "-input=false", "-backend=false"); err != nil {
		t.Fatal(err)
	}
	if _, err := run("plan", "-input=false", "-out=plan.bin"); err != nil {
		return nil, nil, err
	}
	shown, err := run("show", "-json", "plan.bin")
	if err != nil {
		t.Fatal(err)
	}

	var plan struct {
		PlannedValues struct {
			Outputs map[string]struct {
				Value interface{} `json:"value"`
			} `json:"outputs"`
			RootModule struct {
				Resources []struct {
					Type   string                 `json:"type"`
					Name   string                 `json:"name"`
					Index  interface{}            `json:"index"`
					Values map[string]interface{} `json:"values"`
				} `json:"resources"`
			} `json:"root_module"`
		} `json:"planned_values"`
	}
	if err := json.Unmarshal(shown, &plan); err != nil {
		t.Fatal(err)
	}
	documents := map[string]string{}
	for _, r := range plan.PlannedValues.RootModule.Resources {
		if r.Type != "kubectl_manifest" || (r.Name != "namespace" && r.Name != "workloads") {
			continue
		}
		var object document
		if err := yaml.Unmarshal([]byte(r.Values["yaml_body"].(string)), &object); err != nil {
			t.Fatal(err)
		}
		documents[documentID(object)] = normalized(t, object)
	}
	outputs := map[string]interface{}{}
	for name, o := range plan.PlannedValues.Outputs {
		outputs[name] = o.Value
	}
	return documents, outputs, nil
}

// normalized is a value's canonical JSON, so integers and floats, and key
// order, never read as differences.
func normalized(t *testing.T, value interface{}) string {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var generic interface{}
	if err := json.Unmarshal(raw, &generic); err != nil {
		t.Fatal(err)
	}
	canonical, _ := json.Marshal(generic)
	return string(canonical)
}
