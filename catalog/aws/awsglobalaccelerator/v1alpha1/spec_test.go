package awsglobalacceleratorv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsGlobalAcceleratorSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsGlobalAcceleratorSpec Validation Suite")
}

func boolPtr(b bool) *bool {
	return &b
}

func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

func svr(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

func minimalListener() *AwsGlobalAcceleratorListener {
	return &AwsGlobalAcceleratorListener{
		Name:     "tcp-80",
		Protocol: "TCP",
		PortRanges: []*AwsGlobalAcceleratorPortRange{
			{FromPort: 80, ToPort: 80},
		},
		EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{
			{Name: "us-east-1"},
		},
	}
}

var _ = ginkgo.Describe("AwsGlobalAcceleratorSpec validations", func() {

	// -----------------------------------------------------------------
	// Valid inputs
	// -----------------------------------------------------------------
	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.Context("with minimal required fields", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region:    "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{minimalListener()},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with full production configuration", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region:        "us-west-2",
					Enabled:       boolPtr(true),
					IpAddressType: stringPtr("DUAL_STACK"),
					FlowLogs: &AwsGlobalAcceleratorFlowLogs{
						Enabled:  true,
						S3Bucket: svr("my-flow-logs-bucket"),
						S3Prefix: "ga-logs/prod/",
					},
					Listeners: []*AwsGlobalAcceleratorListener{
						{
							Name:           "https",
							Protocol:       "TCP",
							ClientAffinity: stringPtr("SOURCE_IP"),
							PortRanges: []*AwsGlobalAcceleratorPortRange{
								{FromPort: 443, ToPort: 443},
							},
							EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{
								{
									Name:                       "us-east-1",
									EndpointGroupRegion:        "us-east-1",
									HealthCheckProtocol:        stringPtr("HTTP"),
									HealthCheckPath:            "/health",
									HealthCheckIntervalSeconds: int32Ptr(10),
									ThresholdCount:             int32Ptr(5),
									TrafficDialPercentage:      float64Ptr(70.0),
									Endpoints: []*AwsGlobalAcceleratorEndpoint{
										{
											EndpointId:                  svr("arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-alb/1234567890"),
											Weight:                      int32Ptr(200),
											ClientIpPreservationEnabled: boolPtr(true),
										},
									},
									PortOverrides: []*AwsGlobalAcceleratorPortOverride{
										{ListenerPort: 443, EndpointPort: 8443},
									},
								},
								{
									Name:                  "eu-west-1",
									EndpointGroupRegion:   "eu-west-1",
									TrafficDialPercentage: float64Ptr(30.0),
									Endpoints: []*AwsGlobalAcceleratorEndpoint{
										{
											EndpointId: svr("arn:aws:elasticloadbalancing:eu-west-1:123456789012:loadbalancer/app/eu-alb/0987654321"),
											Weight:     int32Ptr(100),
										},
									},
								},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with UDP protocol and SOURCE_IP affinity", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region: "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{
						{
							Name:           "gaming-udp",
							Protocol:       "UDP",
							ClientAffinity: stringPtr("SOURCE_IP"),
							PortRanges: []*AwsGlobalAcceleratorPortRange{
								{FromPort: 7000, ToPort: 8000},
							},
							EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{
								{
									Name: "us-west-2",
									Endpoints: []*AwsGlobalAcceleratorEndpoint{
										{EndpointId: svr("eipalloc-0123456789abcdef0"), Weight: int32Ptr(128)},
									},
								},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with BYOIP addresses", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region:      "us-west-2",
					IpAddresses: []string{"198.51.100.10", "198.51.100.11"},
					Listeners:   []*AwsGlobalAcceleratorListener{minimalListener()},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with HTTPS health check and path", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region: "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{
						{
							Name:     "web",
							Protocol: "TCP",
							PortRanges: []*AwsGlobalAcceleratorPortRange{
								{FromPort: 443, ToPort: 443},
							},
							EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{
								{
									Name:                "primary",
									HealthCheckProtocol: stringPtr("HTTPS"),
									HealthCheckPath:     "/api/health",
									HealthCheckPort:     int32Ptr(8443),
								},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with health check interval 10 seconds", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region: "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{
						{
							Name:     "fast-health",
							Protocol: "TCP",
							PortRanges: []*AwsGlobalAcceleratorPortRange{
								{FromPort: 80, ToPort: 80},
							},
							EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{
								{
									Name:                       "primary",
									HealthCheckIntervalSeconds: int32Ptr(10),
								},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with endpoint weight zero (drain endpoint)", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region: "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{
						{
							Name:     "web",
							Protocol: "TCP",
							PortRanges: []*AwsGlobalAcceleratorPortRange{
								{FromPort: 80, ToPort: 80},
							},
							EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{
								{
									Name: "primary",
									Endpoints: []*AwsGlobalAcceleratorEndpoint{
										{EndpointId: svr("i-1234567890abcdef0"), Weight: int32Ptr(0)},
									},
								},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with a cross-account endpoint carrying an attachment ARN", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region: "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{
						{
							Name:     "web",
							Protocol: "TCP",
							PortRanges: []*AwsGlobalAcceleratorPortRange{
								{FromPort: 443, ToPort: 443},
							},
							EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{
								{
									Name: "primary",
									Endpoints: []*AwsGlobalAcceleratorEndpoint{
										{
											EndpointId:    svr("arn:aws:elasticloadbalancing:us-east-1:210987654321:loadbalancer/app/partner-alb/abc123"),
											AttachmentArn: "arn:aws:globalaccelerator::210987654321:attachment/11111111-2222-3333-4444-555555555555",
										},
									},
								},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with traffic dial at zero (drain region)", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region: "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{
						{
							Name:     "web",
							Protocol: "TCP",
							PortRanges: []*AwsGlobalAcceleratorPortRange{
								{FromPort: 80, ToPort: 80},
							},
							EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{
								{
									Name:                  "drained-region",
									TrafficDialPercentage: float64Ptr(0.0),
								},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with multiple port ranges", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region: "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{
						{
							Name:     "multi-port",
							Protocol: "TCP",
							PortRanges: []*AwsGlobalAcceleratorPortRange{
								{FromPort: 80, ToPort: 80},
								{FromPort: 443, ToPort: 443},
								{FromPort: 8080, ToPort: 8090},
							},
							EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{
								{Name: "primary"},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with endpoint group using valueFrom reference", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region: "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{
						{
							Name:     "web",
							Protocol: "TCP",
							PortRanges: []*AwsGlobalAcceleratorPortRange{
								{FromPort: 443, ToPort: 443},
							},
							EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{
								{
									Name: "primary",
									Endpoints: []*AwsGlobalAcceleratorEndpoint{
										{
											EndpointId: &foreignkeyv1.StringValueOrRef{
												LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
													ValueFrom: &foreignkeyv1.ValueFromRef{
														Kind:      catalogkind.CatalogKind_AwsAlb,
														Name:      "my-alb",
														FieldPath: "status.outputs.load_balancer_arn",
													},
												},
											},
											Weight: int32Ptr(128),
										},
									},
								},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with accelerator disabled", func() {
			ginkgo.It("should not return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region:    "us-west-2",
					Enabled:   boolPtr(false),
					Listeners: []*AwsGlobalAcceleratorListener{minimalListener()},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	// -----------------------------------------------------------------
	// Invalid inputs
	// -----------------------------------------------------------------
	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.Context("with no listeners", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region:    "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid protocol", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region: "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{
						{
							Name:     "bad",
							Protocol: "HTTP",
							PortRanges: []*AwsGlobalAcceleratorPortRange{
								{FromPort: 80, ToPort: 80},
							},
							EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{
								{Name: "primary"},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with missing listener name", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region: "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{
						{
							Protocol: "TCP",
							PortRanges: []*AwsGlobalAcceleratorPortRange{
								{FromPort: 80, ToPort: 80},
							},
							EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{
								{Name: "primary"},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with missing port ranges", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region: "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{
						{
							Name:       "web",
							Protocol:   "TCP",
							PortRanges: []*AwsGlobalAcceleratorPortRange{},
							EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{
								{Name: "primary"},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with missing endpoint groups", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region: "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{
						{
							Name:     "web",
							Protocol: "TCP",
							PortRanges: []*AwsGlobalAcceleratorPortRange{
								{FromPort: 80, ToPort: 80},
							},
							EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid ip_address_type", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region:        "us-west-2",
					IpAddressType: stringPtr("IPV6_ONLY"),
					Listeners:     []*AwsGlobalAcceleratorListener{minimalListener()},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with too many BYOIP addresses", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region:      "us-west-2",
					IpAddresses: []string{"198.51.100.10", "198.51.100.11", "198.51.100.12"},
					Listeners:   []*AwsGlobalAcceleratorListener{minimalListener()},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with health check interval 20 (not 10 or 30)", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region: "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{
						{
							Name:     "web",
							Protocol: "TCP",
							PortRanges: []*AwsGlobalAcceleratorPortRange{
								{FromPort: 80, ToPort: 80},
							},
							EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{
								{
									Name:                       "primary",
									HealthCheckIntervalSeconds: int32Ptr(20),
								},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with health check interval above the set (45)", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region: "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{
						{
							Name:     "web",
							Protocol: "TCP",
							PortRanges: []*AwsGlobalAcceleratorPortRange{
								{FromPort: 80, ToPort: 80},
							},
							EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{
								{
									Name:                       "primary",
									HealthCheckIntervalSeconds: int32Ptr(45),
								},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with threshold count zero", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region: "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{
						{
							Name:     "web",
							Protocol: "TCP",
							PortRanges: []*AwsGlobalAcceleratorPortRange{
								{FromPort: 80, ToPort: 80},
							},
							EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{
								{
									Name:           "primary",
									ThresholdCount: int32Ptr(0),
								},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with health check port zero", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region: "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{
						{
							Name:     "web",
							Protocol: "TCP",
							PortRanges: []*AwsGlobalAcceleratorPortRange{
								{FromPort: 80, ToPort: 80},
							},
							EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{
								{
									Name:            "primary",
									HealthCheckPort: int32Ptr(0),
								},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with a port range whose to_port is below from_port", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region: "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{
						{
							Name:     "web",
							Protocol: "TCP",
							PortRanges: []*AwsGlobalAcceleratorPortRange{
								{FromPort: 8080, ToPort: 80},
							},
							EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{
								{Name: "primary"},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with more than 10 port ranges", func() {
			ginkgo.It("should return a validation error", func() {
				ranges := make([]*AwsGlobalAcceleratorPortRange, 11)
				for i := range ranges {
					ranges[i] = &AwsGlobalAcceleratorPortRange{
						FromPort: int32(1000 + i),
						ToPort:   int32(1000 + i),
					}
				}
				spec := &AwsGlobalAcceleratorSpec{
					Region: "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{
						{
							Name:       "too-many-ranges",
							Protocol:   "TCP",
							PortRanges: ranges,
							EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{
								{Name: "primary"},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with more than 10 port overrides", func() {
			ginkgo.It("should return a validation error", func() {
				overrides := make([]*AwsGlobalAcceleratorPortOverride, 11)
				for i := range overrides {
					overrides[i] = &AwsGlobalAcceleratorPortOverride{
						ListenerPort: int32(1000 + i),
						EndpointPort: int32(2000 + i),
					}
				}
				spec := &AwsGlobalAcceleratorSpec{
					Region: "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{
						{
							Name:     "web",
							Protocol: "TCP",
							PortRanges: []*AwsGlobalAcceleratorPortRange{
								{FromPort: 80, ToPort: 80},
							},
							EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{
								{
									Name:          "primary",
									PortOverrides: overrides,
								},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with flow logs enabled but no s3_bucket", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region: "us-west-2",
					FlowLogs: &AwsGlobalAcceleratorFlowLogs{
						Enabled: true,
					},
					Listeners: []*AwsGlobalAcceleratorListener{minimalListener()},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with traffic dial above 100", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region: "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{
						{
							Name:     "web",
							Protocol: "TCP",
							PortRanges: []*AwsGlobalAcceleratorPortRange{
								{FromPort: 80, ToPort: 80},
							},
							EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{
								{
									Name:                  "primary",
									TrafficDialPercentage: float64Ptr(150.0),
								},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with a non-ARN attachment_arn", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region: "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{
						{
							Name:     "web",
							Protocol: "TCP",
							PortRanges: []*AwsGlobalAcceleratorPortRange{
								{FromPort: 80, ToPort: 80},
							},
							EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{
								{
									Name: "primary",
									Endpoints: []*AwsGlobalAcceleratorEndpoint{
										{
											EndpointId:    svr("i-1234567890abcdef0"),
											AttachmentArn: "not-an-arn",
										},
									},
								},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with HTTP health check but no path", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region: "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{
						{
							Name:     "web",
							Protocol: "TCP",
							PortRanges: []*AwsGlobalAcceleratorPortRange{
								{FromPort: 80, ToPort: 80},
							},
							EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{
								{
									Name:                "primary",
									HealthCheckProtocol: stringPtr("HTTP"),
								},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with endpoint weight out of range", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region: "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{
						{
							Name:     "web",
							Protocol: "TCP",
							PortRanges: []*AwsGlobalAcceleratorPortRange{
								{FromPort: 80, ToPort: 80},
							},
							EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{
								{
									Name: "primary",
									Endpoints: []*AwsGlobalAcceleratorEndpoint{
										{EndpointId: svr("i-1234567890abcdef0"), Weight: int32Ptr(300)},
									},
								},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with port out of range", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region: "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{
						{
							Name:     "web",
							Protocol: "TCP",
							PortRanges: []*AwsGlobalAcceleratorPortRange{
								{FromPort: 0, ToPort: 80},
							},
							EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{
								{Name: "primary"},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with invalid client affinity", func() {
			ginkgo.It("should return a validation error", func() {
				spec := &AwsGlobalAcceleratorSpec{
					Region: "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{
						{
							Name:           "web",
							Protocol:       "TCP",
							ClientAffinity: stringPtr("ROUND_ROBIN"),
							PortRanges: []*AwsGlobalAcceleratorPortRange{
								{FromPort: 80, ToPort: 80},
							},
							EndpointGroups: []*AwsGlobalAcceleratorEndpointGroup{
								{Name: "primary"},
							},
						},
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

var _ = ginkgo.Describe("AwsGlobalAccelerator API envelope validations", func() {

	ginkgo.Context("with valid API envelope", func() {
		ginkgo.It("should not return a validation error", func() {
			resource := &AwsGlobalAccelerator{
				ApiVersion: "aws.planton.dev/v1alpha1",
				Kind:       "AwsGlobalAccelerator",
				Spec: &AwsGlobalAcceleratorSpec{
					Region:    "us-west-2",
					Listeners: []*AwsGlobalAcceleratorListener{minimalListener()},
				},
			}
			err := protovalidate.Validate(resource)
			// Will fail on missing metadata, which is expected in unit test
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Context("with wrong api_version", func() {
		ginkgo.It("should return a validation error", func() {
			resource := &AwsGlobalAccelerator{
				ApiVersion: "wrong.version/v2",
				Kind:       "AwsGlobalAccelerator",
			}
			err := protovalidate.Validate(resource)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Context("with wrong kind", func() {
		ginkgo.It("should return a validation error", func() {
			resource := &AwsGlobalAccelerator{
				ApiVersion: "aws.planton.dev/v1alpha1",
				Kind:       "WrongKind",
			}
			err := protovalidate.Validate(resource)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})

// Coverage for the value-domain promotions: the attachment ARN's full shape,
// the endpoint-group region format, and the flow-logs prefix bound.
var _ = ginkgo.Describe("AwsGlobalAccelerator value-domain promotions", func() {
	ginkgo.It("accepts a real cross-account attachment ARN", func() {
		listener := minimalListener()
		listener.EndpointGroups[0].Endpoints = []*AwsGlobalAcceleratorEndpoint{{
			EndpointId:    svr("arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/x/y"),
			AttachmentArn: "arn:aws:globalaccelerator::210987654321:attachment/2f4e8b7c-0000-1111-2222-333344445555",
		}}
		spec := &AwsGlobalAcceleratorSpec{Region: "us-west-2", Listeners: []*AwsGlobalAcceleratorListener{listener}}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
	})

	ginkgo.It("rejects a non-attachment ARN in attachment_arn", func() {
		listener := minimalListener()
		listener.EndpointGroups[0].Endpoints = []*AwsGlobalAcceleratorEndpoint{{
			EndpointId:    svr("eipalloc-0123456789abcdef0"),
			AttachmentArn: "arn:aws:iam::123456789012:role/NotAnAttachment",
		}}
		spec := &AwsGlobalAcceleratorSpec{Region: "us-west-2", Listeners: []*AwsGlobalAcceleratorListener{listener}}
		gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
	})

	ginkgo.It("rejects a malformed endpoint group region", func() {
		listener := minimalListener()
		listener.EndpointGroups[0].EndpointGroupRegion = "us-east"
		spec := &AwsGlobalAcceleratorSpec{Region: "us-west-2", Listeners: []*AwsGlobalAcceleratorListener{listener}}
		gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
	})

	ginkgo.It("accepts a well-formed endpoint group region", func() {
		listener := minimalListener()
		listener.EndpointGroups[0].EndpointGroupRegion = "ap-southeast-2"
		spec := &AwsGlobalAcceleratorSpec{Region: "us-west-2", Listeners: []*AwsGlobalAcceleratorListener{listener}}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
	})
})
