package kubernetesciliumv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesCilium(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesCilium Suite")
}

func int32Ptr(i int32) *int32    { return &i }
func boolPtr(b bool) *bool       { return &b }
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

var _ = ginkgo.Describe("KubernetesCilium Validation Tests", func() {
	var input *KubernetesCilium

	ginkgo.BeforeEach(func() {
		input = &KubernetesCilium{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesCilium",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-cilium",
			},
			Spec: &KubernetesCiliumSpec{
				Namespace: literal("kube-system"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec should not return a validation error", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "kube-system", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("kind-style posture (kubernetes IPAM + kube-proxy replacement) should be valid", func() {
			input.Spec.Ipam = &KubernetesCiliumIpam{Mode: stringPtr("kubernetes")}
			input.Spec.KubeProxyReplacement = true
			input.Spec.K8SServiceHost = "planton-e2e-cilium-control-plane"
			input.Spec.K8SServicePort = int32Ptr(6443)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("EKS chaining posture (aws-cni, non-exclusive) should be valid", func() {
			input.Spec.Cni = &KubernetesCiliumCni{
				ChainingMode: stringPtr("aws-cni"),
				Exclusive:    boolPtr(false),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("generic-veth chaining with a target network should be valid", func() {
			input.Spec.Cni = &KubernetesCiliumCni{
				ChainingMode:   stringPtr("generic-veth"),
				ChainingTarget: "incumbent-cni",
				Exclusive:      boolPtr(false),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("AWS ENI primary-CNI posture should be valid", func() {
			input.Spec.Ipam = &KubernetesCiliumIpam{Mode: stringPtr("eni")}
			input.Spec.Cloud = &KubernetesCiliumCloudIntegration{AwsEni: true}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("native routing with CIDR and direct node routes should be valid", func() {
			input.Spec.Routing = &KubernetesCiliumRouting{
				Mode:                  stringPtr("native"),
				Ipv4NativeRoutingCidr: "10.0.0.0/8",
				AutoDirectNodeRoutes:  true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("full hubble stack should be valid", func() {
			input.Spec.Hubble = &KubernetesCiliumHubble{
				Enabled: boolPtr(true),
				Relay:   true,
				Ui:      true,
				Metrics: []string{"dns:query;ignoreAAAA", "drop", "http"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("wireguard encryption with node encryption should be valid", func() {
			input.Spec.Encryption = &KubernetesCiliumEncryption{
				Enabled:        true,
				Type:           stringPtr("wireguard"),
				NodeEncryption: true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("gateway api with kube-proxy replacement should be valid", func() {
			input.Spec.KubeProxyReplacement = true
			input.Spec.K8SServiceHost = "api.cluster.example.com"
			input.Spec.GatewayApi = true
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("cluster-pool IPAM tuning should be valid", func() {
			input.Spec.Ipam = &KubernetesCiliumIpam{
				Mode:                    stringPtr("cluster-pool"),
				ClusterPoolIpv4PodCidrs: []string{"10.42.0.0/16"},
				ClusterPoolIpv4MaskSize: int32Ptr(25),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("missing namespace should fail", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("gateway_api without kube_proxy_replacement should fail", func() {
			input.Spec.GatewayApi = true
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("kube_proxy_replacement without k8s_service_host should fail", func() {
			input.Spec.KubeProxyReplacement = true
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("chaining with exclusive left true should fail", func() {
			input.Spec.Cni = &KubernetesCiliumCni{ChainingMode: stringPtr("aws-cni")}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("unknown chaining mode should fail (closed enum)", func() {
			input.Spec.Cni = &KubernetesCiliumCni{ChainingMode: stringPtr("calico"), Exclusive: boolPtr(false)}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("unknown IPAM mode should fail (closed enum)", func() {
			input.Spec.Ipam = &KubernetesCiliumIpam{Mode: stringPtr("static")}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("malformed cluster-pool CIDR should fail", func() {
			input.Spec.Ipam = &KubernetesCiliumIpam{ClusterPoolIpv4PodCidrs: []string{"10.42.0.0"}}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("out-of-range cluster-pool mask size should fail", func() {
			input.Spec.Ipam = &KubernetesCiliumIpam{ClusterPoolIpv4MaskSize: int32Ptr(31)}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("unknown routing mode should fail (closed enum)", func() {
			input.Spec.Routing = &KubernetesCiliumRouting{Mode: stringPtr("hybrid")}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("native-routing knobs in tunnel mode should fail", func() {
			input.Spec.Routing = &KubernetesCiliumRouting{
				Mode:                  stringPtr("tunnel"),
				Ipv4NativeRoutingCidr: "10.0.0.0/8",
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("malformed native routing CIDR should fail", func() {
			input.Spec.Routing = &KubernetesCiliumRouting{
				Mode:                  stringPtr("native"),
				Ipv4NativeRoutingCidr: "not-a-cidr",
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("unknown tunnel protocol should fail (closed enum)", func() {
			input.Spec.Routing = &KubernetesCiliumRouting{TunnelProtocol: stringPtr("gre")}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("two cloud integrations at once should fail", func() {
			input.Spec.Cloud = &KubernetesCiliumCloudIntegration{AwsEni: true, Gke: true}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("hubble ui without relay should fail", func() {
			input.Spec.Hubble = &KubernetesCiliumHubble{Ui: true}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("hubble components with hubble disabled should fail", func() {
			input.Spec.Hubble = &KubernetesCiliumHubble{Enabled: boolPtr(false), Relay: true}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("hubble metrics service monitor without metrics should fail", func() {
			input.Spec.Hubble = &KubernetesCiliumHubble{MetricsServiceMonitor: true}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("unknown encryption type should fail (closed enum)", func() {
			input.Spec.Encryption = &KubernetesCiliumEncryption{Enabled: true, Type: stringPtr("tls")}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("node encryption with ipsec should fail (wireguard only)", func() {
			input.Spec.Encryption = &KubernetesCiliumEncryption{
				Enabled:        true,
				Type:           stringPtr("ipsec"),
				NodeEncryption: true,
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("node encryption without enabled should fail", func() {
			input.Spec.Encryption = &KubernetesCiliumEncryption{NodeEncryption: true, Type: stringPtr("wireguard")}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("explicit type with enabled false should be tolerated (defaulting middleware shape)", func() {
			input.Spec.Encryption = &KubernetesCiliumEncryption{Type: stringPtr("wireguard")}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("bbr without bandwidth manager enabled should fail", func() {
			input.Spec.BandwidthManager = &KubernetesCiliumBandwidthManager{Bbr: true}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("unknown policy enforcement mode should fail (closed enum)", func() {
			input.Spec.PolicyEnforcementMode = stringPtr("strict")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("zero operator replicas should fail (gte 1)", func() {
			input.Spec.Operator = &KubernetesCiliumOperator{Replicas: int32Ptr(0)}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("prometheus service monitor without enabled should fail", func() {
			input.Spec.Prometheus = &KubernetesCiliumPrometheus{ServiceMonitor: true}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("uppercase cluster name should fail the pattern", func() {
			input.Spec.ClusterName = stringPtr("Prod-East")
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("out-of-range k8s service port should fail", func() {
			input.Spec.K8SServicePort = int32Ptr(70000)
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
