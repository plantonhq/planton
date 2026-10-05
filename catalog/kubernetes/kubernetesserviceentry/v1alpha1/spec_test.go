package kubernetesserviceentryv1alpha1

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

func TestKubernetesServiceEntry(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesServiceEntry Suite")
}

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

// ptr returns a pointer to v, for setting proto3 `optional` scalar fields.
func ptr[T any](v T) *T { return &v }

// httpsPort returns a minimal valid HTTPS service port.
func httpsPort() *KubernetesServiceEntryPort {
	return &KubernetesServiceEntryPort{Number: 443, Name: "https", Protocol: ptr("TLS")}
}

var _ = ginkgo.Describe("KubernetesServiceEntry DYNAMIC_DNS rules", func() {
	var input *KubernetesServiceEntry

	// A valid DYNAMIC_DNS entry: wildcard hosts, HTTP-family ports, no
	// addresses, no endpoints (mirrors the istiod webhook's contract).
	ginkgo.BeforeEach(func() {
		input = &KubernetesServiceEntry{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesServiceEntry",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "dynamic-dns-egress",
			},
			Spec: &KubernetesServiceEntrySpec{
				Namespace:  literal("default"),
				Hosts:      []string{"*.example.com"},
				Resolution: ptr("DYNAMIC_DNS"),
				Location:   ptr("MESH_EXTERNAL"),
				Ports: []*KubernetesServiceEntryPort{
					{Number: 443, Name: "https", Protocol: ptr("TLS")},
				},
			},
		}
	})

	ginkgo.It("accepts the valid dynamic-DNS forward-proxy shape", func() {
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
	})

	ginkgo.It("rejects a non-wildcard host", func() {
		input.Spec.Hosts = []string{"api.example.com"}
		gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects addresses — destinations derive from the wildcard hosts", func() {
		input.Spec.Addresses = []string{"240.0.0.1"}
		gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects endpoints — destination IPs resolve from the requested hostname", func() {
		input.Spec.Endpoints = []*KubernetesServiceEntryEndpoint{
			{Address: ptr("10.0.0.1")},
		}
		gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a TCP port — only HTTP, TLS, GRPC, and HTTP2 are supported", func() {
		input.Spec.Ports = []*KubernetesServiceEntryPort{
			{Number: 5432, Name: "postgres", Protocol: ptr("TCP")},
		}
		gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a port with no protocol (defaults to TCP semantics)", func() {
		input.Spec.Ports = []*KubernetesServiceEntryPort{
			{Number: 443, Name: "https"},
		}
		gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
	})
})

