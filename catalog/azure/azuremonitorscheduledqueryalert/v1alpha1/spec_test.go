package azuremonitorscheduledqueryalertv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureMonitorScheduledQueryAlertSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureMonitorScheduledQueryAlertSpec Validation Tests")
}

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

// buildValidQueryAlert returns a minimal valid resource (a row-count
// condition on a workspace scope); tests mutate copies of it.
func buildValidQueryAlert() *AzureMonitorScheduledQueryAlert {
	return &AzureMonitorScheduledQueryAlert{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureMonitorScheduledQueryAlert",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-query-alert",
		},
		Spec: &AzureMonitorScheduledQueryAlertSpec{
			Region:        "eastus",
			ResourceGroup: literal("test-resource-group"),
			AlertName:     "error-spike",
			Scope:         literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/law"),
			Criteria: []*AzureMonitorScheduledQueryAlertCriteria{
				{
					Query:                 "AzureDiagnostics | where Level == \"Error\"",
					TimeAggregationMethod: AzureMonitorScheduledQueryAlertTimeAggregation_COUNT,
					Operator:              AzureMonitorScheduledQueryAlertOperator_GREATER_THAN,
					Threshold:             5,
				},
			},
			Action: &AzureMonitorScheduledQueryAlertAction{
				ActionGroupIds: []*foreignkeyv1.StringValueOrRef{
					literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Insights/actionGroups/oncall"),
				},
			},
		},
	}
}

