package kubernetesingressnginxv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestKubernetesIngressNginxSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesIngressNginxSpec Validation Suite")
}

var _ = ginkgo.Describe("KubernetesIngressNginxSpec validations", func() {
	var spec *KubernetesIngressNginxSpec

	ginkgo.BeforeEach(func() {
		spec = &KubernetesIngressNginxSpec{
			TargetCluster: &kubernetes.KubernetesClusterSelector{
				ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
				ClusterName: "test-cluster",
			},
			Namespace: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "ingress-nginx",
				},
			},
			ChartVersion: "4.11.1",
			Internal:     false,
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("spec with default external LoadBalancer", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with internal LoadBalancer", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.Internal = true
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with GKE configuration", func() {
			ginkgo.It("should not return a validation error with static IP", func() {
				spec.ProviderConfig = &KubernetesIngressNginxSpec_Gke{
					Gke: &KubernetesIngressNginxGkeConfig{
						StaticIpName: "my-static-ip",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with subnetwork for internal LB", func() {
				spec.Internal = true
				spec.ProviderConfig = &KubernetesIngressNginxSpec_Gke{
					Gke: &KubernetesIngressNginxGkeConfig{
						SubnetworkSelfLink: "projects/my-project/regions/us-west1/subnetworks/my-subnet",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with EKS configuration", func() {
			ginkgo.It("should not return a validation error with security groups", func() {
				spec.ProviderConfig = &KubernetesIngressNginxSpec_Eks{
					Eks: &KubernetesIngressNginxEksConfig{},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with IRSA role override", func() {
				spec.ProviderConfig = &KubernetesIngressNginxSpec_Eks{
					Eks: &KubernetesIngressNginxEksConfig{
						IrsaRoleArnOverride: "arn:aws:iam::123456789012:role/ingress-nginx",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with AKS configuration", func() {
			ginkgo.It("should not return a validation error with managed identity", func() {
				spec.ProviderConfig = &KubernetesIngressNginxSpec_Aks{
					Aks: &KubernetesIngressNginxAksConfig{
						ManagedIdentityClientId: "12345678-1234-1234-1234-123456789012",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with public IP name", func() {
				spec.ProviderConfig = &KubernetesIngressNginxSpec_Aks{
					Aks: &KubernetesIngressNginxAksConfig{
						PublicIpName: "my-public-ip",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("spec without provider config", func() {
			ginkgo.It("should not return a validation error for generic cluster", func() {
				// No provider config set - should work for generic clusters
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with empty chart version", func() {
			ginkgo.It("should not return a validation error (uses default)", func() {
				spec.ChartVersion = ""
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	// Note: The spec.proto does not currently have required validations on target_cluster
	// If validations are added in the future, add corresponding tests here
})
