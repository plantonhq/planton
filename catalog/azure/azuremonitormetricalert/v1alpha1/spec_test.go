package azuremonitormetricalertv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureMonitorMetricAlertSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureMonitorMetricAlertSpec Validation Tests")
}

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

// buildValidMetricAlert returns a minimal valid resource (one static
// criterion on a storage account scope); tests mutate copies of it.
func buildValidMetricAlert() *AzureMonitorMetricAlert {
	return &AzureMonitorMetricAlert{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureMonitorMetricAlert",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-metric-alert",
		},
		Spec: &AzureMonitorMetricAlertSpec{
			ResourceGroup: literal("test-resource-group"),
			AlertName:     "storage-availability",
			Scopes: []*foreignkeyv1.StringValueOrRef{
				literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/sa"),
			},
			StaticCriteria: []*AzureMonitorMetricAlertStaticCriteria{
				{
					MetricNamespace: "Microsoft.Storage/storageAccounts",
					MetricName:      "Availability",
					Aggregation:     AzureMonitorMetricAlertAggregation_AVERAGE,
					Operator:        AzureMonitorMetricAlertOperator_LESS_THAN,
					Threshold:       99.9,
				},
			},
			Actions: []*AzureMonitorMetricAlertAction{
				{ActionGroupId: literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Insights/actionGroups/oncall")},
			},
		},
	}
}

