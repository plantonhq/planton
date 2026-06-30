package kubernetesgatewayapicrdsv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
)

func TestKubernetesGatewayApiCrdsSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesGatewayApiCrdsSpec Validation Suite")
}

var _ = ginkgo.Describe("KubernetesGatewayApiCrdsSpec Validation Tests", func() {
	var spec *KubernetesGatewayApiCrdsSpec

	ginkgo.BeforeEach(func() {
		spec = &KubernetesGatewayApiCrdsSpec{
			TargetCluster: &kubernetes.KubernetesClusterSelector{
				ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
				ClusterName: "test-cluster",
			},
		}
	})

	ginkgo.Describe("Basic Validation", func() {
		ginkgo.Context("with minimal valid spec", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Version Validation", func() {
		ginkgo.Context("with valid version v1.2.1", func() {
			ginkgo.It("should not return a validation error", func() {
				version := "v1.2.1"
				spec.Version = &version
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with valid version v1.3.0", func() {
			ginkgo.It("should not return a validation error", func() {
				version := "v1.3.0"
				spec.Version = &version
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with valid version v1.0.0-alpha", func() {
			ginkgo.It("should not return a validation error", func() {
				version := "v1.0.0-alpha"
				spec.Version = &version
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid version missing v prefix", func() {
			ginkgo.It("should return a validation error", func() {
				version := "1.2.1"
				spec.Version = &version
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid version format", func() {
			ginkgo.It("should return a validation error", func() {
				version := "invalid"
				spec.Version = &version
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with empty version", func() {
			ginkgo.It("should return a validation error", func() {
				version := ""
				spec.Version = &version
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Install Channel Validation", func() {
		ginkgo.Context("with standard channel", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.InstallChannel = &KubernetesGatewayApiCrdsSpec_InstallChannel{
					Channel: KubernetesGatewayApiCrdsSpec_InstallChannel_standard,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with experimental channel", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.InstallChannel = &KubernetesGatewayApiCrdsSpec_InstallChannel{
					Channel: KubernetesGatewayApiCrdsSpec_InstallChannel_experimental,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with unspecified channel", func() {
			ginkgo.It("should not return a validation error (defaults to standard behavior)", func() {
				spec.InstallChannel = &KubernetesGatewayApiCrdsSpec_InstallChannel{
					Channel: KubernetesGatewayApiCrdsSpec_InstallChannel_gateway_api_install_channel_unspecified,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with no install_channel set", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.InstallChannel = nil
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Target Cluster Validation", func() {
		ginkgo.Context("with GKE cluster", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.TargetCluster = &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "my-gke-cluster",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with EKS cluster", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.TargetCluster = &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_AwsEksCluster,
					ClusterName: "my-eks-cluster",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with AKS cluster", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.TargetCluster = &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_AzureAksCluster,
					ClusterName: "my-aks-cluster",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Complete Configuration", func() {
		ginkgo.Context("with all fields specified", func() {
			ginkgo.It("should not return a validation error", func() {
				version := "v1.3.0"
				spec.Version = &version
				spec.InstallChannel = &KubernetesGatewayApiCrdsSpec_InstallChannel{
					Channel: KubernetesGatewayApiCrdsSpec_InstallChannel_experimental,
				}
				spec.TargetCluster = &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "production-cluster",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
