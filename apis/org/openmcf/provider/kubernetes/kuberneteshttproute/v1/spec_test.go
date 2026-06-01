package kuberneteshttproutev1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestKubernetesHttpRoute(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesHttpRoute Suite")
}

func stringPtr(s string) *string { return &s }
func int32Ptr(i int32) *int32    { return &i }
func boolPtr(b bool) *bool       { return &b }

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func valueFrom(kind cloudresourcekind.CloudResourceKind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
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
func serviceBackend(name string, port int32) *KubernetesHttpRouteBackendRef {
	return &KubernetesHttpRouteBackendRef{
		Name: name,
		Port: int32Ptr(port),
	}
}

var _ = ginkgo.Describe("KubernetesHttpRoute Validation Tests", func() {
	var input *KubernetesHttpRoute

	ginkgo.BeforeEach(func() {
		input = &KubernetesHttpRoute{
			ApiVersion: "kubernetes.openmcf.org/v1",
			Kind:       "KubernetesHttpRoute",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-route",
			},
			Spec: &KubernetesHttpRouteSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: literal("app-ns"),
				ParentRefs: []*kubernetes.KubernetesGatewayApiParentReference{
					{Name: "my-gateway"},
				},
				Rules: []*KubernetesHttpRouteRule{
					{
						BackendRefs: []*KubernetesHttpRouteBackendRef{
							serviceBackend("web", 80),
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

		ginkgo.It("hostname and path-prefix routing should be valid", func() {
			input.Spec.Hostnames = []string{"app.example.com", "*.example.com"}
			input.Spec.Rules[0].Matches = []*KubernetesHttpRouteMatch{
				{Path: &KubernetesHttpRoutePathMatch{Type: stringPtr("PathPrefix"), Value: stringPtr("/api")}},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("weighted canary across two backends should be valid", func() {
			input.Spec.Rules[0].BackendRefs = []*KubernetesHttpRouteBackendRef{
				{Name: "web-stable", Port: int32Ptr(80), Weight: int32Ptr(90)},
				{Name: "web-canary", Port: int32Ptr(80), Weight: int32Ptr(10)},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("header, query-param, and method matches should be valid", func() {
			input.Spec.Rules[0].Matches = []*KubernetesHttpRouteMatch{
				{
					Path:        &KubernetesHttpRoutePathMatch{Type: stringPtr("Exact"), Value: stringPtr("/checkout")},
					Method:      stringPtr("POST"),
					Headers:     []*KubernetesHttpRouteHeaderMatch{{Name: "x-version", Value: "v2"}},
					QueryParams: []*KubernetesHttpRouteQueryParamMatch{{Name: "debug", Value: "true"}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("request header modifier filter should be valid", func() {
			input.Spec.Rules[0].Filters = []*KubernetesHttpRouteFilter{
				{
					Type: "RequestHeaderModifier",
					RequestHeaderModifier: &KubernetesHttpRouteHeaderFilter{
						Set:    []*KubernetesHttpRouteHeader{{Name: "x-env", Value: "prod"}},
						Add:    []*KubernetesHttpRouteHeader{{Name: "x-trace", Value: "on"}},
						Remove: []string{"x-internal"},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("request redirect filter (no backends) should be valid", func() {
			input.Spec.Rules[0] = &KubernetesHttpRouteRule{
				Filters: []*KubernetesHttpRouteFilter{
					{
						Type: "RequestRedirect",
						RequestRedirect: &KubernetesHttpRouteRequestRedirectFilter{
							Scheme:     stringPtr("https"),
							StatusCode: int32Ptr(301),
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("url rewrite with replace-prefix and a single PathPrefix match should be valid", func() {
			input.Spec.Rules[0].Matches = []*KubernetesHttpRouteMatch{
				{Path: &KubernetesHttpRoutePathMatch{Type: stringPtr("PathPrefix"), Value: stringPtr("/old")}},
			}
			input.Spec.Rules[0].Filters = []*KubernetesHttpRouteFilter{
				{
					Type: "URLRewrite",
					UrlRewrite: &KubernetesHttpRouteUrlRewriteFilter{
						Path: &KubernetesHttpRoutePathModifier{
							Type:               "ReplacePrefixMatch",
							ReplacePrefixMatch: stringPtr("/new"),
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("request mirror filter with percent should be valid", func() {
			input.Spec.Rules[0].Filters = []*KubernetesHttpRouteFilter{
				{
					Type: "RequestMirror",
					RequestMirror: &KubernetesHttpRouteRequestMirrorFilter{
						BackendRef: &kubernetes.KubernetesGatewayApiBackendObjectReference{Name: "mirror-svc", Port: int32Ptr(80)},
						Percent:    int32Ptr(10),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("request mirror filter with fraction should be valid", func() {
			input.Spec.Rules[0].Filters = []*KubernetesHttpRouteFilter{
				{
					Type: "RequestMirror",
					RequestMirror: &KubernetesHttpRouteRequestMirrorFilter{
						BackendRef: &kubernetes.KubernetesGatewayApiBackendObjectReference{Name: "mirror-svc", Port: int32Ptr(80)},
						Fraction:   &kubernetes.KubernetesGatewayApiFraction{Numerator: 5, Denominator: int32Ptr(1000)},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("CORS filter should be valid", func() {
			input.Spec.Rules[0].Filters = []*KubernetesHttpRouteFilter{
				{
					Type: "CORS",
					Cors: &KubernetesHttpRouteCorsFilter{
						AllowOrigins:     []string{"https://app.example.com"},
						AllowMethods:     []string{"GET", "POST"},
						AllowHeaders:     []string{"x-auth"},
						AllowCredentials: boolPtr(true),
						MaxAge:           int32Ptr(600),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("CORS filter with wildcard-only origin should be valid", func() {
			input.Spec.Rules[0].Filters = []*KubernetesHttpRouteFilter{
				{
					Type: "CORS",
					Cors: &KubernetesHttpRouteCorsFilter{
						AllowOrigins: []string{"*"},
						AllowMethods: []string{"*"},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("extension ref filter should be valid", func() {
			input.Spec.Rules[0].Filters = []*KubernetesHttpRouteFilter{
				{
					Type: "ExtensionRef",
					ExtensionRef: &kubernetes.KubernetesGatewayApiLocalObjectReference{
						Group: "networking.example.io",
						Kind:  "MyFilter",
						Name:  "rate-limit",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a fully-specified parent ref (group/kind/section_name/port) should be valid", func() {
			input.Spec.ParentRefs = []*kubernetes.KubernetesGatewayApiParentReference{
				{
					Group:       stringPtr("gateway.networking.k8s.io"),
					Kind:        stringPtr("Gateway"),
					Name:        "my-gateway",
					SectionName: stringPtr("https"),
					Port:        int32Ptr(443),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a parent ref with the core API group (empty string) should be valid", func() {
			input.Spec.ParentRefs = []*kubernetes.KubernetesGatewayApiParentReference{
				{Group: stringPtr(""), Kind: stringPtr("Service"), Name: "my-svc"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("timeouts with backend_request <= request should be valid", func() {
			input.Spec.Rules[0].Timeouts = &KubernetesHttpRouteTimeouts{
				Request:        stringPtr("10s"),
				BackendRequest: stringPtr("5s"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a zero request timeout should not trigger the backend comparison", func() {
			input.Spec.Rules[0].Timeouts = &KubernetesHttpRouteTimeouts{
				Request:        stringPtr("0s"),
				BackendRequest: stringPtr("30s"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a valueFrom reference should be valid", func() {
			input.Spec.Namespace = valueFrom(cloudresourcekind.CloudResourceKind_KubernetesNamespace, "app-ns", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("nil spec should return an error", func() {
			input.Spec = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("nil metadata should return an error", func() {
			input.Metadata = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("missing namespace should return an error", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("empty rules should return an error", func() {
			input.Spec.Rules = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("parent ref without a name should return an error", func() {
			input.Spec.ParentRefs = []*kubernetes.KubernetesGatewayApiParentReference{{}}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a parent ref with a malformed group should return an error", func() {
			input.Spec.ParentRefs = []*kubernetes.KubernetesGatewayApiParentReference{
				{Name: "my-gateway", Group: stringPtr("Invalid_Group")},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a parent ref with a malformed kind should return an error", func() {
			input.Spec.ParentRefs = []*kubernetes.KubernetesGatewayApiParentReference{
				{Name: "my-gateway", Kind: stringPtr("1BadKind")},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a parent ref name longer than 253 chars should return an error", func() {
			input.Spec.ParentRefs = []*kubernetes.KubernetesGatewayApiParentReference{
				{Name: strings.Repeat("a", 254)},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a parent ref with a malformed section_name should return an error", func() {
			input.Spec.ParentRefs = []*kubernetes.KubernetesGatewayApiParentReference{
				{Name: "my-gateway", SectionName: stringPtr("Bad_Section")},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("backend ref without a name should return an error", func() {
			input.Spec.Rules[0].BackendRefs = []*KubernetesHttpRouteBackendRef{{Port: int32Ptr(80)}}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a mirror filter backend ref with a malformed group should return an error", func() {
			input.Spec.Rules[0].Filters = []*KubernetesHttpRouteFilter{
				{
					Type: "RequestMirror",
					RequestMirror: &KubernetesHttpRouteRequestMirrorFilter{
						BackendRef: &kubernetes.KubernetesGatewayApiBackendObjectReference{
							Name: "mirror-svc", Port: int32Ptr(80), Group: stringPtr("Bad_Group"),
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a mirror filter backend ref with a malformed kind should return an error", func() {
			input.Spec.Rules[0].Filters = []*KubernetesHttpRouteFilter{
				{
					Type: "RequestMirror",
					RequestMirror: &KubernetesHttpRouteRequestMirrorFilter{
						BackendRef: &kubernetes.KubernetesGatewayApiBackendObjectReference{
							Name: "mirror-svc", Port: int32Ptr(80), Kind: stringPtr("bad/kind"),
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("an extension ref without a kind should return an error", func() {
			input.Spec.Rules[0].Filters = []*KubernetesHttpRouteFilter{
				{
					Type: "ExtensionRef",
					ExtensionRef: &kubernetes.KubernetesGatewayApiLocalObjectReference{
						Group: "networking.example.io", Name: "rate-limit",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("an extension ref with a malformed group should return an error", func() {
			input.Spec.Rules[0].Filters = []*KubernetesHttpRouteFilter{
				{
					Type: "ExtensionRef",
					ExtensionRef: &kubernetes.KubernetesGatewayApiLocalObjectReference{
						Group: "Bad_Group", Kind: "MyFilter", Name: "rate-limit",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("an uppercase hostname should return an error", func() {
			input.Spec.Hostnames = []string{"App.Example.com"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("an Exact path not starting with '/' should return an error", func() {
			input.Spec.Rules[0].Matches = []*KubernetesHttpRouteMatch{
				{Path: &KubernetesHttpRoutePathMatch{Type: stringPtr("Exact"), Value: stringPtr("api")}},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a PathPrefix value containing '//' should return an error", func() {
			input.Spec.Rules[0].Matches = []*KubernetesHttpRouteMatch{
				{Path: &KubernetesHttpRoutePathMatch{Type: stringPtr("PathPrefix"), Value: stringPtr("/a//b")}},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("an unknown path match type should return an error", func() {
			input.Spec.Rules[0].Matches = []*KubernetesHttpRouteMatch{
				{Path: &KubernetesHttpRoutePathMatch{Type: stringPtr("Glob"), Value: stringPtr("/a")}},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a header match without a name should return an error", func() {
			input.Spec.Rules[0].Matches = []*KubernetesHttpRouteMatch{
				{Headers: []*KubernetesHttpRouteHeaderMatch{{Value: "v2"}}},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("duplicate header match names within a match should return an error", func() {
			input.Spec.Rules[0].Matches = []*KubernetesHttpRouteMatch{
				{Headers: []*KubernetesHttpRouteHeaderMatch{
					{Name: "x-version", Value: "v1"},
					{Name: "x-version", Value: "v2"},
				}},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("an unknown HTTP method should return an error", func() {
			input.Spec.Rules[0].Matches = []*KubernetesHttpRouteMatch{
				{Method: stringPtr("FETCH")},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("an unknown filter type should return an error", func() {
			input.Spec.Rules[0].Filters = []*KubernetesHttpRouteFilter{{Type: "Rumpelstiltskin"}}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a filter whose populated field does not match its type should return an error", func() {
			input.Spec.Rules[0].Filters = []*KubernetesHttpRouteFilter{
				{
					Type:                  "RequestRedirect",
					RequestHeaderModifier: &KubernetesHttpRouteHeaderFilter{Remove: []string{"x"}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a filter whose type has no matching config should return an error", func() {
			input.Spec.Rules[0].Filters = []*KubernetesHttpRouteFilter{{Type: "RequestRedirect"}}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("combining RequestRedirect and URLRewrite on one rule should return an error", func() {
			input.Spec.Rules[0] = &KubernetesHttpRouteRule{
				Filters: []*KubernetesHttpRouteFilter{
					{Type: "RequestRedirect", RequestRedirect: &KubernetesHttpRouteRequestRedirectFilter{StatusCode: int32Ptr(302)}},
					{Type: "URLRewrite", UrlRewrite: &KubernetesHttpRouteUrlRewriteFilter{Hostname: stringPtr("example.com")}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("repeating the RequestHeaderModifier filter should return an error", func() {
			input.Spec.Rules[0].Filters = []*KubernetesHttpRouteFilter{
				{Type: "RequestHeaderModifier", RequestHeaderModifier: &KubernetesHttpRouteHeaderFilter{Remove: []string{"a"}}},
				{Type: "RequestHeaderModifier", RequestHeaderModifier: &KubernetesHttpRouteHeaderFilter{Remove: []string{"b"}}},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a RequestRedirect filter together with backend_refs should return an error", func() {
			input.Spec.Rules[0].Filters = []*KubernetesHttpRouteFilter{
				{Type: "RequestRedirect", RequestRedirect: &KubernetesHttpRouteRequestRedirectFilter{StatusCode: int32Ptr(302)}},
			}
			// baseline rule still has a backend ref, which conflicts with redirect
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a redirect with an unsupported status code should return an error", func() {
			input.Spec.Rules[0] = &KubernetesHttpRouteRule{
				Filters: []*KubernetesHttpRouteFilter{
					{Type: "RequestRedirect", RequestRedirect: &KubernetesHttpRouteRequestRedirectFilter{StatusCode: int32Ptr(404)}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a redirect with an unsupported scheme should return an error", func() {
			input.Spec.Rules[0] = &KubernetesHttpRouteRule{
				Filters: []*KubernetesHttpRouteFilter{
					{Type: "RequestRedirect", RequestRedirect: &KubernetesHttpRouteRequestRedirectFilter{Scheme: stringPtr("ftp")}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a path modifier of type ReplaceFullPath without a value should return an error", func() {
			input.Spec.Rules[0] = &KubernetesHttpRouteRule{
				Filters: []*KubernetesHttpRouteFilter{
					{Type: "URLRewrite", UrlRewrite: &KubernetesHttpRouteUrlRewriteFilter{
						Path: &KubernetesHttpRoutePathModifier{Type: "ReplaceFullPath"},
					}},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a request mirror specifying both percent and fraction should return an error", func() {
			input.Spec.Rules[0].Filters = []*KubernetesHttpRouteFilter{
				{
					Type: "RequestMirror",
					RequestMirror: &KubernetesHttpRouteRequestMirrorFilter{
						BackendRef: &kubernetes.KubernetesGatewayApiBackendObjectReference{Name: "mirror-svc", Port: int32Ptr(80)},
						Percent:    int32Ptr(10),
						Fraction:   &kubernetes.KubernetesGatewayApiFraction{Numerator: 5, Denominator: int32Ptr(1000)},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("CORS allow_origins mixing '*' with other origins should return an error", func() {
			input.Spec.Rules[0].Filters = []*KubernetesHttpRouteFilter{
				{
					Type: "CORS",
					Cors: &KubernetesHttpRouteCorsFilter{
						AllowOrigins: []string{"*", "https://app.example.com"},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("timeouts with backend_request greater than request should return an error", func() {
			input.Spec.Rules[0].Timeouts = &KubernetesHttpRouteTimeouts{
				Request:        stringPtr("5s"),
				BackendRequest: stringPtr("10s"),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("a malformed duration timeout should return an error", func() {
			input.Spec.Rules[0].Timeouts = &KubernetesHttpRouteTimeouts{
				Request: stringPtr("10seconds"),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