var _ = ginkgo.Describe("KubernetesServiceEntry Validation Tests", func() {
	var input *KubernetesServiceEntry

	ginkgo.BeforeEach(func() {
		input = &KubernetesServiceEntry{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesServiceEntry",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-service-entry",
			},
			Spec: &KubernetesServiceEntrySpec{
				Namespace:  literal("default"),
				Hosts:      []string{"api.external-example.com"},
				Location:   ptr("MESH_EXTERNAL"),
				Resolution: ptr("DNS"),
				Ports:      []*KubernetesServiceEntryPort{httpsPort()},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("with a MESH_EXTERNAL DNS external API", func() {
			ginkgo.It("should not return a validation error", func() {
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with location and resolution omitted (upstream defaults)", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Location = nil
				input.Spec.Resolution = nil
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with STATIC resolution and static endpoints + addresses", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Resolution = ptr("STATIC")
				input.Spec.Addresses = []string{"10.0.0.0/24"}
				input.Spec.Endpoints = []*KubernetesServiceEntryEndpoint{
					{Address: ptr("10.0.0.5"), Ports: map[string]uint32{"https": 8443}},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a MESH_INTERNAL workload selector", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Location = ptr("MESH_INTERNAL")
				input.Spec.Resolution = ptr("STATIC")
				input.Spec.WorkloadSelector = &kubernetes.KubernetesIstioApiNetworkingWorkloadSelector{
					Labels: map[string]string{"app": "legacy-vm"},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a single DNS_ROUND_ROBIN endpoint", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Resolution = ptr("DNS_ROUND_ROBIN")
				input.Spec.Endpoints = []*KubernetesServiceEntryEndpoint{
					{Address: ptr("api.external-example.com")},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a unix domain socket endpoint (no ports)", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Resolution = ptr("STATIC")
				input.Spec.Endpoints = []*KubernetesServiceEntryEndpoint{
					{Address: ptr("unix:///var/run/app/app.sock")},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with an endpoint identified only by network", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Resolution = ptr("STATIC")
				input.Spec.Endpoints = []*KubernetesServiceEntryEndpoint{
					{Network: ptr("vpc-west"), Labels: map[string]string{"tier": "db"}},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with multiple distinctly-named/numbered ports", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Ports = []*KubernetesServiceEntryPort{
					{Number: 80, Name: "http", Protocol: ptr("HTTP")},
					{Number: 443, Name: "https", Protocol: ptr("HTTPS"), TargetPort: ptr(uint32(8443))},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with the namespace resolved via a valueFrom reference", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "finance-ns", "spec.name")
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
				input.Kind = "KubernetesDestinationRule"
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

	ginkgo.Describe("When hosts are invalid", func() {
		ginkgo.Context("with an empty hosts list", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Hosts = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a bare wildcard host", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Hosts = []string{"*"}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a partial wildcard host (allowed upstream)", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Hosts = []string{"*.external-example.com"}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When ports are invalid", func() {
		ginkgo.Context("with a duplicate port number", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Ports = []*KubernetesServiceEntryPort{
					{Number: 443, Name: "https", Protocol: ptr("HTTPS")},
					{Number: 443, Name: "tls", Protocol: ptr("TLS")},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a duplicate port name", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Ports = []*KubernetesServiceEntryPort{
					{Number: 443, Name: "https", Protocol: ptr("HTTPS")},
					{Number: 8443, Name: "https", Protocol: ptr("HTTPS")},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a port number out of range", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Ports = []*KubernetesServiceEntryPort{
					{Number: 70000, Name: "bad"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a missing port name", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Ports = []*KubernetesServiceEntryPort{
					{Number: 443},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with an unsupported protocol", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Ports = []*KubernetesServiceEntryPort{
					{Number: 443, Name: "https", Protocol: ptr("WEBSOCKET")},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a target_port out of range", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Ports = []*KubernetesServiceEntryPort{
					{Number: 443, Name: "https", TargetPort: ptr(uint32(70000))},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When location or resolution are invalid", func() {
		ginkgo.Context("with an unknown location", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Location = ptr("MESH_SIDECAR")
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with an unknown resolution", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Resolution = ptr("ROUND_ROBIN")
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When attachment and resolution rules are violated", func() {
		ginkgo.Context("with both workload_selector and endpoints set", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Resolution = ptr("STATIC")
				input.Spec.WorkloadSelector = &kubernetes.KubernetesIstioApiNetworkingWorkloadSelector{
					Labels: map[string]string{"app": "legacy-vm"},
				}
				input.Spec.Endpoints = []*KubernetesServiceEntryEndpoint{
					{Address: ptr("10.0.0.5")},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a CIDR address under DNS resolution", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Addresses = []string{"10.0.0.0/24"}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with endpoints under NONE resolution", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Resolution = ptr("NONE")
				input.Spec.Endpoints = []*KubernetesServiceEntryEndpoint{
					{Address: ptr("10.0.0.5")},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with multiple endpoints under DNS_ROUND_ROBIN", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Resolution = ptr("DNS_ROUND_ROBIN")
				input.Spec.Endpoints = []*KubernetesServiceEntryEndpoint{
					{Address: ptr("a.example.com")},
					{Address: ptr("b.example.com")},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When an endpoint is invalid", func() {
		ginkgo.Context("with neither an address nor a network", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Resolution = ptr("STATIC")
				input.Spec.Endpoints = []*KubernetesServiceEntryEndpoint{
					{Labels: map[string]string{"tier": "db"}},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a unix:// address carrying ports", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Resolution = ptr("STATIC")
				input.Spec.Endpoints = []*KubernetesServiceEntryEndpoint{
					{Address: ptr("unix:///var/run/app.sock"), Ports: map[string]uint32{"https": 8443}},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a unix:// address that is a directory", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Resolution = ptr("STATIC")
				input.Spec.Endpoints = []*KubernetesServiceEntryEndpoint{
					{Address: ptr("unix:///var/run/app/")},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When the workload selector labels are invalid", func() {
		ginkgo.Context("with a wildcard in a label value", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Location = ptr("MESH_INTERNAL")
				input.Spec.Resolution = ptr("STATIC")
				input.Spec.WorkloadSelector = &kubernetes.KubernetesIstioApiNetworkingWorkloadSelector{
					Labels: map[string]string{"app": "legacy-*"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})
