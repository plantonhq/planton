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

	v1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesagentsubstrate/v1alpha1"
	"github.com/plantonhq/planton/pkg/failure"
	"github.com/plantonhq/planton/pkg/iac/tofu/generators"
	"github.com/plantonhq/planton/shared"
	"sigs.k8s.io/yaml"
)

// TestEnginesApplyTheSameDocuments is the cross-engine parity proof: for
// each spec, the Terraform module's planned documents (a real `tofu plan`
// of iac/tf, its yaml_body per kubectl_manifest) must equal the documents
// this module composes, object for object, and the planned outputs these
// outputs. The two engines patch upstream's files with separate code, so
// this is the guard that they install the same thing. The generated Secrets
// are typed resources, outside the comparison; their pools are proven by
// TestBothEnginesPoolsParseWithUpstreamsCode.
//
// It plans against a cluster (the plan-time ownership, stamp and definition
// reads) and fetches the published release, so it runs when asked:
// AGENT_SUBSTRATE_TF_PARITY=1, with KUBE_CONFIG_PATH naming a cluster that
// holds no Substrate install and `tofu` on PATH.
func TestEnginesApplyTheSameDocuments(t *testing.T) {
	if os.Getenv("AGENT_SUBSTRATE_TF_PARITY") != "1" {
		t.Skip("set AGENT_SUBSTRATE_TF_PARITY=1 (with KUBE_CONFIG_PATH and tofu on PATH) to compare the engines' documents")
	}
	if os.Getenv("KUBE_CONFIG_PATH") == "" {
		t.Fatal("KUBE_CONFIG_PATH must name the cluster the plan reads")
	}
	pluginCache := t.TempDir()

	agentgateway := minimalSpec()
	agentgateway.Router = &v1.KubernetesAgentSubstrateRouter{Dataplane: v1.KubernetesAgentSubstrateRouter_agentgateway, Replicas: i32(2)}
	agentgateway.Egress.CredentialProvider = &v1.KubernetesAgentSubstrateCredentialProvider{Kind: &v1.KubernetesAgentSubstrateCredentialProvider_External{
		External: &v1.KubernetesAgentSubstrateExternalCredentialProvider{Name: "secretmanager.googleapis.com", Address: "gsm.ate-system.svc:50051"},
	}}
	agentgateway.Database = &v1.KubernetesAgentSubstrateDatabase{Kind: &v1.KubernetesAgentSubstrateDatabase_External{External: &v1.KubernetesAgentSubstrateExternalDatabase{
		ReadWriteConnectionString: &v1.KubernetesAgentSubstrateSecretKeyRef{Name: literal("substrate-db"), Key: "rw"},
		ServerCaCertificate:       "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----\n",
	}}}

	envoyEverything := minimalSpec()
	envoyEverything.ClusterSize = v1.KubernetesAgentSubstrateSpec_size10
	envoyEverything.CordonControlPlane = true
	envoyEverything.ImageRegistry = "mirror.example.com"
	envoyEverything.Egress = &v1.KubernetesAgentSubstrateEgress{
		CredentialProvider: &v1.KubernetesAgentSubstrateCredentialProvider{Kind: &v1.KubernetesAgentSubstrateCredentialProvider_KubernetesSecrets{
			KubernetesSecrets: &v1.KubernetesAgentSubstrateKubernetesSecretsProvider{Grants: []*v1.KubernetesAgentSubstrateNamespaceGrant{{Atespace: "team-a", AllowedNamespaces: []string{"team-a", "shared"}}}},
		}},
		EnvoyConcurrency:         i32(4),
		AdditionalExtProcService: &v1.KubernetesAgentSubstrateExtProcService{Namespace: "policy", Service: "filter", Port: 9443},
	}

	cases := map[string]*v1.KubernetesAgentSubstrateSpec{
		"minimal":                    minimalSpec(),
		"gke with cloud sql and gcs": gkeSpec(),
		"agentgateway, external provider and database":           agentgateway,
		"envoy with size10, cordon, mirror, provider and filter": envoyEverything,
		"every setting": fullSurfaceSpec(),
	}
	names := make([]string, 0, len(cases))
	for name := range cases {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		spec := cases[name]
		// The test cluster holds no Substrate namespace, so each case owns it;
		// TestTerraformRefusesAMissingNamespace covers the other arm.
		spec.CreateNamespace = true
		t.Run(name, func(t *testing.T) {
			planned, plannedOutputs, err := terraformPlan(t, spec, pluginCache)
			if err != nil {
				t.Fatal(err)
			}
			s := resolveSettings(spec)
			s.identityLabels = identityLabels(parityMetadata)
			files, err := fetchRelease(s.version)
			if err != nil {
				t.Fatal(err)
			}
			composed, err := compose(s, files)
			if err != nil {
				t.Fatal(err)
			}
			want := map[string]string{}
			for _, d := range composed.all() {
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
			for output, value := range outputValues(s, composed.images) {
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
	if os.Getenv("AGENT_SUBSTRATE_TF_PARITY") != "1" {
		t.Skip("set AGENT_SUBSTRATE_TF_PARITY=1 (with KUBE_CONFIG_PATH and tofu on PATH) to compare the engines' refusals")
	}
	_, _, err := terraformPlan(t, minimalSpec(), t.TempDir())
	if err == nil {
		t.Fatal("a plan with create_namespace false must be refused while ate-system is missing")
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
func terraformPlan(t *testing.T, spec *v1.KubernetesAgentSubstrateSpec, pluginCache string) (map[string]string, map[string]interface{}, error) {
	t.Helper()
	dir := t.TempDir()
	if err := copyTerraformModule(dir); err != nil {
		t.Fatal(err)
	}
	manifest := &v1.KubernetesAgentSubstrate{
		ApiVersion: "kubernetes.planton.dev/v1alpha1",
		Kind:       "KubernetesAgentSubstrate",
		Metadata:   parityMetadata,
		Spec:       spec,
	}
	if err := generators.WriteVarFile(manifest, filepath.Join(dir, "terraform.tfvars")); err != nil {
		t.Fatal(err)
	}
	if _, err := tofu(dir, pluginCache, "init", "-input=false", "-backend=false", "-lockfile=readonly"); err != nil {
		t.Fatal(err)
	}
	if _, err := tofu(dir, pluginCache, "plan", "-input=false", "-out=plan.bin"); err != nil {
		return nil, nil, err
	}
	shown, err := tofu(dir, pluginCache, "show", "-json", "plan.bin")
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
		body, ok := r.Values["yaml_body"].(string)
		if r.Type != "kubectl_manifest" || !ok {
			continue
		}
		var object document
		if err := yaml.Unmarshal([]byte(body), &object); err != nil {
			t.Fatal(err)
		}
		if str(object, "kind") == "CustomResourceDefinition" {
			continue
		}
		documents[documentID(object)] = normalized(t, object)
	}
	outputs := map[string]interface{}{}
	for name, o := range plan.PlannedValues.Outputs {
		outputs[name] = o.Value
	}
	return documents, outputs, nil
}

// copyTerraformModule copies iac/tf (its .tf files and provider lock) into
// dir.
func copyTerraformModule(dir string) error {
	sources, err := filepath.Glob(filepath.Join("..", "..", "tf", "*.tf"))
	if err != nil || len(sources) == 0 {
		return fmt.Errorf("the Terraform module's files: %v", err)
	}
	sources = append(sources, filepath.Join("..", "..", "tf", ".terraform.lock.hcl"))
	for _, source := range sources {
		raw, err := os.ReadFile(source)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, filepath.Base(source)), raw, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func tofu(dir, pluginCache string, args ...string) ([]byte, error) {
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
