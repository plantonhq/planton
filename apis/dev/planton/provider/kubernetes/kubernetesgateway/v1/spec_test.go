package kubernetesgatewayv1

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

func TestKubernetesGateway(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesGateway Suite")
}

func stringPtr(s string) *string { return &s }

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

// httpListener returns a minimal valid HTTP listener.
func httpListener(name string, port int32) *KubernetesGatewayListener {
	return &KubernetesGatewayListener{
		Name:     name,
		Port:     port,
		Protocol: "HTTP",
	}
}

var _ = ginkgo.Describe("KubernetesGateway Validation Tests", func() {
	var input *KubernetesGateway

	ginkgo.BeforeEach(func() {
		input = &KubernetesGateway{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesGateway",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-gateway",
			},
			Spec: &KubernetesGatewaySpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace:        literal("istio-ingress"),
				GatewayClassName: literal("istio"),
				Listeners: []*KubernetesGatewayListener{
					httpListener("http", 80),
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

		ginkgo.Context("with an HTTPS listener terminating TLS via a certificate ref", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Listeners = []*KubernetesGatewayListener{
					{
						Name:     "https",
						Hostname: stringPtr("app.example.com"),
						Port:     443,
						Protocol: "HTTPS",
						Tls: &KubernetesGatewayListenerTlsConfig{
							Mode: stringPtr("Terminate"),
							CertificateRefs: []*kubernetes.KubernetesGatewayApiSecretObjectReference{
								{Name: "app-tls"},
							},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a TLS passthrough listener (no certificate required)", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Listeners = []*KubernetesGatewayListener{
					{
						Name:     "tls",
						Hostname: stringPtr("passthrough.example.com"),
						Port:     443,
						Protocol: "TLS",
						Tls:      &KubernetesGatewayListenerTlsConfig{Mode: stringPtr("Passthrough")},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with multiple distinct protocol listeners", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Listeners = []*KubernetesGatewayListener{
					httpListener("http", 80),
					{
						Name:     "https",
						Port:     443,
						Protocol: "HTTPS",
						Tls: &KubernetesGatewayListenerTlsConfig{
							Mode:            stringPtr("Terminate"),
							CertificateRefs: []*kubernetes.KubernetesGatewayApiSecretObjectReference{{Name: "app-tls"}},
						},
					},
					{Name: "postgres", Port: 5432, Protocol: "TCP"},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with namespace and gateway_class_name resolved via valueFrom references", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Namespace = valueFrom(cloudresourcekind.CloudResourceKind_KubernetesNamespace, "ingress-ns", "spec.name")
				input.Spec.GatewayClassName = valueFrom(cloudresourcekind.CloudResourceKind_KubernetesGatewayClass, "istio", "status.outputs.gateway_class_name")
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with allowed_routes restricting kinds and namespaces", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Listeners[0].AllowedRoutes = &KubernetesGatewayAllowedRoutes{
					Namespaces: &KubernetesGatewayRouteNamespaces{From: stringPtr("Same")},
					Kinds:      []*KubernetesGatewayRouteGroupKind{{Kind: "HTTPRoute"}},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a label selector for route namespaces", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Listeners[0].AllowedRoutes = &KubernetesGatewayAllowedRoutes{
					Namespaces: &KubernetesGatewayRouteNamespaces{
						From: stringPtr("Selector"),
						Selector: &KubernetesGatewayLabelSelector{
							MatchLabels: map[string]string{"team": "platform"},
							MatchExpressions: []*KubernetesGatewayLabelSelectorRequirement{
								{Key: "tier", Operator: "In", Values: []string{"web", "api"}},
							},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with requested addresses and infrastructure metadata", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Addresses = []*KubernetesGatewayAddress{
					{Type: stringPtr("IPAddress"), Value: "203.0.113.10"},
				}
				input.Spec.Infrastructure = &KubernetesGatewayInfrastructure{
					Labels:      map[string]string{"app.kubernetes.io/part-of": "leftbin"},
					Annotations: map[string]string{"example.com/team": "platform"},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with gateway-level frontend mutual TLS validation", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Tls = &KubernetesGatewayTlsConfig{
					Frontend: &KubernetesGatewayFrontendTlsConfig{
						Default: &KubernetesGatewayFrontendTlsValidationConfig{
							Validation: &KubernetesGatewayFrontendTlsValidation{
								CaCertificateRefs: []*kubernetes.KubernetesGatewayApiObjectReference{
									{Group: "", Kind: "ConfigMap", Name: "client-ca"},
								},
								Mode: stringPtr("AllowValidOnly"),
							},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with allowed_listeners restricting ListenerSet attachment", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.AllowedListeners = &KubernetesGatewayAllowedListeners{
					Namespaces: &KubernetesGatewayListenerNamespaces{From: stringPtr("None")},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("with no listeners", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with tls set on an HTTP listener", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners[0].Tls = &KubernetesGatewayListenerTlsConfig{Mode: stringPtr("Terminate")}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with a TLS listener missing its tls mode", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners = []*KubernetesGatewayListener{
					{Name: "tls", Port: 443, Protocol: "TLS"},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with an HTTPS listener set to Passthrough mode", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners = []*KubernetesGatewayListener{
					{
						Name:     "https",
						Port:     443,
						Protocol: "HTTPS",
						Tls:      &KubernetesGatewayListenerTlsConfig{Mode: stringPtr("Passthrough")},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with a hostname on a TCP listener", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners = []*KubernetesGatewayListener{
					{Name: "tcp", Hostname: stringPtr("nope.example.com"), Port: 5432, Protocol: "TCP"},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with duplicate listener names", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners = []*KubernetesGatewayListener{
					httpListener("dup", 80),
					httpListener("dup", 81),
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with a duplicate port/protocol/hostname combination", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners = []*KubernetesGatewayListener{
					httpListener("first", 80),
					httpListener("second", 80),
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with an invalid protocol value", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners[0].Protocol = "not a protocol"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with a listener port above the valid range", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners[0].Port = 70000
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with a Terminate listener missing both certificate_refs and options", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners = []*KubernetesGatewayListener{
					{
						Name:     "https",
						Port:     443,
						Protocol: "HTTPS",
						Tls:      &KubernetesGatewayListenerTlsConfig{Mode: stringPtr("Terminate")},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with an invalid listener tls mode", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners = []*KubernetesGatewayListener{
					{
						Name:     "https",
						Port:     443,
						Protocol: "HTTPS",
						Tls: &KubernetesGatewayListenerTlsConfig{
							Mode:            stringPtr("Reterminate"),
							CertificateRefs: []*kubernetes.KubernetesGatewayApiSecretObjectReference{{Name: "app-tls"}},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with an invalid address type", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Addresses = []*KubernetesGatewayAddress{
					{Type: stringPtr("Bad Type"), Value: "203.0.113.10"},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with an invalid label selector operator", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners[0].AllowedRoutes = &KubernetesGatewayAllowedRoutes{
					Namespaces: &KubernetesGatewayRouteNamespaces{
						From: stringPtr("Selector"),
						Selector: &KubernetesGatewayLabelSelector{
							MatchExpressions: []*KubernetesGatewayLabelSelectorRequirement{
								{Key: "tier", Operator: "Contains", Values: []string{"web"}},
							},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with frontend TLS validation but no CA certificate refs", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Tls = &KubernetesGatewayTlsConfig{
					Frontend: &KubernetesGatewayFrontendTlsConfig{
						Default: &KubernetesGatewayFrontendTlsValidationConfig{
							Validation: &KubernetesGatewayFrontendTlsValidation{},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with duplicate per-port frontend TLS configurations", func() {
			ginkgo.It("should return a validation error", func() {
				validation := &KubernetesGatewayFrontendTlsValidationConfig{
					Validation: &KubernetesGatewayFrontendTlsValidation{
						CaCertificateRefs: []*kubernetes.KubernetesGatewayApiObjectReference{
							{Group: "", Kind: "ConfigMap", Name: "client-ca"},
						},
					},
				}
				input.Spec.Tls = &KubernetesGatewayTlsConfig{
					Frontend: &KubernetesGatewayFrontendTlsConfig{
						Default: validation,
						PerPort: []*KubernetesGatewayTlsPortConfig{
							{Port: 443, Tls: validation},
							{Port: 443, Tls: validation},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with an empty gateway_class_name", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.GatewayClassName = literal("")
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with an empty namespace", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Namespace = literal("")
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with a certificate ref whose group is malformed", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners = []*KubernetesGatewayListener{
					{
						Name:     "https",
						Port:     443,
						Protocol: "HTTPS",
						Tls: &KubernetesGatewayListenerTlsConfig{
							Mode:            stringPtr("Terminate"),
							CertificateRefs: []*kubernetes.KubernetesGatewayApiSecretObjectReference{{Name: "app-tls", Group: stringPtr("Bad_Group")}},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with a certificate ref whose kind is malformed", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Listeners = []*KubernetesGatewayListener{
					{
						Name:     "https",
						Port:     443,
						Protocol: "HTTPS",
						Tls: &KubernetesGatewayListenerTlsConfig{
							Mode:            stringPtr("Terminate"),
							CertificateRefs: []*kubernetes.KubernetesGatewayApiSecretObjectReference{{Name: "app-tls", Kind: stringPtr("bad/kind")}},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with a CA certificate ref whose kind is missing (now required)", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Tls = &KubernetesGatewayTlsConfig{
					Frontend: &KubernetesGatewayFrontendTlsConfig{
						Default: &KubernetesGatewayFrontendTlsValidationConfig{
							Validation: &KubernetesGatewayFrontendTlsValidation{
								CaCertificateRefs: []*kubernetes.KubernetesGatewayApiObjectReference{
									{Group: "", Name: "client-ca"},
								},
							},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with a CA certificate ref whose group is malformed", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Tls = &KubernetesGatewayTlsConfig{
					Frontend: &KubernetesGatewayFrontendTlsConfig{
						Default: &KubernetesGatewayFrontendTlsValidationConfig{
							Validation: &KubernetesGatewayFrontendTlsValidation{
								CaCertificateRefs: []*kubernetes.KubernetesGatewayApiObjectReference{
									{Group: "Bad_Group", Kind: "ConfigMap", Name: "client-ca"},
								},
							},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("missing spec", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("missing metadata", func() {
			ginkgo.It("should return a validation error", func() {
				input.Metadata = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})
	})
})
