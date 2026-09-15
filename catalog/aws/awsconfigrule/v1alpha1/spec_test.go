package awsconfigrulev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAwsConfigRuleSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsConfigRuleSpec Validation Suite")
}

func svr(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

// minimalManagedRule is the smallest valid instance: a managed rule.
func minimalManagedRule() *AwsConfigRuleSpec {
	return &AwsConfigRuleSpec{
		Region:  "us-west-2",
		Managed: &AwsConfigRuleManagedSource{RuleIdentifier: "S3_BUCKET_VERSIONING_ENABLED"},
	}
}

var _ = ginkgo.Describe("AwsConfigRuleSpec validations", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("accepts the minimal managed rule", func() {
			gomega.Expect(protovalidate.Validate(minimalManagedRule())).To(gomega.BeNil())
		})

		ginkgo.It("accepts a custom lambda rule with source details", func() {
			spec := &AwsConfigRuleSpec{
				Region: "us-west-2",
				CustomLambda: &AwsConfigRuleCustomLambdaSource{
					FunctionArn: svr("arn:aws:lambda:us-west-2:123456789012:function:rule-eval"),
					SourceDetails: []*AwsConfigRuleSourceDetail{{
						MessageType: "ConfigurationItemChangeNotification",
					}},
				},
			}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a custom policy rule", func() {
			spec := &AwsConfigRuleSpec{
				Region: "us-west-2",
				CustomPolicy: &AwsConfigRuleCustomPolicySource{
					PolicyRuntime: "guard-2.x.x",
					PolicyText:    "rule s3_encrypted { true }",
				},
			}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an organization block switched off beside account-only settings (the rule deploys account-scoped)", func() {
			spec := minimalManagedRule()
			spec.Organization = &AwsConfigRuleOrganization{Enabled: proto.Bool(false), ExcludedAccounts: []string{"123456789012"}}
			spec.EvaluationModes = []string{"DETECTIVE"}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an organization managed rule without trigger types", func() {
			spec := minimalManagedRule()
			spec.Organization = &AwsConfigRuleOrganization{
				ExcludedAccounts: []string{"123456789012"},
			}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an organization custom lambda rule with trigger types", func() {
			spec := &AwsConfigRuleSpec{
				Region: "us-west-2",
				CustomLambda: &AwsConfigRuleCustomLambdaSource{
					FunctionArn: svr("arn:aws:lambda:us-west-2:123456789012:function:rule-eval"),
				},
				Organization: &AwsConfigRuleOrganization{
					TriggerTypes: []string{"ConfigurationItemChangeNotification"},
				},
			}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("accepts scope and evaluation modes on an account rule", func() {
			spec := minimalManagedRule()
			spec.Scope = &AwsConfigRuleScope{
				ComplianceResourceTypes: []string{"AWS::S3::Bucket"},
				TagKey:                  "env",
				TagValue:                "prod",
			}
			spec.EvaluationModes = []string{"DETECTIVE", "PROACTIVE"}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("accepts automatic remediation with the retry contract", func() {
			spec := minimalManagedRule()
			spec.Remediation = &AwsConfigRuleRemediation{
				Automatic:                true,
				TargetId:                 "AWS-DisableS3BucketPublicReadWrite",
				MaximumAutomaticAttempts: 3,
				RetryAttemptSeconds:      60,
				Parameters: []*AwsConfigRuleRemediationParameter{
					{Name: "S3BucketName", ResourceValue: "RESOURCE_ID"},
					{Name: "SnsTopicArn", StaticValue: "arn:aws:sns:us-west-2:123456789012:alerts"},
				},
			}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("rejects a missing region", func() {
			spec := minimalManagedRule()
			spec.Region = ""
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects zero rule sources", func() {
			spec := &AwsConfigRuleSpec{Region: "us-west-2"}
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects two rule sources", func() {
			spec := minimalManagedRule()
			spec.CustomPolicy = &AwsConfigRuleCustomPolicySource{
				PolicyRuntime: "guard-2.x.x",
				PolicyText:    "rule x { true }",
			}
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an organization managed rule WITH trigger types", func() {
			spec := minimalManagedRule()
			spec.Organization = &AwsConfigRuleOrganization{
				TriggerTypes: []string{"ConfigurationItemChangeNotification"},
			}
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an organization custom rule WITHOUT trigger types", func() {
			spec := &AwsConfigRuleSpec{
				Region: "us-west-2",
				CustomLambda: &AwsConfigRuleCustomLambdaSource{
					FunctionArn: svr("arn:aws:lambda:us-west-2:123456789012:function:rule-eval"),
				},
				Organization: &AwsConfigRuleOrganization{},
			}
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects ScheduledNotification on an organization custom policy rule", func() {
			spec := &AwsConfigRuleSpec{
				Region: "us-west-2",
				CustomPolicy: &AwsConfigRuleCustomPolicySource{
					PolicyRuntime: "guard-2.x.x",
					PolicyText:    "rule x { true }",
				},
				Organization: &AwsConfigRuleOrganization{
					TriggerTypes: []string{"ScheduledNotification"},
				},
			}
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects debug log accounts on a non-policy rule", func() {
			spec := minimalManagedRule()
			spec.Organization = &AwsConfigRuleOrganization{
				DebugLogDeliveryAccounts: []string{"123456789012"},
			}
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects source details on an organization lambda rule", func() {
			spec := &AwsConfigRuleSpec{
				Region: "us-west-2",
				CustomLambda: &AwsConfigRuleCustomLambdaSource{
					FunctionArn: svr("arn:aws:lambda:us-west-2:123456789012:function:rule-eval"),
					SourceDetails: []*AwsConfigRuleSourceDetail{{
						MessageType: "ConfigurationItemChangeNotification",
					}},
				},
				Organization: &AwsConfigRuleOrganization{
					TriggerTypes: []string{"ConfigurationItemChangeNotification"},
				},
			}
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects remediation on an organization rule", func() {
			spec := minimalManagedRule()
			spec.Organization = &AwsConfigRuleOrganization{}
			spec.Remediation = &AwsConfigRuleRemediation{TargetId: "AWS-Doc"}
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects evaluation modes on an organization rule", func() {
			spec := minimalManagedRule()
			spec.Organization = &AwsConfigRuleOrganization{}
			spec.EvaluationModes = []string{"DETECTIVE"}
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a pinned resource without exactly one type", func() {
			spec := minimalManagedRule()
			spec.Scope = &AwsConfigRuleScope{ComplianceResourceId: "i-0123456789abcdef0"}
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a tag value without its key", func() {
			spec := minimalManagedRule()
			spec.Scope = &AwsConfigRuleScope{TagValue: "prod"}
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects automatic remediation without the retry contract", func() {
			spec := minimalManagedRule()
			spec.Remediation = &AwsConfigRuleRemediation{
				Automatic: true,
				TargetId:  "AWS-DisableS3BucketPublicReadWrite",
			}
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a remediation parameter with two value forms", func() {
			spec := minimalManagedRule()
			spec.Remediation = &AwsConfigRuleRemediation{
				TargetId: "AWS-Doc",
				Parameters: []*AwsConfigRuleRemediationParameter{{
					Name:         "S3BucketName",
					StaticValue:  "one",
					StaticValues: []string{"two"},
				}},
			}
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a bad guard runtime", func() {
			spec := &AwsConfigRuleSpec{
				Region: "us-west-2",
				CustomPolicy: &AwsConfigRuleCustomPolicySource{
					PolicyRuntime: "guard-3.x.x",
					PolicyText:    "rule x { true }",
				},
			}
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})
	})
})
