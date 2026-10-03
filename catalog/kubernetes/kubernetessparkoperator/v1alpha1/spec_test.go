package kubernetessparkoperatorv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesSparkOperator(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesSparkOperator Suite")
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

var _ = ginkgo.Describe("KubernetesSparkOperator Validation Tests", func() {
	var input *KubernetesSparkOperator

	ginkgo.BeforeEach(func() {
		input = &KubernetesSparkOperator{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesSparkOperator",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "spark-operator",
			},
			Spec: &KubernetesSparkOperatorSpec{
				Namespace: literal("spark-system"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error (every optional block omitted)", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "spark-system", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("the cluster-wide workload posture (empty workload block) should be valid", func() {
			input.Spec.Workload = &KubernetesSparkOperatorWorkload{}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("the fenced workload posture with explicit namespaces should be valid", func() {
			input.Spec.Workload = &KubernetesSparkOperatorWorkload{
				Namespaces:     []string{"data-pipelines", "ml-jobs"},
				ServiceAccount: stringPtr("spark"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("multi-replica (leader-elected) operators should be valid", func() {
			input.Spec.Replicas = int32Ptr(2)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("dynamic config with hot properties should be valid", func() {
			input.Spec.DynamicConfig = &KubernetesSparkOperatorDynamicConfig{
				Enabled: true,
				Properties: map[string]string{
					"spark.kubernetes.operator.reconciler.intervalSeconds": "60",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("full surface should be valid", func() {
			input.Spec.CreateNamespace = true
			input.Spec.ChartVersion = stringPtr("1.8.0")
			input.Spec.Replicas = int32Ptr(2)
			input.Spec.Workload = &KubernetesSparkOperatorWorkload{
				Namespaces: []string{"data-pipelines"},
			}
			input.Spec.OperatorProperties = map[string]string{
				"spark.kubernetes.operator.reconciler.intervalSeconds": "30",
			}
			input.Spec.JvmArgs = "-XX:+UseParallelGC"
			input.Spec.ImageRegistry = "mirror.example.com"
			input.Spec.ImagePullSecrets = []string{"mirror-pull"}
			input.Spec.Scheduling = &KubernetesSparkOperatorScheduling{
				NodeSelector: map[string]string{"workload": "system"},
			}
			input.Spec.HelmValues = "operatorDeployment:\n  operatorPod:\n    priorityClassName: system-cluster-critical\n"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("a namespace-less spec should fail", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero replicas should fail", func() {
			input.Spec.Replicas = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
