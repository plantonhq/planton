package kubernetestcproutev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesTcpRoute(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesTcpRoute Suite")
}

func int32Ptr(i int32) *int32    { return &i }
func stringPtr(s string) *string { return &s }

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// serviceBackend returns a minimal valid backend reference to a Service.
func serviceBackend(name string, port int32) *kubernetes.KubernetesGatewayApiBackendRef {
	return &kubernetes.KubernetesGatewayApiBackendRef{
		Name: literal(name),
		Port: int32Ptr(port),
	}
}

var _ = ginkgo.Describe("KubernetesTcpRoute Validation Tests", func() {
	var input *KubernetesTcpRoute

	ginkgo.BeforeEach(func() {
		input = &KubernetesTcpRoute{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesTcpRoute",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-tcp-route",
			},
			Spec: &KubernetesTcpRouteSpec{
				Namespace: literal("app-ns"),
				ParentRefs: []*kubernetes.KubernetesGatewayApiParentReference{
					{Name: literal("my-gateway")},
				},
				Rules: []*KubernetesTcpRouteRule{
					{
						BackendRefs: []*kubernetes.KubernetesGatewayApiBackendRef{
							serviceBackend("tcp-svc", 5432),
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

		ginkgo.It("named rule should be valid", func() {
			input.Spec.Rules[0].Name = stringPtr("forward")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("multiple weighted backends should be valid", func() {
			input.Spec.Rules[0].BackendRefs = []*kubernetes.KubernetesGatewayApiBackendRef{
				{Name: literal("tcp-stable"), Port: int32Ptr(5432), Weight: int32Ptr(90)},
				{Name: literal("tcp-canary"), Port: int32Ptr(5432), Weight: int32Ptr(10)},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("multiple rules should be valid (max_items=16)", func() {
			input.Spec.Rules = []*KubernetesTcpRouteRule{
				{BackendRefs: []*kubernetes.KubernetesGatewayApiBackendRef{serviceBackend("a", 5432)}},
				{BackendRefs: []*kubernetes.KubernetesGatewayApiBackendRef{serviceBackend("b", 5433)}},
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

		ginkgo.It("rule with no backend_refs should fail (min_items=1)", func() {
			input.Spec.Rules[0].BackendRefs = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("backend ref without a name should fail", func() {
			input.Spec.Rules[0].BackendRefs = []*kubernetes.KubernetesGatewayApiBackendRef{
				{Port: int32Ptr(5432)},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("backend ref with out-of-range port should fail", func() {
			input.Spec.Rules[0].BackendRefs = []*kubernetes.KubernetesGatewayApiBackendRef{
				{Name: literal("tcp-svc"), Port: int32Ptr(70000)},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("backend ref with invalid group pattern should fail", func() {
			input.Spec.Rules[0].BackendRefs = []*kubernetes.KubernetesGatewayApiBackendRef{
				{Name: literal("tcp-svc"), Group: stringPtr("Bad_Group"), Port: int32Ptr(5432)},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("invalid rule name pattern should fail", func() {
			input.Spec.Rules[0].Name = stringPtr("Bad_Name")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a parent ref with a malformed kind should fail", func() {
			input.Spec.ParentRefs = []*kubernetes.KubernetesGatewayApiParentReference{
				{Name: literal("my-gateway"), Kind: stringPtr("bad/kind")},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