var _ = ginkgo.Describe("AzureMonitorMetricAlertSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should not return a validation error for minimal valid fields", func() {
			err := protovalidate.Validate(buildValidMetricAlert())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a full static configuration with dimensions", func() {
			input := buildValidMetricAlert()
			input.Spec.Description = "storage availability dropped"
			input.Spec.Severity = proto.Int32(1)
			input.Spec.Frequency = proto.String("PT5M")
			input.Spec.WindowSize = proto.String("PT15M")
			input.Spec.AutoMitigate = proto.Bool(true)
			input.Spec.StaticCriteria[0].Dimensions = []*AzureMonitorMetricAlertDimension{
				{
					Name:     "ApiName",
					Operator: AzureMonitorMetricAlertDimensionOperator_INCLUDE,
					Values:   []string{"GetBlob", "PutBlob"},
				},
			}
			input.Spec.Tags = map[string]string{"team": "platform"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a zero threshold", func() {
			input := buildValidMetricAlert()
			input.Spec.StaticCriteria[0].Operator = AzureMonitorMetricAlertOperator_GREATER_THAN
			input.Spec.StaticCriteria[0].Threshold = 0
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept dynamic criteria", func() {
			input := buildValidMetricAlert()
			input.Spec.StaticCriteria = nil
			input.Spec.DynamicCriteria = &AzureMonitorMetricAlertDynamicCriteria{
				MetricNamespace:        "Microsoft.Storage/storageAccounts",
				MetricName:             "Transactions",
				Aggregation:            AzureMonitorMetricAlertAggregation_TOTAL,
				Operator:               AzureMonitorMetricAlertOperator_GREATER_OR_LESS_THAN,
				AlertSensitivity:       AzureMonitorMetricAlertSensitivity_MEDIUM,
				EvaluationTotalCount:   proto.Int32(4),
				EvaluationFailureCount: proto.Int32(3),
				IgnoreDataBefore:       "2026-01-15T00:00:00Z",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept web-test availability criteria", func() {
			input := buildValidMetricAlert()
			input.Spec.StaticCriteria = nil
			input.Spec.WebTestAvailabilityCriteria = &AzureMonitorMetricAlertWebTestCriteria{
				WebTestId:           literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Insights/webTests/homepage"),
				ComponentId:         literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Insights/components/app"),
				FailedLocationCount: 3,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept multi-scope with target type and location", func() {
			input := buildValidMetricAlert()
			input.Spec.Scopes = append(input.Spec.Scopes,
				literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/sb"))
			input.Spec.TargetResourceType = "Microsoft.Storage/storageAccounts"
			input.Spec.TargetResourceLocation = "eastus"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept every frequency and window value", func() {
			input := buildValidMetricAlert()
			for _, f := range []string{"PT1M", "PT5M", "PT15M", "PT30M", "PT1H"} {
				input.Spec.Frequency = proto.String(f)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
			for _, w := range []string{"PT1M", "PT5M", "PT15M", "PT30M", "PT1H", "PT6H", "PT12H", "P1D"} {
				input.Spec.WindowSize = proto.String(w)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing alert name", func() {
			input := buildValidMetricAlert()
			input.Spec.AlertName = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject empty scopes", func() {
			input := buildValidMetricAlert()
			input.Spec.Scopes = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a rule with no condition family", func() {
			input := buildValidMetricAlert()
			input.Spec.StaticCriteria = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("exactly one condition family"))
		})

		ginkgo.It("should reject a rule with two condition families", func() {
			input := buildValidMetricAlert()
			input.Spec.DynamicCriteria = &AzureMonitorMetricAlertDynamicCriteria{
				MetricNamespace:  "Microsoft.Storage/storageAccounts",
				MetricName:       "Transactions",
				Aggregation:      AzureMonitorMetricAlertAggregation_TOTAL,
				Operator:         AzureMonitorMetricAlertOperator_GREATER_THAN,
				AlertSensitivity: AzureMonitorMetricAlertSensitivity_MEDIUM,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject the dynamic-only operator on static criteria", func() {
			input := buildValidMetricAlert()
			input.Spec.StaticCriteria[0].Operator = AzureMonitorMetricAlertOperator_GREATER_OR_LESS_THAN
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("static criteria"))
		})

		ginkgo.It("should reject a static-only operator on dynamic criteria", func() {
			input := buildValidMetricAlert()
			input.Spec.StaticCriteria = nil
			input.Spec.DynamicCriteria = &AzureMonitorMetricAlertDynamicCriteria{
				MetricNamespace:  "Microsoft.Storage/storageAccounts",
				MetricName:       "Transactions",
				Aggregation:      AzureMonitorMetricAlertAggregation_TOTAL,
				Operator:         AzureMonitorMetricAlertOperator_EQUALS,
				AlertSensitivity: AzureMonitorMetricAlertSensitivity_MEDIUM,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("dynamic criteria"))
		})

		ginkgo.It("should reject an invalid frequency", func() {
			input := buildValidMetricAlert()
			input.Spec.Frequency = proto.String("PT2M")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid window size", func() {
			input := buildValidMetricAlert()
			input.Spec.WindowSize = proto.String("P2D")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a severity above 4", func() {
			input := buildValidMetricAlert()
			input.Spec.Severity = proto.Int32(5)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a dimension without values", func() {
			input := buildValidMetricAlert()
			input.Spec.StaticCriteria[0].Dimensions = []*AzureMonitorMetricAlertDimension{
				{Name: "ApiName", Operator: AzureMonitorMetricAlertDimensionOperator_INCLUDE},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a dimension without an operator", func() {
			input := buildValidMetricAlert()
			input.Spec.StaticCriteria[0].Dimensions = []*AzureMonitorMetricAlertDimension{
				{Name: "ApiName", Values: []string{"GetBlob"}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed ignore_data_before timestamp", func() {
			input := buildValidMetricAlert()
			input.Spec.StaticCriteria = nil
			input.Spec.DynamicCriteria = &AzureMonitorMetricAlertDynamicCriteria{
				MetricNamespace:  "Microsoft.Storage/storageAccounts",
				MetricName:       "Transactions",
				Aggregation:      AzureMonitorMetricAlertAggregation_TOTAL,
				Operator:         AzureMonitorMetricAlertOperator_GREATER_THAN,
				AlertSensitivity: AzureMonitorMetricAlertSensitivity_MEDIUM,
				IgnoreDataBefore: "January 15, 2026",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("RFC 3339"))
		})

		ginkgo.It("should reject a zero failed location count on web-test criteria", func() {
			input := buildValidMetricAlert()
			input.Spec.StaticCriteria = nil
			input.Spec.WebTestAvailabilityCriteria = &AzureMonitorMetricAlertWebTestCriteria{
				WebTestId:           literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Insights/webTests/homepage"),
				ComponentId:         literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Insights/components/app"),
				FailedLocationCount: 0,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an action without an action group", func() {
			input := buildValidMetricAlert()
			input.Spec.Actions = []*AzureMonitorMetricAlertAction{{}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
