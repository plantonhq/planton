package awscloudfrontv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsCloudFront(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsCloudFront Suite")
}

var _ = ginkgo.Describe("AwsCloudFront", func() {

	var input *AwsCloudFront

	// The baseline is the minimal modern shape: one S3 origin with a created
	// OAC and a default behavior on a managed cache policy.
	ginkgo.BeforeEach(func() {
		input = &AwsCloudFront{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudFront",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-cdn",
			},
			Spec: &AwsCloudFrontSpec{
				Region: "us-east-1",
				Origins: []*AwsCloudFrontOrigin{
					{
						OriginId:   "s3-assets",
						DomainName: "assets.s3.us-east-1.amazonaws.com",
						S3Origin: &AwsCloudFrontS3Origin{
							CreateOriginAccessControl: true,
						},
					},
				},
				DefaultCacheBehavior: &AwsCloudFrontCacheBehavior{
					TargetOriginId:       "s3-assets",
					ViewerProtocolPolicy: "redirect-to-https",
					CachePolicyId:        "658327ea-f89d-4fab-a63d-7e88639e58f6",
				},
			},
		}
	})

	ginkgo.Context("when valid input is passed", func() {
		ginkgo.It("should not return a validation error", func() {
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("cache_tag_header_name (lowercase format)", func() {
		ginkgo.It("should accept a lowercase header name", func() {
			input.Spec.CacheTagHeaderName = "cache-tag"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		// AWS stores the header name lowercased while the provider passes
		// the value through verbatim -- a mixed-case value would re-plan as
		// a perpetual cosmetic diff on both engines (live-verified).
		ginkgo.It("should reject a mixed-case header name", func() {
			input.Spec.CacheTagHeaderName = "Cache-Tag"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Aliases (aliases_require_certificate)", func() {
		ginkgo.It("should reject aliases without a viewer certificate", func() {
			input.Spec.Aliases = []string{"cdn.example.com"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept aliases with the ACM arm", func() {
			input.Spec.Aliases = []string{"cdn.example.com"}
			input.Spec.ViewerCertificate = &AwsCloudFrontViewerCertificate{
				AcmCertificateArn: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:acm:us-east-1:123456789012:certificate/abc"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a wildcard alias", func() {
			input.Spec.Aliases = []string{"*.example.com"}
			input.Spec.ViewerCertificate = &AwsCloudFrontViewerCertificate{
				IamCertificateId: "ASCAJRRE5XYZ",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject an alias that is not domain-shaped", func() {
			input.Spec.Aliases = []string{"not a domain"}
			input.Spec.ViewerCertificate = &AwsCloudFrontViewerCertificate{
				IamCertificateId: "ASCAJRRE5XYZ",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Origin identity (origin_ids_unique)", func() {
		ginkgo.It("should reject duplicate origin IDs", func() {
			input.Spec.Origins = append(input.Spec.Origins, &AwsCloudFrontOrigin{
				OriginId:   "s3-assets",
				DomainName: "other.s3.us-east-1.amazonaws.com",
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an origin group ID colliding with an origin ID", func() {
			input.Spec.Origins = append(input.Spec.Origins, &AwsCloudFrontOrigin{
				OriginId:   "backup",
				DomainName: "backup.s3.us-east-1.amazonaws.com",
			})
			input.Spec.OriginGroups = []*AwsCloudFrontOriginGroup{
				{
					OriginGroupId:       "s3-assets",
					MemberOriginIds:     []string{"s3-assets", "backup"},
					FailoverStatusCodes: []int32{500, 502},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Behavior targets (default_behavior_target_resolves / ordered_behavior_targets_resolve)", func() {
		ginkgo.It("should reject a default behavior targeting an undeclared origin", func() {
			input.Spec.DefaultCacheBehavior.TargetOriginId = "no-such-origin"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept a default behavior targeting an origin group", func() {
			input.Spec.Origins = append(input.Spec.Origins, &AwsCloudFrontOrigin{
				OriginId:   "backup",
				DomainName: "backup.s3.us-east-1.amazonaws.com",
			})
			input.Spec.OriginGroups = []*AwsCloudFrontOriginGroup{
				{
					OriginGroupId:       "failover-pair",
					MemberOriginIds:     []string{"s3-assets", "backup"},
					FailoverStatusCodes: []int32{500, 502, 503, 504},
				},
			}
			input.Spec.DefaultCacheBehavior.TargetOriginId = "failover-pair"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject an ordered behavior targeting an undeclared origin", func() {
			input.Spec.OrderedCacheBehaviors = []*AwsCloudFrontOrderedCacheBehavior{
				{
					PathPattern: "/api/*",
					Behavior: &AwsCloudFrontCacheBehavior{
						TargetOriginId:       "no-such-origin",
						ViewerProtocolPolicy: "https-only",
						CachePolicyId:        "4135ea2d-6df8-44a3-9df3-4b5a84be39ad",
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Origin groups (origin_group_members_resolve + shape)", func() {
		ginkgo.BeforeEach(func() {
			input.Spec.Origins = append(input.Spec.Origins, &AwsCloudFrontOrigin{
				OriginId:   "backup",
				DomainName: "backup.s3.us-east-1.amazonaws.com",
			})
		})

		ginkgo.It("should reject a member that is not a declared origin", func() {
			input.Spec.OriginGroups = []*AwsCloudFrontOriginGroup{
				{
					OriginGroupId:       "failover-pair",
					MemberOriginIds:     []string{"s3-assets", "ghost"},
					FailoverStatusCodes: []int32{500},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a single-member group", func() {
			input.Spec.OriginGroups = []*AwsCloudFrontOriginGroup{
				{
					OriginGroupId:       "failover-pair",
					MemberOriginIds:     []string{"s3-assets"},
					FailoverStatusCodes: []int32{500},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a failover status code AWS does not support", func() {
			input.Spec.OriginGroups = []*AwsCloudFrontOriginGroup{
				{
					OriginGroupId:       "failover-pair",
					MemberOriginIds:     []string{"s3-assets", "backup"},
					FailoverStatusCodes: []int32{418},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Origin type arms (at_most_one_origin_type)", func() {
		ginkgo.It("should reject an origin with both S3 and custom arms", func() {
			input.Spec.Origins[0].CustomOrigin = &AwsCloudFrontCustomOrigin{
				ProtocolPolicy: "https-only",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept a custom origin with timeouts", func() {
			input.Spec.Origins[0].S3Origin = nil
			input.Spec.Origins[0].CustomOrigin = &AwsCloudFrontCustomOrigin{
				ProtocolPolicy:          "https-only",
				ReadTimeoutSeconds:      60,
				KeepaliveTimeoutSeconds: 10,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject a custom origin with an unknown protocol policy", func() {
			input.Spec.Origins[0].S3Origin = nil
			input.Spec.Origins[0].CustomOrigin = &AwsCloudFrontCustomOrigin{
				ProtocolPolicy: "tls-only",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept a VPC origin", func() {
			input.Spec.Origins[0].S3Origin = nil
			input.Spec.Origins[0].VpcOrigin = &AwsCloudFrontVpcOrigin{
				VpcOriginId: "vo-0123456789abcdef0",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Origin access control (oac_attach_or_create / at_most_one_access_arm)", func() {
		ginkgo.It("should reject creating an OAC while also attaching one by ID", func() {
			input.Spec.Origins[0].OriginAccessControlId = "E2EXAMPLEOAC"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept attaching an existing OAC to a custom origin", func() {
			input.Spec.Origins[0].S3Origin = nil
			input.Spec.Origins[0].CustomOrigin = &AwsCloudFrontCustomOrigin{
				ProtocolPolicy: "https-only",
			}
			input.Spec.Origins[0].OriginAccessControlId = "E2EXAMPLEOAC"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject creating an OAC alongside a legacy OAI", func() {
			input.Spec.Origins[0].S3Origin = &AwsCloudFrontS3Origin{
				CreateOriginAccessControl: true,
				OriginAccessIdentity:      "origin-access-identity/cloudfront/E2EXAMPLEOAI",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept an existing legacy OAI path", func() {
			input.Spec.Origins[0].S3Origin = &AwsCloudFrontS3Origin{
				OriginAccessIdentity: "origin-access-identity/cloudfront/E2EXAMPLEOAI",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed OAI path", func() {
			input.Spec.Origins[0].S3Origin = &AwsCloudFrontS3Origin{
				OriginAccessIdentity: "E2EXAMPLEOAI",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Caching generations (exactly_one_caching_generation)", func() {
		ginkgo.It("should reject a behavior with neither a cache policy nor forwarded values", func() {
			input.Spec.DefaultCacheBehavior.CachePolicyId = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a behavior with both generations", func() {
			input.Spec.DefaultCacheBehavior.ForwardedValues = &AwsCloudFrontForwardedValues{
				CookiesForward: "none",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept the legacy generation with TTLs", func() {
			input.Spec.DefaultCacheBehavior.CachePolicyId = ""
			input.Spec.DefaultCacheBehavior.ForwardedValues = &AwsCloudFrontForwardedValues{
				QueryString:    true,
				CookiesForward: "none",
			}
			input.Spec.DefaultCacheBehavior.DefaultTtlSeconds = 3600
			input.Spec.DefaultCacheBehavior.MaxTtlSeconds = 86400
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("TTL coupling (ttls_require_forwarded_values)", func() {
		ginkgo.It("should reject TTLs alongside a cache policy", func() {
			input.Spec.DefaultCacheBehavior.DefaultTtlSeconds = 3600
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Forwarded values (whitelist_names_require_whitelist_mode)", func() {
		ginkgo.BeforeEach(func() {
			input.Spec.DefaultCacheBehavior.CachePolicyId = ""
			input.Spec.DefaultCacheBehavior.ForwardedValues = &AwsCloudFrontForwardedValues{
				CookiesForward: "none",
			}
		})

		ginkgo.It("should reject whitelisted cookies outside whitelist mode", func() {
			input.Spec.DefaultCacheBehavior.ForwardedValues.WhitelistedCookieNames = []string{"session"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept whitelisted cookies in whitelist mode", func() {
			input.Spec.DefaultCacheBehavior.ForwardedValues.CookiesForward = "whitelist"
			input.Spec.DefaultCacheBehavior.ForwardedValues.WhitelistedCookieNames = []string{"session"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown cookie forwarding mode", func() {
			input.Spec.DefaultCacheBehavior.ForwardedValues.CookiesForward = "some"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Edge function associations", func() {
		ginkgo.It("should accept a viewer-request CloudFront Function", func() {
			input.Spec.DefaultCacheBehavior.FunctionAssociations = []*AwsCloudFrontFunctionAssociation{
				{
					EventType:   "viewer-request",
					FunctionArn: "arn:aws:cloudfront::123456789012:function/url-rewrites",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject a CloudFront Function on an origin event", func() {
			input.Spec.DefaultCacheBehavior.FunctionAssociations = []*AwsCloudFrontFunctionAssociation{
				{
					EventType:   "origin-request",
					FunctionArn: "arn:aws:cloudfront::123456789012:function/url-rewrites",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept an origin-request Lambda@Edge version ARN", func() {
			input.Spec.DefaultCacheBehavior.LambdaFunctionAssociations = []*AwsCloudFrontLambdaFunctionAssociation{
				{
					EventType: "origin-request",
					LambdaArn: "arn:aws:lambda:us-east-1:123456789012:function:origin-rewrites:3",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject a Lambda@Edge ARN without a version", func() {
			input.Spec.DefaultCacheBehavior.LambdaFunctionAssociations = []*AwsCloudFrontLambdaFunctionAssociation{
				{
					EventType: "origin-request",
					LambdaArn: "arn:aws:lambda:us-east-1:123456789012:function:origin-rewrites",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a Lambda@Edge function outside us-east-1", func() {
			input.Spec.DefaultCacheBehavior.LambdaFunctionAssociations = []*AwsCloudFrontLambdaFunctionAssociation{
				{
					EventType: "origin-request",
					LambdaArn: "arn:aws:lambda:eu-west-1:123456789012:function:origin-rewrites:3",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Viewer certificate (at_most_one_certificate_arm)", func() {
		ginkgo.It("should reject both certificate arms at once", func() {
			input.Spec.ViewerCertificate = &AwsCloudFrontViewerCertificate{
				AcmCertificateArn: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:acm:us-east-1:123456789012:certificate/abc"},
				},
				IamCertificateId: "ASCAJRRE5XYZ",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown minimum protocol version", func() {
			input.Spec.ViewerCertificate = &AwsCloudFrontViewerCertificate{
				IamCertificateId:       "ASCAJRRE5XYZ",
				MinimumProtocolVersion: "TLSv1.3_2026",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Custom error responses (page_requires_response_code)", func() {
		ginkgo.It("should accept mapping S3's 403 to a custom 404 page", func() {
			input.Spec.CustomErrorResponses = []*AwsCloudFrontCustomErrorResponse{
				{
					ErrorCode:        403,
					ResponseCode:     404,
					ResponsePagePath: "/errors/404.html",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject a custom page without a response code", func() {
			input.Spec.CustomErrorResponses = []*AwsCloudFrontCustomErrorResponse{
				{
					ErrorCode:        404,
					ResponsePagePath: "/errors/404.html",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an error code CloudFront cannot intercept", func() {
			input.Spec.CustomErrorResponses = []*AwsCloudFrontCustomErrorResponse{
				{ErrorCode: 401},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Geo restriction", func() {
		ginkgo.It("should accept a whitelist of countries", func() {
			input.Spec.GeoRestriction = &AwsCloudFrontGeoRestriction{
				RestrictionType: "whitelist",
				Locations:       []string{"US", "DE", "IN"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject a lowercase country code", func() {
			input.Spec.GeoRestriction = &AwsCloudFrontGeoRestriction{
				RestrictionType: "blacklist",
				Locations:       []string{"us"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a restriction without locations", func() {
			input.Spec.GeoRestriction = &AwsCloudFrontGeoRestriction{
				RestrictionType: "whitelist",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Logging", func() {
		ginkgo.It("should accept an S3 bucket domain name", func() {
			input.Spec.Logging = &AwsCloudFrontLogging{
				Bucket: "my-logs.s3.amazonaws.com",
				Prefix: "cdn/",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a regional S3 bucket domain name", func() {
			input.Spec.Logging = &AwsCloudFrontLogging{
				Bucket: "my-logs.s3.us-west-2.amazonaws.com",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject a bare bucket name", func() {
			input.Spec.Logging = &AwsCloudFrontLogging{
				Bucket: "my-logs",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept an empty bucket (v1 delivery off, cookie preference kept)", func() {
			input.Spec.Logging = &AwsCloudFrontLogging{
				IncludeCookies: true,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Continuous deployment (continuous_deployment_arms_exclusive / exactly_one_traffic_arm)", func() {
		stagingDNS := func() []*foreignkeyv1.StringValueOrRef {
			return []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "d111111abcdef8.cloudfront.net"}},
			}
		}

		ginkgo.It("should accept a weight-routed policy", func() {
			input.Spec.ContinuousDeployment = &AwsCloudFrontContinuousDeployment{
				StagingDistributionDnsNames: stagingDNS(),
				SingleWeight: &AwsCloudFrontContinuousDeploymentSingleWeight{
					Weight: 0.10,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a header-routed policy with stickiness-free shape", func() {
			input.Spec.ContinuousDeployment = &AwsCloudFrontContinuousDeployment{
				StagingDistributionDnsNames: stagingDNS(),
				SingleHeader: &AwsCloudFrontContinuousDeploymentSingleHeader{
					Header: "aws-cf-cd-canary",
					Value:  "team-a",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject owning a policy while also attaching one by ID", func() {
			input.Spec.ContinuousDeploymentPolicyId = "cdp-example123"
			input.Spec.ContinuousDeployment = &AwsCloudFrontContinuousDeployment{
				StagingDistributionDnsNames: stagingDNS(),
				SingleWeight:                &AwsCloudFrontContinuousDeploymentSingleWeight{Weight: 0.05},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a policy with no routing arm", func() {
			input.Spec.ContinuousDeployment = &AwsCloudFrontContinuousDeployment{
				StagingDistributionDnsNames: stagingDNS(),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a policy with both routing arms", func() {
			input.Spec.ContinuousDeployment = &AwsCloudFrontContinuousDeployment{
				StagingDistributionDnsNames: stagingDNS(),
				SingleWeight:                &AwsCloudFrontContinuousDeploymentSingleWeight{Weight: 0.05},
				SingleHeader: &AwsCloudFrontContinuousDeploymentSingleHeader{
					Header: "aws-cf-cd-canary",
					Value:  "team-a",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a policy without a staging distribution", func() {
			input.Spec.ContinuousDeployment = &AwsCloudFrontContinuousDeployment{
				SingleWeight: &AwsCloudFrontContinuousDeploymentSingleWeight{Weight: 0.05},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a weight over AWS's 15% cap", func() {
			input.Spec.ContinuousDeployment = &AwsCloudFrontContinuousDeployment{
				StagingDistributionDnsNames: stagingDNS(),
				SingleWeight:                &AwsCloudFrontContinuousDeploymentSingleWeight{Weight: 0.2},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept the maximum 15% weight with session stickiness", func() {
			input.Spec.ContinuousDeployment = &AwsCloudFrontContinuousDeployment{
				StagingDistributionDnsNames: stagingDNS(),
				SingleWeight: &AwsCloudFrontContinuousDeploymentSingleWeight{
					Weight: 0.15,
					SessionStickiness: &AwsCloudFrontSessionStickiness{
						IdleTtlSeconds:    300,
						MaximumTtlSeconds: 3600,
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject stickiness with idle TTL above the maximum TTL", func() {
			input.Spec.ContinuousDeployment = &AwsCloudFrontContinuousDeployment{
				StagingDistributionDnsNames: stagingDNS(),
				SingleWeight: &AwsCloudFrontContinuousDeploymentSingleWeight{
					Weight: 0.05,
					SessionStickiness: &AwsCloudFrontSessionStickiness{
						IdleTtlSeconds:    3600,
						MaximumTtlSeconds: 300,
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject stickiness TTLs outside 300-3600", func() {
			input.Spec.ContinuousDeployment = &AwsCloudFrontContinuousDeployment{
				StagingDistributionDnsNames: stagingDNS(),
				SingleWeight: &AwsCloudFrontContinuousDeploymentSingleWeight{
					Weight: 0.05,
					SessionStickiness: &AwsCloudFrontSessionStickiness{
						IdleTtlSeconds:    60,
						MaximumTtlSeconds: 300,
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a routing header without the aws-cf-cd- prefix", func() {
			input.Spec.ContinuousDeployment = &AwsCloudFrontContinuousDeployment{
				StagingDistributionDnsNames: stagingDNS(),
				SingleHeader: &AwsCloudFrontContinuousDeploymentSingleHeader{
					Header: "x-canary",
					Value:  "team-a",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept the staging flag on its own", func() {
			input.Spec.Staging = true
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Viewer mTLS (trust_store_knobs_require_trust_store)", func() {
		ginkgo.It("should accept required mode with a trust store", func() {
			input.Spec.ViewerMtls = &AwsCloudFrontViewerMtls{
				Mode:                       "required",
				TrustStoreId:               "ts-example123",
				AdvertiseTrustStoreCaNames: true,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept passthrough mode without a trust store", func() {
			input.Spec.ViewerMtls = &AwsCloudFrontViewerMtls{
				Mode: "passthrough",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject trust-store knobs without a trust store", func() {
			input.Spec.ViewerMtls = &AwsCloudFrontViewerMtls{
				Mode:                    "required",
				IgnoreCertificateExpiry: true,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown mode", func() {
			input.Spec.ViewerMtls = &AwsCloudFrontViewerMtls{
				Mode: "strict",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("New origin leaves", func() {
		ginkgo.BeforeEach(func() {
			input.Spec.Origins[0].S3Origin = nil
			input.Spec.Origins[0].CustomOrigin = &AwsCloudFrontCustomOrigin{
				ProtocolPolicy: "https-only",
			}
		})

		ginkgo.It("should accept a dualstack origin with origin-side mTLS", func() {
			input.Spec.Origins[0].CustomOrigin.IpAddressType = "dualstack"
			input.Spec.Origins[0].CustomOrigin.MtlsClientCertificateArn = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:acm:us-east-1:123456789012:certificate/def"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown origin IP address type", func() {
			input.Spec.Origins[0].CustomOrigin.IpAddressType = "ipv5"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept a completion timeout covering the read timeout", func() {
			input.Spec.Origins[0].CustomOrigin.ReadTimeoutSeconds = 60
			input.Spec.Origins[0].ResponseCompletionTimeoutSeconds = 120
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject a completion timeout below the read timeout", func() {
			input.Spec.Origins[0].CustomOrigin.ReadTimeoutSeconds = 60
			input.Spec.Origins[0].ResponseCompletionTimeoutSeconds = 30
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed cross-account VPC origin owner", func() {
			input.Spec.Origins[0].CustomOrigin = nil
			input.Spec.Origins[0].VpcOrigin = &AwsCloudFrontVpcOrigin{
				VpcOriginId:    "vo-0123456789abcdef0",
				OwnerAccountId: "not-an-account",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept a cross-account VPC origin", func() {
			input.Spec.Origins[0].CustomOrigin = nil
			input.Spec.Origins[0].VpcOrigin = &AwsCloudFrontVpcOrigin{
				VpcOriginId:    "vo-0123456789abcdef0",
				OwnerAccountId: "210987654321",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Current-generation TLS policies", func() {
		ginkgo.It("should accept the TLSv1.3_2025 security policy", func() {
			input.Spec.ViewerCertificate = &AwsCloudFrontViewerCertificate{
				IamCertificateId:       "ASCAJRRE5XYZ",
				MinimumProtocolVersion: "TLSv1.3_2025",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept the static-ip support method", func() {
			input.Spec.ViewerCertificate = &AwsCloudFrontViewerCertificate{
				IamCertificateId: "ASCAJRRE5XYZ",
				SslSupportMethod: "static-ip",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Distribution-level knobs", func() {
		ginkgo.It("should reject an unknown price class", func() {
			input.Spec.PriceClass = "PriceClass_50"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept http2and3", func() {
			input.Spec.HttpVersion = "http2and3"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown http version", func() {
			input.Spec.HttpVersion = "http4"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a comment over 128 characters", func() {
			long := make([]byte, 129)
			for i := range long {
				long[i] = 'x'
			}
			input.Spec.Comment = string(long)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should require at least one origin", func() {
			input.Spec.Origins = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})
