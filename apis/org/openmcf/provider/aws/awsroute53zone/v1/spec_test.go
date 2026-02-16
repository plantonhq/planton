package awsroute53zonev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	dnsrecordtype "github.com/plantonhq/openmcf/apis/org/openmcf/shared/networking/enums/dnsrecordtype"
)

func TestAwsRoute53ZoneSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsRoute53ZoneSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("AwsRoute53ZoneSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("minimal valid public zone", func() {
			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &AwsRoute53Zone{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsRoute53Zone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "example.com",
					},
					Spec: &AwsRoute53ZoneSpec{
						// Public zone by default (is_private=false)
						// No records required
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("public zone with basic DNS records", func() {
			ginkgo.It("should not return a validation error for A records", func() {
				input := &AwsRoute53Zone{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsRoute53Zone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "example.com",
					},
					Spec: &AwsRoute53ZoneSpec{
						Records: []*Route53DnsRecord{
							{
								RecordType: dnsrecordtype.DnsRecordType_A,
								Name:       "www.example.com",
								Values:     []string{"192.0.2.1", "192.0.2.2"},
								TtlSeconds: 300,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for CNAME records", func() {
				input := &AwsRoute53Zone{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsRoute53Zone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "example.com",
					},
					Spec: &AwsRoute53ZoneSpec{
						Records: []*Route53DnsRecord{
							{
								RecordType: dnsrecordtype.DnsRecordType_CNAME,
								Name:       "blog.example.com",
								Values:     []string{"example.hosting.com"},
								TtlSeconds: 300,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for MX records", func() {
				input := &AwsRoute53Zone{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsRoute53Zone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "example.com",
					},
					Spec: &AwsRoute53ZoneSpec{
						Records: []*Route53DnsRecord{
							{
								RecordType: dnsrecordtype.DnsRecordType_MX,
								Name:       "example.com",
								Values:     []string{"10 mail1.example.com", "20 mail2.example.com"},
								TtlSeconds: 3600,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for TXT records", func() {
				input := &AwsRoute53Zone{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsRoute53Zone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "example.com",
					},
					Spec: &AwsRoute53ZoneSpec{
						Records: []*Route53DnsRecord{
							{
								RecordType: dnsrecordtype.DnsRecordType_TXT,
								Name:       "example.com",
								Values:     []string{"v=spf1 include:_spf.example.com ~all"},
								TtlSeconds: 300,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("alias records", func() {
			ginkgo.It("should not return a validation error for alias to CloudFront", func() {
				input := &AwsRoute53Zone{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsRoute53Zone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "example.com",
					},
					Spec: &AwsRoute53ZoneSpec{
						Records: []*Route53DnsRecord{
							{
								RecordType: dnsrecordtype.DnsRecordType_A,
								Name:       "example.com",
								AliasTarget: &Route53AliasTarget{
									DnsName:              "d1234abcd.cloudfront.net",
									HostedZoneId:         "Z2FDTNDATAQYW2",
									EvaluateTargetHealth: false,
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for alias to ALB", func() {
				input := &AwsRoute53Zone{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsRoute53Zone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "example.com",
					},
					Spec: &AwsRoute53ZoneSpec{
						Records: []*Route53DnsRecord{
							{
								RecordType: dnsrecordtype.DnsRecordType_A,
								Name:       "api.example.com",
								AliasTarget: &Route53AliasTarget{
									DnsName:              "my-alb-1234567890.us-east-1.elb.amazonaws.com",
									HostedZoneId:         "Z35SXDOTRQ7X7K",
									EvaluateTargetHealth: true,
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("routing policies", func() {
			ginkgo.It("should not return a validation error for weighted routing", func() {
				input := &AwsRoute53Zone{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsRoute53Zone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "example.com",
					},
					Spec: &AwsRoute53ZoneSpec{
						Records: []*Route53DnsRecord{
							{
								RecordType:    dnsrecordtype.DnsRecordType_A,
								Name:          "app.example.com",
								Values:        []string{"192.0.2.1"},
								TtlSeconds:    300,
								SetIdentifier: "weight-70",
								RoutingPolicy: &Route53RoutingPolicy{
									Policy: &Route53RoutingPolicy_Weighted{
										Weighted: &Route53WeightedRoutingPolicy{
											Weight: 70,
										},
									},
								},
							},
							{
								RecordType:    dnsrecordtype.DnsRecordType_A,
								Name:          "app.example.com",
								Values:        []string{"192.0.2.2"},
								TtlSeconds:    300,
								SetIdentifier: "weight-30",
								RoutingPolicy: &Route53RoutingPolicy{
									Policy: &Route53RoutingPolicy_Weighted{
										Weighted: &Route53WeightedRoutingPolicy{
											Weight: 30,
										},
									},
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for latency-based routing", func() {
				input := &AwsRoute53Zone{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsRoute53Zone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "example.com",
					},
					Spec: &AwsRoute53ZoneSpec{
						Records: []*Route53DnsRecord{
							{
								RecordType:    dnsrecordtype.DnsRecordType_A,
								Name:          "global.example.com",
								Values:        []string{"192.0.2.1"},
								TtlSeconds:    60,
								SetIdentifier: "us-east-1",
								RoutingPolicy: &Route53RoutingPolicy{
									Policy: &Route53RoutingPolicy_Latency{
										Latency: &Route53LatencyRoutingPolicy{
											Region: "us-east-1",
										},
									},
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for failover routing", func() {
				input := &AwsRoute53Zone{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsRoute53Zone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "example.com",
					},
					Spec: &AwsRoute53ZoneSpec{
						Records: []*Route53DnsRecord{
							{
								RecordType:    dnsrecordtype.DnsRecordType_A,
								Name:          "api.example.com",
								Values:        []string{"192.0.2.1"},
								TtlSeconds:    60,
								SetIdentifier: "primary",
								HealthCheckId: "abc123",
								RoutingPolicy: &Route53RoutingPolicy{
									Policy: &Route53RoutingPolicy_Failover{
										Failover: &Route53FailoverRoutingPolicy{
											Type: Route53FailoverRoutingPolicy_PRIMARY,
										},
									},
								},
							},
							{
								RecordType:    dnsrecordtype.DnsRecordType_A,
								Name:          "api.example.com",
								Values:        []string{"192.0.2.2"},
								TtlSeconds:    60,
								SetIdentifier: "secondary",
								RoutingPolicy: &Route53RoutingPolicy{
									Policy: &Route53RoutingPolicy_Failover{
										Failover: &Route53FailoverRoutingPolicy{
											Type: Route53FailoverRoutingPolicy_SECONDARY,
										},
									},
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for geolocation routing", func() {
				input := &AwsRoute53Zone{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsRoute53Zone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "example.com",
					},
					Spec: &AwsRoute53ZoneSpec{
						Records: []*Route53DnsRecord{
							{
								RecordType:    dnsrecordtype.DnsRecordType_A,
								Name:          "www.example.com",
								Values:        []string{"192.0.2.1"},
								TtlSeconds:    300,
								SetIdentifier: "europe",
								RoutingPolicy: &Route53RoutingPolicy{
									Policy: &Route53RoutingPolicy_Geolocation{
										Geolocation: &Route53GeolocationRoutingPolicy{
											Continent: "EU",
										},
									},
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("private zones", func() {
			ginkgo.It("should not return a validation error for private zone with VPC associations", func() {
				input := &AwsRoute53Zone{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsRoute53Zone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "internal.example.com",
					},
					Spec: &AwsRoute53ZoneSpec{
						IsPrivate: true,
						VpcAssociations: []*Route53VpcAssociation{
							{
								VpcId:     &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "vpc-12345678"}},
								VpcRegion: "us-east-1",
							},
						},
						Records: []*Route53DnsRecord{
							{
								RecordType: dnsrecordtype.DnsRecordType_A,
								Name:       "db.internal.example.com",
								Values:     []string{"10.0.1.100"},
								TtlSeconds: 300,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("advanced features", func() {
			ginkgo.It("should not return a validation error for zone with DNSSEC enabled", func() {
				input := &AwsRoute53Zone{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsRoute53Zone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "secure.example.com",
					},
					Spec: &AwsRoute53ZoneSpec{
						EnableDnssec: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for zone with query logging", func() {
				input := &AwsRoute53Zone{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsRoute53Zone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "monitored.example.com",
					},
					Spec: &AwsRoute53ZoneSpec{
						EnableQueryLogging: true,
						QueryLogGroupName:  "/aws/route53/monitored.example.com",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("invalid record name", func() {
			ginkgo.It("should return a validation error for invalid domain name", func() {
				input := &AwsRoute53Zone{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsRoute53Zone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "example.com",
					},
					Spec: &AwsRoute53ZoneSpec{
						Records: []*Route53DnsRecord{
							{
								RecordType: dnsrecordtype.DnsRecordType_A,
								Name:       "invalid_domain!@#", // Invalid characters
								Values:     []string{"192.0.2.1"},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("weighted routing validation", func() {
			ginkgo.It("should return a validation error for weight > 255", func() {
				input := &AwsRoute53Zone{
					ApiVersion: "aws.openmcf.org/v1",
					Kind:       "AwsRoute53Zone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "example.com",
					},
					Spec: &AwsRoute53ZoneSpec{
						Records: []*Route53DnsRecord{
							{
								RecordType:    dnsrecordtype.DnsRecordType_A,
								Name:          "app.example.com",
								Values:        []string{"192.0.2.1"},
								SetIdentifier: "heavy",
								RoutingPolicy: &Route53RoutingPolicy{
									Policy: &Route53RoutingPolicy_Weighted{
										Weighted: &Route53WeightedRoutingPolicy{
											Weight: 300, // Invalid: > 255
										},
									},
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
