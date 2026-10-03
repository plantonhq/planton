package module

import (
	"fmt"
	"strings"
	"testing"

	kubernetestempov1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetestempo/v1alpha1"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

// These tests pin what the r2 storage arm renders into the chart values:
// the endpoint host follows the bucket's jurisdiction (without a scheme,
// as Tempo's S3 client takes it), the region is R2's only one, the key
// pair leaves the configuration for the module-owned Secret, and the pod
// carries the Secret's fingerprint so a rotated key rolls it.
// The OpenTofu twin is held to the same shapes by the live r2 scenario,
// which writes traces into a real R2 bucket through both engines' installs.

const testSecretAccessKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func r2Locals(jurisdiction, secret string) *Locals {
	return initializeLocals(nil, &kubernetestempov1alpha1.KubernetesTempoIacInput{
		Target: &kubernetestempov1alpha1.KubernetesTempo{
			Metadata: &shared.CatalogObjectMetadata{Name: "hub-traces"},
			Spec: &kubernetestempov1alpha1.KubernetesTempoSpec{
				Namespace: literal("observability"),
				Storage: &kubernetestempov1alpha1.KubernetesTempoStorage{
					Backend: &kubernetestempov1alpha1.KubernetesTempoStorage_R2{
						R2: &kubernetestempov1alpha1.KubernetesTempoR2Storage{
							AccountId:    literal("074755a78d8e8f77c119a90a125e8a06"),
							Jurisdiction: literal(jurisdiction),
							Bucket:       literal("traces-bucket"),
							Credentials: &kubernetestempov1alpha1.KubernetesTempoR2Credentials{
								AccessKeyId:     literal("0f1e2d3c4b5a69788796a5b4c3d2e1f0"),
								SecretAccessKey: literal(secret),
							},
						},
					},
				},
			},
		},
	})
}

func renderTempo(t *testing.T, locals *Locals) map[string]interface{} {
	t.Helper()
	values, err := buildHelmValues(locals)
	if err != nil {
		t.Fatalf("buildHelmValues: %v", err)
	}
	return values
}

func traceStorage(values map[string]interface{}) map[string]interface{} {
	return values["tempo"].(map[string]interface{})["storage"].(map[string]interface{})["trace"].(map[string]interface{})
}

func TestR2ComposesTheS3DialectFromR2sVocabulary(t *testing.T) {
	trace := traceStorage(renderTempo(t, r2Locals("", testSecretAccessKey)))
	if trace["backend"] != "s3" {
		t.Errorf("trace backend = %v, want s3 (R2 speaks S3)", trace["backend"])
	}
	s3 := trace["s3"].(map[string]interface{})
	want := map[string]interface{}{
		"bucket":         "traces-bucket",
		"endpoint":       "074755a78d8e8f77c119a90a125e8a06.r2.cloudflarestorage.com",
		"region":         "auto",
		"forcepathstyle": true,
		"access_key":     "${TEMPO_S3_ACCESS_KEY_ID}",
		"secret_key":     "${TEMPO_S3_SECRET_ACCESS_KEY}",
	}
	for key, value := range want {
		if s3[key] != value {
			t.Errorf("tempo.storage.trace.s3.%s = %v, want %v", key, s3[key], value)
		}
	}
	if _, insecure := s3["insecure"]; insecure {
		t.Error("the r2 arm renders insecure; R2 is reached over TLS")
	}
}

func TestR2JurisdictionSelectsItsOwnHost(t *testing.T) {
	endpoint := traceStorage(renderTempo(t, r2Locals("eu", testSecretAccessKey)))["s3"].(map[string]interface{})["endpoint"]
	if endpoint != "074755a78d8e8f77c119a90a125e8a06.eu.r2.cloudflarestorage.com" {
		t.Errorf("endpoint for an eu bucket = %v; a jurisdictional bucket is served only through its own host", endpoint)
	}
}

func TestR2KeysLeaveTheValuesForTheModuleSecret(t *testing.T) {
	locals := r2Locals("", testSecretAccessKey)
	values := renderTempo(t, locals)
	if strings.Contains(fmt.Sprintf("%v", values), testSecretAccessKey) {
		t.Error("the R2 secret access key appears in the rendered chart values")
	}
	tempo := values["tempo"].(map[string]interface{})
	refs := map[string]string{}
	for _, entry := range tempo["extraEnv"].([]interface{}) {
		e := entry.(map[string]interface{})
		ref := e["valueFrom"].(map[string]interface{})["secretKeyRef"].(map[string]interface{})
		refs[e["name"].(string)] = ref["name"].(string) + "/" + ref["key"].(string)
	}
	if refs["TEMPO_S3_ACCESS_KEY_ID"] != "hub-traces-r2-credentials/access-key-id" ||
		refs["TEMPO_S3_SECRET_ACCESS_KEY"] != "hub-traces-r2-credentials/secret-access-key" {
		t.Errorf("credential variables read %v, want the module-owned hub-traces-r2-credentials Secret", refs)
	}
	if locals.R2.SecretData["secret-access-key"] != testSecretAccessKey {
		t.Error("the module-owned Secret does not carry the resolved secret access key")
	}
	if fmt.Sprintf("%v", tempo["extraArgs"]) != "map[config.expand-env:true]" {
		t.Errorf("tempo.extraArgs = %v; without expansion the ${VAR} placeholders reach S3 verbatim", tempo["extraArgs"])
	}
}

func TestR2RotatedKeyChangesThePodTemplate(t *testing.T) {
	checksum := func(secret string) string {
		annotations := renderTempo(t, r2Locals("", secret))["podAnnotations"].(map[string]interface{})
		return annotations["checksum/credentials"].(string)
	}
	if checksum(testSecretAccessKey) == checksum(strings.Repeat("f", 64)) {
		t.Error("a rotated R2 key left the pod template unchanged; Tempo would keep the old key until someone restarts it")
	}
	if checksum(testSecretAccessKey) != checksum(testSecretAccessKey) {
		t.Error("the checksum is not deterministic; every apply would roll Tempo")
	}
}
