package awsroute53dnsrecordv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAwsRoute53DnsRecordSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsRoute53DnsRecordSpec Custom Validation Tests")
}

// Helper to create StringValueOrRef with literal value
func stringValue(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

var _ = ginkgo.Describe("AwsRoute53DnsRecordSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_route53_dns_record", func() {

			ginkgo.It("should not return a validation error for minimal valid A record", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-a-record",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						ZoneId: stringValue("Z1234567890ABC"),
						Name:   "www.example.com",
						Type:   AwsRoute53DnsRecordSpec_A,
						Ttl:    300,
						Values: []string{"192.0.2.1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for AAAA record", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-aaaa-record",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						ZoneId: stringValue("Z1234567890ABC"),
						Name:   "www.example.com",
						Type:   AwsRoute53DnsRecordSpec_AAAA,
						Ttl:    300,
						Values: []string{"2001:db8::1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for CNAME record", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-cname-record",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						ZoneId: stringValue("Z1234567890ABC"),
						Name:   "app.example.com",
						Type:   AwsRoute53DnsRecordSpec_CNAME,
						Ttl:    300,
						Values: []string{"target.example.com"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for MX record", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-mx-record",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						ZoneId: stringValue("Z1234567890ABC"),
						Name:   "example.com",
						Type:   AwsRoute53DnsRecordSpec_MX,
						Ttl:    3600,
						Values: []string{"10 mail1.example.com", "20 mail2.example.com"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for TXT record", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-txt-record",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						ZoneId: stringValue("Z1234567890ABC"),
						Name:   "example.com",
						Type:   AwsRoute53DnsRecordSpec_TXT,
						Ttl:    300,
						Values: []string{"v=spf1 include:_spf.google.com ~all"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for A record with multiple values", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-multi-a-record",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						ZoneId: stringValue("Z1234567890ABC"),
						Name:   "www.example.com",
						Type:   AwsRoute53DnsRecordSpec_A,
						Ttl:    300,
						Values: []string{"192.0.2.1", "192.0.2.2", "192.0.2.3"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for alias record to CloudFront", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-alias-record",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						ZoneId: stringValue("Z1234567890ABC"),
						Name:   "example.com",
						Type:   AwsRoute53DnsRecordSpec_A,
						AliasTarget: &AwsRoute53AliasTarget{
							DnsName: stringValue("d1234abcd.cloudfront.net"),
							ZoneId:  stringValue("Z2FDTNDATAQYW2"),
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for alias record to ALB", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-alb-alias-record",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						ZoneId: stringValue("Z1234567890ABC"),
						Name:   "api.example.com",
						Type:   AwsRoute53DnsRecordSpec_A,
						AliasTarget: &AwsRoute53AliasTarget{
							DnsName:              stringValue("my-alb-1234567890.us-east-1.elb.amazonaws.com"),
							ZoneId:               stringValue("Z35SXDOTRQ7X7K"),
							EvaluateTargetHealth: true,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for wildcard record", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-wildcard-record",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						ZoneId: stringValue("Z1234567890ABC"),
						Name:   "*.example.com",
						Type:   AwsRoute53DnsRecordSpec_A,
						Ttl:    300,
						Values: []string{"192.0.2.1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for weighted routing policy", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-weighted-record",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						ZoneId: stringValue("Z1234567890ABC"),
						Name:   "www.example.com",
						Type:   AwsRoute53DnsRecordSpec_A,
						Ttl:    300,
						Values: []string{"192.0.2.1"},
						RoutingPolicy: &AwsRoute53RoutingPolicy{
							Policy: &AwsRoute53RoutingPolicy_Weighted{
								Weighted: &AwsRoute53WeightedPolicy{
									Weight: 70,
								},
							},
						},
						SetIdentifier: "primary",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for latency routing policy", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-latency-record",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						ZoneId: stringValue("Z1234567890ABC"),
						Name:   "api.example.com",
						Type:   AwsRoute53DnsRecordSpec_A,
						Ttl:    60,
						Values: []string{"192.0.2.1"},
						RoutingPolicy: &AwsRoute53RoutingPolicy{
							Policy: &AwsRoute53RoutingPolicy_Latency{
								Latency: &AwsRoute53LatencyPolicy{
									Region: "us-east-1",
								},
							},
						},
						SetIdentifier: "us-east-1",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for failover routing policy", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-failover-primary",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						ZoneId: stringValue("Z1234567890ABC"),
						Name:   "www.example.com",
						Type:   AwsRoute53DnsRecordSpec_A,
						Ttl:    60,
						Values: []string{"192.0.2.1"},
						RoutingPolicy: &AwsRoute53RoutingPolicy{
							Policy: &AwsRoute53RoutingPolicy_Failover{
								Failover: &AwsRoute53FailoverPolicy{
									FailoverType: AwsRoute53FailoverPolicy_primary,
								},
							},
						},
						SetIdentifier: "primary",
						HealthCheckId: "abcd1234-5678-90ab-cdef-example",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for geolocation routing policy with country", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-geo-record",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						ZoneId: stringValue("Z1234567890ABC"),
						Name:   "www.example.com",
						Type:   AwsRoute53DnsRecordSpec_A,
						Ttl:    300,
						Values: []string{"192.0.2.1"},
						RoutingPolicy: &AwsRoute53RoutingPolicy{
							Policy: &AwsRoute53RoutingPolicy_Geolocation{
								Geolocation: &AwsRoute53GeolocationPolicy{
									Country: "US",
								},
							},
						},
						SetIdentifier: "us",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for geolocation routing with continent", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-geo-continent-record",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						ZoneId: stringValue("Z1234567890ABC"),
						Name:   "www.example.com",
						Type:   AwsRoute53DnsRecordSpec_A,
						Ttl:    300,
						Values: []string{"192.0.2.1"},
						RoutingPolicy: &AwsRoute53RoutingPolicy{
							Policy: &AwsRoute53RoutingPolicy_Geolocation{
								Geolocation: &AwsRoute53GeolocationPolicy{
									Continent: "EU",
								},
							},
						},
						SetIdentifier: "europe",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for CAA record", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-caa-record",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						ZoneId: stringValue("Z1234567890ABC"),
						Name:   "example.com",
						Type:   AwsRoute53DnsRecordSpec_CAA,
						Ttl:    3600,
						Values: []string{"0 issue \"letsencrypt.org\""},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("aws_route53_dns_record", func() {

			ginkgo.It("should return a validation error when zone_id is missing", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						Name:   "www.example.com",
						Type:   AwsRoute53DnsRecordSpec_A,
						Ttl:    300,
						Values: []string{"192.0.2.1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						ZoneId: stringValue("Z1234567890ABC"),
						Type:   AwsRoute53DnsRecordSpec_A,
						Ttl:    300,
						Values: []string{"192.0.2.1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when type is unspecified", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						ZoneId: stringValue("Z1234567890ABC"),
						Name:   "www.example.com",
						Type:   AwsRoute53DnsRecordSpec_record_type_unspecified,
						Ttl:    300,
						Values: []string{"192.0.2.1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when neither values nor alias_target is specified", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						ZoneId: stringValue("Z1234567890ABC"),
						Name:   "www.example.com",
						Type:   AwsRoute53DnsRecordSpec_A,
						Ttl:    300,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when both values and alias_target are specified", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						ZoneId: stringValue("Z1234567890ABC"),
						Name:   "www.example.com",
						Type:   AwsRoute53DnsRecordSpec_A,
						Ttl:    300,
						Values: []string{"192.0.2.1"},
						AliasTarget: &AwsRoute53AliasTarget{
							DnsName: stringValue("d1234abcd.cloudfront.net"),
							ZoneId:  stringValue("Z2FDTNDATAQYW2"),
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for TTL exceeding max", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						ZoneId: stringValue("Z1234567890ABC"),
						Name:   "www.example.com",
						Type:   AwsRoute53DnsRecordSpec_A,
						Ttl:    700000,
						Values: []string{"192.0.2.1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for alias_target missing dns_name", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						ZoneId: stringValue("Z1234567890ABC"),
						Name:   "example.com",
						Type:   AwsRoute53DnsRecordSpec_A,
						AliasTarget: &AwsRoute53AliasTarget{
							ZoneId: stringValue("Z2FDTNDATAQYW2"),
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for alias_target missing zone_id", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						ZoneId: stringValue("Z1234567890ABC"),
						Name:   "example.com",
						Type:   AwsRoute53DnsRecordSpec_A,
						AliasTarget: &AwsRoute53AliasTarget{
							DnsName: stringValue("d1234abcd.cloudfront.net"),
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for weighted routing without set_identifier", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						ZoneId: stringValue("Z1234567890ABC"),
						Name:   "www.example.com",
						Type:   AwsRoute53DnsRecordSpec_A,
						Ttl:    300,
						Values: []string{"192.0.2.1"},
						RoutingPolicy: &AwsRoute53RoutingPolicy{
							Policy: &AwsRoute53RoutingPolicy_Weighted{
								Weighted: &AwsRoute53WeightedPolicy{
									Weight: 70,
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for latency routing missing region", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						ZoneId: stringValue("Z1234567890ABC"),
						Name:   "www.example.com",
						Type:   AwsRoute53DnsRecordSpec_A,
						Ttl:    300,
						Values: []string{"192.0.2.1"},
						RoutingPolicy: &AwsRoute53RoutingPolicy{
							Policy: &AwsRoute53RoutingPolicy_Latency{
								Latency: &AwsRoute53LatencyPolicy{},
							},
						},
						SetIdentifier: "us-east",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for weighted routing with weight exceeding max", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						ZoneId: stringValue("Z1234567890ABC"),
						Name:   "www.example.com",
						Type:   AwsRoute53DnsRecordSpec_A,
						Ttl:    300,
						Values: []string{"192.0.2.1"},
						RoutingPolicy: &AwsRoute53RoutingPolicy{
							Policy: &AwsRoute53RoutingPolicy_Weighted{
								Weighted: &AwsRoute53WeightedPolicy{
									Weight: 300,
								},
							},
						},
						SetIdentifier: "primary",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for invalid name pattern", func() {
				input := &AwsRoute53DnsRecord{
					ApiVersion: "aws.planton.dev/v1",
					Kind:       "AwsRoute53DnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &AwsRoute53DnsRecordSpec{
						Region: "us-east-1",
						ZoneId: stringValue("Z1234567890ABC"),
						Name:   "invalid name with spaces",
						Type:   AwsRoute53DnsRecordSpec_A,
						Ttl:    300,
						Values: []string{"192.0.2.1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
