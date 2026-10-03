package cloudflareworkflowv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestCloudflareWorkflowSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareWorkflowSpec Custom Validation Tests")
}

const testAccountID = "0da42c8d2132a9ddaf714f9e7c920711"

func scriptRef() *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "order-processor"},
	}
}

func validWorkflow(spec *CloudflareWorkflowSpec) *CloudflareWorkflow {
	return &CloudflareWorkflow{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareWorkflow",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-workflow",
		},
		Spec: spec,
	}
}

var _ = ginkgo.Describe("CloudflareWorkflowSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal registration", func() {
			input := validWorkflow(&CloudflareWorkflowSpec{
				AccountId:    testAccountID,
				WorkflowName: "order-fulfillment",
				ClassName:    "OrderFulfillment",
				ScriptName:   scriptRef(),
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept retention, limits, and schedules together", func() {
			input := validWorkflow(&CloudflareWorkflowSpec{
				AccountId:    testAccountID,
				WorkflowName: "nightly-reconcile",
				ClassName:    "NightlyReconcile",
				ScriptName:   scriptRef(),
				DefaultRetention: &CloudflareWorkflowRetention{
					ErrorRetention:   "7 days",
					SuccessRetention: "86400000",
				},
				Limits:    &CloudflareWorkflowLimits{Steps: proto.Int64(512)},
				Schedules: []*CloudflareWorkflowSchedule{{Cron: "0 3 * * *"}},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a malformed account id", func() {
			input := validWorkflow(&CloudflareWorkflowSpec{
				AccountId:    "not-an-account",
				WorkflowName: "order-fulfillment",
				ClassName:    "OrderFulfillment",
				ScriptName:   scriptRef(),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an empty workflow name", func() {
			input := validWorkflow(&CloudflareWorkflowSpec{
				AccountId:  testAccountID,
				ClassName:  "OrderFulfillment",
				ScriptName: scriptRef(),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing class name", func() {
			input := validWorkflow(&CloudflareWorkflowSpec{
				AccountId:    testAccountID,
				WorkflowName: "order-fulfillment",
				ScriptName:   scriptRef(),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing script reference", func() {
			input := validWorkflow(&CloudflareWorkflowSpec{
				AccountId:    testAccountID,
				WorkflowName: "order-fulfillment",
				ClassName:    "OrderFulfillment",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a zero step limit -- Cloudflare requires at least 1", func() {
			input := validWorkflow(&CloudflareWorkflowSpec{
				AccountId:    testAccountID,
				WorkflowName: "order-fulfillment",
				ClassName:    "OrderFulfillment",
				ScriptName:   scriptRef(),
				Limits:       &CloudflareWorkflowLimits{Steps: proto.Int64(0)},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a schedule without a cron expression", func() {
			input := validWorkflow(&CloudflareWorkflowSpec{
				AccountId:    testAccountID,
				WorkflowName: "order-fulfillment",
				ClassName:    "OrderFulfillment",
				ScriptName:   scriptRef(),
				Schedules:    []*CloudflareWorkflowSchedule{{}},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
