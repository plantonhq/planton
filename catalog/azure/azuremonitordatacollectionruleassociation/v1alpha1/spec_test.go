package azuremonitordatacollectionruleassociationv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureMonitorDataCollectionRuleAssociationSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureMonitorDataCollectionRuleAssociationSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	testVmId  = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Compute/virtualMachines/app-vm"
	testDcrId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/obs-rg/providers/Microsoft.Insights/dataCollectionRules/linux-logs"
	testDceId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/obs-rg/providers/Microsoft.Insights/dataCollectionEndpoints/obs-dce"
)

// validResource returns a valid rule-binding association that
// individual cases mutate into the shape under test.
func validResource() *AzureMonitorDataCollectionRuleAssociation {
	return &AzureMonitorDataCollectionRuleAssociation{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureMonitorDataCollectionRuleAssociation",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-dcra",
		},
		Spec: &AzureMonitorDataCollectionRuleAssociationSpec{
			TargetResourceId:     literal(testVmId),
			Name:                 "app-vm-linux-logs",
			DataCollectionRuleId: literal(testDcrId),
		},
	}
}

var _ = ginkgo.Describe("AzureMonitorDataCollectionRuleAssociationSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_monitor_data_collection_rule_association", func() {

			ginkgo.It("should not return a validation error for a rule binding with a name", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an endpoint binding without a name", func() {
				input := validResource()
				input.Spec.Name = ""
				input.Spec.DataCollectionRuleId = nil
				input.Spec.DataCollectionEndpointId = literal(testDceId)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an endpoint binding carrying an explicit name", func() {
				input := validResource()
				input.Spec.Name = "configurationAccessEndpoint"
				input.Spec.DataCollectionRuleId = nil
				input.Spec.DataCollectionEndpointId = literal(testDceId)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a description", func() {
				input := validResource()
				input.Spec.Description = "attaches the app VM to the Linux logs rule"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_monitor_data_collection_rule_association", func() {

			ginkgo.It("should reject a missing target resource id", func() {
				input := validResource()
				input.Spec.TargetResourceId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an association binding BOTH a rule and an endpoint", func() {
				input := validResource()
				input.Spec.DataCollectionEndpointId = literal(testDceId)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("exactly one"))
			})

			ginkgo.It("should reject an association binding NEITHER a rule nor an endpoint", func() {
				input := validResource()
				input.Spec.DataCollectionRuleId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("exactly one"))
			})

			ginkgo.It("should reject a rule binding without a name", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("name is required"))
			})
		})
	})
})
