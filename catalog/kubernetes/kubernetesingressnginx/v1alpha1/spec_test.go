package kubernetesingressnginxv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesIngressNginx(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesIngressNginx Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func int32p(v int32) *int32 { return &v }

var _ = ginkgo.Describe("KubernetesIngressNginx Validation Tests", func() {
	var input *KubernetesIngressNginx

	ginkgo.BeforeEach(func() {
		input = &KubernetesIngressNginx{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesIngressNginx",
			Metadata:   &shared.CatalogObjectMetadata{Name: "ingress-nginx"},
			Spec: &KubernetesIngressNginxSpec{
				Namespace:       literal("ingress-nginx"),
				CreateNamespace: true,
			},
		}
	})

	ginkgo.Describe("valid configurations", func() {
		ginkgo.It("accepts a minimal install", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts a second controller with its own class", func() {
			input.Spec.IngressClass = &KubernetesIngressNginxIngressClass{
				Name: func() *string { s := "nginx-internal"; return &s }(),
			}
			input.Spec.Service = &KubernetesIngressNginxService{
				Annotations: map[string]string{
					"service.beta.kubernetes.io/aws-load-balancer-scheme": "internal",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts autoscaling with a valid range", func() {
			input.Spec.Autoscaling = &KubernetesIngressNginxAutoscaling{
				Enabled:     true,
				MinReplicas: int32p(2),
				MaxReplicas: int32p(10),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts the internal dual-LB service with annotations", func() {
			input.Spec.Service = &KubernetesIngressNginxService{
				Internal: &KubernetesIngressNginxInternalService{
					Enabled: true,
					Annotations: map[string]string{
						"service.beta.kubernetes.io/azure-load-balancer-internal": "true",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts host networking on its own", func() {
			input.Spec.HostNetwork = true
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts host ports on their own", func() {
			input.Spec.HostPorts = true
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts TCP/UDP service exposure in the documented format", func() {
			input.Spec.TcpServices = map[string]string{"5432": "default/postgres:5432"}
			input.Spec.UdpServices = map[string]string{"53": "kube-system/kube-dns:53"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts a TCP target with PROXY protocol suffix", func() {
			input.Spec.TcpServices = map[string]string{"5432": "default/postgres:5432::PROXY"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts load balancer source ranges as CIDRs", func() {
			input.Spec.Service = &KubernetesIngressNginxService{
				LoadBalancerSourceRanges: []string{"10.0.0.0/8", "203.0.113.4/32"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts a default TLS certificate reference", func() {
			input.Spec.DefaultTlsCertificate = &KubernetesIngressNginxDefaultTlsCertificate{
				SecretName: literal("wildcard-tls"),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})

		ginkgo.It("accepts metrics with a ServiceMonitor", func() {
			input.Spec.Metrics = &KubernetesIngressNginxMetrics{
				Enabled:        true,
				ServiceMonitor: true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.Succeed())
		})
	})

	ginkgo.Describe("required fields and contracts", func() {
		ginkgo.It("rejects a missing namespace", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an empty namespace value", func() {
			input.Spec.Namespace = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: ""},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects host_network combined with host_ports", func() {
			input.Spec.HostNetwork = true
			input.Spec.HostPorts = true
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects autoscaling min above max", func() {
			input.Spec.Autoscaling = &KubernetesIngressNginxAutoscaling{
				Enabled:     true,
				MinReplicas: int32p(5),
				MaxReplicas: int32p(2),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an enabled internal service without annotations", func() {
			input.Spec.Service = &KubernetesIngressNginxService{
				Internal: &KubernetesIngressNginxInternalService{Enabled: true},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an invalid ingress class name", func() {
			bad := "Nginx_Internal"
			input.Spec.IngressClass = &KubernetesIngressNginxIngressClass{Name: &bad}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a non-numeric TCP service port key", func() {
			input.Spec.TcpServices = map[string]string{"postgres": "default/postgres:5432"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an out-of-range TCP service port key", func() {
			input.Spec.TcpServices = map[string]string{"70000": "default/postgres:5432"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a malformed TCP service target", func() {
			input.Spec.TcpServices = map[string]string{"5432": "postgres:5432"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a malformed UDP service target", func() {
			input.Spec.UdpServices = map[string]string{"53": "kube-dns"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a non-CIDR load balancer source range", func() {
			input.Spec.Service = &KubernetesIngressNginxService{
				LoadBalancerSourceRanges: []string{"10.0.0.0"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a ServiceMonitor without metrics enabled", func() {
			input.Spec.Metrics = &KubernetesIngressNginxMetrics{
				ServiceMonitor: true,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a default TLS certificate without a secret name", func() {
			input.Spec.DefaultTlsCertificate = &KubernetesIngressNginxDefaultTlsCertificate{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects zero replicas", func() {
			input.Spec.Replicas = int32p(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects a webhook timeout above the API server maximum", func() {
			input.Spec.AdmissionWebhooks = &KubernetesIngressNginxAdmissionWebhooks{
				TimeoutSeconds: int32p(31),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})

		ginkgo.It("rejects an out-of-range autoscaling CPU target", func() {
			input.Spec.Autoscaling = &KubernetesIngressNginxAutoscaling{
				Enabled:                     true,
				TargetCpuUtilizationPercent: int32p(150),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.Succeed())
		})
	})
})
