package azureapplicationinsightsv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureApplicationInsightsSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureApplicationInsightsSpec Validation Tests")
}

// buildValidAppInsights returns a minimal valid resource; tests mutate copies
// of it to probe individual rules.
func buildValidAppInsights() *AzureApplicationInsights {
	return &AzureApplicationInsights{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureApplicationInsights",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-appinsights",
		},
		Spec: &AzureApplicationInsightsSpec{
			Region: "eastus",
			ResourceGroup: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "test-resource-group",
				},
			},
			ApplicationInsightsName: "test-appinsights",
			WorkspaceId: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/test-law",
				},
			},
		},
	}
}

var _ = ginkgo.Describe("AzureApplicationInsightsSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should not return a validation error for minimal valid fields", func() {
			err := protovalidate.Validate(buildValidAppInsights())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a full production configuration", func() {
			input := buildValidAppInsights()
			input.Metadata.Org = "mycompany"
			input.Metadata.Env = "production"
			input.Spec.ApplicationType = AzureApplicationInsightsApplicationType_WEB
			input.Spec.RetentionInDays = proto.Int32(365)
			input.Spec.DailyDataCapInGb = proto.Float64(50)
			input.Spec.DailyDataCapNotificationsEnabled = proto.Bool(true)
			input.Spec.SamplingPercentage = proto.Float64(50)
			input.Spec.IpMaskingEnabled = proto.Bool(true)
			input.Spec.LocalAuthenticationEnabled = proto.Bool(false)
			input.Spec.InternetIngestionEnabled = proto.Bool(false)
			input.Spec.InternetQueryEnabled = proto.Bool(false)
			input.Spec.ForceCustomerStorageForProfiler = true
			input.Spec.Tags = map[string]string{"cost-center": "platform"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept every application type", func() {
			input := buildValidAppInsights()
			for _, appType := range []AzureApplicationInsightsApplicationType{
				AzureApplicationInsightsApplicationType_WEB,
				AzureApplicationInsightsApplicationType_JAVA,
				AzureApplicationInsightsApplicationType_NODE_JS,
				AzureApplicationInsightsApplicationType_OTHER,
				AzureApplicationInsightsApplicationType_IOS,
				AzureApplicationInsightsApplicationType_PHONE,
				AzureApplicationInsightsApplicationType_STORE,
				AzureApplicationInsightsApplicationType_MOBILE_CENTER,
			} {
				input.Spec.ApplicationType = appType
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("should accept every allowed retention value", func() {
			input := buildValidAppInsights()
			for _, days := range []int32{30, 60, 90, 120, 180, 270, 365, 550, 730} {
				input.Spec.RetentionInDays = proto.Int32(days)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("should accept sampling percentage boundaries", func() {
			input := buildValidAppInsights()
			input.Spec.SamplingPercentage = proto.Float64(0)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			input.Spec.SamplingPercentage = proto.Float64(100)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a zero daily data cap", func() {
			input := buildValidAppInsights()
			input.Spec.DailyDataCapInGb = proto.Float64(0)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a workspace reference by valueFrom", func() {
			input := buildValidAppInsights()
			input.Spec.WorkspaceId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{Name: "platform-law"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing region", func() {
			input := buildValidAppInsights()
			input.Spec.Region = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing resource group", func() {
			input := buildValidAppInsights()
			input.Spec.ResourceGroup = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing name", func() {
			input := buildValidAppInsights()
			input.Spec.ApplicationInsightsName = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a name longer than 260 characters", func() {
			input := buildValidAppInsights()
			name := make([]byte, 261)
			for i := range name {
				name[i] = 'a'
			}
			input.Spec.ApplicationInsightsName = string(name)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing workspace reference", func() {
			input := buildValidAppInsights()
			input.Spec.WorkspaceId = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an undefined application type enum number", func() {
			input := buildValidAppInsights()
			input.Spec.ApplicationType = AzureApplicationInsightsApplicationType(99)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a retention value outside Azure's fixed set", func() {
			input := buildValidAppInsights()
			input.Spec.RetentionInDays = proto.Int32(45)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a negative daily data cap", func() {
			input := buildValidAppInsights()
			input.Spec.DailyDataCapInGb = proto.Float64(-1)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a sampling percentage above 100", func() {
			input := buildValidAppInsights()
			input.Spec.SamplingPercentage = proto.Float64(100.5)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a negative sampling percentage", func() {
			input := buildValidAppInsights()
			input.Spec.SamplingPercentage = proto.Float64(-5)
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})
