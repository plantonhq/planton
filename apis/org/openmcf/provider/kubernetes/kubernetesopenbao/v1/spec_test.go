package kubernetesopenbaov1

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

func TestKubernetesOpenBao(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesOpenBao Suite")
}

var _ = ginkgo.Describe("KubernetesOpenBao Custom Validation Tests", func() {
	var input *KubernetesOpenBao

	ginkgo.BeforeEach(func() {
		input = &KubernetesOpenBao{
			ApiVersion: "kubernetes.openmcf.org/v1",
			Kind:       "KubernetesOpenBao",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-openbao",
			},
			Spec: &KubernetesOpenBaoSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "test-namespace",
					},
				},
				CreateNamespace: true,
				ServerContainer: &KubernetesOpenBaoServerContainer{
					Replicas:        1,
					DataStorageSize: "10Gi",
					Resources: &kubernetes.ContainerResources{
						Limits: &kubernetes.CpuMemory{
							Cpu:    "500m",
							Memory: "256Mi",
						},
						Requests: &kubernetes.CpuMemory{
							Cpu:    "100m",
							Memory: "128Mi",
						},
					},
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("standalone openbao deployment", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("standalone openbao with ingress enabled", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Ingress = &KubernetesOpenBaoIngress{
					Enabled:  true,
					Hostname: "openbao.example.com",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("ha mode openbao deployment", func() {
			ginkgo.It("should not return a validation error", func() {
				haReplicas := int32(3)
				input.Spec.HighAvailability = &KubernetesOpenBaoHighAvailability{
					Enabled:  true,
					Replicas: &haReplicas,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("openbao with injector enabled", func() {
			ginkgo.It("should not return a validation error", func() {
				injectorReplicas := int32(2)
				input.Spec.Injector = &KubernetesOpenBaoInjector{
					Enabled:  true,
					Replicas: &injectorReplicas,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("standalone openbao with GCP KMS auto-unseal", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
					Seal: &KubernetesOpenBaoAutoUnseal_GcpKms{
						GcpKms: &KubernetesOpenBaoGcpKmsSeal{
							Project: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-project"},
							},
							Region: "us-central1",
							KeyRing: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-keyring"},
							},
							CryptoKey: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-key"},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("standalone openbao with AWS KMS auto-unseal", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
					Seal: &KubernetesOpenBaoAutoUnseal_AwsKms{
						AwsKms: &KubernetesOpenBaoAwsKmsSeal{
							Region:   "us-east-1",
							KmsKeyId: "arn:aws:kms:us-east-1:111122223333:key/example-key-id",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("standalone openbao with Azure Key Vault auto-unseal", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
					Seal: &KubernetesOpenBaoAutoUnseal_AzureKeyVault{
						AzureKeyVault: &KubernetesOpenBaoAzureKeyVaultSeal{
							VaultName: "my-keyvault",
							KeyName:   "unseal-key",
							TenantId:  "00000000-0000-0000-0000-000000000000",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("standalone openbao with Transit auto-unseal", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
					Seal: &KubernetesOpenBaoAutoUnseal_Transit{
						Transit: &KubernetesOpenBaoTransitSeal{
							Address:         "https://vault.example.com:8200",
							KeyName:         "autounseal",
							MountPath:       "transit/",
							TokenSecretName: "vault-transit-token",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("ha openbao with GCP KMS auto-unseal", func() {
			ginkgo.It("should not return a validation error", func() {
				haReplicas := int32(3)
				input.Spec.HighAvailability = &KubernetesOpenBaoHighAvailability{
					Enabled:  true,
					Replicas: &haReplicas,
				}
				input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
					Seal: &KubernetesOpenBaoAutoUnseal_GcpKms{
						GcpKms: &KubernetesOpenBaoGcpKmsSeal{
							Project: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-project"},
							},
							Region: "us-central1",
							KeyRing: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-keyring"},
							},
							CryptoKey: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-key"},
							},
							WorkloadIdentityServiceAccount: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "openbao@my-project.iam.gserviceaccount.com"},
							},
						},
					},
				}
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

		ginkgo.Context("invalid data storage size format", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.ServerContainer.DataStorageSize = "invalid"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("server replicas below minimum", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.ServerContainer.Replicas = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("server replicas above maximum", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.ServerContainer.Replicas = 11
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("ha replicas below minimum", func() {
			ginkgo.It("should return a validation error", func() {
				haReplicas := int32(1)
				input.Spec.HighAvailability = &KubernetesOpenBaoHighAvailability{
					Enabled:  true,
					Replicas: &haReplicas,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("ingress enabled without hostname", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Ingress = &KubernetesOpenBaoIngress{
					Enabled:  true,
					Hostname: "",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("injector replicas above maximum", func() {
			ginkgo.It("should return a validation error", func() {
				injectorReplicas := int32(6)
				input.Spec.Injector = &KubernetesOpenBaoInjector{
					Enabled:  true,
					Replicas: &injectorReplicas,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("gcp kms auto-unseal missing project", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
					Seal: &KubernetesOpenBaoAutoUnseal_GcpKms{
						GcpKms: &KubernetesOpenBaoGcpKmsSeal{
							Region: "us-central1",
							KeyRing: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-keyring"},
							},
							CryptoKey: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-key"},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("gcp kms auto-unseal missing key_ring", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
					Seal: &KubernetesOpenBaoAutoUnseal_GcpKms{
						GcpKms: &KubernetesOpenBaoGcpKmsSeal{
							Project: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-project"},
							},
							Region: "us-central1",
							CryptoKey: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-key"},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("gcp kms auto-unseal missing crypto_key", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
					Seal: &KubernetesOpenBaoAutoUnseal_GcpKms{
						GcpKms: &KubernetesOpenBaoGcpKmsSeal{
							Project: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-project"},
							},
							Region: "us-central1",
							KeyRing: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-keyring"},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("gcp kms auto-unseal missing region", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
					Seal: &KubernetesOpenBaoAutoUnseal_GcpKms{
						GcpKms: &KubernetesOpenBaoGcpKmsSeal{
							Project: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-project"},
							},
							KeyRing: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-keyring"},
							},
							CryptoKey: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-key"},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("aws kms auto-unseal missing region", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
					Seal: &KubernetesOpenBaoAutoUnseal_AwsKms{
						AwsKms: &KubernetesOpenBaoAwsKmsSeal{
							KmsKeyId: "arn:aws:kms:us-east-1:111122223333:key/example-key-id",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("aws kms auto-unseal missing kms_key_id", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
					Seal: &KubernetesOpenBaoAutoUnseal_AwsKms{
						AwsKms: &KubernetesOpenBaoAwsKmsSeal{
							Region: "us-east-1",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("azure key vault auto-unseal missing vault_name", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
					Seal: &KubernetesOpenBaoAutoUnseal_AzureKeyVault{
						AzureKeyVault: &KubernetesOpenBaoAzureKeyVaultSeal{
							KeyName:  "unseal-key",
							TenantId: "00000000-0000-0000-0000-000000000000",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("azure key vault auto-unseal missing key_name", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
					Seal: &KubernetesOpenBaoAutoUnseal_AzureKeyVault{
						AzureKeyVault: &KubernetesOpenBaoAzureKeyVaultSeal{
							VaultName: "my-keyvault",
							TenantId:  "00000000-0000-0000-0000-000000000000",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("azure key vault auto-unseal missing tenant_id", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
					Seal: &KubernetesOpenBaoAutoUnseal_AzureKeyVault{
						AzureKeyVault: &KubernetesOpenBaoAzureKeyVaultSeal{
							VaultName: "my-keyvault",
							KeyName:   "unseal-key",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("transit auto-unseal missing address", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
					Seal: &KubernetesOpenBaoAutoUnseal_Transit{
						Transit: &KubernetesOpenBaoTransitSeal{
							KeyName:         "autounseal",
							TokenSecretName: "vault-transit-token",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("transit auto-unseal missing key_name", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
					Seal: &KubernetesOpenBaoAutoUnseal_Transit{
						Transit: &KubernetesOpenBaoTransitSeal{
							Address:         "https://vault.example.com:8200",
							TokenSecretName: "vault-transit-token",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("transit auto-unseal missing token_secret_name", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.AutoUnseal = &KubernetesOpenBaoAutoUnseal{
					Seal: &KubernetesOpenBaoAutoUnseal_Transit{
						Transit: &KubernetesOpenBaoTransitSeal{
							Address: "https://vault.example.com:8200",
							KeyName: "autounseal",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
