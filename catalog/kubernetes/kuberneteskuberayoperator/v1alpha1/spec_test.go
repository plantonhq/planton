package kuberneteskuberayoperatorv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesKubeRayOperator(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesKubeRayOperator Suite")
}

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

var _ = ginkgo.Describe("KubernetesKubeRayOperator Validation Tests", func() {
	var input *KubernetesKubeRayOperator

	ginkgo.BeforeEach(func() {
		input = &KubernetesKubeRayOperator{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesKubeRayOperator",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "kuberay-operator",
			},
			Spec: &KubernetesKubeRayOperatorSpec{
				Namespace: literal("ray-system"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error (every optional block omitted)", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "ray-system", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("each supported batch scheduler should be valid", func() {
			for _, s := range []string{"volcano", "yunikorn", "scheduler-plugins", ""} {
				input.Spec.BatchScheduler = s
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("feature gates should be valid", func() {
			input.Spec.FeatureGates = []*KubernetesKubeRayOperatorFeatureGate{
				{Name: "RayJobDeletionPolicy", Enabled: true},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("full surface should be valid", func() {
			input.Spec.CreateNamespace = true
			input.Spec.ChartVersion = stringPtr("1.6.2")
			input.Spec.WatchNamespaces = []string{"ml-team-a", "ml-team-b"}
			input.Spec.LeaderElectionEnabled = boolPtr(true)
			input.Spec.BatchScheduler = "yunikorn"
			input.Spec.MetricsEnabled = boolPtr(true)
			input.Spec.ServiceMonitorEnabled = true
			input.Spec.ImageRegistry = "mirror.example.com"
			input.Spec.HelmValues = "logging:\n  stdoutEncoder: console\n"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("a namespace-less spec should fail", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unsupported batch scheduler should fail", func() {
			input.Spec.BatchScheduler = "volcano-legacy"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a feature gate without a name should fail", func() {
			input.Spec.FeatureGates = []*KubernetesKubeRayOperatorFeatureGate{{Enabled: true}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
