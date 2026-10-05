package kuberneteskedav1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesKeda(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesKeda Suite")
}

func int32Ptr(i int32) *int32    { return &i }
func boolPtr(b bool) *bool       { return &b }
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

var _ = ginkgo.Describe("KubernetesKeda Validation Tests", func() {
	var input *KubernetesKeda

	ginkgo.BeforeEach(func() {
		input = &KubernetesKeda{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesKeda",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-keda",
			},
			Spec: &KubernetesKedaSpec{
				Namespace: literal("keda"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "keda", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("HA posture with sized components should be valid", func() {
			input.Spec.Operator = &KubernetesKedaComponent{Replicas: int32Ptr(2)}
			input.Spec.MetricsServer = &KubernetesKedaComponent{Replicas: int32Ptr(2)}
			input.Spec.Webhooks = &KubernetesKedaWebhooks{
				Replicas:      int32Ptr(2),
				FailurePolicy: stringPtr("Fail"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("IRSA pod identity should be valid", func() {
			input.Spec.PodIdentity = &KubernetesKedaPodIdentity{
				AwsIrsa: &KubernetesKedaAwsIrsa{
					Enabled: true,
					RoleArn: "arn:aws:iam::123456789012:role/keda-scalers",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("Azure workload identity should be valid", func() {
			input.Spec.PodIdentity = &KubernetesKedaPodIdentity{
				AzureWorkloadIdentity: &KubernetesKedaAzureWorkloadIdentity{
					Enabled:  true,
					ClientId: "11111111-2222-3333-4444-555555555555",
					TenantId: "66666666-7777-8888-9999-000000000000",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("GCP workload identity should be valid", func() {
			input.Spec.PodIdentity = &KubernetesKedaPodIdentity{
				GcpWorkloadIdentity: &KubernetesKedaGcpWorkloadIdentity{
					Enabled:             true,
					ServiceAccountEmail: "keda-scalers@my-project.iam.gserviceaccount.com",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("cert-manager certificates with an issuer reference should be valid", func() {
			input.Spec.Certificates = &KubernetesKedaCertificates{
				Type: stringPtr("cert_manager"),
				CertManagerIssuer: &kubernetes.CertManagerIssuerRef{
					IssuerType: &kubernetes.CertManagerIssuerRef_ClusterIssuer{
						ClusterIssuer: &kubernetes.CertManagerClusterIssuerRef{
							Name: valueFrom(catalogkind.CatalogKind_KubernetesClusterIssuer, "platform-ca", "status.outputs.cluster_issuer_name"),
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace-fenced watch with tuned timeout should be valid", func() {
			input.Spec.WatchNamespace = "team-a"
			input.Spec.HttpTimeoutMs = int32Ptr(5000)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("crds keep-on-uninstall disabled deliberately should be valid", func() {
			input.Spec.Crds = &KubernetesKedaCrds{KeepOnUninstall: boolPtr(false)}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("missing namespace should fail", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("crds install false with keep_on_uninstall true should fail", func() {
			input.Spec.Crds = &KubernetesKedaCrds{
				Install:         boolPtr(false),
				KeepOnUninstall: boolPtr(true),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("zero operator replicas should fail (gte 1)", func() {
			input.Spec.Operator = &KubernetesKedaComponent{Replicas: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("unknown webhook failure policy should fail (closed enum)", func() {
			input.Spec.Webhooks = &KubernetesKedaWebhooks{FailurePolicy: stringPtr("Warn")}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("IRSA enabled without role_arn should fail", func() {
			input.Spec.PodIdentity = &KubernetesKedaPodIdentity{
				AwsIrsa: &KubernetesKedaAwsIrsa{Enabled: true},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("malformed IRSA role ARN should fail", func() {
			input.Spec.PodIdentity = &KubernetesKedaPodIdentity{
				AwsIrsa: &KubernetesKedaAwsIrsa{
					Enabled: true,
					RoleArn: "role/keda-scalers",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("Azure workload identity without tenant should fail", func() {
			input.Spec.PodIdentity = &KubernetesKedaPodIdentity{
				AzureWorkloadIdentity: &KubernetesKedaAzureWorkloadIdentity{
					Enabled:  true,
					ClientId: "11111111-2222-3333-4444-555555555555",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("GCP workload identity enabled without email should fail", func() {
			input.Spec.PodIdentity = &KubernetesKedaPodIdentity{
				GcpWorkloadIdentity: &KubernetesKedaGcpWorkloadIdentity{Enabled: true},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("malformed GCP service account email should fail", func() {
			input.Spec.PodIdentity = &KubernetesKedaPodIdentity{
				GcpWorkloadIdentity: &KubernetesKedaGcpWorkloadIdentity{
					Enabled:             true,
					ServiceAccountEmail: "keda@example.com",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("unknown certificates type should fail (closed enum)", func() {
			input.Spec.Certificates = &KubernetesKedaCertificates{Type: stringPtr("vault")}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("issuer reference with operator certificates type should fail", func() {
			input.Spec.Certificates = &KubernetesKedaCertificates{
				Type: stringPtr("operator"),
				CertManagerIssuer: &kubernetes.CertManagerIssuerRef{
					IssuerType: &kubernetes.CertManagerIssuerRef_Issuer{
						Issuer: &kubernetes.CertManagerNamespacedIssuerRef{Name: literal("platform-ca")},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("cert-manager issuer that names no grain should fail", func() {
			input.Spec.Certificates = &KubernetesKedaCertificates{
				Type:              stringPtr("cert_manager"),
				CertManagerIssuer: &kubernetes.CertManagerIssuerRef{},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("An issuer is required"))
		})

		ginkgo.It("cert-manager issuer arm without a name should fail", func() {
			input.Spec.Certificates = &KubernetesKedaCertificates{
				Type: stringPtr("cert_manager"),
				CertManagerIssuer: &kubernetes.CertManagerIssuerRef{
					IssuerType: &kubernetes.CertManagerIssuerRef_Issuer{Issuer: &kubernetes.CertManagerNamespacedIssuerRef{}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("zero http timeout should fail (gte 1)", func() {
			input.Spec.HttpTimeoutMs = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("prometheus service monitor without enabled should fail", func() {
			input.Spec.Prometheus = &KubernetesKedaPrometheus{ServiceMonitor: true}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