var _ = ginkgo.Describe("AzureMonitorScheduledQueryAlertSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should not return a validation error for minimal valid fields", func() {
			err := protovalidate.Validate(buildValidQueryAlert())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a metric-measurement condition with dimensions and failing periods", func() {
			input := buildValidQueryAlert()
			input.Spec.DisplayName = "API latency"
			input.Spec.Description = "p95 latency over 500ms"
			input.Spec.Severity = proto.Int32(1)
			input.Spec.EvaluationFrequency = proto.String("PT15M")
			input.Spec.WindowDuration = proto.String("PT30M")
			input.Spec.Criteria = []*AzureMonitorScheduledQueryAlertCriteria{
				{
					Query:                 "AppRequests | summarize avg(DurationMs) by bin(TimeGenerated, 5m), AppRoleName",
					TimeAggregationMethod: AzureMonitorScheduledQueryAlertTimeAggregation_AVERAGE,
					Operator:              AzureMonitorScheduledQueryAlertOperator_GREATER_THAN,
					Threshold:             500,
					MetricMeasureColumn:   "avg_DurationMs",
					Dimensions: []*AzureMonitorScheduledQueryAlertDimension{
						{
							Name:     "AppRoleName",
							Operator: AzureMonitorScheduledQueryAlertDimensionOperator_INCLUDE,
							Values:   []string{"*"},
						},
					},
					FailingPeriods: &AzureMonitorScheduledQueryAlertFailingPeriods{
						MinimumFailingPeriodsToTriggerAlert: 3,
						NumberOfEvaluationPeriods:           5,
					},
				},
			}
			input.Spec.Tags = map[string]string{"team": "platform"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a mute duration when auto-mitigation is off", func() {
			input := buildValidQueryAlert()
			input.Spec.MuteActionsAfterAlertDuration = "PT30M"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept auto-mitigation without a mute duration", func() {
			input := buildValidQueryAlert()
			input.Spec.AutoMitigationEnabled = true
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a system-assigned identity", func() {
			input := buildValidQueryAlert()
			input.Spec.Identity = &AzureMonitorScheduledQueryAlertIdentity{
				Type: AzureMonitorScheduledQueryAlertIdentityType_SYSTEM_ASSIGNED,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a user-assigned identity with an identity id", func() {
			input := buildValidQueryAlert()
			input.Spec.Identity = &AzureMonitorScheduledQueryAlertIdentity{
				Type: AzureMonitorScheduledQueryAlertIdentityType_USER_ASSIGNED,
				UserAssignedIdentityIds: []*foreignkeyv1.StringValueOrRef{
					literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai"),
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a query time range override and target resource types", func() {
			input := buildValidQueryAlert()
			input.Spec.QueryTimeRangeOverride = "P2D"
			input.Spec.TargetResourceTypes = []string{"Microsoft.Compute/virtualMachines"}
			input.Spec.Criteria[0].ResourceIdColumn = "_ResourceId"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept every evaluation frequency and window duration", func() {
			input := buildValidQueryAlert()
			for _, f := range []string{"PT1M", "PT5M", "PT10M", "PT15M", "PT30M", "PT45M", "PT1H", "PT2H", "PT3H", "PT4H", "PT5H", "PT6H", "P1D"} {
				input.Spec.EvaluationFrequency = proto.String(f)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
			for _, w := range []string{"PT1M", "PT5M", "PT10M", "PT15M", "PT30M", "PT45M", "PT1H", "PT2H", "PT3H", "PT4H", "PT5H", "PT6H", "P1D", "P2D"} {
				input.Spec.WindowDuration = proto.String(w)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing region", func() {
			input := buildValidQueryAlert()
			input.Spec.Region = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing scope", func() {
			input := buildValidQueryAlert()
			input.Spec.Scope = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject empty criteria", func() {
			input := buildValidQueryAlert()
			input.Spec.Criteria = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a criterion without a query", func() {
			input := buildValidQueryAlert()
			input.Spec.Criteria[0].Query = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a criterion without an aggregation", func() {
			input := buildValidQueryAlert()
			input.Spec.Criteria[0].TimeAggregationMethod = AzureMonitorScheduledQueryAlertTimeAggregation_azure_monitor_scheduled_query_alert_time_aggregation_unspecified
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a criterion without an operator", func() {
			input := buildValidQueryAlert()
			input.Spec.Criteria[0].Operator = AzureMonitorScheduledQueryAlertOperator_azure_monitor_scheduled_query_alert_operator_unspecified
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid evaluation frequency", func() {
			input := buildValidQueryAlert()
			input.Spec.EvaluationFrequency = proto.String("PT2M")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid window duration", func() {
			input := buildValidQueryAlert()
			input.Spec.WindowDuration = proto.String("P3D")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid mute duration", func() {
			input := buildValidQueryAlert()
			input.Spec.MuteActionsAfterAlertDuration = "PT20M"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject auto-mitigation combined with a mute duration", func() {
			input := buildValidQueryAlert()
			input.Spec.AutoMitigationEnabled = true
			input.Spec.MuteActionsAfterAlertDuration = "PT30M"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("mutually exclusive"))
		})

		ginkgo.It("should reject an invalid query time range override", func() {
			input := buildValidQueryAlert()
			input.Spec.QueryTimeRangeOverride = "PT7M"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a severity above 4", func() {
			input := buildValidQueryAlert()
			input.Spec.Severity = proto.Int32(5)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject failing periods where the minimum exceeds the examined count", func() {
			input := buildValidQueryAlert()
			input.Spec.Criteria[0].FailingPeriods = &AzureMonitorScheduledQueryAlertFailingPeriods{
				MinimumFailingPeriodsToTriggerAlert: 5,
				NumberOfEvaluationPeriods:           3,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("cannot exceed"))
		})

		ginkgo.It("should reject failing periods outside 1-6", func() {
			input := buildValidQueryAlert()
			input.Spec.Criteria[0].FailingPeriods = &AzureMonitorScheduledQueryAlertFailingPeriods{
				MinimumFailingPeriodsToTriggerAlert: 1,
				NumberOfEvaluationPeriods:           7,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a dimension without values", func() {
			input := buildValidQueryAlert()
			input.Spec.Criteria[0].Dimensions = []*AzureMonitorScheduledQueryAlertDimension{
				{Name: "AppRoleName", Operator: AzureMonitorScheduledQueryAlertDimensionOperator_INCLUDE},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject USER_ASSIGNED identity without identity ids", func() {
			input := buildValidQueryAlert()
			input.Spec.Identity = &AzureMonitorScheduledQueryAlertIdentity{
				Type: AzureMonitorScheduledQueryAlertIdentityType_USER_ASSIGNED,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an empty target resource type entry", func() {
			input := buildValidQueryAlert()
			input.Spec.TargetResourceTypes = []string{""}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
