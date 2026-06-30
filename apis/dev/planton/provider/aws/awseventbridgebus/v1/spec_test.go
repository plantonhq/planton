package awseventbridgebusv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAwsEventBridgeBusSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsEventBridgeBusSpec Validation Suite")
}

// helper to create a StringValueOrRef with a literal value.
func strRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

var _ = ginkgo.Describe("AwsEventBridgeBusSpec validations", func() {
	var spec *AwsEventBridgeBusSpec

	ginkgo.BeforeEach(func() {
		spec = &AwsEventBridgeBusSpec{
			Region: "us-east-1",
		}
	})

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal custom bus (all defaults)", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a bus with description", func() {
		spec.Description = "Order processing event bus"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a bus with KMS encryption", func() {
		spec.KmsKeyIdentifier = strRef("arn:aws:kms:us-east-1:123456789012:key/mrk-abc123")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a bus with dead letter config", func() {
		spec.DeadLetterConfig = &AwsEventBridgeBusDeadLetterConfig{
			Arn: strRef("arn:aws:sqs:us-east-1:123456789012:my-bus-dlq"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a bus with log config (level only)", func() {
		spec.LogConfig = &AwsEventBridgeBusLogConfig{
			Level: "ERROR",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a bus with log config (level + include_detail)", func() {
		spec.LogConfig = &AwsEventBridgeBusLogConfig{
			Level:         "TRACE",
			IncludeDetail: "FULL",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a valid partner event source name", func() {
		spec.EventSourceName = "aws.partner/example.com/tenant123/orders"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a fully configured production bus", func() {
		spec.Description = "Production order events"
		spec.KmsKeyIdentifier = strRef("arn:aws:kms:us-east-1:123456789012:key/mrk-abc123")
		spec.DeadLetterConfig = &AwsEventBridgeBusDeadLetterConfig{
			Arn: strRef("arn:aws:sqs:us-east-1:123456789012:order-bus-dlq"),
		}
		spec.LogConfig = &AwsEventBridgeBusLogConfig{
			Level:         "ERROR",
			IncludeDetail: "NONE",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Field-level validations
	// -------------------------------------------------------------------------

	ginkgo.It("fails when description exceeds 512 characters", func() {
		spec.Description = string(make([]byte, 513))
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: event_source_name_pattern
	// -------------------------------------------------------------------------

	ginkgo.It("fails when event_source_name does not start with aws.partner/", func() {
		spec.EventSourceName = "custom.source/example/events"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when event_source_name has fewer than 2 path segments after aws.partner", func() {
		spec.EventSourceName = "aws.partner/example"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts event_source_name with exactly 2 path segments", func() {
		spec.EventSourceName = "aws.partner/example.com/events"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Dead letter config validations
	// -------------------------------------------------------------------------

	ginkgo.It("fails when dead_letter_config.arn is missing", func() {
		spec.DeadLetterConfig = &AwsEventBridgeBusDeadLetterConfig{}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Log config validations
	// -------------------------------------------------------------------------

	ginkgo.It("fails when log_config.level is missing", func() {
		spec.LogConfig = &AwsEventBridgeBusLogConfig{}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when log_config.level has an invalid value", func() {
		spec.LogConfig = &AwsEventBridgeBusLogConfig{
			Level: "DEBUG",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when log_config.include_detail has an invalid value", func() {
		spec.LogConfig = &AwsEventBridgeBusLogConfig{
			Level:         "INFO",
			IncludeDetail: "PARTIAL",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts log_config with all valid level values", func() {
		for _, level := range []string{"OFF", "ERROR", "INFO", "TRACE"} {
			spec.LogConfig = &AwsEventBridgeBusLogConfig{
				Level: level,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil(), "expected level %q to be valid", level)
		}
	})

	ginkgo.It("accepts log_config with all valid include_detail values", func() {
		for _, detail := range []string{"NONE", "FULL"} {
			spec.LogConfig = &AwsEventBridgeBusLogConfig{
				Level:         "INFO",
				IncludeDetail: detail,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil(), "expected include_detail %q to be valid", detail)
		}
	})
})
