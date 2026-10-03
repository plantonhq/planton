package kuberneteskyvernov1alpha1

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

func TestKubernetesKyverno(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesKyverno Suite")
}

func int32Ptr(i int32) *int32 { return &i }
func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }

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

var _ = ginkgo.Describe("KubernetesKyverno Validation Tests", func() {
	var input *KubernetesKyverno

	ginkgo.BeforeEach(func() {
		input = &KubernetesKyverno{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesKyverno",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "kyverno",
			},
			Spec: &KubernetesKyvernoSpec{
				Namespace: literal("kyverno"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("a minimal spec should be valid (chart defaults)", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "kyverno", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a crds block with keep-on-uninstall should be valid", func() {
			input.Spec.Crds = &KubernetesKyvernoCrds{
				Install:          boolPtr(true),
				KeepOnUninstall:  true,
				MigrationEnabled: boolPtr(false),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a config block with filter edits should be valid", func() {
			input.Spec.Config = &KubernetesKyvernoConfig{
				WebhookExcludeNamespaces:      []string{"platform-system"},
				ResourceFiltersInclude:        []string{"[Secret,ci-cache,*]"},
				ResourceFiltersExclude:        []string{"[Node,*,*]"},
				ExcludeGroups:                 []string{"system:nodes"},
				ExcludeUsernames:              []string{"!system:kube-scheduler"},
				DefaultRegistry:               "mirror.example.com",
				EnableDefaultRegistryMutation: boolPtr(false),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a features block should be valid", func() {
			input.Spec.Features = &KubernetesKyvernoFeatures{
				ForceFailurePolicyIgnore: true,
				BackgroundScan: &KubernetesKyvernoBackgroundScan{
					Enabled:  boolPtr(true),
					Workers:  int32Ptr(4),
					Interval: "30m",
				},
				GenerateValidatingAdmissionPolicy: boolPtr(false),
				LoggingFormat:                     strPtr("json"),
				LoggingVerbosity:                  int32Ptr(4),
				OmitEventTypes:                    []string{"PolicyApplied", "PolicySkipped"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("an HA admission controller should be valid", func() {
			input.Spec.AdmissionController = &KubernetesKyvernoAdmissionController{
				Replicas: int32Ptr(3),
				Resources: &kubernetes.ContainerResources{
					Requests: &kubernetes.CpuMemory{Cpu: "100m", Memory: "256Mi"},
					Limits:   &kubernetes.CpuMemory{Cpu: "1", Memory: "512Mi"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("admission controller autoscaling should be valid", func() {
			input.Spec.AdmissionController = &KubernetesKyvernoAdmissionController{
				Autoscaling: &KubernetesKyvernoHpa{
					MinReplicas:                    int32Ptr(2),
					MaxReplicas:                    10,
					TargetCpuUtilizationPercentage: int32Ptr(75),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("disabling optional controllers should be valid", func() {
			input.Spec.BackgroundController = &KubernetesKyvernoOptionalController{Enabled: boolPtr(false)}
			input.Spec.ReportsController = &KubernetesKyvernoOptionalController{Enabled: boolPtr(false)}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("cert-manager certificates with an issuer reference should be valid", func() {
			input.Spec.Certificates = &KubernetesKyvernoCertificates{
				CertManager: &KubernetesKyvernoCertManagerCertificates{
					IssuerName: valueFrom(catalogkind.CatalogKind_KubernetesClusterIssuer, "platform-ca", "metadata.name"),
					IssuerKind: strPtr("ClusterIssuer"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("cert-manager certificates with the chart's own self-signed issuer should be valid", func() {
			input.Spec.Certificates = &KubernetesKyvernoCertificates{
				CertManager: &KubernetesKyvernoCertManagerCertificates{},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("metrics with a service monitor should be valid", func() {
			input.Spec.Metrics = &KubernetesKyvernoMetrics{ServiceMonitor: true}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("a missing namespace should be invalid", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an uppercase webhook-exclude namespace should be invalid", func() {
			input.Spec.Config = &KubernetesKyvernoConfig{
				WebhookExcludeNamespaces: []string{"Platform-System"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown logging format should be invalid", func() {
			input.Spec.Features = &KubernetesKyvernoFeatures{LoggingFormat: strPtr("yaml")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("logging verbosity above 10 should be invalid", func() {
			input.Spec.Features = &KubernetesKyvernoFeatures{LoggingVerbosity: int32Ptr(11)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown omit-event type should be invalid", func() {
			input.Spec.Features = &KubernetesKyvernoFeatures{OmitEventTypes: []string{"PolicyIgnored"}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a background-scan interval with a calendar unit should be invalid", func() {
			input.Spec.Features = &KubernetesKyvernoFeatures{
				BackgroundScan: &KubernetesKyvernoBackgroundScan{Interval: "1d"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero background-scan workers should be invalid", func() {
			input.Spec.Features = &KubernetesKyvernoFeatures{
				BackgroundScan: &KubernetesKyvernoBackgroundScan{Workers: int32Ptr(0)},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero admission-controller replicas should be invalid (the chart rejects 0)", func() {
			input.Spec.AdmissionController = &KubernetesKyvernoAdmissionController{Replicas: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("autoscaling without max replicas should be invalid", func() {
			input.Spec.AdmissionController = &KubernetesKyvernoAdmissionController{
				Autoscaling: &KubernetesKyvernoHpa{MinReplicas: int32Ptr(1)},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("autoscaling target CPU above 100 should be invalid", func() {
			input.Spec.AdmissionController = &KubernetesKyvernoAdmissionController{
				Autoscaling: &KubernetesKyvernoHpa{
					MaxReplicas:                    5,
					TargetCpuUtilizationPercentage: int32Ptr(101),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero replicas on an optional controller should be invalid", func() {
			input.Spec.CleanupController = &KubernetesKyvernoOptionalController{Replicas: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown cert-manager issuer kind should be invalid", func() {
			input.Spec.Certificates = &KubernetesKyvernoCertificates{
				CertManager: &KubernetesKyvernoCertManagerCertificates{
					IssuerKind: strPtr("SelfSigned"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
