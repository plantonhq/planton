package generators

import (
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2/hclparse"

	testkubernetesv1 "github.com/plantonhq/planton/apis/dev/planton/provider/_test/testcloudresourcekubernetes/v1"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	kubernetesredisv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/kubernetesredis/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestProtoToTFVars_NamespaceFlattened(t *testing.T) {
	msg := &testkubernetesv1.TestCloudResourceKubernetes{
		ApiVersion: "_test.planton.dev/v1",
		Kind:       "TestCloudResourceKubernetes",
		Spec: &testkubernetesv1.TestCloudResourceKubernetesSpec{
			Namespace: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "e2e-test-ns",
				},
			},
			CreateNamespace: true,
			Schedule:        stringPtr("*/5 * * * *"),
		},
	}

	got, err := ProtoToTFVars(msg)
	if err != nil {
		t.Fatalf("ProtoToTFVars: %v", err)
	}

	if strings.Contains(got, `"value" = "e2e-test-ns"`) {
		t.Errorf("namespace should be flattened to plain string, got nested object:\n%s", got)
	}
	if !strings.Contains(got, `"namespace" = "e2e-test-ns"`) &&
		!strings.Contains(got, `namespace = "e2e-test-ns"`) {
		t.Errorf("namespace should appear as flat string assignment, got:\n%s", got)
	}

	parser := hclparse.NewParser()
	_, diags := parser.ParseHCL([]byte(got), "test.tfvars")
	if diags.HasErrors() {
		t.Errorf("generated tfvars is not valid HCL: %s\n%s", diags.Error(), got)
	}
}

func TestProtoToTFVars_TargetClusterSkipped(t *testing.T) {
	msg := &testkubernetesv1.TestCloudResourceKubernetes{
		ApiVersion: "_test.planton.dev/v1",
		Kind:       "TestCloudResourceKubernetes",
		Spec: &testkubernetesv1.TestCloudResourceKubernetesSpec{
			TargetCluster: &kubernetes.KubernetesClusterSelector{
				ClusterName: "test-cluster",
			},
			Namespace: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "test-ns",
				},
			},
		},
	}

	got, err := ProtoToTFVars(msg)
	if err != nil {
		t.Fatalf("ProtoToTFVars: %v", err)
	}

	if strings.Contains(got, "target_cluster") || strings.Contains(got, "targetCluster") {
		t.Errorf("target_cluster should be skipped, got:\n%s", got)
	}
}

func TestProtoToTFVars_MapRefValuesFlattened(t *testing.T) {
	msg := &testkubernetesv1.TestCloudResourceKubernetes{
		ApiVersion: "_test.planton.dev/v1",
		Kind:       "TestCloudResourceKubernetes",
		Spec: &testkubernetesv1.TestCloudResourceKubernetesSpec{
			Namespace: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "ns",
				},
			},
			RefMap: map[string]*foreignkeyv1.StringValueOrRef{
				"DB_HOST": {
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "localhost",
					},
				},
			},
		},
	}

	got, err := ProtoToTFVars(msg)
	if err != nil {
		t.Fatalf("ProtoToTFVars: %v", err)
	}

	if strings.Contains(got, `"value" = "localhost"`) {
		t.Errorf("map value should be flattened, got nested object:\n%s", got)
	}
	if !strings.Contains(got, `"DB_HOST" = "localhost"`) {
		t.Errorf("DB_HOST should be flat string in map, got:\n%s", got)
	}
}

func TestProtoToTFVars_ApiVersionKindSkipped(t *testing.T) {
	msg := &testkubernetesv1.TestCloudResourceKubernetes{
		ApiVersion: "_test.planton.dev/v1",
		Kind:       "TestCloudResourceKubernetes",
		Spec: &testkubernetesv1.TestCloudResourceKubernetesSpec{
			Namespace: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "ns",
				},
			},
		},
	}

	got, err := ProtoToTFVars(msg)
	if err != nil {
		t.Fatalf("ProtoToTFVars: %v", err)
	}

	if strings.Contains(got, "api_version") || strings.Contains(got, "apiVersion") {
		t.Errorf("apiVersion should be skipped, got:\n%s", got)
	}
	if strings.Contains(got, "kind =") {
		t.Errorf("kind should be skipped, got:\n%s", got)
	}
}

func TestProtoToTFVars_Redis_BackwardCompatible(t *testing.T) {
	msg := &kubernetesredisv1.KubernetesRedis{
		ApiVersion: "kubernetes.planton.dev/v1",
		Kind:       "KubernetesRedis",
		Metadata: &shared.CloudResourceMetadata{
			Name: "red-one",
			Labels: map[string]string{
				"env": "production",
			},
		},
		Spec: &kubernetesredisv1.KubernetesRedisSpec{
			Container: &kubernetesredisv1.KubernetesRedisContainer{
				DiskSize:           "2Gi",
				PersistenceEnabled: true,
				Replicas:           1,
				Resources: &kubernetes.ContainerResources{
					Limits: &kubernetes.CpuMemory{
						Cpu:    "1000m",
						Memory: "1Gi",
					},
					Requests: &kubernetes.CpuMemory{
						Cpu:    "50m",
						Memory: "100Mi",
					},
				},
			},
		},
	}

	got, err := ProtoToTFVars(msg)
	if err != nil {
		t.Fatalf("ProtoToTFVars: %v", err)
	}

	if !strings.Contains(got, `"name" = "red-one"`) {
		t.Errorf("metadata.name missing, got:\n%s", got)
	}

	parser := hclparse.NewParser()
	_, diags := parser.ParseHCL([]byte(got), "test.tfvars")
	if diags.HasErrors() {
		t.Errorf("generated tfvars is not valid HCL: %s\n%s", diags.Error(), got)
	}
}
