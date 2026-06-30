package ocinetworkloadbalancerv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestOciNetworkLoadBalancerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciNetworkLoadBalancerSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidNlb() *OciNetworkLoadBalancer {
	return &OciNetworkLoadBalancer{
		ApiVersion: "oci.planton.dev/v1",
		Kind:       "OciNetworkLoadBalancer",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-nlb",
		},
		Spec: &OciNetworkLoadBalancerSpec{
			CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
			SubnetId:      newStringValueOrRef("ocid1.subnet.oc1.iad.example"),
			BackendSets: []*OciNetworkLoadBalancerSpec_BackendSet{
				{
					Name:   "tcp-backend",
					Policy: OciNetworkLoadBalancerSpec_BackendSet_five_tuple,
					HealthChecker: &OciNetworkLoadBalancerSpec_HealthChecker{
						Protocol: OciNetworkLoadBalancerSpec_HealthChecker_tcp,
					},
				},
			},
			Listeners: []*OciNetworkLoadBalancerSpec_Listener{
				{
					Name:                  "tcp-listener",
					Port:                  80,
					Protocol:              OciNetworkLoadBalancerSpec_Listener_tcp,
					DefaultBackendSetName: "tcp-backend",
				},
			},
		},
	}
}

var _ = ginkgo.Describe("OciNetworkLoadBalancerSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_network_load_balancer", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := minimalValidNlb()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with display name set", func() {
				input := minimalValidNlb()
				input.Spec.DisplayName = "My Production NLB"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for private NLB", func() {
				input := minimalValidNlb()
				input.Spec.IsPrivate = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with source destination preservation", func() {
				input := minimalValidNlb()
				input.Spec.IsPreserveSourceDestination = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with symmetric hash enabled", func() {
				input := minimalValidNlb()
				input.Spec.IsPreserveSourceDestination = true
				input.Spec.IsSymmetricHashEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with NSGs", func() {
				input := minimalValidNlb()
				input.Spec.NetworkSecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
					newStringValueOrRef("ocid1.networksecuritygroup.oc1.iad.example1"),
					newStringValueOrRef("ocid1.networksecuritygroup.oc1.iad.example2"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with nlb_ip_version set", func() {
				input := minimalValidNlb()
				input.Spec.NlbIpVersion = "IPV4_AND_IPV6"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with reserved IPs", func() {
				input := minimalValidNlb()
				input.Spec.ReservedIps = []*OciNetworkLoadBalancerSpec_ReservedIp{
					{Id: "ocid1.publicip.oc1.iad.example"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with assigned IPv6", func() {
				input := minimalValidNlb()
				input.Spec.AssignedIpv6 = "2607:9b80:9a0a:9a7e:abcd:ef01:2345:6789"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with assigned private IPv4", func() {
				input := minimalValidNlb()
				input.Spec.AssignedPrivateIpv4 = "10.0.0.100"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with three_tuple policy", func() {
				input := minimalValidNlb()
				input.Spec.BackendSets[0].Policy = OciNetworkLoadBalancerSpec_BackendSet_three_tuple
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with two_tuple policy", func() {
				input := minimalValidNlb()
				input.Spec.BackendSets[0].Policy = OciNetworkLoadBalancerSpec_BackendSet_two_tuple
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with HTTP health checker", func() {
				input := minimalValidNlb()
				input.Spec.BackendSets[0].HealthChecker = &OciNetworkLoadBalancerSpec_HealthChecker{
					Protocol:   OciNetworkLoadBalancerSpec_HealthChecker_http,
					Port:       8080,
					UrlPath:    "/health",
					ReturnCode: 200,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with HTTPS health checker", func() {
				input := minimalValidNlb()
				input.Spec.BackendSets[0].HealthChecker = &OciNetworkLoadBalancerSpec_HealthChecker{
					Protocol:   OciNetworkLoadBalancerSpec_HealthChecker_https,
					Port:       443,
					UrlPath:    "/ready",
					ReturnCode: 200,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with UDP health checker and probe data", func() {
				input := minimalValidNlb()
				input.Spec.BackendSets[0].HealthChecker = &OciNetworkLoadBalancerSpec_HealthChecker{
					Protocol:     OciNetworkLoadBalancerSpec_HealthChecker_udp,
					Port:         5353,
					RequestData:  "AQAAAQAAAAAAAAZ",
					ResponseData: "AQAAAQABAAAAAA==",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with DNS health checker", func() {
				input := minimalValidNlb()
				input.Spec.BackendSets[0].HealthChecker = &OciNetworkLoadBalancerSpec_HealthChecker{
					Protocol: OciNetworkLoadBalancerSpec_HealthChecker_dns,
					Port:     53,
					DnsHealthCheck: &OciNetworkLoadBalancerSpec_DnsHealthCheck{
						DomainName:        "health.example.com",
						QueryClass:        "IN",
						QueryType:         "A",
						Rcodes:            []string{"NOERROR"},
						TransportProtocol: "UDP",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with health checker custom intervals", func() {
				input := minimalValidNlb()
				input.Spec.BackendSets[0].HealthChecker = &OciNetworkLoadBalancerSpec_HealthChecker{
					Protocol:          OciNetworkLoadBalancerSpec_HealthChecker_http,
					Port:              8080,
					UrlPath:           "/health",
					ReturnCode:        200,
					ResponseBodyRegex: ".*OK.*",
					IntervalInMillis:  5000,
					TimeoutInMillis:   2000,
					Retries:           5,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with backends using ip_address", func() {
				input := minimalValidNlb()
				input.Spec.BackendSets[0].Backends = []*OciNetworkLoadBalancerSpec_Backend{
					{IpAddress: "10.0.1.10", Port: 8080, Weight: 1},
					{IpAddress: "10.0.1.11", Port: 8080, Weight: 2},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with backends using target_id", func() {
				input := minimalValidNlb()
				input.Spec.BackendSets[0].Backends = []*OciNetworkLoadBalancerSpec_Backend{
					{TargetId: "ocid1.instance.oc1.iad.example", Port: 8080},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with backend flags", func() {
				input := minimalValidNlb()
				input.Spec.BackendSets[0].Backends = []*OciNetworkLoadBalancerSpec_Backend{
					{IpAddress: "10.0.1.10", Port: 8080, Weight: 1},
					{IpAddress: "10.0.1.11", Port: 8080, IsBackup: true},
					{IpAddress: "10.0.1.12", Port: 8080, IsDrain: true},
					{IpAddress: "10.0.1.13", Port: 8080, IsOffline: true},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with backend set failover features", func() {
				input := minimalValidNlb()
				input.Spec.BackendSets[0].IsFailOpen = true
				input.Spec.BackendSets[0].IsInstantFailoverEnabled = true
				input.Spec.BackendSets[0].IsInstantFailoverTcpResetEnabled = true
				input.Spec.BackendSets[0].IsPreserveSource = true
				input.Spec.BackendSets[0].AreOperationallyActiveBackendsPreferred = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with UDP listener", func() {
				input := minimalValidNlb()
				input.Spec.Listeners[0].Protocol = OciNetworkLoadBalancerSpec_Listener_udp
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with TCP_AND_UDP listener", func() {
				input := minimalValidNlb()
				input.Spec.Listeners[0].Protocol = OciNetworkLoadBalancerSpec_Listener_tcp_and_udp
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with ANY listener protocol", func() {
				input := minimalValidNlb()
				input.Spec.Listeners[0].Protocol = OciNetworkLoadBalancerSpec_Listener_any
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with listener idle timeouts and PPv2", func() {
				input := minimalValidNlb()
				input.Spec.Listeners[0].TcpIdleTimeout = 300
				input.Spec.Listeners[0].UdpIdleTimeout = 120
				input.Spec.Listeners[0].L3IpIdleTimeout = 200
				input.Spec.Listeners[0].IsPpv2Enabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via value_from ref", func() {
				input := minimalValidNlb()
				input.Spec.CompartmentId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-compartment",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with subnet_id via value_from ref", func() {
				input := minimalValidNlb()
				input.Spec.SubnetId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-subnet",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple backend sets and listeners", func() {
				input := minimalValidNlb()
				input.Spec.BackendSets = []*OciNetworkLoadBalancerSpec_BackendSet{
					{
						Name:   "tcp-backend",
						Policy: OciNetworkLoadBalancerSpec_BackendSet_five_tuple,
						HealthChecker: &OciNetworkLoadBalancerSpec_HealthChecker{
							Protocol: OciNetworkLoadBalancerSpec_HealthChecker_tcp,
						},
					},
					{
						Name:   "udp-backend",
						Policy: OciNetworkLoadBalancerSpec_BackendSet_three_tuple,
						HealthChecker: &OciNetworkLoadBalancerSpec_HealthChecker{
							Protocol: OciNetworkLoadBalancerSpec_HealthChecker_udp,
							Port:     5353,
						},
					},
				}
				input.Spec.Listeners = []*OciNetworkLoadBalancerSpec_Listener{
					{
						Name:                  "tcp-listener",
						Port:                  80,
						Protocol:              OciNetworkLoadBalancerSpec_Listener_tcp,
						DefaultBackendSetName: "tcp-backend",
					},
					{
						Name:                  "udp-listener",
						Port:                  5353,
						Protocol:              OciNetworkLoadBalancerSpec_Listener_udp,
						DefaultBackendSetName: "udp-backend",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_network_load_balancer", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidNlb()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidNlb()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidNlb()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciNetworkLoadBalancer{
					ApiVersion: "oci.planton.dev/v1",
					Kind:       "OciNetworkLoadBalancer",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-nlb"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidNlb()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when subnet_id is missing", func() {
				input := minimalValidNlb()
				input.Spec.SubnetId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when backend_sets is empty", func() {
				input := minimalValidNlb()
				input.Spec.BackendSets = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when listeners is empty", func() {
				input := minimalValidNlb()
				input.Spec.Listeners = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when backend set name is empty", func() {
				input := minimalValidNlb()
				input.Spec.BackendSets[0].Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when backend set policy is unspecified", func() {
				input := minimalValidNlb()
				input.Spec.BackendSets[0].Policy = OciNetworkLoadBalancerSpec_BackendSet_policy_unspecified
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when health checker is missing", func() {
				input := minimalValidNlb()
				input.Spec.BackendSets[0].HealthChecker = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when health checker protocol is unspecified", func() {
				input := minimalValidNlb()
				input.Spec.BackendSets[0].HealthChecker.Protocol = OciNetworkLoadBalancerSpec_HealthChecker_protocol_unspecified
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when listener name is empty", func() {
				input := minimalValidNlb()
				input.Spec.Listeners[0].Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when listener port is zero", func() {
				input := minimalValidNlb()
				input.Spec.Listeners[0].Port = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when listener port exceeds 65535", func() {
				input := minimalValidNlb()
				input.Spec.Listeners[0].Port = 70000
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when listener protocol is unspecified", func() {
				input := minimalValidNlb()
				input.Spec.Listeners[0].Protocol = OciNetworkLoadBalancerSpec_Listener_protocol_unspecified
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when listener default_backend_set_name is empty", func() {
				input := minimalValidNlb()
				input.Spec.Listeners[0].DefaultBackendSetName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when backend port is zero", func() {
				input := minimalValidNlb()
				input.Spec.BackendSets[0].Backends = []*OciNetworkLoadBalancerSpec_Backend{
					{IpAddress: "10.0.1.10", Port: 0},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when backend port exceeds 65535", func() {
				input := minimalValidNlb()
				input.Spec.BackendSets[0].Backends = []*OciNetworkLoadBalancerSpec_Backend{
					{IpAddress: "10.0.1.10", Port: 70000},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when reserved IP id is empty", func() {
				input := minimalValidNlb()
				input.Spec.ReservedIps = []*OciNetworkLoadBalancerSpec_ReservedIp{
					{Id: ""},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when DNS health check domain_name is empty", func() {
				input := minimalValidNlb()
				input.Spec.BackendSets[0].HealthChecker = &OciNetworkLoadBalancerSpec_HealthChecker{
					Protocol: OciNetworkLoadBalancerSpec_HealthChecker_dns,
					DnsHealthCheck: &OciNetworkLoadBalancerSpec_DnsHealthCheck{
						DomainName: "",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
