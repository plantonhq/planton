package kubernetesopenfgav1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestKubernetesOpenFga(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesOpenFga Suite")
}

var _ = ginkgo.Describe("KubernetesOpenFga Custom Validation Tests", func() {
	var input *KubernetesOpenFga

	ginkgo.BeforeEach(func() {
		input = &KubernetesOpenFga{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesOpenFga",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-openfga",
			},
			Spec: &KubernetesOpenFgaSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "test-namespace",
					},
				},
				Container: &KubernetesOpenFgaContainer{
					Replicas: 1,
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
				Ingress: &KubernetesOpenFgaIngress{
					Enabled:  true,
					Hostname: "test-openfga.example.com",
				},
				Datastore: &KubernetesOpenFgaDataStore{
					Engine:   "postgres",
					Host:     "localhost",
					Database: "testdb",
					Username: "user",
					Password: &kubernetes.KubernetesSensitiveValue{
						SensitiveValue: &kubernetes.KubernetesSensitiveValue_Value{
							Value: "pass",
						},
					},
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("with plain string password", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with secret reference password", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Datastore.Password = &kubernetes.KubernetesSensitiveValue{
					SensitiveValue: &kubernetes.KubernetesSensitiveValue_SecretRef{
						SecretRef: &kubernetes.KubernetesSecretKeyRef{
							Name: "openfga-db-credentials",
							Key:  "password",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with custom port and SSL enabled", func() {
			ginkgo.It("should not return a validation error", func() {
				port := int32(5433)
				input.Spec.Datastore.Port = &port
				input.Spec.Datastore.IsSecure = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with mysql engine", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Datastore.Engine = "mysql"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("with invalid engine", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Datastore.Engine = "invalid"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with missing host", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Datastore.Host = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with missing database", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Datastore.Database = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with missing username", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Datastore.Username = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with missing password", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Datastore.Password = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid port", func() {
			ginkgo.It("should return a validation error for port 0", func() {
				port := int32(0)
				input.Spec.Datastore.Port = &port
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for port > 65535", func() {
				port := int32(65536)
				input.Spec.Datastore.Port = &port
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
