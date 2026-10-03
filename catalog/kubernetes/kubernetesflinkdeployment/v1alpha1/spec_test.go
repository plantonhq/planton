package kubernetesflinkdeploymentv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesFlinkDeployment(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesFlinkDeployment Suite")
}

func int32Ptr(i int32) *int32    { return &i }
func stringPtr(s string) *string { return &s }

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func valueFrom(kind catalogkind.CatalogKind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{
				Kind:      kind,
				Name:      name,
				FieldPath: fieldPath,
			},
		},
	}
}

func s3Block() *KubernetesFlinkDeploymentS3 {
	return &KubernetesFlinkDeploymentS3{
		Endpoint: literal("http://objects-s3.data.svc.cluster.local:8333"),
		AccessKeySecret: &KubernetesFlinkDeploymentSecretSelector{
			Name: literal("objects-s3-secret"),
			Key:  "admin_access_key_id",
		},
		SecretKeySecret: &KubernetesFlinkDeploymentSecretSelector{
			Name: literal("objects-s3-secret"),
			Key:  "admin_secret_access_key",
		},
	}
}

var _ = ginkgo.Describe("KubernetesFlinkDeployment Validation Tests", func() {
	var input *KubernetesFlinkDeployment

	ginkgo.BeforeEach(func() {
		input = &KubernetesFlinkDeployment{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesFlinkDeployment",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "orders-pipeline",
			},
			Spec: &KubernetesFlinkDeploymentSpec{
				Namespace:    literal("stream-processing"),
				FlinkVersion: "v2_1",
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("a minimal SESSION cluster (no job) should not return a validation error", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "stream-processing", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("every CR flink_version enum value should be valid", func() {
			for _, v := range []string{"v1_13", "v1_14", "v1_15", "v1_16", "v1_17", "v1_18", "v1_19", "v1_20", "v2_0", "v2_1", "v2_2"} {
				input.Spec.FlinkVersion = v
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("a stateless APPLICATION cluster needs no state block", func() {
			input.Spec.Job = &KubernetesFlinkDeploymentJob{
				JarUri: "local:///opt/flink/examples/streaming/StateMachineExample.jar",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a last-state job with a checkpoints dir should be valid", func() {
			input.Spec.Job = &KubernetesFlinkDeploymentJob{
				JarUri:      "local:///opt/pipeline/orders.jar",
				UpgradeMode: stringPtr("last-state"),
			}
			input.Spec.State = &KubernetesFlinkDeploymentState{
				CheckpointsDir: "s3://flink-state/checkpoints/orders",
				S3:             s3Block(),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a savepoint job with both dirs should be valid", func() {
			input.Spec.Job = &KubernetesFlinkDeploymentJob{
				JarUri:      "local:///opt/pipeline/orders.jar",
				UpgradeMode: stringPtr("savepoint"),
				Parallelism: int32Ptr(4),
			}
			input.Spec.State = &KubernetesFlinkDeploymentState{
				CheckpointsDir: "s3://flink-state/checkpoints/orders",
				SavepointsDir:  "s3://flink-state/savepoints/orders",
				S3:             s3Block(),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("JobManager standbys with HA enabled should be valid", func() {
			input.Spec.JobManager = &KubernetesFlinkDeploymentJobManager{Replicas: int32Ptr(2)}
			input.Spec.State = &KubernetesFlinkDeploymentState{
				HighAvailability: &KubernetesFlinkDeploymentHighAvailability{
					Enabled:    true,
					StorageDir: "s3://flink-state/ha/orders",
				},
				S3: s3Block(),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("S3 composing a KubernetesSeaweedFs by reference should be valid", func() {
			input.Spec.State = &KubernetesFlinkDeploymentState{
				S3: &KubernetesFlinkDeploymentS3{
					Endpoint: valueFrom(catalogkind.CatalogKind_KubernetesSeaweedFs, "objects", "status.outputs.s3_endpoint"),
					AccessKeySecret: &KubernetesFlinkDeploymentSecretSelector{
						Name: valueFrom(catalogkind.CatalogKind_KubernetesSeaweedFs, "objects", "status.outputs.s3_credentials_secret_name"),
						Key:  "admin_access_key_id",
					},
					SecretKeySecret: &KubernetesFlinkDeploymentSecretSelector{
						Name: valueFrom(catalogkind.CatalogKind_KubernetesSeaweedFs, "objects", "status.outputs.s3_credentials_secret_name"),
						Key:  "admin_secret_access_key",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("full surface should be valid", func() {
			input.Spec.CreateNamespace = true
			input.Spec.Image = "registry.example.com/pipelines/orders:2.1"
			input.Spec.Job = &KubernetesFlinkDeploymentJob{
				JarUri:                "local:///opt/pipeline/orders.jar",
				EntryClass:            "com.example.OrdersPipeline",
				Args:                  []string{"--source", "kafka"},
				Parallelism:           int32Ptr(8),
				State:                 stringPtr("running"),
				UpgradeMode:           stringPtr("savepoint"),
				AllowNonRestoredState: false,
				SavepointTriggerNonce: 1,
			}
			input.Spec.JobManager = &KubernetesFlinkDeploymentJobManager{Replicas: int32Ptr(2)}
			input.Spec.TaskManager = &KubernetesFlinkDeploymentTaskManager{}
			input.Spec.FlinkConfiguration = map[string]string{
				"taskmanager.numberOfTaskSlots": "2",
			}
			input.Spec.State = &KubernetesFlinkDeploymentState{
				CheckpointsDir: "s3://flink-state/checkpoints/orders",
				SavepointsDir:  "s3://flink-state/savepoints/orders",
				HighAvailability: &KubernetesFlinkDeploymentHighAvailability{
					Enabled:    true,
					StorageDir: "s3://flink-state/ha/orders",
				},
				S3: s3Block(),
			}
			input.Spec.Mode = stringPtr("native")
			input.Spec.ServiceAccount = stringPtr("flink")
			input.Spec.LogConfiguration = map[string]string{
				"log4j-console.properties": "rootLogger.level = INFO",
			}
			input.Spec.Scheduling = &KubernetesFlinkDeploymentScheduling{
				NodeSelector: map[string]string{"workload": "stream"},
			}
			input.Spec.RestartNonce = 3
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("a namespace-less spec should fail", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a spec without flink_version should fail", func() {
			input.Spec.FlinkVersion = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a dotted flink_version (not the CR enum form) should fail", func() {
			input.Spec.FlinkVersion = "2.1"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a job without jar_uri should fail", func() {
			input.Spec.Job = &KubernetesFlinkDeploymentJob{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a last-state job WITHOUT a checkpoints dir should fail (the operator's own rule)", func() {
			input.Spec.Job = &KubernetesFlinkDeploymentJob{
				JarUri:      "local:///opt/pipeline/orders.jar",
				UpgradeMode: stringPtr("last-state"),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a savepoint job WITHOUT a savepoints dir should fail (the operator's own rule)", func() {
			input.Spec.Job = &KubernetesFlinkDeploymentJob{
				JarUri:      "local:///opt/pipeline/orders.jar",
				UpgradeMode: stringPtr("savepoint"),
			}
			input.Spec.State = &KubernetesFlinkDeploymentState{
				CheckpointsDir: "s3://flink-state/checkpoints/orders",
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("JobManager standbys WITHOUT HA should fail (standbys coordinate through HA metadata)", func() {
			input.Spec.JobManager = &KubernetesFlinkDeploymentJobManager{Replicas: int32Ptr(2)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("HA without a storage dir should fail", func() {
			input.Spec.State = &KubernetesFlinkDeploymentState{
				HighAvailability: &KubernetesFlinkDeploymentHighAvailability{Enabled: true},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an S3 block without credentials should fail", func() {
			input.Spec.State = &KubernetesFlinkDeploymentState{
				S3: &KubernetesFlinkDeploymentS3{
					Endpoint: literal("http://objects-s3.data.svc.cluster.local:8333"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unsupported upgrade mode should fail", func() {
			input.Spec.Job = &KubernetesFlinkDeploymentJob{
				JarUri:      "local:///opt/pipeline/orders.jar",
				UpgradeMode: stringPtr("blue-green"),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unsupported mode should fail", func() {
			input.Spec.Mode = stringPtr("session")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
