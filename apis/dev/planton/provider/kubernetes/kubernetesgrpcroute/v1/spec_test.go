package kubernetesgrpcroutev1

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

func TestKubernetesGrpcRoute(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesGrpcRoute Suite")
}

func stringPtr(s string) *string { return &s }
func int32Ptr(i int32) *int32    { return &i }

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// serviceBackend returns a minimal valid backend reference to a Service.
func serviceBackend(name string, port int32) *KubernetesGrpcRouteBackendRef {
	return &KubernetesGrpcRouteBackendRef{
		Name: name,
		Port: int32Ptr(port),
	}
}

var _ = ginkgo.Describe("KubernetesGrpcRoute Validation Tests", func() {
	var input *KubernetesGrpcRoute

	ginkgo.BeforeEach(func() {
		input = &KubernetesGrpcRoute{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesGrpcRoute",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-grpc-route",
			},
			Spec: &KubernetesGrpcRouteSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: literal("app-ns"),
				ParentRefs: []*kubernetes.KubernetesGatewayApiParentReference{
					{Name: "my-gateway"},
				},
				Rules: []*KubernetesGrpcRouteRule{
					{
						BackendRefs: []*KubernetesGrpcRouteBackendRef{
							serviceBackend("grpc-svc", 9000),
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

		ginkgo.It("hostname and service/method routing should be valid", func() {
			input.Spec.Hostnames = []string{"api.example.com", "*.example.com"}
			input.Spec.Rules[0].Matches = []*KubernetesGrpcRouteMatch{
				{Method: &KubernetesGrpcRouteMethodMatch{
					Type:    stringPtr("Exact"),
					Service: stringPtr("helloworld.Greeter"),
					Method:  stringPtr("SayHello"),
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("service-only method match should be valid", func() {
			input.Spec.Rules[0].Matches = []*KubernetesGrpcRouteMatch{
				{Method: &KubernetesGrpcRouteMethodMatch{Service: stringPtr("helloworld.Greeter")}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("regular-expression method match should not enforce the exact-char regex", func() {
			input.Spec.Rules[0].Matches = []*KubernetesGrpcRouteMatch{
				{Method: &KubernetesGrpcRouteMethodMatch{
					Type:    stringPtr("RegularExpression"),
					Service: stringPtr("helloworld\\..*"),
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("header match should be valid", func() {
			input.Spec.Rules[0].Matches = []*KubernetesGrpcRouteMatch{
				{Headers: []*KubernetesGrpcRouteHeaderMatch{{Name: "x-version", Value: "v2"}}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("weighted canary across two backends should be valid", func() {
			input.Spec.Rules[0].BackendRefs = []*KubernetesGrpcRouteBackendRef{
				{Name: "grpc-stable", Port: int32Ptr(9000), Weight: int32Ptr(90)},
				{Name: "grpc-canary", Port: int32Ptr(9000), Weight: int32Ptr(10)},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("request header modifier filter should be valid", func() {
			input.Spec.Rules[0].Filters = []*KubernetesGrpcRouteFilter{
				{
					Type: "RequestHeaderModifier",
					RequestHeaderModifier: &KubernetesGrpcRouteHeaderFilter{
						Set:    []*KubernetesGrpcRouteHeader{{Name: "x-env", Value: "prod"}},
						Add:    []*KubernetesGrpcRouteHeader{{Name: "x-trace", Value: "on"}},
						Remove: []string{"x-internal"},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("request mirror filter with percent should be valid", func() {
			input.Spec.Rules[0].Filters = []*KubernetesGrpcRouteFilter{
				{
					Type: "RequestMirror",
					RequestMirror: &KubernetesGrpcRouteRequestMirrorFilter{
						BackendRef: &kubernetes.KubernetesGatewayApiBackendObjectReference{
							Name: "grpc-mirror",
							Port: int32Ptr(9000),
						},
						Percent: int32Ptr(10),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("extension ref filter should be valid", func() {
			input.Spec.Rules[0].Filters = []*KubernetesGrpcRouteFilter{
				{
					Type: "ExtensionRef",
					ExtensionRef: &kubernetes.KubernetesGatewayApiLocalObjectReference{
						Group: "example.com",
						Kind:  "MyGrpcFilter",
						Name:  "custom",
					},
				},
			}
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

		ginkgo.It("invalid hostname should fail", func() {
			input.Spec.Hostnames = []string{"Not_A_Host"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("invalid method match type should fail", func() {
			input.Spec.Rules[0].Matches = []*KubernetesGrpcRouteMatch{
				{Method: &KubernetesGrpcRouteMethodMatch{Type: stringPtr("Prefix"), Service: stringPtr("foo.Bar")}},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("Exact method match with neither service nor method should fail", func() {
			input.Spec.Rules[0].Matches = []*KubernetesGrpcRouteMatch{
				{Method: &KubernetesGrpcRouteMethodMatch{Type: stringPtr("Exact")}},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("Exact method match with invalid service chars should fail", func() {
			input.Spec.Rules[0].Matches = []*KubernetesGrpcRouteMatch{
				{Method: &KubernetesGrpcRouteMethodMatch{Type: stringPtr("Exact"), Service: stringPtr("foo/bar")}},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("duplicate header match names should fail", func() {
			input.Spec.Rules[0].Matches = []*KubernetesGrpcRouteMatch{
				{Headers: []*KubernetesGrpcRouteHeaderMatch{
					{Name: "x-dup", Value: "a"},
					{Name: "x-dup", Value: "b"},
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("invalid header match name pattern should fail", func() {
			input.Spec.Rules[0].Matches = []*KubernetesGrpcRouteMatch{
				{Headers: []*KubernetesGrpcRouteHeaderMatch{{Name: "bad header", Value: "v"}}},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("filter type/config mismatch should fail", func() {
			input.Spec.Rules[0].Filters = []*KubernetesGrpcRouteFilter{
				{
					Type: "RequestMirror",
					RequestHeaderModifier: &KubernetesGrpcRouteHeaderFilter{
						Set: []*KubernetesGrpcRouteHeader{{Name: "x", Value: "y"}},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("unknown filter type should fail", func() {
			input.Spec.Rules[0].Filters = []*KubernetesGrpcRouteFilter{
				{Type: "RequestRedirect"},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("repeated RequestHeaderModifier filter should fail", func() {
			input.Spec.Rules[0].Filters = []*KubernetesGrpcRouteFilter{
				{Type: "RequestHeaderModifier", RequestHeaderModifier: &KubernetesGrpcRouteHeaderFilter{Set: []*KubernetesGrpcRouteHeader{{Name: "a", Value: "1"}}}},
				{Type: "RequestHeaderModifier", RequestHeaderModifier: &KubernetesGrpcRouteHeaderFilter{Set: []*KubernetesGrpcRouteHeader{{Name: "b", Value: "2"}}}},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("request mirror with both percent and fraction should fail", func() {
			input.Spec.Rules[0].Filters = []*KubernetesGrpcRouteFilter{
				{
					Type: "RequestMirror",
					RequestMirror: &KubernetesGrpcRouteRequestMirrorFilter{
						BackendRef: &kubernetes.KubernetesGatewayApiBackendObjectReference{Name: "m", Port: int32Ptr(9000)},
						Percent:    int32Ptr(10),
						Fraction:   &kubernetes.KubernetesGatewayApiFraction{Numerator: 5},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("backend ref weight above max should fail", func() {
			input.Spec.Rules[0].BackendRefs = []*KubernetesGrpcRouteBackendRef{
				{Name: "grpc-svc", Port: int32Ptr(9000), Weight: int32Ptr(1000001)},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("backend ref missing name should fail", func() {
			input.Spec.Rules[0].BackendRefs = []*KubernetesGrpcRouteBackendRef{
				{Port: int32Ptr(9000)},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("too many parent refs should fail (max 32)", func() {
			refs := make([]*kubernetes.KubernetesGatewayApiParentReference, 0, 33)
			for i := 0; i < 33; i++ {
				refs = append(refs, &kubernetes.KubernetesGatewayApiParentReference{Name: "gw"})
			}
			input.Spec.ParentRefs = refs
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a parent ref with a malformed kind should fail", func() {
			input.Spec.ParentRefs = []*kubernetes.KubernetesGatewayApiParentReference{
				{Name: "my-gateway", Kind: stringPtr("bad/kind")},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a mirror filter backend ref with a malformed group should fail", func() {
			input.Spec.Rules[0].Filters = []*KubernetesGrpcRouteFilter{
				{
					Type: "RequestMirror",
					RequestMirror: &KubernetesGrpcRouteRequestMirrorFilter{
						BackendRef: &kubernetes.KubernetesGatewayApiBackendObjectReference{
							Name: "m", Port: int32Ptr(9000), Group: stringPtr("Bad_Group"),
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("an extension ref without a kind should fail", func() {
			input.Spec.Rules[0].Filters = []*KubernetesGrpcRouteFilter{
				{
					Type: "ExtensionRef",
					ExtensionRef: &kubernetes.KubernetesGatewayApiLocalObjectReference{
						Group: "example.com", Name: "custom",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
