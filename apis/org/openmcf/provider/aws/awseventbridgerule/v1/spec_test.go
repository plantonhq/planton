package awseventbridgerulev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestAwsEventBridgeRuleSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsEventBridgeRuleSpec Validation Suite")
}

// helper to create a StringValueOrRef with a literal value.
func strRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

// helper to create a minimal valid event pattern.
func minimalEventPattern() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]interface{}{
		"source": []interface{}{"aws.ec2"},
	})
	return s
}

// helper to create a minimal valid target.
func minimalTarget() *AwsEventBridgeTarget {
	return &AwsEventBridgeTarget{
		Name: "my-target",
		Arn:  strRef("arn:aws:lambda:us-east-1:123456789012:function:my-func"),
	}
}

var _ = ginkgo.Describe("AwsEventBridgeRuleSpec validations", func() {
	var spec *AwsEventBridgeRuleSpec

	ginkgo.BeforeEach(func() {
		// Minimal valid spec: event pattern + one target.
		spec = &AwsEventBridgeRuleSpec{
			Region:       "us-west-2",
			EventPattern: minimalEventPattern(),
			Targets:      []*AwsEventBridgeTarget{minimalTarget()},
		}
	})

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal event pattern rule with one target", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a schedule-based rule with one target", func() {
		spec.EventPattern = nil
		spec.ScheduleExpression = "rate(5 minutes)"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a rule with description", func() {
		spec.Description = "Route EC2 state change events to Lambda"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a rule with event_bus_name referencing a custom bus", func() {
		spec.EventBusName = strRef("my-custom-bus")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a rule with state ENABLED", func() {
		spec.State = "ENABLED"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a rule with state DISABLED", func() {
		spec.State = "DISABLED"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a rule with multiple targets", func() {
		spec.Targets = []*AwsEventBridgeTarget{
			{
				Name: "lambda-target",
				Arn:  strRef("arn:aws:lambda:us-east-1:123456789012:function:processor"),
			},
			{
				Name: "sqs-target",
				Arn:  strRef("arn:aws:sqs:us-east-1:123456789012:event-queue"),
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a fully configured production rule", func() {
		spec.Description = "Route order events to processing pipeline"
		spec.EventBusName = strRef("order-events")
		spec.State = "ENABLED"
		spec.Targets = []*AwsEventBridgeTarget{
			{
				Name:    "order-processor",
				Arn:     strRef("arn:aws:lambda:us-east-1:123456789012:function:order-processor"),
				Input:   `{"source": "eventbridge"}`,
				RoleArn: strRef("arn:aws:iam::123456789012:role/eb-invoke-lambda"),
				DeadLetterConfig: &AwsEventBridgeTargetDeadLetterConfig{
					Arn: strRef("arn:aws:sqs:us-east-1:123456789012:order-dlq"),
				},
				RetryPolicy: &AwsEventBridgeTargetRetryPolicy{
					MaximumEventAgeInSeconds: 3600,
					MaximumRetryAttempts:     10,
				},
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: event_pattern_or_schedule_required
	// -------------------------------------------------------------------------

	ginkgo.It("fails when neither event_pattern nor schedule_expression is set", func() {
		spec.EventPattern = nil
		spec.ScheduleExpression = ""
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when both event_pattern and schedule_expression are set", func() {
		spec.ScheduleExpression = "rate(1 hour)"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: state_valid_values
	// -------------------------------------------------------------------------

	ginkgo.It("fails when state has an invalid value", func() {
		spec.State = "PAUSED"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: targets_not_empty
	// -------------------------------------------------------------------------

	ginkgo.It("fails when targets is empty", func() {
		spec.Targets = []*AwsEventBridgeTarget{}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Field-level validations
	// -------------------------------------------------------------------------

	ginkgo.It("fails when description exceeds 512 characters", func() {
		spec.Description = string(make([]byte, 513))
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when schedule_expression exceeds 256 characters", func() {
		spec.EventPattern = nil
		spec.ScheduleExpression = string(make([]byte, 257))
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Target validations
	// -------------------------------------------------------------------------

	ginkgo.It("fails when target name is missing", func() {
		spec.Targets = []*AwsEventBridgeTarget{
			{
				Arn: strRef("arn:aws:lambda:us-east-1:123456789012:function:my-func"),
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when target arn is missing", func() {
		spec.Targets = []*AwsEventBridgeTarget{
			{
				Name: "my-target",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when target name exceeds 64 characters", func() {
		spec.Targets = []*AwsEventBridgeTarget{
			{
				Name: string(make([]byte, 65)),
				Arn:  strRef("arn:aws:lambda:us-east-1:123456789012:function:my-func"),
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when target name has invalid characters", func() {
		spec.Targets = []*AwsEventBridgeTarget{
			{
				Name: "my target!",
				Arn:  strRef("arn:aws:lambda:us-east-1:123456789012:function:my-func"),
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts target name with valid special characters", func() {
		spec.Targets = []*AwsEventBridgeTarget{
			{
				Name: "my-target_v2.0",
				Arn:  strRef("arn:aws:lambda:us-east-1:123456789012:function:my-func"),
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Target input mutual exclusion
	// -------------------------------------------------------------------------

	ginkgo.It("accepts target with input only", func() {
		spec.Targets[0].Input = `{"key": "value"}`
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts target with input_path only", func() {
		spec.Targets[0].InputPath = "$.detail"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts target with input_transformer only", func() {
		spec.Targets[0].InputTransformer = &AwsEventBridgeInputTransformer{
			InputPaths:    map[string]string{"instance": "$.detail.instance-id"},
			InputTemplate: `"Instance <instance> changed"`,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails when both input and input_path are set on a target", func() {
		spec.Targets[0].Input = `{"key": "value"}`
		spec.Targets[0].InputPath = "$.detail"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when both input and input_transformer are set on a target", func() {
		spec.Targets[0].Input = `{"key": "value"}`
		spec.Targets[0].InputTransformer = &AwsEventBridgeInputTransformer{
			InputTemplate: `"<key>"`,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when both input_path and input_transformer are set on a target", func() {
		spec.Targets[0].InputPath = "$.detail"
		spec.Targets[0].InputTransformer = &AwsEventBridgeInputTransformer{
			InputTemplate: `"<key>"`,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Input transformer validations
	// -------------------------------------------------------------------------

	ginkgo.It("fails when input_transformer is missing input_template", func() {
		spec.Targets[0].InputTransformer = &AwsEventBridgeInputTransformer{
			InputPaths: map[string]string{"instance": "$.detail.instance-id"},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when input_template exceeds 8192 characters", func() {
		spec.Targets[0].InputTransformer = &AwsEventBridgeInputTransformer{
			InputTemplate: string(make([]byte, 8193)),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Dead letter config validations
	// -------------------------------------------------------------------------

	ginkgo.It("accepts target with dead letter config", func() {
		spec.Targets[0].DeadLetterConfig = &AwsEventBridgeTargetDeadLetterConfig{
			Arn: strRef("arn:aws:sqs:us-east-1:123456789012:my-dlq"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails when dead_letter_config.arn is missing", func() {
		spec.Targets[0].DeadLetterConfig = &AwsEventBridgeTargetDeadLetterConfig{}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Retry policy validations
	// -------------------------------------------------------------------------

	ginkgo.It("accepts valid retry policy", func() {
		spec.Targets[0].RetryPolicy = &AwsEventBridgeTargetRetryPolicy{
			MaximumEventAgeInSeconds: 3600,
			MaximumRetryAttempts:     10,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts retry policy with zero retries (disable retries)", func() {
		spec.Targets[0].RetryPolicy = &AwsEventBridgeTargetRetryPolicy{
			MaximumEventAgeInSeconds: 60,
			MaximumRetryAttempts:     0,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts retry policy at maximum bounds", func() {
		spec.Targets[0].RetryPolicy = &AwsEventBridgeTargetRetryPolicy{
			MaximumEventAgeInSeconds: 86400,
			MaximumRetryAttempts:     185,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("fails when maximum_event_age_in_seconds is below minimum", func() {
		spec.Targets[0].RetryPolicy = &AwsEventBridgeTargetRetryPolicy{
			MaximumEventAgeInSeconds: 30,
			MaximumRetryAttempts:     10,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when maximum_event_age_in_seconds exceeds maximum", func() {
		spec.Targets[0].RetryPolicy = &AwsEventBridgeTargetRetryPolicy{
			MaximumEventAgeInSeconds: 100000,
			MaximumRetryAttempts:     10,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when maximum_retry_attempts exceeds 185", func() {
		spec.Targets[0].RetryPolicy = &AwsEventBridgeTargetRetryPolicy{
			MaximumEventAgeInSeconds: 3600,
			MaximumRetryAttempts:     200,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// SQS config
	// -------------------------------------------------------------------------

	ginkgo.It("accepts target with sqs_config", func() {
		spec.Targets[0].SqsConfig = &AwsEventBridgeTargetSqsConfig{
			MessageGroupId: "order-group",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})
})
