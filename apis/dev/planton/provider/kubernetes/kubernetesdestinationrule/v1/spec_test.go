package kubernetesdestinationrulev1

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

func TestKubernetesDestinationRule(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesDestinationRule Suite")
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

var _ = ginkgo.Describe("KubernetesDestinationRule Validation Tests", func() {
	var input *KubernetesDestinationRule

	ginkgo.BeforeEach(func() {
		input = &KubernetesDestinationRule{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesDestinationRule",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-destination-rule",
			},
			Spec: &KubernetesDestinationRuleSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: literal("default"),
				Host:      "reviews.prod.svc.cluster.local",
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("with only host (minimal)", func() {
			ginkgo.It("should not return a validation error", func() {
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a simple load balancer + outlier detection + connection pool", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.TrafficPolicy = &KubernetesDestinationRuleTrafficPolicy{
					LoadBalancer: &KubernetesDestinationRuleLoadBalancerSettings{Simple: ptr("LEAST_REQUEST")},
					ConnectionPool: &KubernetesDestinationRuleConnectionPoolSettings{
						Tcp:  &KubernetesDestinationRuleTcpSettings{MaxConnections: ptr(int32(100)), ConnectTimeout: ptr("30ms")},
						Http: &KubernetesDestinationRuleHttpSettings{Http2MaxRequests: ptr(int32(1000)), H2UpgradePolicy: ptr("UPGRADE")},
					},
					OutlierDetection: &KubernetesDestinationRuleOutlierDetection{
						Consecutive_5XxErrors: ptr(uint32(7)),
						Interval:              ptr("5m"),
						BaseEjectionTime:      ptr("15m"),
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with consistent-hash (HTTP cookie) load balancing", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.TrafficPolicy = &KubernetesDestinationRuleTrafficPolicy{
					LoadBalancer: &KubernetesDestinationRuleLoadBalancerSettings{
						ConsistentHash: &KubernetesDestinationRuleConsistentHashLb{
							HttpCookie: &KubernetesDestinationRuleHttpCookie{Name: "user", Ttl: ptr("0s")},
							RingHash:   &KubernetesDestinationRuleRingHash{MinimumRingSize: ptr(uint64(1024))},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with mutual TLS via credential_name", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.TrafficPolicy = &KubernetesDestinationRuleTrafficPolicy{
					Tls: &KubernetesDestinationRuleClientTlsSettings{
						Mode:           ptr("MUTUAL"),
						CredentialName: ptr("client-credential"),
						Sni:            ptr("example.com"),
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with port-level settings reusing the shared port selector", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.TrafficPolicy = &KubernetesDestinationRuleTrafficPolicy{
					PortLevelSettings: []*KubernetesDestinationRulePortTrafficPolicy{
						{
							Port:         &kubernetes.KubernetesIstioApiPortSelector{Number: 443},
							LoadBalancer: &KubernetesDestinationRuleLoadBalancerSettings{Simple: ptr("ROUND_ROBIN")},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a subset overriding the traffic policy", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Subsets = []*KubernetesDestinationRuleSubset{
					{
						Name:          "v2",
						Labels:        map[string]string{"version": "v2"},
						TrafficPolicy: &KubernetesDestinationRuleTrafficPolicy{LoadBalancer: &KubernetesDestinationRuleLoadBalancerSettings{Simple: ptr("RANDOM")}},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with locality failover and warmup", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.TrafficPolicy = &KubernetesDestinationRuleTrafficPolicy{
					LoadBalancer: &KubernetesDestinationRuleLoadBalancerSettings{
						Warmup: &KubernetesDestinationRuleWarmupConfiguration{Duration: "30s", MinimumPercent: ptr(float64(10)), Aggression: ptr(float64(2))},
						LocalityLbSetting: &KubernetesDestinationRuleLocalityLbSetting{
							Failover: []*KubernetesDestinationRuleLocalityFailover{{From: ptr("us-east"), To: ptr("eu-west")}},
							Enabled:  ptr(true),
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a tunnel and proxy protocol", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.TrafficPolicy = &KubernetesDestinationRuleTrafficPolicy{
					Tunnel:        &KubernetesDestinationRuleTunnelSettings{Protocol: ptr("CONNECT"), TargetHost: "10.0.0.1", TargetPort: 8443},
					ProxyProtocol: &KubernetesDestinationRuleProxyProtocol{Version: ptr("V2")},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a workload selector (matchLabels)", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.WorkloadSelector = &kubernetes.KubernetesIstioApiWorkloadSelector{
					MatchLabels: map[string]string{"app": "reviews"},
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
				input.Kind = "KubernetesServiceEntry"
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

		ginkgo.Context("without a host", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Host = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When union (oneof) exclusivity is violated", func() {
		ginkgo.Context("with both simple and consistent_hash set", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.TrafficPolicy = &KubernetesDestinationRuleTrafficPolicy{
					LoadBalancer: &KubernetesDestinationRuleLoadBalancerSettings{
						Simple:         ptr("ROUND_ROBIN"),
						ConsistentHash: &KubernetesDestinationRuleConsistentHashLb{UseSourceIp: ptr(true)},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with two consistent-hash keys set", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.TrafficPolicy = &KubernetesDestinationRuleTrafficPolicy{
					LoadBalancer: &KubernetesDestinationRuleLoadBalancerSettings{
						ConsistentHash: &KubernetesDestinationRuleConsistentHashLb{
							HttpHeaderName: ptr("x-user"),
							UseSourceIp:    ptr(true),
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with both ring_hash and maglev set", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.TrafficPolicy = &KubernetesDestinationRuleTrafficPolicy{
					LoadBalancer: &KubernetesDestinationRuleLoadBalancerSettings{
						ConsistentHash: &KubernetesDestinationRuleConsistentHashLb{
							HttpHeaderName: ptr("x-user"),
							RingHash:       &KubernetesDestinationRuleRingHash{MinimumRingSize: ptr(uint64(1024))},
							Maglev:         &KubernetesDestinationRuleMagLev{TableSize: ptr(uint64(65537))},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with both warmup and warmup_duration_secs set", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.TrafficPolicy = &KubernetesDestinationRuleTrafficPolicy{
					LoadBalancer: &KubernetesDestinationRuleLoadBalancerSettings{
						WarmupDurationSecs: ptr("30s"),
						Warmup:             &KubernetesDestinationRuleWarmupConfiguration{Duration: "30s"},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with both locality distribute and failover set", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.TrafficPolicy = &KubernetesDestinationRuleTrafficPolicy{
					LoadBalancer: &KubernetesDestinationRuleLoadBalancerSettings{
						LocalityLbSetting: &KubernetesDestinationRuleLocalityLbSetting{
							Distribute: []*KubernetesDestinationRuleLocalityDistribute{{From: ptr("us-west/zone1/*"), To: map[string]uint32{"us-west/zone1/*": 100}}},
							Failover:   []*KubernetesDestinationRuleLocalityFailover{{From: ptr("us-east"), To: ptr("eu-west")}},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When closed-set enum values are invalid", func() {
		ginkgo.Context("with an unknown simple LB algorithm", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.TrafficPolicy = &KubernetesDestinationRuleTrafficPolicy{
					LoadBalancer: &KubernetesDestinationRuleLoadBalancerSettings{Simple: ptr("MAGIC")},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with an unknown h2_upgrade_policy", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.TrafficPolicy = &KubernetesDestinationRuleTrafficPolicy{
					ConnectionPool: &KubernetesDestinationRuleConnectionPoolSettings{
						Http: &KubernetesDestinationRuleHttpSettings{H2UpgradePolicy: ptr("MAYBE")},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with an unknown TLS mode", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.TrafficPolicy = &KubernetesDestinationRuleTrafficPolicy{
					Tls: &KubernetesDestinationRuleClientTlsSettings{Mode: ptr("ONE_WAY")},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with an unknown tunnel protocol", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.TrafficPolicy = &KubernetesDestinationRuleTrafficPolicy{
					Tunnel: &KubernetesDestinationRuleTunnelSettings{Protocol: ptr("GRPC"), TargetHost: "10.0.0.1", TargetPort: 8443},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with an unknown proxy protocol version", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.TrafficPolicy = &KubernetesDestinationRuleTrafficPolicy{
					ProxyProtocol: &KubernetesDestinationRuleProxyProtocol{Version: ptr("V3")},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When duration fields are invalid", func() {
		ginkgo.Context("with a connect_timeout below the 1ms minimum", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.TrafficPolicy = &KubernetesDestinationRuleTrafficPolicy{
					ConnectionPool: &KubernetesDestinationRuleConnectionPoolSettings{
						Tcp: &KubernetesDestinationRuleTcpSettings{ConnectTimeout: ptr("0s")},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a malformed outlier-detection interval", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.TrafficPolicy = &KubernetesDestinationRuleTrafficPolicy{
					OutlierDetection: &KubernetesDestinationRuleOutlierDetection{Interval: ptr("5 minutes")},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a fractional connect_timeout above the minimum (allowed)", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.TrafficPolicy = &KubernetesDestinationRuleTrafficPolicy{
					ConnectionPool: &KubernetesDestinationRuleConnectionPoolSettings{
						Tcp: &KubernetesDestinationRuleTcpSettings{ConnectTimeout: ptr("1.5s")},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When warmup bounds are violated", func() {
		ginkgo.Context("with minimum_percent above 100", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.TrafficPolicy = &KubernetesDestinationRuleTrafficPolicy{
					LoadBalancer: &KubernetesDestinationRuleLoadBalancerSettings{
						Warmup: &KubernetesDestinationRuleWarmupConfiguration{Duration: "30s", MinimumPercent: ptr(float64(150))},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with aggression below 1", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.TrafficPolicy = &KubernetesDestinationRuleTrafficPolicy{
					LoadBalancer: &KubernetesDestinationRuleLoadBalancerSettings{
						Warmup: &KubernetesDestinationRuleWarmupConfiguration{Duration: "30s", Aggression: ptr(float64(0.5))},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a missing warmup duration", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.TrafficPolicy = &KubernetesDestinationRuleTrafficPolicy{
					LoadBalancer: &KubernetesDestinationRuleLoadBalancerSettings{
						Warmup: &KubernetesDestinationRuleWarmupConfiguration{MinimumPercent: ptr(float64(10))},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When required nested fields are missing", func() {
		ginkgo.Context("with an HTTP cookie missing its name", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.TrafficPolicy = &KubernetesDestinationRuleTrafficPolicy{
					LoadBalancer: &KubernetesDestinationRuleLoadBalancerSettings{
						ConsistentHash: &KubernetesDestinationRuleConsistentHashLb{
							HttpCookie: &KubernetesDestinationRuleHttpCookie{Ttl: ptr("1h")},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a tunnel missing its target_host", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.TrafficPolicy = &KubernetesDestinationRuleTrafficPolicy{
					Tunnel: &KubernetesDestinationRuleTunnelSettings{Protocol: ptr("CONNECT"), TargetPort: 8443},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a tunnel target_port out of range", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.TrafficPolicy = &KubernetesDestinationRuleTrafficPolicy{
					Tunnel: &KubernetesDestinationRuleTunnelSettings{TargetHost: "10.0.0.1", TargetPort: 70000},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with a subset missing its name", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Subsets = []*KubernetesDestinationRuleSubset{
					{Labels: map[string]string{"version": "v2"}},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When the workload selector labels are invalid", func() {
		ginkgo.Context("with a wildcard in a label value", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.WorkloadSelector = &kubernetes.KubernetesIstioApiWorkloadSelector{
					MatchLabels: map[string]string{"app": "reviews-*"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})
