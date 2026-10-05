package kubernetesudproutev1alpha1

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

func TestKubernetesUdpRoute(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesUdpRoute Suite")
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

// serviceBackend returns a minimal valid backend reference to a Service.
func serviceBackend(name string, port int32) *kubernetes.KubernetesGatewayApiBackendRef {
	return &kubernetes.KubernetesGatewayApiBackendRef{
		Name: literal(name),
		Port: int32Ptr(port),
	}
}

var _ = ginkgo.Describe("KubernetesUdpRoute Validation Tests", func() {
	var input *KubernetesUdpRoute

	ginkgo.BeforeEach(func() {
		input = &KubernetesUdpRoute{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesUdpRoute",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-udp-route",
			},
			Spec: &KubernetesUdpRouteSpec{
				Namespace: literal("app-ns"),
				ParentRefs: []*kubernetes.KubernetesGatewayApiParentReference{
					{Name: literal("my-gateway")},
				},
				Rules: []*KubernetesUdpRouteRule{
					{
						BackendRefs: []*kubernetes.KubernetesGatewayApiBackendRef{
							serviceBackend("udp-svc", 53),
						},
					},
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal route should not return a validation error", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("full surface should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "app-ns", "spec.name")
			input.Spec.ParentRefs = []*kubernetes.KubernetesGatewayApiParentReference{
				{
					Group:       stringPtr("gateway.networking.k8s.io"),
					Kind:        stringPtr("Gateway"),
					Namespace:   stringPtr("ingress"),
					Name:        valueFrom(catalogkind.CatalogKind_KubernetesGateway, "my-gateway", "status.outputs.gateway_name"),
					SectionName: stringPtr("udp"),
					Port:        int32Ptr(53),
				},
			}
			input.Spec.Rules = []*KubernetesUdpRouteRule{
				{
					Name: stringPtr("forward"),
					BackendRefs: []*kubernetes.KubernetesGatewayApiBackendRef{
						{Name: literal("udp-stable"), Port: int32Ptr(53), Weight: int32Ptr(90)},
						{Name: literal("udp-canary"), Port: int32Ptr(53), Weight: int32Ptr(10)},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("named rule should be valid", func() {
			input.Spec.Rules[0].Name = stringPtr("forward")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("multiple rules should be valid (max_items=16)", func() {
			input.Spec.Rules = []*KubernetesUdpRouteRule{
				{BackendRefs: []*kubernetes.KubernetesGatewayApiBackendRef{serviceBackend("a", 53)}},
				{BackendRefs: []*kubernetes.KubernetesGatewayApiBackendRef{serviceBackend("b", 5353)}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("route without parent_refs should be valid (attachment optional)", func() {
			input.Spec.ParentRefs = nil
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("missing namespace should fail", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("zero rules should fail (min_items=1)", func() {
			input.Spec.Rules = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("more than 16 rules should fail (max_items=16)", func() {
			rules := make([]*KubernetesUdpRouteRule, 0, 17)
			for i := 0; i < 17; i++ {
				rules = append(rules, &KubernetesUdpRouteRule{
					BackendRefs: []*kubernetes.KubernetesGatewayApiBackendRef{serviceBackend("udp-svc", 53)},
				})
			}
			input.Spec.Rules = rules
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rule with no backend_refs should fail (min_items=1)", func() {
			input.Spec.Rules[0].BackendRefs = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("backend ref without a name should fail", func() {
			input.Spec.Rules[0].BackendRefs = []*kubernetes.KubernetesGatewayApiBackendRef{
				{Port: int32Ptr(53)},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("backend ref with out-of-range port should fail", func() {
			input.Spec.Rules[0].BackendRefs = []*kubernetes.KubernetesGatewayApiBackendRef{
				{Name: literal("udp-svc"), Port: int32Ptr(70000)},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("backend ref with out-of-range weight should fail", func() {
			input.Spec.Rules[0].BackendRefs = []*kubernetes.KubernetesGatewayApiBackendRef{
				{Name: literal("udp-svc"), Port: int32Ptr(53), Weight: int32Ptr(1000001)},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("invalid rule name pattern should fail", func() {
			input.Spec.Rules[0].Name = stringPtr("Bad_Name")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a parent ref without a name should fail", func() {
			input.Spec.ParentRefs = []*kubernetes.KubernetesGatewayApiParentReference{
				{Kind: stringPtr("Gateway")},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
