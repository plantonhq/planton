package kuberneteskafkatopicv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesKafkaTopic(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesKafkaTopic Suite")
}

func int32Ptr(i int32) *int32 { return &i }

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

var _ = ginkgo.Describe("KubernetesKafkaTopic Validation Tests", func() {
	var input *KubernetesKafkaTopic

	ginkgo.BeforeEach(func() {
		input = &KubernetesKafkaTopic{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesKafkaTopic",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "orders",
			},
			Spec: &KubernetesKafkaTopicSpec{
				Namespace:    literal("kafka"),
				KafkaCluster: literal("my-kafka"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("kafka_cluster as a reference should be valid", func() {
			input.Spec.KafkaCluster = valueFrom(catalogkind.CatalogKind_KubernetesKafka, "my-kafka", "status.outputs.cluster_name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a topic-name override with dots, underscores and uppercase should be valid", func() {
			input.Spec.TopicName = "orders.v1_DLQ"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("full surface (partitions, replicas, config) should be valid", func() {
			input.Spec.Partitions = int32Ptr(12)
			input.Spec.Replicas = int32Ptr(3)
			input.Spec.Config = map[string]string{
				"retention.ms":   "604800000",
				"cleanup.policy": "compact",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("missing namespace should fail (required)", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("missing kafka_cluster should fail (required)", func() {
			input.Spec.KafkaCluster = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a topic name with a slash should fail (spec.topic_name.format)", func() {
			input.Spec.TopicName = "orders/v1"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a topic name over 249 characters should fail (spec.topic_name.format)", func() {
			input.Spec.TopicName = strings.Repeat("a", 250)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero partitions should fail (gte 1)", func() {
			input.Spec.Partitions = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero replicas should fail (gte 1)", func() {
			input.Spec.Replicas = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("replicas above 32767 should fail (Kafka's int16 bound)", func() {
			input.Spec.Replicas = int32Ptr(40000)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
