package kubernetesstrimzikafkaoperatorv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesStrimziKafkaOperator(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesStrimziKafkaOperator Suite")
}

func int32Ptr(i int32) *int32    { return &i }
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

var _ = ginkgo.Describe("KubernetesStrimziKafkaOperator Validation Tests", func() {
	var input *KubernetesStrimziKafkaOperator

	ginkgo.BeforeEach(func() {
		input = &KubernetesStrimziKafkaOperator{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesStrimziKafkaOperator",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-strimzi-operator",
			},
			Spec: &KubernetesStrimziKafkaOperatorSpec{
				Namespace: literal("strimzi-system"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error (every optional block omitted)", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "strimzi-system", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("any_namespace watch scope alone should be valid", func() {
			input.Spec.Watch = &KubernetesStrimziKafkaOperatorWatch{AnyNamespace: true}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a namespace-fence watch scope alone should be valid", func() {
			input.Spec.Watch = &KubernetesStrimziKafkaOperatorWatch{Namespaces: []string{"team-a", "team-b"}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("every log level in the operator's vocabulary should be valid", func() {
			for _, level := range []string{"ERROR", "WARN", "INFO", "DEBUG", "TRACE"} {
				input.Spec.LogLevel = stringPtr(level)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("full surface (replicas, intervals, gates, image, scheduling) should be valid", func() {
			input.Spec.Replicas = int32Ptr(2)
			input.Spec.FullReconciliationIntervalMs = int32Ptr(180000)
			input.Spec.OperationTimeoutMs = int32Ptr(420000)
			input.Spec.FeatureGates = "+SomeGate"
			input.Spec.KubernetesServiceDnsDomain = stringPtr("cluster.local")
			input.Spec.Image = &KubernetesStrimziKafkaOperatorImage{
				Registry:   "mirror.example.com",
				Repository: "strimzi",
				Tag:        "1.1.0",
			}
			input.Spec.NodeSelector = map[string]string{"kubernetes.io/os": "linux"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("missing namespace should fail (required)", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("any_namespace combined with a namespace fence should fail (spec.watch.any_namespace_xor_namespaces)", func() {
			input.Spec.Watch = &KubernetesStrimziKafkaOperatorWatch{
				AnyNamespace: true,
				Namespaces:   []string{"team-a"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown log level should fail (spec.log_level_enum)", func() {
			input.Spec.LogLevel = stringPtr("VERBOSE")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero replicas should fail (replicas gte 1)", func() {
			input.Spec.Replicas = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a sub-second reconciliation interval should fail (gte 1000ms)", func() {
			input.Spec.FullReconciliationIntervalMs = int32Ptr(500)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a sub-second operation timeout should fail (gte 1000ms)", func() {
			input.Spec.OperationTimeoutMs = int32Ptr(999)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
