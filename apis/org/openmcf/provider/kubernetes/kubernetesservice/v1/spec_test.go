package kubernetesservicev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestKubernetesServiceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesServiceSpec Validation Suite")
}

var _ = ginkgo.Describe("KubernetesServiceSpec validations", func() {

	// Helper to create a pointer to a service type enum value.
	svcType := func(t KubernetesServiceSpec_KubernetesServiceType) *KubernetesServiceSpec_KubernetesServiceType {
		return &t
	}

	// Helper to create a pointer to an external traffic policy enum value.
	extPolicy := func(p KubernetesServiceSpec_KubernetesServiceExternalTrafficPolicy) *KubernetesServiceSpec_KubernetesServiceExternalTrafficPolicy {
		return &p
	}

	// Helper to create a pointer to a session affinity enum value.
	sessAffinity := func(s KubernetesServiceSpec_KubernetesServiceSessionAffinity) *KubernetesServiceSpec_KubernetesServiceSessionAffinity {
		return &s
	}

	// Helper to create a pointer to a protocol enum value.
	proto := func(p KubernetesServicePort_KubernetesServiceProtocol) *KubernetesServicePort_KubernetesServiceProtocol {
		return &p
	}

	ginkgo.Context("When valid specs are provided", func() {

		ginkgo.It("accepts a minimal ClusterIP service", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "default",
				Name:      "my-service",
				Selector:  map[string]string{"app": "web"},
				Ports: []*KubernetesServicePort{
					{Port: 80},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a ClusterIP service with explicit type", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "production",
				Name:      "api-gateway",
				Type:      svcType(KubernetesServiceSpec_cluster_ip),
				Selector:  map[string]string{"app": "api", "tier": "frontend"},
				Ports: []*KubernetesServicePort{
					{Name: "http", Port: 80, TargetPort: "8080", Protocol: proto(KubernetesServicePort_TCP)},
					{Name: "https", Port: 443, TargetPort: "8443", Protocol: proto(KubernetesServicePort_TCP)},
				},
				Labels:      map[string]string{"team": "platform"},
				Annotations: map[string]string{"description": "API gateway service"},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a NodePort service", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "staging",
				Name:      "web-nodeport",
				Type:      svcType(KubernetesServiceSpec_node_port),
				Selector:  map[string]string{"app": "web"},
				Ports: []*KubernetesServicePort{
					{Port: 80, TargetPort: "8080", NodePort: 30080},
				},
				ExternalTrafficPolicy: extPolicy(KubernetesServiceSpec_local),
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a LoadBalancer service", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "production",
				Name:      "web-lb",
				Type:      svcType(KubernetesServiceSpec_load_balancer),
				Selector:  map[string]string{"app": "web"},
				Ports: []*KubernetesServicePort{
					{Name: "http", Port: 80, TargetPort: "8080"},
					{Name: "https", Port: 443, TargetPort: "8443"},
				},
				ExternalTrafficPolicy:    extPolicy(KubernetesServiceSpec_cluster),
				SessionAffinity:          sessAffinity(KubernetesServiceSpec_client_ip),
				LoadBalancerSourceRanges: []string{"203.0.113.0/24", "10.0.0.0/8"},
				Annotations: map[string]string{
					"service.beta.kubernetes.io/aws-load-balancer-type": "nlb",
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts an ExternalName service", func() {
			spec := &KubernetesServiceSpec{
				Namespace:       "production",
				Name:            "external-db",
				Type:            svcType(KubernetesServiceSpec_external_name),
				ExternalDnsName: "my-database.us-east-1.rds.amazonaws.com",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a headless service", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "production",
				Name:      "statefulset-headless",
				Headless:  true,
				Selector:  map[string]string{"app": "cassandra"},
				Ports: []*KubernetesServicePort{
					{Port: 9042, TargetPort: "cql"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a headless service with explicit ClusterIP type", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "data",
				Name:      "etcd-headless",
				Type:      svcType(KubernetesServiceSpec_cluster_ip),
				Headless:  true,
				Selector:  map[string]string{"app": "etcd"},
				Ports: []*KubernetesServicePort{
					{Name: "client", Port: 2379},
					{Name: "peer", Port: 2380},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a service with UDP protocol", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "kube-system",
				Name:      "dns-service",
				Selector:  map[string]string{"app": "coredns"},
				Ports: []*KubernetesServicePort{
					{Name: "dns-udp", Port: 53, Protocol: proto(KubernetesServicePort_UDP)},
					{Name: "dns-tcp", Port: 53, Protocol: proto(KubernetesServicePort_TCP)},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a NodePort service with auto-allocated node port", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "staging",
				Name:      "web-auto-nodeport",
				Type:      svcType(KubernetesServiceSpec_node_port),
				Selector:  map[string]string{"app": "web"},
				Ports: []*KubernetesServicePort{
					{Port: 80, NodePort: 0},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a service with named target port", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "default",
				Name:      "named-port-svc",
				Selector:  map[string]string{"app": "web"},
				Ports: []*KubernetesServicePort{
					{Port: 80, TargetPort: "http"},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("When invalid specs are provided", func() {

		ginkgo.It("rejects empty service name", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "default",
				Name:      "",
				Ports: []*KubernetesServicePort{
					{Port: 80},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects empty namespace", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "",
				Name:      "my-service",
				Ports: []*KubernetesServicePort{
					{Port: 80},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects service name with uppercase letters", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "default",
				Name:      "MyService",
				Ports: []*KubernetesServicePort{
					{Port: 80},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects service name with underscores", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "default",
				Name:      "my_service",
				Ports: []*KubernetesServicePort{
					{Port: 80},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects service name starting with hyphen", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "default",
				Name:      "-my-service",
				Ports: []*KubernetesServicePort{
					{Port: 80},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects service name ending with hyphen", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "default",
				Name:      "my-service-",
				Ports: []*KubernetesServicePort{
					{Port: 80},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects service name longer than 63 characters", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "default",
				Name:      "this-is-a-very-long-service-name-that-exceeds-sixty-three-characters-limit",
				Ports: []*KubernetesServicePort{
					{Port: 80},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects namespace with uppercase letters", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "Default",
				Name:      "my-service",
				Ports: []*KubernetesServicePort{
					{Port: 80},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects port number 0", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "default",
				Name:      "my-service",
				Ports: []*KubernetesServicePort{
					{Port: 0},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects port number exceeding 65535", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "default",
				Name:      "my-service",
				Ports: []*KubernetesServicePort{
					{Port: 70000},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects node port outside valid range", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "default",
				Name:      "my-service",
				Type:      svcType(KubernetesServiceSpec_node_port),
				Selector:  map[string]string{"app": "web"},
				Ports: []*KubernetesServicePort{
					{Port: 80, NodePort: 8080},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects node port just below valid range", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "default",
				Name:      "my-service",
				Type:      svcType(KubernetesServiceSpec_node_port),
				Selector:  map[string]string{"app": "web"},
				Ports: []*KubernetesServicePort{
					{Port: 80, NodePort: 29999},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects node port just above valid range", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "default",
				Name:      "my-service",
				Type:      svcType(KubernetesServiceSpec_node_port),
				Selector:  map[string]string{"app": "web"},
				Ports: []*KubernetesServicePort{
					{Port: 80, NodePort: 32768},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects ExternalName type without external_dns_name", func() {
			spec := &KubernetesServiceSpec{
				Namespace:       "default",
				Name:            "external-svc",
				Type:            svcType(KubernetesServiceSpec_external_name),
				ExternalDnsName: "",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects headless with NodePort type", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "default",
				Name:      "headless-nodeport",
				Type:      svcType(KubernetesServiceSpec_node_port),
				Headless:  true,
				Selector:  map[string]string{"app": "web"},
				Ports: []*KubernetesServicePort{
					{Port: 80},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects headless with LoadBalancer type", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "default",
				Name:      "headless-lb",
				Type:      svcType(KubernetesServiceSpec_load_balancer),
				Headless:  true,
				Selector:  map[string]string{"app": "web"},
				Ports: []*KubernetesServicePort{
					{Port: 80},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects non-ExternalName service without ports", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "default",
				Name:      "no-ports",
				Type:      svcType(KubernetesServiceSpec_cluster_ip),
				Selector:  map[string]string{"app": "web"},
				Ports:     []*KubernetesServicePort{},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects ClusterIP service with no ports (default type)", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "default",
				Name:      "no-ports-default",
				Selector:  map[string]string{"app": "web"},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects port name longer than 63 characters", func() {
			spec := &KubernetesServiceSpec{
				Namespace: "default",
				Name:      "long-port-name",
				Selector:  map[string]string{"app": "web"},
				Ports: []*KubernetesServicePort{
					{Name: "this-is-a-very-long-port-name-that-exceeds-the-sixty-three-character-limit", Port: 80},
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
