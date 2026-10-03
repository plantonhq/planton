package kuberneteslistenersetv1alpha1

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

func TestKubernetesListenerSet(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesListenerSet Suite")
}

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

// httpListener returns a minimal valid HTTP listener.
func httpListener(name string, port int32) *KubernetesListenerSetListener {
	return &KubernetesListenerSetListener{
		Name:     name,
		Port:     port,
		Protocol: "HTTP",
	}
}

var _ = ginkgo.Describe("KubernetesListenerSet Validation Tests", func() {
	var input *KubernetesListenerSet

	ginkgo.BeforeEach(func() {
		input = &KubernetesListenerSet{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesListenerSet",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-listener-set",
			},
			Spec: &KubernetesListenerSetSpec{
				Namespace: literal("team-ns"),
				ParentRef: &kubernetes.KubernetesGatewayApiParentGatewayReference{
					Name: literal("shared-gateway"),
				},
				Listeners: []*KubernetesListenerSetListener{
					httpListener("http", 8080),
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("with a single HTTP listener", func() {
			ginkgo.It("should not return a validation error", func() {
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with the full surface exercised", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "team-ns", "spec.name")
				input.Spec.ParentRef = &kubernetes.KubernetesGatewayApiParentGatewayReference{
					Group:     stringPtr("gateway.networking.k8s.io"),
					Kind:      stringPtr("Gateway"),
					Namespace: stringPtr("ingress"),
					Name:      valueFrom(catalogkind.CatalogKind_KubernetesGateway, "shared-gateway", "status.outputs.gateway_name"),
				}
				input.Spec.Listeners = []*KubernetesListenerSetListener{
					{
						Name:     "https",
						Hostname: stringPtr("app.example.com"),
						Port:     443,
						Protocol: "HTTPS",
						Tls: &kubernetes.KubernetesGatewayApiListenerTlsConfig{
							Mode: stringPtr("Terminate"),
							CertificateRefs: []*kubernetes.KubernetesGatewayApiSecretObjectReference{
								{Name: literal("app-tls")},
							},
							Options: map[string]string{"example.com/min-tls-version": "1.3"},
						},
						AllowedRoutes: &kubernetes.KubernetesGatewayApiAllowedRoutes{
							Namespaces: &kubernetes.KubernetesGatewayApiRouteNamespaces{
								From: stringPtr("Selector"),
								Selector: &kubernetes.KubernetesGatewayApiLabelSelector{
									MatchLabels: map[string]string{"team": "platform"},
									MatchExpressions: []*kubernetes.KubernetesGatewayApiLabelSelectorRequirement{
										{Key: "tier", Operator: "In", Values: []string{"web", "api"}},
									},
								},
							},
							Kinds: []*kubernetes.KubernetesGatewayApiRouteGroupKind{{Kind: "HTTPRoute"}},
						},
					},
					{
						Name:     "tls-passthrough",
						Hostname: stringPtr("*.passthrough.example.com"),
						Port:     8443,
						Protocol: "TLS",
						Tls:      &kubernetes.KubernetesGatewayApiListenerTlsConfig{Mode: stringPtr("Passthrough")},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("with no namespace", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Namespace = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with no parent_ref", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.ParentRef = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with a parent_ref missing its name", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.ParentRef = &kubernetes.KubernetesGatewayApiParentGatewayReference{
					Kind: stringPtr("Gateway"),
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with no listeners", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with duplicate listener names", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners = []*KubernetesListenerSetListener{
					httpListener("dup", 8080),
					httpListener("dup", 8081),
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with a duplicate port/protocol/hostname combination", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners = []*KubernetesListenerSetListener{
					httpListener("first", 8080),
					httpListener("second", 8080),
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with tls set on an HTTP listener", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners[0].Tls = &kubernetes.KubernetesGatewayApiListenerTlsConfig{
					Mode:            stringPtr("Terminate"),
					CertificateRefs: []*kubernetes.KubernetesGatewayApiSecretObjectReference{{Name: literal("app-tls")}},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with a hostname on a TCP listener", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners = []*KubernetesListenerSetListener{
					{Name: "tcp", Hostname: stringPtr("nope.example.com"), Port: 5432, Protocol: "TCP"},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with an HTTPS listener set to Passthrough mode", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners = []*KubernetesListenerSetListener{
					{
						Name:     "https",
						Port:     443,
						Protocol: "HTTPS",
						Tls:      &kubernetes.KubernetesGatewayApiListenerTlsConfig{Mode: stringPtr("Passthrough")},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with a TLS listener missing its tls mode", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners = []*KubernetesListenerSetListener{
					{Name: "tls", Port: 8443, Protocol: "TLS"},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with a Terminate listener missing both certificate_refs and options", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners = []*KubernetesListenerSetListener{
					{
						Name:     "https",
						Port:     443,
						Protocol: "HTTPS",
						Tls:      &kubernetes.KubernetesGatewayApiListenerTlsConfig{Mode: stringPtr("Terminate")},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with an invalid listener tls mode", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners = []*KubernetesListenerSetListener{
					{
						Name:     "https",
						Port:     443,
						Protocol: "HTTPS",
						Tls: &kubernetes.KubernetesGatewayApiListenerTlsConfig{
							Mode:            stringPtr("Reterminate"),
							CertificateRefs: []*kubernetes.KubernetesGatewayApiSecretObjectReference{{Name: literal("app-tls")}},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with an invalid allowed_routes from value", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners[0].AllowedRoutes = &kubernetes.KubernetesGatewayApiAllowedRoutes{
					Namespaces: &kubernetes.KubernetesGatewayApiRouteNamespaces{From: stringPtr("Everywhere")},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with an invalid label selector operator", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners[0].AllowedRoutes = &kubernetes.KubernetesGatewayApiAllowedRoutes{
					Namespaces: &kubernetes.KubernetesGatewayApiRouteNamespaces{
						From: stringPtr("Selector"),
						Selector: &kubernetes.KubernetesGatewayApiLabelSelector{
							MatchExpressions: []*kubernetes.KubernetesGatewayApiLabelSelectorRequirement{
								{Key: "tier", Operator: "Contains", Values: []string{"web"}},
							},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with an invalid listener name pattern", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners[0].Name = "Bad_Name"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with a listener port above the valid range", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners[0].Port = 70000
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})
	})
})
