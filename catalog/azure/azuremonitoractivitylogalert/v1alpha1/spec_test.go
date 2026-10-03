package azuremonitoractivitylogalertv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureMonitorActivityLogAlertSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureMonitorActivityLogAlertSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func ref(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

func validResource() *AzureMonitorActivityLogAlert {
	return &AzureMonitorActivityLogAlert{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureMonitorActivityLogAlert",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-ala",
		},
		Spec: &AzureMonitorActivityLogAlertSpec{
			ResourceGroup: literal("test-rg"),
			Name:          "test-ala",
			Scopes:        []*foreignkeyv1.StringValueOrRef{ref("watched-rg")},
			Criteria: &AzureMonitorActivityLogAlertCriteria{
				Category: AzureMonitorActivityLogAlertCategory_ADMINISTRATIVE,
			},
			Actions: []*AzureMonitorActivityLogAlertAction{
				{ActionGroupId: ref("ops-action-group")},
			},
		},
	}
}

var _ = ginkgo.Describe("AzureMonitorActivityLogAlertSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_monitor_activity_log_alert", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an administrative alert with levels and operation", func() {
				input := validResource()
				input.Spec.Criteria.Levels = []AzureMonitorActivityLogAlertLevel{AzureMonitorActivityLogAlertLevel_ERROR, AzureMonitorActivityLogAlertLevel_CRITICAL}
				input.Spec.Criteria.OperationName = "Microsoft.Compute/virtualMachines/delete"
				input.Spec.Criteria.Statuses = []string{"Succeeded"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a service-health alert", func() {
				input := validResource()
				input.Spec.Criteria = &AzureMonitorActivityLogAlertCriteria{
					Category: AzureMonitorActivityLogAlertCategory_SERVICE_HEALTH,
					ServiceHealth: &AzureMonitorActivityLogAlertServiceHealth{
						Events:   []AzureMonitorActivityLogAlertServiceHealthEvent{AzureMonitorActivityLogAlertServiceHealthEvent_INCIDENT},
						Services: []string{"Virtual Machines"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a resource-health alert", func() {
				input := validResource()
				input.Spec.Criteria = &AzureMonitorActivityLogAlertCriteria{
					Category: AzureMonitorActivityLogAlertCategory_RESOURCE_HEALTH,
					ResourceHealth: &AzureMonitorActivityLogAlertResourceHealth{
						Current: []AzureMonitorActivityLogAlertHealthStatus{AzureMonitorActivityLogAlertHealthStatus_UNAVAILABLE},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a recommendation alert by category+impact", func() {
				input := validResource()
				input.Spec.Criteria = &AzureMonitorActivityLogAlertCriteria{
					Category:               AzureMonitorActivityLogAlertCategory_RECOMMENDATION,
					RecommendationCategory: AzureMonitorActivityLogAlertRecommendationCategory_COST,
					RecommendationImpact:   AzureMonitorActivityLogAlertRecommendationImpact_HIGH,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the global location and disabled flag", func() {
				input := validResource()
				input.Spec.Location = AzureMonitorActivityLogAlertLocation_GLOBAL
				disabled := false
				input.Spec.Enabled = &disabled
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_monitor_activity_log_alert", func() {

			ginkgo.It("should return a validation error when scopes are empty", func() {
				input := validResource()
				input.Spec.Scopes = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when criteria is missing", func() {
				input := validResource()
				input.Spec.Criteria = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the category is unspecified", func() {
				input := validResource()
				input.Spec.Criteria.Category = AzureMonitorActivityLogAlertCategory_azure_monitor_activity_log_alert_category_unspecified
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when caller is combined with service_health", func() {
				input := validResource()
				input.Spec.Criteria = &AzureMonitorActivityLogAlertCriteria{
					Category:      AzureMonitorActivityLogAlertCategory_SERVICE_HEALTH,
					Caller:        "someone@example.com",
					ServiceHealth: &AzureMonitorActivityLogAlertServiceHealth{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_health and service_health coexist", func() {
				input := validResource()
				input.Spec.Criteria = &AzureMonitorActivityLogAlertCriteria{
					Category:       AzureMonitorActivityLogAlertCategory_RESOURCE_HEALTH,
					ResourceHealth: &AzureMonitorActivityLogAlertResourceHealth{},
					ServiceHealth:  &AzureMonitorActivityLogAlertServiceHealth{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when recommendation_type is combined with recommendation_category", func() {
				input := validResource()
				input.Spec.Criteria = &AzureMonitorActivityLogAlertCriteria{
					Category:               AzureMonitorActivityLogAlertCategory_RECOMMENDATION,
					RecommendationType:     "some-type",
					RecommendationCategory: AzureMonitorActivityLogAlertRecommendationCategory_COST,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the action group is missing", func() {
				input := validResource()
				input.Spec.Actions = []*AzureMonitorActivityLogAlertAction{{}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := validResource()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
