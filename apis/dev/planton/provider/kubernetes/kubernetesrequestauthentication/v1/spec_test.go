package kubernetesrequestauthenticationv1

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

func TestKubernetesRequestAuthentication(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesRequestAuthentication Suite")
}

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

// ptr returns a pointer to v, for setting proto3 `optional` scalar fields.
func ptr[T any](v T) *T { return &v }

// jwtRule returns a minimal valid JWT rule: an issuer plus a jwks_uri.
func jwtRule(issuer, jwksURI string) *KubernetesRequestAuthenticationJwtRule {
	return &KubernetesRequestAuthenticationJwtRule{
		Issuer:  issuer,
		JwksUri: ptr(jwksURI),
	}
}

// targetRef returns a same-namespace PolicyTargetReference (namespace left empty,
// as upstream requires in the 1.26 line).
func targetRef(kind, name string) *kubernetes.KubernetesIstioApiPolicyTargetReference {
	return &kubernetes.KubernetesIstioApiPolicyTargetReference{
		Kind: kind,
		Name: name,
	}
}

var _ = ginkgo.Describe("KubernetesRequestAuthentication Validation Tests", func() {
	var input *KubernetesRequestAuthentication

	ginkgo.BeforeEach(func() {
		input = &KubernetesRequestAuthentication{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesRequestAuthentication",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-request-authentication",
			},
			Spec: &KubernetesRequestAuthenticationSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: literal("default"),
				JwtRules: []*KubernetesRequestAuthenticationJwtRule{
					jwtRule("https://accounts.example.com", "https://accounts.example.com/.well-known/jwks.json"),
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("with a single issuer + jwks_uri rule (no selector)", func() {
			ginkgo.It("should not return a validation error", func() {
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with no jwt_rules at all (a no-op policy)", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.JwtRules = nil
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a selector targeting specific workloads", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Selector = &kubernetes.KubernetesIstioApiWorkloadSelector{
					MatchLabels: map[string]string{"app": "finance"},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with target_refs instead of a selector", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.TargetRefs = []*kubernetes.KubernetesIstioApiPolicyTargetReference{
					targetRef("Gateway", "edge-gateway"),
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with an inline jwks instead of a jwks_uri", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.JwtRules = []*KubernetesRequestAuthenticationJwtRule{
					{
						Issuer: "https://accounts.example.com",
						Jwks:   ptr(`{"keys":[]}`),
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a fully-populated jwt rule", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.JwtRules = []*KubernetesRequestAuthenticationJwtRule{
					{
						Issuer:    "https://accounts.example.com",
						Audiences: []string{"bookstore.example.com"},
						JwksUri:   ptr("https://accounts.example.com/jwks.json"),
						FromHeaders: []*KubernetesRequestAuthenticationJwtHeader{
							{Name: "x-jwt-assertion", Prefix: ptr("Bearer ")},
						},
						FromParams:            []string{"access_token"},
						FromCookies:           []string{"auth-token"},
						OutputPayloadToHeader: ptr("x-jwt-payload"),
						ForwardOriginalToken:  ptr(true),
						OutputClaimToHeaders: []*KubernetesRequestAuthenticationClaimToHeader{
							{Header: "x-jwt-group", Claim: "groups"},
							{Header: "x-jwt-nested", Claim: "nested.key.group"},
						},
						Timeout: ptr("5s"),
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with various valid timeout durations", func() {
			ginkgo.It("should accept 1ms, 1500ms, 5s, and 1m30s", func() {
				for _, d := range []string{"1ms", "1500ms", "5s", "1m30s"} {
					input.Spec.JwtRules[0].Timeout = ptr(d)
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "timeout %s should be valid", d)
				}
			})
		})

		ginkgo.Context("with the namespace resolved via a valueFrom reference", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Namespace = valueFrom(cloudresourcekind.CloudResourceKind_KubernetesNamespace, "finance-ns", "spec.name")
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When the envelope is invalid", func() {
		ginkgo.Context("with the wrong api_version", func() {
			ginkgo.It("should return a validation error", func() {
				input.ApiVersion = "v1"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with the wrong kind", func() {
			ginkgo.It("should return a validation error", func() {
				input.Kind = "KubernetesPeerAuthentication"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("without metadata", func() {
			ginkgo.It("should return a validation error", func() {
				input.Metadata = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("without a spec", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("without a namespace", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Namespace = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When selector and target_refs are both set", func() {
		ginkgo.It("should return a validation error (at most one attachment)", func() {
			input.Spec.Selector = &kubernetes.KubernetesIstioApiWorkloadSelector{
				MatchLabels: map[string]string{"app": "finance"},
			}
			input.Spec.TargetRefs = []*kubernetes.KubernetesIstioApiPolicyTargetReference{
				targetRef("Gateway", "edge-gateway"),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When too many target_refs are provided", func() {
		ginkgo.It("should reject more than 16 entries", func() {
			refs := make([]*kubernetes.KubernetesIstioApiPolicyTargetReference, 0, 17)
			for i := 0; i < 17; i++ {
				refs = append(refs, targetRef("Service", "svc"))
			}
			input.Spec.TargetRefs = refs
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When a target_ref is invalid", func() {
		ginkgo.Context("with a cross-namespace reference", func() {
			ginkgo.It("should return a validation error", func() {
				ref := targetRef("Gateway", "edge-gateway")
				ref.Namespace = "other-ns"
				input.Spec.TargetRefs = []*kubernetes.KubernetesIstioApiPolicyTargetReference{ref}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("without a kind", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.TargetRefs = []*kubernetes.KubernetesIstioApiPolicyTargetReference{
					{Name: "edge-gateway"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("without a name", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.TargetRefs = []*kubernetes.KubernetesIstioApiPolicyTargetReference{
					{Kind: "Gateway"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When a jwt rule is invalid", func() {
		ginkgo.Context("with both jwks_uri and jwks set", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.JwtRules = []*KubernetesRequestAuthenticationJwtRule{
					{
						Issuer:  "https://accounts.example.com",
						JwksUri: ptr("https://accounts.example.com/jwks.json"),
						Jwks:    ptr(`{"keys":[]}`),
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a missing issuer", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.JwtRules = []*KubernetesRequestAuthenticationJwtRule{
					{JwksUri: ptr("https://accounts.example.com/jwks.json")},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a jwks_uri using a non-http(s) scheme", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.JwtRules = []*KubernetesRequestAuthenticationJwtRule{
					{
						Issuer:  "https://accounts.example.com",
						JwksUri: ptr("ftp://accounts.example.com/jwks.json"),
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a timeout below the 1ms minimum", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.JwtRules[0].Timeout = ptr("0s")
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a malformed timeout duration", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.JwtRules[0].Timeout = ptr("not-a-duration")
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with an empty audience entry", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.JwtRules[0].Audiences = []string{""}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When a from_header is invalid", func() {
		ginkgo.Context("without a header name", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.JwtRules[0].FromHeaders = []*KubernetesRequestAuthenticationJwtHeader{
					{Prefix: ptr("Bearer ")},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When an output_claim_to_header is invalid", func() {
		ginkgo.Context("with a header name containing illegal characters", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.JwtRules[0].OutputClaimToHeaders = []*KubernetesRequestAuthenticationClaimToHeader{
					{Header: "x bad header", Claim: "groups"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("without a claim", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.JwtRules[0].OutputClaimToHeaders = []*KubernetesRequestAuthenticationClaimToHeader{
					{Header: "x-jwt-group"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When the selector match labels are invalid", func() {
		ginkgo.Context("with a wildcard in a label key", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Selector = &kubernetes.KubernetesIstioApiWorkloadSelector{
					MatchLabels: map[string]string{"app*": "finance"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})
