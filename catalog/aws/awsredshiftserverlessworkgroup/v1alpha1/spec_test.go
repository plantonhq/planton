package awsredshiftserverlessworkgroupv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsRedshiftServerlessWorkgroupSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsRedshiftServerlessWorkgroupSpec Custom Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// threeSubnets returns the minimum valid subnet list (three literal
// subnet IDs -- Redshift Serverless requires three distinct AZs).
func threeSubnets() []*foreignkeyv1.StringValueOrRef {
	return []*foreignkeyv1.StringValueOrRef{
		literal("subnet-11111111"),
		literal("subnet-22222222"),
		literal("subnet-33333333"),
	}
}

// validWorkgroup returns a minimal valid workgroup that individual tests
// mutate into specific scenarios.
func validWorkgroup() *AwsRedshiftServerlessWorkgroup {
	return &AwsRedshiftServerlessWorkgroup{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsRedshiftServerlessWorkgroup",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-workgroup",
		},
		Spec: &AwsRedshiftServerlessWorkgroupSpec{
			Region:        "us-west-2",
			NamespaceName: literal("analytics-namespace"),
			SubnetIds:     threeSubnets(),
		},
	}
}

var _ = ginkgo.Describe("AwsRedshiftServerlessWorkgroupSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("accepts a minimal workgroup on the AWS capacity default", func() {
			err := protovalidate.Validate(validWorkgroup())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a fixed-baseline workgroup with a spend cap", func() {
			input := validWorkgroup()
			input.Spec.BaseCapacity = 8
			input.Spec.MaxCapacity = 64
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a price-performance-targeted workgroup", func() {
			input := validWorkgroup()
			input.Spec.PricePerformanceTarget = &AwsRedshiftServerlessWorkgroupPricePerformanceTarget{
				Enabled: true,
				Level:   25,
			}
			input.Spec.MaxCapacity = 128
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a governed workgroup with endpoints, limits, and a custom domain", func() {
			input := validWorkgroup()
			input.Spec.CustomDomain = &AwsRedshiftServerlessWorkgroupCustomDomain{
				DomainName:     "warehouse.example.com",
				CertificateArn: literal("arn:aws:acm:us-west-2:123456789012:certificate/abc"),
			}
			input.Spec.EndpointAccesses = []*AwsRedshiftServerlessWorkgroupEndpointAccess{
				{EndpointName: "analytics-consumers"},
				{EndpointName: "etl-consumers", SubnetIds: threeSubnets()},
			}
			input.Spec.UsageLimits = []*AwsRedshiftServerlessWorkgroupUsageLimit{
				{UsageType: "serverless-compute", Amount: 100, Period: "daily", BreachAction: "deactivate"},
				{UsageType: "serverless-compute", Amount: 2000, Period: "monthly"},
				{UsageType: "cross-region-datasharing", Amount: 5},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a disabled price-performance target alongside a fixed baseline", func() {
			input := validWorkgroup()
			input.Spec.BaseCapacity = 32
			input.Spec.PricePerformanceTarget = &AwsRedshiftServerlessWorkgroupPricePerformanceTarget{
				Enabled: false,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a default-VPC workgroup with no subnets", func() {
			input := validWorkgroup()
			input.Spec.SubnetIds = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a fully composed production workgroup", func() {
			input := validWorkgroup()
			input.Spec.BaseCapacity = 128
			input.Spec.MaxCapacity = 512
			input.Spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{literal("sg-12345678")}
			input.Spec.EnhancedVpcRouting = true
			input.Spec.Port = 5439
			input.Spec.TrackName = "current"
			input.Spec.ConfigParameters = []*AwsRedshiftServerlessWorkgroupConfigParameter{
				{Name: "require_ssl", Value: "true"},
				{Name: "max_query_execution_time", Value: "3600"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("rejects a missing namespace reference", func() {
			input := validWorkgroup()
			input.Spec.NamespaceName = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a fixed baseline alongside an enabled price-performance target (base_capacity_xor_price_performance)", func() {
			input := validWorkgroup()
			input.Spec.BaseCapacity = 32
			input.Spec.PricePerformanceTarget = &AwsRedshiftServerlessWorkgroupPricePerformanceTarget{
				Enabled: true,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("price-performance dial"))
		})

		ginkgo.It("rejects a negative base_capacity (base_capacity_positive)", func() {
			input := validWorkgroup()
			input.Spec.BaseCapacity = -8
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a negative max_capacity (max_capacity_positive)", func() {
			input := validWorkgroup()
			input.Spec.MaxCapacity = -1
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a spend cap below the baseline (max_capacity_covers_base)", func() {
			input := validWorkgroup()
			input.Spec.BaseCapacity = 128
			input.Spec.MaxCapacity = 64
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("ceiling below the baseline"))
		})

		ginkgo.It("rejects an invalid price-performance level", func() {
			input := validWorkgroup()
			input.Spec.PricePerformanceTarget = &AwsRedshiftServerlessWorkgroupPricePerformanceTarget{
				Enabled: true,
				Level:   60,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects fewer than three subnets (subnets_span_three_azs)", func() {
			input := validWorkgroup()
			input.Spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
				literal("subnet-11111111"),
				literal("subnet-22222222"),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("three distinct availability zones"))
		})

		ginkgo.It("rejects a port outside the serverless ranges (port_in_serverless_ranges)", func() {
			input := validWorkgroup()
			input.Spec.Port = 5439
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())

			input.Spec.Port = 5500
			err = protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("5431-5455"))
		})

		ginkgo.It("rejects an unknown config parameter name", func() {
			input := validWorkgroup()
			input.Spec.ConfigParameters = []*AwsRedshiftServerlessWorkgroupConfigParameter{
				{Name: "wlm_json_configuration", Value: "{}"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a config parameter without a value", func() {
			input := validWorkgroup()
			input.Spec.ConfigParameters = []*AwsRedshiftServerlessWorkgroupConfigParameter{
				{Name: "require_ssl"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a track name with invalid characters", func() {
			input := validWorkgroup()
			input.Spec.TrackName = "current-track"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a custom domain without a certificate", func() {
			input := validWorkgroup()
			input.Spec.CustomDomain = &AwsRedshiftServerlessWorkgroupCustomDomain{
				DomainName: "warehouse.example.com",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects duplicate endpoint access names (endpoint_access_names_unique)", func() {
			input := validWorkgroup()
			input.Spec.EndpointAccesses = []*AwsRedshiftServerlessWorkgroupEndpointAccess{
				{EndpointName: "analytics-consumers"},
				{EndpointName: "analytics-consumers"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("endpoint access names must be unique"))
		})

		ginkgo.It("rejects an endpoint with no resolvable subnets (endpoint_access_subnets_resolvable)", func() {
			input := validWorkgroup()
			input.Spec.SubnetIds = nil
			input.Spec.EndpointAccesses = []*AwsRedshiftServerlessWorkgroupEndpointAccess{
				{EndpointName: "analytics-consumers"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("requires the workgroup to declare subnet_ids"))
		})

		ginkgo.It("rejects duplicate usage limit pairs (usage_limits_unique)", func() {
			input := validWorkgroup()
			input.Spec.UsageLimits = []*AwsRedshiftServerlessWorkgroupUsageLimit{
				{UsageType: "serverless-compute", Amount: 100},
				{UsageType: "serverless-compute", Amount: 200},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("usage_limits must be unique"))
		})

		ginkgo.It("rejects a provisioned-vocabulary breach action (deactivate, not disable)", func() {
			input := validWorkgroup()
			input.Spec.UsageLimits = []*AwsRedshiftServerlessWorkgroupUsageLimit{
				{UsageType: "serverless-compute", Amount: 100, BreachAction: "disable"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a non-positive usage limit amount", func() {
			input := validWorkgroup()
			input.Spec.UsageLimits = []*AwsRedshiftServerlessWorkgroupUsageLimit{
				{UsageType: "serverless-compute", Amount: 0},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})
