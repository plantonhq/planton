package gcpdnsrecordv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestGcpDnsRecordSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpDnsRecordSpec Validation Tests")
}

// literal wraps a string in a StringValueOrRef literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// f64 returns a pointer to the given float64 (for optional-presence fields).
func f64(value float64) *float64 {
	return &value
}

// literals wraps strings in StringValueOrRef literal values.
func literals(values ...string) []*foreignkeyv1.StringValueOrRef {
	result := make([]*foreignkeyv1.StringValueOrRef, len(values))
	for i, v := range values {
		result[i] = literal(v)
	}
	return result
}

// baseRecord returns a valid minimal record that individual cases mutate.
func baseRecord() *GcpDnsRecord {
	return &GcpDnsRecord{
		ApiVersion: "gcp.planton.dev/v1alpha1",
		Kind:       "GcpDnsRecord",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-dns-record",
		},
		Spec: &GcpDnsRecordSpec{
			ProjectId:   literal("test-project-123"),
			ManagedZone: literal("example-zone"),
			Type:        "A",
			Name:        literal("www.example.com."),
			Values:      literals("192.0.2.1"),
		},
	}
}

var _ = ginkgo.Describe("GcpDnsRecordSpec Validation Tests", func() {

	ginkgo.Describe("Valid configurations", func() {
		ginkgo.Context("static value records", func() {

			ginkgo.It("should accept a minimal A record", func() {
				gomega.Expect(protovalidate.Validate(baseRecord())).To(gomega.BeNil())
			})

			ginkgo.It("should accept a CNAME record", func() {
				input := baseRecord()
				input.Spec.Type = "CNAME"
				input.Spec.Name = literal("alias.example.com.")
				input.Spec.Values = literals("target.example.com.")
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept multiple values (round-robin)", func() {
				input := baseRecord()
				input.Spec.Values = literals("192.0.2.1", "192.0.2.2", "192.0.2.3")
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a custom TTL", func() {
				ttl := int32(3600)
				input := baseRecord()
				input.Spec.Type = "TXT"
				input.Spec.Name = literal("example.com.")
				input.Spec.Values = literals("v=spf1 include:_spf.google.com ~all")
				input.Spec.TtlSeconds = &ttl
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept TTL of zero (no caching)", func() {
				ttl := int32(0)
				input := baseRecord()
				input.Spec.TtlSeconds = &ttl
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept the two-day NS-convention TTL", func() {
				ttl := int32(172800)
				input := baseRecord()
				input.Spec.Type = "NS"
				input.Spec.Values = literals("ns-cloud-a1.googledomains.com.")
				input.Spec.TtlSeconds = &ttl
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a wildcard record", func() {
				input := baseRecord()
				input.Spec.Name = literal("*.example.com.")
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept underscore service labels", func() {
				input := baseRecord()
				input.Spec.Type = "TXT"
				input.Spec.Name = literal("_dmarc.example.com.")
				input.Spec.Values = literals("v=DMARC1; p=none")
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept modern record types beyond the classic set", func() {
				for _, recordType := range []string{"HTTPS", "SVCB", "DS", "DNSKEY", "TLSA", "SSHFP", "NAPTR", "CAA", "SOA", "PTR", "SRV", "MX", "AAAA"} {
					input := baseRecord()
					input.Spec.Type = recordType
					input.Spec.Values = literals("placeholder")
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "type %s should be accepted", recordType)
				}
			})

			ginkgo.It("should accept an omitted project_id (ambient project)", func() {
				input := baseRecord()
				input.Spec.ProjectId = nil
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("routing policy records", func() {

			ginkgo.It("should accept a weighted round-robin policy", func() {
				input := baseRecord()
				input.Spec.Values = nil
				input.Spec.RoutingPolicy = &GcpDnsRecordRoutingPolicy{
					Wrr: []*GcpDnsRecordWrrPolicyItem{
						{Weight: f64(80), Values: []*foreignkeyv1.StringValueOrRef{literal("192.0.2.1")}},
						{Weight: f64(20), Values: []*foreignkeyv1.StringValueOrRef{literal("192.0.2.2")}},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a zero-weight WRR entry (staged target)", func() {
				input := baseRecord()
				input.Spec.Values = nil
				input.Spec.RoutingPolicy = &GcpDnsRecordRoutingPolicy{
					Wrr: []*GcpDnsRecordWrrPolicyItem{
						{Weight: f64(100), Values: []*foreignkeyv1.StringValueOrRef{literal("192.0.2.1")}},
						{Weight: f64(0), Values: []*foreignkeyv1.StringValueOrRef{literal("192.0.2.2")}},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a geolocation policy with fencing", func() {
				input := baseRecord()
				input.Spec.Values = nil
				input.Spec.RoutingPolicy = &GcpDnsRecordRoutingPolicy{
					Geo: []*GcpDnsRecordGeoPolicyItem{
						{Location: "us-east1", Values: []*foreignkeyv1.StringValueOrRef{literal("192.0.2.1")}},
						{Location: "europe-west3", Values: []*foreignkeyv1.StringValueOrRef{literal("192.0.2.2")}},
					},
					EnableGeoFencing: true,
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a geo policy with health-checked internal load balancers", func() {
				input := baseRecord()
				input.Spec.Values = nil
				input.Spec.RoutingPolicy = &GcpDnsRecordRoutingPolicy{
					Geo: []*GcpDnsRecordGeoPolicyItem{
						{
							Location: "us-central1",
							HealthCheckedTargets: &GcpDnsRecordHealthCheckedTargets{
								InternalLoadBalancers: []*GcpDnsRecordInternalLoadBalancerTarget{
									{
										IpAddress:        literal("10.0.0.10"),
										IpProtocol:       "tcp",
										LoadBalancerType: "regionalL4ilb",
										NetworkUrl:       literal("https://www.googleapis.com/compute/v1/projects/p/global/networks/n"),
										Port:             "80",
										Project:          literal("test-project-123"),
										Region:           "us-central1",
									},
								},
							},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a primary-backup failover policy", func() {
				trickle := 0.1
				input := baseRecord()
				input.Spec.Values = nil
				input.Spec.RoutingPolicy = &GcpDnsRecordRoutingPolicy{
					PrimaryBackup: &GcpDnsRecordPrimaryBackupPolicy{
						Primary: &GcpDnsRecordHealthCheckedTargets{
							InternalLoadBalancers: []*GcpDnsRecordInternalLoadBalancerTarget{
								{
									IpAddress:  literal("10.0.0.10"),
									IpProtocol: "tcp",
									NetworkUrl: literal("https://www.googleapis.com/compute/v1/projects/p/global/networks/n"),
									Port:       "80",
									Project:    literal("test-project-123"),
									Region:     "us-central1",
								},
							},
						},
						BackupGeo: []*GcpDnsRecordGeoPolicyItem{
							{Location: "us-east1", Values: []*foreignkeyv1.StringValueOrRef{literal("192.0.2.9")}},
						},
						TrickleRatio: &trickle,
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept external endpoints with a health check reference", func() {
				input := baseRecord()
				input.Spec.Values = nil
				input.Spec.RoutingPolicy = &GcpDnsRecordRoutingPolicy{
					Wrr: []*GcpDnsRecordWrrPolicyItem{
						{
							Weight: f64(1),
							HealthCheckedTargets: &GcpDnsRecordHealthCheckedTargets{
								ExternalEndpoints: []string{"203.0.113.10"},
							},
						},
					},
					HealthCheck: literal("https://www.googleapis.com/compute/v1/projects/p/global/healthChecks/hc"),
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("deletion policy", func() {
			ginkgo.It("should accept each deletion_policy value", func() {
				for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
					input := baseRecord()
					input.Spec.DeletionPolicy = v
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				}
			})
		})
	})

	ginkgo.Describe("Invalid configurations", func() {
		ginkgo.Context("missing required fields", func() {

			ginkgo.It("should reject an invalid deletion_policy", func() {
				input := baseRecord()
				input.Spec.DeletionPolicy = "KEEP"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a missing managed_zone", func() {
				input := baseRecord()
				input.Spec.ManagedZone = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a missing type", func() {
				input := baseRecord()
				input.Spec.Type = ""
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a missing name", func() {
				input := baseRecord()
				input.Spec.Name = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a record with neither values nor routing_policy", func() {
				input := baseRecord()
				input.Spec.Values = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a record with both values and routing_policy", func() {
				input := baseRecord()
				input.Spec.RoutingPolicy = &GcpDnsRecordRoutingPolicy{
					Wrr: []*GcpDnsRecordWrrPolicyItem{{Weight: f64(1), Values: []*foreignkeyv1.StringValueOrRef{literal("192.0.2.5")}}},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid field formats", func() {

			ginkgo.It("should reject a lowercase type", func() {
				input := baseRecord()
				input.Spec.Type = "a"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a name without a trailing dot", func() {
				input := baseRecord()
				input.Spec.Name = literal("www.example.com")
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a negative TTL", func() {
				ttl := int32(-1)
				input := baseRecord()
				input.Spec.TtlSeconds = &ttl
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("routing policy coherence", func() {

			ginkgo.It("should reject an empty routing policy", func() {
				input := baseRecord()
				input.Spec.Values = nil
				input.Spec.RoutingPolicy = &GcpDnsRecordRoutingPolicy{}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a policy mixing wrr and geo", func() {
				input := baseRecord()
				input.Spec.Values = nil
				input.Spec.RoutingPolicy = &GcpDnsRecordRoutingPolicy{
					Wrr: []*GcpDnsRecordWrrPolicyItem{{Weight: f64(1), Values: []*foreignkeyv1.StringValueOrRef{literal("192.0.2.1")}}},
					Geo: []*GcpDnsRecordGeoPolicyItem{{Location: "us-east1", Values: []*foreignkeyv1.StringValueOrRef{literal("192.0.2.2")}}},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject geo fencing on a wrr policy", func() {
				input := baseRecord()
				input.Spec.Values = nil
				input.Spec.RoutingPolicy = &GcpDnsRecordRoutingPolicy{
					Wrr:              []*GcpDnsRecordWrrPolicyItem{{Weight: f64(1), Values: []*foreignkeyv1.StringValueOrRef{literal("192.0.2.1")}}},
					EnableGeoFencing: true,
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a geo entry without a location", func() {
				input := baseRecord()
				input.Spec.Values = nil
				input.Spec.RoutingPolicy = &GcpDnsRecordRoutingPolicy{
					Geo: []*GcpDnsRecordGeoPolicyItem{{Values: []*foreignkeyv1.StringValueOrRef{literal("192.0.2.1")}}},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a primary-backup policy without backup_geo", func() {
				input := baseRecord()
				input.Spec.Values = nil
				input.Spec.RoutingPolicy = &GcpDnsRecordRoutingPolicy{
					PrimaryBackup: &GcpDnsRecordPrimaryBackupPolicy{
						Primary: &GcpDnsRecordHealthCheckedTargets{
							ExternalEndpoints: []string{"203.0.113.10"},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a primary-backup policy without a primary", func() {
				input := baseRecord()
				input.Spec.Values = nil
				input.Spec.RoutingPolicy = &GcpDnsRecordRoutingPolicy{
					PrimaryBackup: &GcpDnsRecordPrimaryBackupPolicy{
						BackupGeo: []*GcpDnsRecordGeoPolicyItem{
							{Location: "us-east1", Values: []*foreignkeyv1.StringValueOrRef{literal("192.0.2.9")}},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a trickle ratio above 1.0", func() {
				trickle := 1.5
				input := baseRecord()
				input.Spec.Values = nil
				input.Spec.RoutingPolicy = &GcpDnsRecordRoutingPolicy{
					PrimaryBackup: &GcpDnsRecordPrimaryBackupPolicy{
						Primary: &GcpDnsRecordHealthCheckedTargets{
							ExternalEndpoints: []string{"203.0.113.10"},
						},
						BackupGeo: []*GcpDnsRecordGeoPolicyItem{
							{Location: "us-east1", Values: []*foreignkeyv1.StringValueOrRef{literal("192.0.2.9")}},
						},
						TrickleRatio: &trickle,
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject empty health-checked targets", func() {
				input := baseRecord()
				input.Spec.Values = nil
				input.Spec.RoutingPolicy = &GcpDnsRecordRoutingPolicy{
					Wrr: []*GcpDnsRecordWrrPolicyItem{
						{Weight: f64(1), HealthCheckedTargets: &GcpDnsRecordHealthCheckedTargets{}},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an internal load balancer target with a bad protocol", func() {
				input := baseRecord()
				input.Spec.Values = nil
				input.Spec.RoutingPolicy = &GcpDnsRecordRoutingPolicy{
					Geo: []*GcpDnsRecordGeoPolicyItem{
						{
							Location: "us-central1",
							HealthCheckedTargets: &GcpDnsRecordHealthCheckedTargets{
								InternalLoadBalancers: []*GcpDnsRecordInternalLoadBalancerTarget{
									{
										IpAddress:  literal("10.0.0.10"),
										IpProtocol: "TCP",
										NetworkUrl: literal("https://www.googleapis.com/compute/v1/projects/p/global/networks/n"),
										Port:       "80",
										Project:    literal("test-project-123"),
									},
								},
							},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an internal load balancer target with a bad type", func() {
				input := baseRecord()
				input.Spec.Values = nil
				input.Spec.RoutingPolicy = &GcpDnsRecordRoutingPolicy{
					Geo: []*GcpDnsRecordGeoPolicyItem{
						{
							Location: "us-central1",
							HealthCheckedTargets: &GcpDnsRecordHealthCheckedTargets{
								InternalLoadBalancers: []*GcpDnsRecordInternalLoadBalancerTarget{
									{
										IpAddress:        literal("10.0.0.10"),
										IpProtocol:       "tcp",
										LoadBalancerType: "l4",
										NetworkUrl:       literal("https://www.googleapis.com/compute/v1/projects/p/global/networks/n"),
										Port:             "80",
										Project:          literal("test-project-123"),
									},
								},
							},
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})
	})
})
