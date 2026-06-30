package kubernetestlsroutev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestKubernetesTlsRoute(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesTlsRoute Suite")
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
		Name: name,
		Port: int32Ptr(port),
	}
}

var _ = ginkgo.Describe("KubernetesTlsRoute Validation Tests", func() {
	var input *KubernetesTlsRoute

	ginkgo.BeforeEach(func() {
		input = &KubernetesTlsRoute{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesTlsRoute",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-tls-route",
			},
			Spec: &KubernetesTlsRouteSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: literal("app-ns"),
				ParentRefs: []*kubernetes.KubernetesGatewayApiParentReference{
					{Name: "my-gateway"},
				},
				Hostnames: []string{"secure.example.com"},
				Rules: []*KubernetesTlsRouteRule{
					{
						BackendRefs: []*kubernetes.KubernetesGatewayApiBackendRef{
							serviceBackend("tls-svc", 8443),
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

		ginkgo.It("wildcard SNI hostname should be valid", func() {
			input.Spec.Hostnames = []string{"*.example.com", "secure.example.com"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("named rule should be valid", func() {
			input.Spec.Rules[0].Name = stringPtr("passthrough")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("weighted backends should be valid", func() {
			input.Spec.Rules[0].BackendRefs = []*kubernetes.KubernetesGatewayApiBackendRef{
				{Name: "tls-stable", Port: int32Ptr(8443), Weight: int32Ptr(90)},
				{Name: "tls-canary", Port: int32Ptr(8443), Weight: int32Ptr(10)},
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

		ginkgo.It("missing hostnames should fail (min_items=1)", func() {
			input.Spec.Hostnames = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("IP address as SNI hostname should fail (RFC 6066)", func() {
			input.Spec.Hostnames = []string{"10.0.0.1"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("invalid hostname pattern should fail", func() {
			input.Spec.Hostnames = []string{"Not_A_Host"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("more than 16 hostnames should fail", func() {
			input.Spec.Hostnames = []string{
				"a.example.com", "b.example.com", "c.example.com", "d.example.com",
				"e.example.com", "f.example.com", "g.example.com", "h.example.com",
				"i.example.com", "j.example.com", "k.example.com", "l.example.com",
				"m.example.com", "n.example.com", "o.example.com", "p.example.com",
				"q.example.com",
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("zero rules should fail (min_items=1)", func() {
			input.Spec.Rules = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("more than one rule should fail (max_items=1)", func() {
			input.Spec.Rules = []*KubernetesTlsRouteRule{
				{BackendRefs: []*kubernetes.KubernetesGatewayApiBackendRef{serviceBackend("a", 8443)}},
				{BackendRefs: []*kubernetes.KubernetesGatewayApiBackendRef{serviceBackend("b", 8443)}},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rule with no backend_refs should fail (min_items=1)", func() {
			input.Spec.Rules[0].BackendRefs = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("backend ref without a name should fail", func() {
			input.Spec.Rules[0].BackendRefs = []*kubernetes.KubernetesGatewayApiBackendRef{
				{Port: int32Ptr(8443)},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("backend ref with out-of-range port should fail", func() {
			input.Spec.Rules[0].BackendRefs = []*kubernetes.KubernetesGatewayApiBackendRef{
				{Name: "tls-svc", Port: int32Ptr(70000)},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("backend ref with invalid kind pattern should fail", func() {
			input.Spec.Rules[0].BackendRefs = []*kubernetes.KubernetesGatewayApiBackendRef{
				{Name: "tls-svc", Kind: stringPtr("bad kind"), Port: int32Ptr(8443)},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("invalid rule name pattern should fail", func() {
			input.Spec.Rules[0].Name = stringPtr("Bad_Name")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a parent ref with a malformed kind should fail", func() {
			input.Spec.ParentRefs = []*kubernetes.KubernetesGatewayApiParentReference{
				{Name: "my-gateway", Kind: stringPtr("bad/kind")},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
