package kuberneteskafkauserv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesKafkaUser(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesKafkaUser Suite")
}

func int32Ptr(i int32) *int32       { return &i }
func float64Ptr(f float64) *float64 { return &f }
func stringPtr(s string) *string    { return &s }

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

// topicAcl returns a minimal valid ACL rule — the base every ACL test
// mutates from.
func topicAcl() *KubernetesKafkaUserAcl {
	return &KubernetesKafkaUserAcl{
		Resource: &KubernetesKafkaUserAclResource{
			Type: "topic",
			Name: "orders",
		},
		Operations: []string{"Read", "Describe"},
	}
}

var _ = ginkgo.Describe("KubernetesKafkaUser Validation Tests", func() {
	var input *KubernetesKafkaUser

	ginkgo.BeforeEach(func() {
		input = &KubernetesKafkaUser{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesKafkaUser",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "orders-service",
			},
			Spec: &KubernetesKafkaUserSpec{
				Namespace:    literal("kafka"),
				KafkaCluster: literal("my-kafka"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error (ACL-only principal shape)", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("kafka_cluster as a reference should be valid", func() {
			input.Spec.KafkaCluster = valueFrom(catalogkind.CatalogKind_KubernetesKafka, "my-kafka", "status.outputs.cluster_name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("every authentication type should be valid", func() {
			for _, authType := range []string{"scram-sha-512", "tls", "tls-external"} {
				input.Spec.Authentication = &KubernetesKafkaUserAuthentication{Type: authType}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("authorization with literal, prefix, group and cluster ACLs should be valid", func() {
			input.Spec.Authorization = &KubernetesKafkaUserAuthorization{
				Acls: []*KubernetesKafkaUserAcl{
					topicAcl(),
					{
						Resource: &KubernetesKafkaUserAclResource{
							Type:        "topic",
							Name:        "analytics.",
							PatternType: stringPtr("prefix"),
						},
						Operations: []string{"Write", "Create", "IdempotentWrite"},
					},
					{
						Resource:   &KubernetesKafkaUserAclResource{Type: "group", Name: "orders-service"},
						Operations: []string{"Read"},
					},
					{
						Resource:   &KubernetesKafkaUserAclResource{Type: "cluster"},
						Operations: []string{"DescribeConfigs"},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("quotas with every field should be valid", func() {
			input.Spec.Quotas = &KubernetesKafkaUserQuotas{
				ProducerByteRate:       int32Ptr(1048576),
				ConsumerByteRate:       int32Ptr(2097152),
				RequestPercentage:      int32Ptr(200),
				ControllerMutationRate: float64Ptr(10),
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

		ginkgo.It("an unknown authentication type should fail (spec.authentication.type_enum — oauth lost its first-class type in Strimzi 1.x)", func() {
			input.Spec.Authentication = &KubernetesKafkaUserAuthentication{Type: "oauth"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an authorization block without ACLs should fail (min_items)", func() {
			input.Spec.Authorization = &KubernetesKafkaUserAuthorization{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown authorization type should fail (simple is Strimzi's only type)", func() {
			input.Spec.Authorization = &KubernetesKafkaUserAuthorization{
				Type: stringPtr("opa"),
				Acls: []*KubernetesKafkaUserAcl{topicAcl()},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an ACL without operations should fail (min_items)", func() {
			acl := topicAcl()
			acl.Operations = nil
			input.Spec.Authorization = &KubernetesKafkaUserAuthorization{Acls: []*KubernetesKafkaUserAcl{acl}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown operation should fail (spec.acl.operations_enum)", func() {
			acl := topicAcl()
			acl.Operations = []string{"Publish"}
			input.Spec.Authorization = &KubernetesKafkaUserAuthorization{Acls: []*KubernetesKafkaUserAcl{acl}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown resource type should fail (spec.acl.resource.type_enum)", func() {
			acl := topicAcl()
			acl.Resource.Type = "queue"
			input.Spec.Authorization = &KubernetesKafkaUserAuthorization{Acls: []*KubernetesKafkaUserAcl{acl}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown pattern type should fail (spec.acl.resource.pattern_enum)", func() {
			acl := topicAcl()
			acl.Resource.PatternType = stringPtr("regex")
			input.Spec.Authorization = &KubernetesKafkaUserAuthorization{Acls: []*KubernetesKafkaUserAcl{acl}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a nameless topic ACL should fail (name_required_unless_cluster)", func() {
			acl := topicAcl()
			acl.Resource.Name = ""
			input.Spec.Authorization = &KubernetesKafkaUserAuthorization{Acls: []*KubernetesKafkaUserAcl{acl}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a negative producer byte rate should fail (gte 0)", func() {
			input.Spec.Quotas = &KubernetesKafkaUserQuotas{ProducerByteRate: int32Ptr(-1)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a negative controller mutation rate should fail (gte 0)", func() {
			input.Spec.Quotas = &KubernetesKafkaUserQuotas{ControllerMutationRate: float64Ptr(-0.5)}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
