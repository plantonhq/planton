package kubernetesauthorizationpolicyv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesAuthorizationPolicy(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesAuthorizationPolicy Suite")
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

// targetRef returns a same-namespace PolicyTargetReference (namespace left empty,
// as upstream requires in the 1.30 line).
func targetRef(kind, name string) *kubernetes.KubernetesIstioApiPolicyTargetReference {
	return &kubernetes.KubernetesIstioApiPolicyTargetReference{
		Kind: kind,
		Name: literal(name),
	}
}

// rule returns a minimal rule allowing GET from a service-account-less source.
func rule() *KubernetesAuthorizationPolicyRule {
	return &KubernetesAuthorizationPolicyRule{
		To: []*KubernetesAuthorizationPolicyRuleTo{
			{Operation: &KubernetesAuthorizationPolicyOperation{Methods: []string{"GET"}}},
		},
	}
}

var _ = ginkgo.Describe("KubernetesAuthorizationPolicy Validation Tests", func() {
	var input *KubernetesAuthorizationPolicy

	ginkgo.BeforeEach(func() {
		input = &KubernetesAuthorizationPolicy{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesAuthorizationPolicy",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-authorization-policy",
			},
			Spec: &KubernetesAuthorizationPolicySpec{
				Namespace: literal("default"),
				Rules:     []*KubernetesAuthorizationPolicyRule{rule()},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("with a single allow rule (no selector, default action)", func() {
			ginkgo.It("should not return a validation error", func() {
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with no rules at all (an allow-nothing policy)", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Rules = nil
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a selector targeting specific workloads", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Selector = &kubernetes.KubernetesIstioApiWorkloadSelector{
					MatchLabels: map[string]string{"app": "httpbin"},
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

		ginkgo.Context("with each non-CUSTOM action value", func() {
			ginkgo.It("should accept ALLOW, DENY, and AUDIT", func() {
				for _, a := range []string{"ALLOW", "DENY", "AUDIT"} {
					input.Spec.Action = ptr(a)
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "action %s should be valid", a)
				}
			})
		})

		ginkgo.Context("with the CUSTOM action and an extension provider", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Action = ptr("CUSTOM")
				input.Spec.Provider = &KubernetesAuthorizationPolicyExtensionProvider{Name: "my-custom-authz"}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a fully-populated rule (from/to/when)", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Rules = []*KubernetesAuthorizationPolicyRule{
					{
						From: []*KubernetesAuthorizationPolicyRuleFrom{
							{Source: &KubernetesAuthorizationPolicySource{
								Principals:        []string{"cluster.local/ns/default/sa/sleep"},
								RequestPrincipals: []string{"https://accounts.google.com/*"},
								Namespaces:        []string{"prod", "test"},
								IpBlocks:          []string{"203.0.113.0/24"},
								NotIpBlocks:       []string{"203.0.113.4"},
								RemoteIpBlocks:    []string{"198.51.100.0/24"},
							}},
							{Source: &KubernetesAuthorizationPolicySource{
								ServiceAccounts:    []string{"default/productpage"},
								NotServiceAccounts: []string{"default/legacy"},
							}},
						},
						To: []*KubernetesAuthorizationPolicyRuleTo{
							{Operation: &KubernetesAuthorizationPolicyOperation{
								Hosts:    []string{"*.example.com"},
								Ports:    []string{"8080"},
								Methods:  []string{"GET", "POST"},
								Paths:    []string{"/info*", "/data"},
								NotPaths: []string{"/admin*"},
							}},
						},
						When: []*KubernetesAuthorizationPolicyCondition{
							{Key: "request.auth.claims[iss]", Values: []string{"https://accounts.google.com"}},
							{Key: "source.ip", NotValues: []string{"203.0.113.4"}},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with service_accounts at the upstream bounds", func() {
			ginkgo.It("should accept 16 entries each up to 320 characters", func() {
				accounts := make([]string, 0, 16)
				for i := 0; i < 16; i++ {
					accounts = append(accounts, "default/sa")
				}
				accounts[0] = "default/" + strings.Repeat("a", 312) // 320 chars total
				input.Spec.Rules = []*KubernetesAuthorizationPolicyRule{
					{From: []*KubernetesAuthorizationPolicyRuleFrom{
						{Source: &KubernetesAuthorizationPolicySource{ServiceAccounts: accounts}},
					}},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
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
				MatchLabels: map[string]string{"app": "httpbin"},
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
					{Name: literal("edge-gateway")},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a malformed group", func() {
			ginkgo.It("should return a validation error", func() {
				ref := targetRef("Gateway", "edge-gateway")
				ref.Group = "Not A Valid Group"
				input.Spec.TargetRefs = []*kubernetes.KubernetesIstioApiPolicyTargetReference{ref}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a missing name", func() {
			ginkgo.It("should return a validation error", func() {
				// name is a StringValueOrRef foreign key: requiredness is presence of
				// the reference. Upstream's 253-char name bound is enforced by the
				// API server at apply (a StringValueOrRef carries no length bound —
				// the same posture as every other foreign key in the catalog).
				ref := targetRef("Gateway", "edge-gateway")
				ref.Name = nil
				input.Spec.TargetRefs = []*kubernetes.KubernetesIstioApiPolicyTargetReference{ref}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When the action is invalid", func() {
		ginkgo.It("should reject a value outside the closed set", func() {
			input.Spec.Action = ptr("PERMIT")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a lowercase value", func() {
			input.Spec.Action = ptr("allow")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("When a source mixes exclusive identity fields", func() {
		ginkgo.Context("with service_accounts and principals together", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Rules = []*KubernetesAuthorizationPolicyRule{
					{From: []*KubernetesAuthorizationPolicyRuleFrom{
						{Source: &KubernetesAuthorizationPolicySource{
							ServiceAccounts: []string{"default/productpage"},
							Principals:      []string{"cluster.local/ns/default/sa/sleep"},
						}},
					}},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with service_accounts and namespaces together", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Rules = []*KubernetesAuthorizationPolicyRule{
					{From: []*KubernetesAuthorizationPolicyRuleFrom{
						{Source: &KubernetesAuthorizationPolicySource{
							ServiceAccounts: []string{"default/productpage"},
							Namespaces:      []string{"prod"},
						}},
					}},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with not_service_accounts and not_principals together", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Rules = []*KubernetesAuthorizationPolicyRule{
					{From: []*KubernetesAuthorizationPolicyRuleFrom{
						{Source: &KubernetesAuthorizationPolicySource{
							NotServiceAccounts: []string{"default/legacy"},
							NotPrincipals:      []string{"cluster.local/ns/default/sa/sleep"},
						}},
					}},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When service_accounts exceed the upstream bounds", func() {
		ginkgo.Context("with more than 16 entries", func() {
			ginkgo.It("should return a validation error", func() {
				accounts := make([]string, 0, 17)
				for i := 0; i < 17; i++ {
					accounts = append(accounts, "default/sa")
				}
				input.Spec.Rules = []*KubernetesAuthorizationPolicyRule{
					{From: []*KubernetesAuthorizationPolicyRuleFrom{
						{Source: &KubernetesAuthorizationPolicySource{ServiceAccounts: accounts}},
					}},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with an entry longer than 320 characters", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Rules = []*KubernetesAuthorizationPolicyRule{
					{From: []*KubernetesAuthorizationPolicyRuleFrom{
						{Source: &KubernetesAuthorizationPolicySource{
							ServiceAccounts: []string{"default/" + strings.Repeat("a", 313)}, // 321 chars
						}},
					}},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When a when-condition is invalid", func() {
		ginkgo.Context("without a key", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Rules = []*KubernetesAuthorizationPolicyRule{
					{When: []*KubernetesAuthorizationPolicyCondition{
						{Values: []string{"https://accounts.google.com"}},
					}},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When an extension provider is invalid", func() {
		ginkgo.Context("with an empty provider name", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Action = ptr("CUSTOM")
				input.Spec.Provider = &KubernetesAuthorizationPolicyExtensionProvider{Name: ""}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When too many rules are provided", func() {
		ginkgo.It("should reject more than 512 entries", func() {
			// Each entry is a valid rule on its own (match_all), so the only
			// thing this policy gets wrong is the count.
			rules := make([]*KubernetesAuthorizationPolicyRule, 0, 513)
			for i := 0; i < 513; i++ {
				rules = append(rules, &KubernetesAuthorizationPolicyRule{MatchAll: ptr(true)})
			}
			input.Spec.Rules = rules
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(strings.Contains(err.Error(), "512")).To(gomega.BeTrue())
		})
	})

	ginkgo.Describe("When a rule says what it matches", func() {
		ginkgo.Context("with match_all: true and no matchers (match every request)", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Rules = []*KubernetesAuthorizationPolicyRule{{MatchAll: ptr(true)}}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with matchers and no match_all", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Rules = []*KubernetesAuthorizationPolicyRule{rule()}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with an empty rule that names neither match_all nor a matcher", func() {
			ginkgo.It("should return a validation error naming the choice", func() {
				input.Spec.Rules = []*KubernetesAuthorizationPolicyRule{{}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(strings.Contains(err.Error(), "a rule must say what it matches")).To(gomega.BeTrue())
			})
		})

		ginkgo.Context("with match_all: true beside a matcher", func() {
			ginkgo.It("should return a validation error naming the contradiction", func() {
				r := rule()
				r.MatchAll = ptr(true)
				input.Spec.Rules = []*KubernetesAuthorizationPolicyRule{r}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(strings.Contains(err.Error(), "match_all: true matches every request")).To(gomega.BeTrue())
			})
		})

		ginkgo.Context("with match_all: false and no matchers", func() {
			ginkgo.It("should return a validation error (a stated false is not a matcher)", func() {
				input.Spec.Rules = []*KubernetesAuthorizationPolicyRule{{MatchAll: ptr(false)}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When the selector match labels are invalid", func() {
		ginkgo.Context("with a wildcard in a label key", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Selector = &kubernetes.KubernetesIstioApiWorkloadSelector{
					MatchLabels: map[string]string{"app*": "httpbin"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})
