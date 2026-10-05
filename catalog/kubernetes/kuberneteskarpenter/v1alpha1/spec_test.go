package kuberneteskarpenterv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesKarpenter(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesKarpenter Suite")
}

func int32Ptr(i int32) *int32    { return &i }
func boolPtr(b bool) *bool       { return &b }
func stringPtr(s string) *string { return &s }

// awsArm wraps the AWS settings into the spec's cloud oneof.
func awsArm(aws *KubernetesKarpenterAws) *KubernetesKarpenterSpec_Aws {
	return &KubernetesKarpenterSpec_Aws{Aws: aws}
}

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

var _ = ginkgo.Describe("KubernetesKarpenter Validation Tests", func() {
	var input *KubernetesKarpenter

	ginkgo.BeforeEach(func() {
		input = &KubernetesKarpenter{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesKarpenter",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-karpenter",
			},
			Spec: &KubernetesKarpenterSpec{
				Namespace: literal("kube-system"),
				Cluster: &KubernetesKarpenterCluster{
					Name: "prod-eks",
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "kube-system", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("https cluster endpoint should be valid", func() {
			input.Spec.Cluster.Endpoint = "https://ABC123.gr7.us-east-1.eks.amazonaws.com"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("EKS control plane with empty endpoint should be valid", func() {
			input.Spec.Cluster.EksControlPlane = true
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("crds keep-on-uninstall disabled deliberately should be valid", func() {
			input.Spec.Crds = &KubernetesKarpenterCrds{
				Install:         boolPtr(false),
				KeepOnUninstall: boolPtr(false),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("crds install false with keep unset should be valid", func() {
			input.Spec.Crds = &KubernetesKarpenterCrds{Install: boolPtr(false)}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("full AWS arm should be valid", func() {
			input.Spec.Cloud = awsArm(&KubernetesKarpenterAws{
				IrsaRoleArn:             literal("arn:aws:iam::123456789012:role/karpenter-controller"),
				InterruptionQueue:       "karpenter-interruptions",
				IsolatedVpc:             true,
				ReservedEnis:            int32Ptr(1),
				EnableZonalShift:        true,
				VmMemoryOverheadPercent: stringPtr("0.075"),
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("GovCloud IRSA role ARN partition should be valid", func() {
			input.Spec.Cloud = awsArm(&KubernetesKarpenterAws{
				IrsaRoleArn: literal("arn:aws-us-gov:iam::123456789012:role/karpenter-controller"),
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("IRSA role ARN as an AwsIamRole reference should be valid", func() {
			// The one-run composition shape: the controller's role deploys in
			// the same run and its ARN flows in by reference.
			input.Spec.Cloud = awsArm(&KubernetesKarpenterAws{
				IrsaRoleArn: valueFrom(catalogkind.CatalogKind_AwsIamRole,
					"karpenter-controller", "status.outputs.role_arn"),
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("vm memory overhead boundary values should be valid", func() {
			input.Spec.Cloud = awsArm(&KubernetesKarpenterAws{VmMemoryOverheadPercent: stringPtr("0")})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			input.Spec.GetAws().VmMemoryOverheadPercent = stringPtr("1.0")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("sized controller with debug log level should be valid", func() {
			input.Spec.Controller = &KubernetesKarpenterController{
				Replicas: int32Ptr(3),
				LogLevel: stringPtr("debug"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("error log level should be valid", func() {
			input.Spec.Controller = &KubernetesKarpenterController{LogLevel: stringPtr("error")}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("compound Go durations for batching should be valid", func() {
			input.Spec.Batching = &KubernetesKarpenterBatching{
				MaxDuration:  stringPtr("1m30s"),
				IdleDuration: stringPtr("2s"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("non-default scheduling posture should be valid", func() {
			input.Spec.Scheduling = &KubernetesKarpenterSchedulingPosture{
				PreferencePolicy: stringPtr("Ignore"),
				MinValuesPolicy:  stringPtr("BestEffort"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("missing namespace should fail", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("missing cluster should fail", func() {
			input.Spec.Cluster = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("empty cluster name should fail", func() {
			input.Spec.Cluster.Name = ""
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("non-https cluster endpoint should fail", func() {
			input.Spec.Cluster.Endpoint = "http://insecure.example.com"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("crds install false with keep_on_uninstall true should fail", func() {
			input.Spec.Crds = &KubernetesKarpenterCrds{
				Install:         boolPtr(false),
				KeepOnUninstall: boolPtr(true),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("IRSA role ARN present but empty should fail (StringValueOrRef contract)", func() {
			// The shared StringValueOrRef rule: a present message must carry a
			// non-empty literal or a reference.
			input.Spec.Cloud = awsArm(&KubernetesKarpenterAws{
				IrsaRoleArn: &foreignkeyv1.StringValueOrRef{},
			})
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("negative reserved ENIs should fail (gte 0)", func() {
			input.Spec.Cloud = awsArm(&KubernetesKarpenterAws{ReservedEnis: int32Ptr(-1)})
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("vm memory overhead above 1 should fail", func() {
			input.Spec.Cloud = awsArm(&KubernetesKarpenterAws{VmMemoryOverheadPercent: stringPtr("7.5")})
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("non-numeric vm memory overhead should fail", func() {
			input.Spec.Cloud = awsArm(&KubernetesKarpenterAws{VmMemoryOverheadPercent: stringPtr("abc")})
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("zero controller replicas should fail (gte 1)", func() {
			input.Spec.Controller = &KubernetesKarpenterController{Replicas: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("unknown log level should fail (closed enum)", func() {
			input.Spec.Controller = &KubernetesKarpenterController{LogLevel: stringPtr("warn")}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("unitless batching max duration should fail", func() {
			input.Spec.Batching = &KubernetesKarpenterBatching{MaxDuration: stringPtr("10")}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("millisecond batching idle duration should fail (s/m/h only)", func() {
			input.Spec.Batching = &KubernetesKarpenterBatching{IdleDuration: stringPtr("500ms")}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("lowercase preference policy should fail (closed enum)", func() {
			input.Spec.Scheduling = &KubernetesKarpenterSchedulingPosture{PreferencePolicy: stringPtr("respect")}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("unknown min values policy should fail (closed enum)", func() {
			input.Spec.Scheduling = &KubernetesKarpenterSchedulingPosture{MinValuesPolicy: stringPtr("Relaxed")}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
