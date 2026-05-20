package kubernetescertmanagerv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestKubernetesCertManager(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesCertManager Suite")
}

var _ = ginkgo.Describe("KubernetesCertManager Validation Tests", func() {
	var input *KubernetesCertManager

	ginkgo.BeforeEach(func() {
		input = &KubernetesCertManager{
			ApiVersion: "kubernetes.openmcf.org/v1",
			Kind:       "KubernetesCertManager",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-cert-manager",
			},
			Spec: &KubernetesCertManagerSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "cert-manager",
					},
				},
				CreateNamespace: true,
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("basic install without workload identity", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with GKE workload identity", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.WorkloadIdentity = &WorkloadIdentityConfig{
					Provider: &WorkloadIdentityConfig_Gke{
						Gke: &GkeWorkloadIdentity{
							ServiceAccountEmail: "cert-manager@my-project.iam.gserviceaccount.com",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with EKS IRSA", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.WorkloadIdentity = &WorkloadIdentityConfig{
					Provider: &WorkloadIdentityConfig_Eks{
						Eks: &EksIrsa{
							RoleArn: "arn:aws:iam::123456789012:role/cert-manager",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with AKS workload identity", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.WorkloadIdentity = &WorkloadIdentityConfig{
					Provider: &WorkloadIdentityConfig_Aks{
						Aks: &AksWorkloadIdentity{
							ClientId: "12345678-1234-1234-1234-123456789012",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with custom versions", func() {
			ginkgo.It("should not return a validation error", func() {
				version := "v1.16.4"
				chartVersion := "v1.16.4"
				input.Spec.KubernetesCertManagerVersion = &version
				input.Spec.HelmChartVersion = &chartVersion
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with skip_install_self_signed_issuer", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.SkipInstallSelfSignedIssuer = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("missing namespace", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Namespace = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("empty namespace value", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Namespace = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("missing metadata", func() {
			ginkgo.It("should return a validation error", func() {
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("GKE workload identity missing service_account_email", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.WorkloadIdentity = &WorkloadIdentityConfig{
					Provider: &WorkloadIdentityConfig_Gke{
						Gke: &GkeWorkloadIdentity{
							ServiceAccountEmail: "",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("EKS IRSA missing role_arn", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.WorkloadIdentity = &WorkloadIdentityConfig{
					Provider: &WorkloadIdentityConfig_Eks{
						Eks: &EksIrsa{
							RoleArn: "",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("AKS workload identity missing client_id", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.WorkloadIdentity = &WorkloadIdentityConfig{
					Provider: &WorkloadIdentityConfig_Aks{
						Aks: &AksWorkloadIdentity{
							ClientId: "",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
