package kubernetesqdrantv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesQdrant(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesQdrant Suite")
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

var _ = ginkgo.Describe("KubernetesQdrant Validation Tests", func() {
	var input *KubernetesQdrant

	ginkgo.BeforeEach(func() {
		input = &KubernetesQdrant{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesQdrant",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-qdrant",
			},
			Spec: &KubernetesQdrantSpec{
				Namespace: literal("vector"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error (every optional block omitted)", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "vector", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a multi-node cluster should be valid", func() {
			input.Spec.Replicas = int32Ptr(3)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a generated read-write API key alone should be valid", func() {
			input.Spec.ApiKey = &KubernetesQdrantApiKey{
				Source: &KubernetesQdrantApiKey_Generate{Generate: true},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a read-only key alongside a read-write key should be valid", func() {
			input.Spec.ApiKey = &KubernetesQdrantApiKey{
				Source: &KubernetesQdrantApiKey_Generate{Generate: true},
			}
			input.Spec.ReadOnlyApiKey = &KubernetesQdrantApiKey{
				Source: &KubernetesQdrantApiKey_ExistingSecret{
					ExistingSecret: &KubernetesQdrantSecretKeyRef{Name: "qdrant-ro-key", Key: "api-key"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("TLS from a certificate reference should be valid", func() {
			input.Spec.Tls = &KubernetesQdrantTls{
				Secret: valueFrom(catalogkind.CatalogKind_KubernetesCertificate, "qdrant-tls", "status.outputs.secret_name"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("full surface (storage, snapshots, keys, tls, scheduling, image) should be valid", func() {
			input.Spec.CreateNamespace = true
			input.Spec.ChartVersion = stringPtr("1.18.2")
			input.Spec.Replicas = int32Ptr(3)
			input.Spec.Resources = &kubernetes.ContainerResources{
				Requests: &kubernetes.CpuMemory{Cpu: "500m", Memory: "2Gi"},
				Limits:   &kubernetes.CpuMemory{Cpu: "2", Memory: "8Gi"},
			}
			input.Spec.Storage = &KubernetesQdrantDataVolume{
				Size:         stringPtr("50Gi"),
				StorageClass: literal("gp3"),
			}
			input.Spec.Snapshots = &KubernetesQdrantSnapshotsVolume{
				Size:         stringPtr("50Gi"),
				StorageClass: literal("sc-cold"),
			}
			input.Spec.ApiKey = &KubernetesQdrantApiKey{
				Source: &KubernetesQdrantApiKey_Generate{Generate: true},
			}
			input.Spec.ReadOnlyApiKey = &KubernetesQdrantApiKey{
				Source: &KubernetesQdrantApiKey_Generate{Generate: true},
			}
			input.Spec.Tls = &KubernetesQdrantTls{Secret: literal("qdrant-tls")}
			input.Spec.Scheduling = &KubernetesQdrantScheduling{
				NodeSelector: map[string]string{"kubernetes.io/os": "linux"},
				Tolerations: []*kubernetes.WorkloadToleration{
					{Key: "dedicated", Operator: "Equal", Value: "vector", Effect: "NoSchedule"},
				},
				PodAntiAffinity:   true,
				PriorityClassName: "high-priority",
			}
			input.Spec.ServiceMonitorEnabled = true
			input.Spec.Image = &KubernetesQdrantImage{
				Repository:      "mirror.example.com/qdrant/qdrant",
				Tag:             "v1.18.2",
				UseUnprivileged: true,
			}
			input.Spec.HelmValues = "config:\n  storage:\n    performance:\n      max_search_threads: 4\n"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("a missing namespace should fail", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("zero replicas should fail", func() {
			input.Spec.Replicas = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a read-only key WITHOUT a read-write key should fail (the cluster would stay open)", func() {
			input.Spec.ReadOnlyApiKey = &KubernetesQdrantApiKey{
				Source: &KubernetesQdrantApiKey_Generate{Generate: true},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("an existing-secret key reference without a name should fail", func() {
			input.Spec.ApiKey = &KubernetesQdrantApiKey{
				Source: &KubernetesQdrantApiKey_ExistingSecret{
					ExistingSecret: &KubernetesQdrantSecretKeyRef{Key: "api-key"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("an existing-secret key reference without a key should fail", func() {
			input.Spec.ApiKey = &KubernetesQdrantApiKey{
				Source: &KubernetesQdrantApiKey_ExistingSecret{
					ExistingSecret: &KubernetesQdrantSecretKeyRef{Name: "qdrant-key"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("TLS without a certificate secret should fail", func() {
			input.Spec.Tls = &KubernetesQdrantTls{}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a malformed storage size should fail", func() {
			input.Spec.Storage = &KubernetesQdrantDataVolume{Size: stringPtr("ten gigs")}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a malformed snapshots size should fail", func() {
			input.Spec.Snapshots = &KubernetesQdrantSnapshotsVolume{Size: stringPtr("50GB!")}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
