package generators

import (
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2/hclparse"

	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes"
	kubernetescronjobv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetescronjob/v1"
	kubernetesredisv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes/kubernetesredis/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestProtoToTFVars_CronJob_NamespaceFlattened(t *testing.T) {
	msg := &kubernetescronjobv1.KubernetesCronJob{
		ApiVersion: "kubernetes.openmcf.org/v1",
		Kind:       "KubernetesCronJob",
		Spec: &kubernetescronjobv1.KubernetesCronJobSpec{
			Namespace: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "e2e-cronjob-ns",
				},
			},
			CreateNamespace: true,
			Schedule:        "*/5 * * * *",
		},
	}

	got, err := ProtoToTFVars(msg)
	if err != nil {
		t.Fatalf("ProtoToTFVars: %v", err)
	}

	// namespace must be a flat string, not {value = "..."}
	if strings.Contains(got, `"value" = "e2e-cronjob-ns"`) {
		t.Errorf("namespace should be flattened to plain string, got nested object:\n%s", got)
	}
	// After flatten, the key is the proto field name ("namespace") and inside
	// the spec map it's at indent > 0 so it gets quoted.
	if !strings.Contains(got, `"namespace" = "e2e-cronjob-ns"`) &&
		!strings.Contains(got, `namespace = "e2e-cronjob-ns"`) {
		t.Errorf("namespace should appear as flat string assignment, got:\n%s", got)
	}

	// Verify HCL parses cleanly
	parser := hclparse.NewParser()
	_, diags := parser.ParseHCL([]byte(got), "test.tfvars")
	if diags.HasErrors() {
		t.Errorf("generated tfvars is not valid HCL: %s\n%s", diags.Error(), got)
	}
}

func TestProtoToTFVars_CronJob_TargetClusterSkipped(t *testing.T) {
	msg := &kubernetescronjobv1.KubernetesCronJob{
		ApiVersion: "kubernetes.openmcf.org/v1",
		Kind:       "KubernetesCronJob",
		Spec: &kubernetescronjobv1.KubernetesCronJobSpec{
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

func TestProtoToTFVars_CronJob_MapVariablesFlattened(t *testing.T) {
	msg := &kubernetescronjobv1.KubernetesCronJob{
		ApiVersion: "kubernetes.openmcf.org/v1",
		Kind:       "KubernetesCronJob",
		Spec: &kubernetescronjobv1.KubernetesCronJobSpec{
			Namespace: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "ns",
				},
			},
			Env: &kubernetescronjobv1.KubernetesCronJobContainerAppEnv{
				Variables: map[string]*foreignkeyv1.StringValueOrRef{
					"DB_HOST": {
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
							Value: "localhost",
						},
					},
				},
			},
		},
	}

	got, err := ProtoToTFVars(msg)
	if err != nil {
		t.Fatalf("ProtoToTFVars: %v", err)
	}

	// DB_HOST should be a flat string inside the variables map,
	// not a nested {value = "localhost"} object.
	if strings.Contains(got, `"value" = "localhost"`) {
		t.Errorf("map value should be flattened, got nested object:\n%s", got)
	}
	if !strings.Contains(got, `"DB_HOST" = "localhost"`) {
		t.Errorf("DB_HOST should be flat string in map, got:\n%s", got)
	}
}

func TestProtoToTFVars_CronJob_ApiVersionKindSkipped(t *testing.T) {
	msg := &kubernetescronjobv1.KubernetesCronJob{
		ApiVersion: "kubernetes.openmcf.org/v1",
		Kind:       "KubernetesCronJob",
		Spec: &kubernetescronjobv1.KubernetesCronJobSpec{
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
		ApiVersion: "kubernetes.openmcf.org/v1",
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

	// Verify basic structure
	if !strings.Contains(got, `"name" = "red-one"`) {
		t.Errorf("metadata.name missing, got:\n%s", got)
	}

	// Verify HCL validity
	parser := hclparse.NewParser()
	_, diags := parser.ParseHCL([]byte(got), "test.tfvars")
	if diags.HasErrors() {
		t.Errorf("generated tfvars is not valid HCL: %s\n%s", diags.Error(), got)
	}
}
