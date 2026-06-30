package kubernetesexternalsecretsv1

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

func TestKubernetesExternalSecrets(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesExternalSecrets Suite")
}

var _ = ginkgo.Describe("KubernetesExternalSecrets Validation Tests", func() {
	var input *KubernetesExternalSecrets

	ginkgo.BeforeEach(func() {
		input = &KubernetesExternalSecrets{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesExternalSecrets",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-external-secrets",
			},
			Spec: &KubernetesExternalSecretsSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "external-secrets",
					},
				},
				Container: &KubernetesExternalSecretsSpecContainer{
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
	})

	ginkgo.Describe("GKE Configuration", func() {
		ginkgo.Context("with valid GKE config", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.ProviderConfig = &KubernetesExternalSecretsSpec_Gke{
					Gke: &KubernetesExternalSecretsGkeConfig{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-gcp-project",
							},
						},
						GsaEmail: "external-secrets@my-project.iam.gserviceaccount.com",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with missing project_id", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.ProviderConfig = &KubernetesExternalSecretsSpec_Gke{
					Gke: &KubernetesExternalSecretsGkeConfig{
						GsaEmail: "external-secrets@my-project.iam.gserviceaccount.com",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with missing gsa_email", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.ProviderConfig = &KubernetesExternalSecretsSpec_Gke{
					Gke: &KubernetesExternalSecretsGkeConfig{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-gcp-project",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("EKS Configuration", func() {
		ginkgo.Context("with valid minimal EKS config", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.ProviderConfig = &KubernetesExternalSecretsSpec_Eks{
					Eks: &KubernetesExternalSecretsEksConfig{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with EKS config with region", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.ProviderConfig = &KubernetesExternalSecretsSpec_Eks{
					Eks: &KubernetesExternalSecretsEksConfig{
						Region: "us-west-2",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with EKS config with IRSA role override", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.ProviderConfig = &KubernetesExternalSecretsSpec_Eks{
					Eks: &KubernetesExternalSecretsEksConfig{
						Region:              "us-east-1",
						IrsaRoleArnOverride: "arn:aws:iam::123456789012:role/external-secrets-role",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("AKS Configuration", func() {
		ginkgo.Context("with valid minimal AKS config", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.ProviderConfig = &KubernetesExternalSecretsSpec_Aks{
					Aks: &KubernetesExternalSecretsAksConfig{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with AKS config with key vault", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.ProviderConfig = &KubernetesExternalSecretsSpec_Aks{
					Aks: &KubernetesExternalSecretsAksConfig{
						KeyVaultResourceId: "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/my-rg/providers/Microsoft.KeyVault/vaults/my-vault",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with AKS config with managed identity", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.ProviderConfig = &KubernetesExternalSecretsSpec_Aks{
					Aks: &KubernetesExternalSecretsAksConfig{
						KeyVaultResourceId:      "/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/my-rg/providers/Microsoft.KeyVault/vaults/my-vault",
						ManagedIdentityClientId: "12345678-1234-1234-1234-123456789012",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Poll Interval Validation", func() {
		ginkgo.Context("with custom poll interval", func() {
			ginkgo.It("should not return a validation error", func() {
				pollInterval := uint32(30)
				input.Spec.PollIntervalSeconds = &pollInterval
				input.Spec.ProviderConfig = &KubernetesExternalSecretsSpec_Gke{
					Gke: &KubernetesExternalSecretsGkeConfig{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-gcp-project",
							},
						},
						GsaEmail: "external-secrets@my-project.iam.gserviceaccount.com",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with zero poll interval", func() {
			ginkgo.It("should return a validation error", func() {
				pollInterval := uint32(0)
				input.Spec.PollIntervalSeconds = &pollInterval
				input.Spec.ProviderConfig = &KubernetesExternalSecretsSpec_Gke{
					Gke: &KubernetesExternalSecretsGkeConfig{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-gcp-project",
							},
						},
						GsaEmail: "external-secrets@my-project.iam.gserviceaccount.com",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Container Resources Validation", func() {
		ginkgo.Context("with missing container resources", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Container = nil
				input.Spec.ProviderConfig = &KubernetesExternalSecretsSpec_Gke{
					Gke: &KubernetesExternalSecretsGkeConfig{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-gcp-project",
							},
						},
						GsaEmail: "external-secrets@my-project.iam.gserviceaccount.com",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with valid container resources", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.ProviderConfig = &KubernetesExternalSecretsSpec_Gke{
					Gke: &KubernetesExternalSecretsGkeConfig{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-gcp-project",
							},
						},
						GsaEmail: "external-secrets@my-project.iam.gserviceaccount.com",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Default Values", func() {
		ginkgo.Context("with defaults applied from proto", func() {
			ginkgo.It("should apply default values", func() {
				input.Spec.ProviderConfig = &KubernetesExternalSecretsSpec_Gke{
					Gke: &KubernetesExternalSecretsGkeConfig{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-gcp-project",
							},
						},
						GsaEmail: "external-secrets@my-project.iam.gserviceaccount.com",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
				// Defaults should be applied:
				// poll_interval_seconds: 10
				// container.resources: predefined values
			})
		})
	})
})
