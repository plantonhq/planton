package module

import (
	"fmt"
	"strings"
	"testing"

	kuberneteslokiv1alpha1 "github.com/plantonhq/planton/catalog/kubernetes/kubernetesloki/v1alpha1"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

// These tests pin what the r2 storage arm renders into the chart values:
// the endpoint follows the bucket's jurisdiction, the region is R2's only
// one, the key pair leaves the configuration for the module-owned Secret,
// every Loki pod carries the Secret's fingerprint so a rotated key rolls
// them, and the schema's object store follows the arm.
// The OpenTofu twin is held to the same shapes by the live r2 scenario,
// which writes logs into a real R2 bucket through both engines' installs.

const testSecretAccessKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func r2Locals(jurisdiction, secret string) *Locals {
	return initializeLocals(nil, &kuberneteslokiv1alpha1.KubernetesLokiIacInput{
		Target: &kuberneteslokiv1alpha1.KubernetesLoki{
			Metadata: &shared.CatalogObjectMetadata{Name: "hub-logs"},
			Spec: &kuberneteslokiv1alpha1.KubernetesLokiSpec{
				Namespace: literal("observability"),
				Storage: &kuberneteslokiv1alpha1.KubernetesLokiStorage{
					Backend: &kuberneteslokiv1alpha1.KubernetesLokiStorage_R2{
						R2: &kuberneteslokiv1alpha1.KubernetesLokiR2Storage{
							AccountId:    literal("074755a78d8e8f77c119a90a125e8a06"),
							Jurisdiction: literal(jurisdiction),
							Bucket:       literal("logs-bucket"),
							Credentials: &kuberneteslokiv1alpha1.KubernetesLokiR2Credentials{
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

func lokiValues(t *testing.T, locals *Locals) map[string]interface{} {
	t.Helper()
	values, err := buildHelmValues(locals)
	if err != nil {
		t.Fatalf("buildHelmValues: %v", err)
	}
	return values["loki"].(map[string]interface{})
}

func TestR2ComposesTheS3DialectFromR2sVocabulary(t *testing.T) {
	loki := lokiValues(t, r2Locals("", testSecretAccessKey))
	storage := loki["storage"].(map[string]interface{})
	if storage["type"] != "s3" {
		t.Errorf("loki.storage.type = %v, want s3 (R2 speaks S3)", storage["type"])
	}
	s3 := storage["s3"].(map[string]interface{})
	want := map[string]interface{}{
		"endpoint":         "https://074755a78d8e8f77c119a90a125e8a06.r2.cloudflarestorage.com",
		"region":           "auto",
		"s3ForcePathStyle": true,
		"accessKeyId":      "${LOKI_S3_ACCESS_KEY_ID}",
		"secretAccessKey":  "${LOKI_S3_SECRET_ACCESS_KEY}",
	}
	for key, value := range want {
		if s3[key] != value {
			t.Errorf("loki.storage.s3.%s = %v, want %v", key, s3[key], value)
		}
	}
	buckets := storage["bucketNames"].(map[string]interface{})
	if buckets["chunks"] != "logs-bucket" || buckets["ruler"] != "logs-bucket" {
		t.Errorf("bucketNames = %v, want chunks and ruler in logs-bucket", buckets)
	}
}

func TestR2JurisdictionSelectsItsOwnHost(t *testing.T) {
	loki := lokiValues(t, r2Locals("us", testSecretAccessKey))
	endpoint := loki["storage"].(map[string]interface{})["s3"].(map[string]interface{})["endpoint"]
	if endpoint != "https://074755a78d8e8f77c119a90a125e8a06.us.r2.cloudflarestorage.com" {
		t.Errorf("endpoint for a us bucket = %v; a jurisdictional bucket is served only through its own host", endpoint)
	}
}

func TestR2KeysLeaveTheValuesForTheModuleSecret(t *testing.T) {
	locals := r2Locals("", testSecretAccessKey)
	values, err := buildHelmValues(locals)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(fmt.Sprintf("%v", values), testSecretAccessKey) {
		t.Error("the R2 secret access key appears in the rendered chart values")
	}
	env := values["defaults"].(map[string]interface{})["extraEnv"].([]interface{})
	refs := map[string]string{}
	for _, entry := range env {
		e := entry.(map[string]interface{})
		ref := e["valueFrom"].(map[string]interface{})["secretKeyRef"].(map[string]interface{})
		refs[e["name"].(string)] = ref["name"].(string) + "/" + ref["key"].(string)
	}
	if refs["LOKI_S3_ACCESS_KEY_ID"] != "hub-logs-r2-credentials/access-key-id" ||
		refs["LOKI_S3_SECRET_ACCESS_KEY"] != "hub-logs-r2-credentials/secret-access-key" {
		t.Errorf("credential variables read %v, want the module-owned hub-logs-r2-credentials Secret", refs)
	}
	if locals.R2.SecretData["secret-access-key"] != testSecretAccessKey {
		t.Error("the module-owned Secret does not carry the resolved secret access key")
	}
	if args := values["defaults"].(map[string]interface{})["extraArgs"]; fmt.Sprintf("%v", args) != "[-config.expand-env=true]" {
		t.Errorf("defaults.extraArgs = %v; without expansion the ${VAR} placeholders reach S3 verbatim", args)
	}
}

func TestR2RotatedKeyChangesEveryLokiPodTemplate(t *testing.T) {
	checksum := func(secret string) string {
		annotations := lokiValues(t, r2Locals("", secret))["podAnnotations"].(map[string]interface{})
		return annotations["checksum/credentials"].(string)
	}
	if checksum(testSecretAccessKey) == checksum(strings.Repeat("f", 64)) {
		t.Error("a rotated R2 key left the pod templates unchanged; Loki would keep the old key until someone restarts it")
	}
	if checksum(testSecretAccessKey) != checksum(testSecretAccessKey) {
		t.Error("the checksum is not deterministic; every apply would roll Loki")
	}
}

func TestR2SchemaUsesTheS3ObjectStore(t *testing.T) {
	rendered := fmt.Sprintf("%v", lokiValues(t, r2Locals("", testSecretAccessKey))["schemaConfig"])
	if !strings.Contains(rendered, "object_store:s3") {
		t.Errorf("the derived schema does not store in s3: %s", rendered)
	}
}
