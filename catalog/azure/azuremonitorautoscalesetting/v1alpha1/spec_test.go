package azuremonitorautoscalesettingv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureMonitorAutoscaleSettingSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureMonitorAutoscaleSettingSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func int32Ptr(v int32) *int32 { return &v }

const testTargetId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Web/serverFarms/app-plan"

// defaultProfile returns a valid rule-less default profile.
func defaultProfile() *AzureMonitorAutoscaleSettingProfile {
	return &AzureMonitorAutoscaleSettingProfile{
		Name: "default",
		Capacity: &AzureMonitorAutoscaleSettingCapacity{
			Minimum: int32Ptr(1),
			Maximum: int32Ptr(4),
			Default: int32Ptr(1),
		},
	}
}

// cpuRule returns a valid scale-out rule on CPU percentage.
func cpuRule() *AzureMonitorAutoscaleSettingRule {
	return &AzureMonitorAutoscaleSettingRule{
		MetricTrigger: &AzureMonitorAutoscaleSettingMetricTrigger{
			MetricName:       "CpuPercentage",
			MetricResourceId: literal(testTargetId),
			TimeGrain:        "PT1M",
			Statistic:        "Average",
			TimeWindow:       "PT10M",
			TimeAggregation:  "Average",
			Operator:         "GreaterThan",
			Threshold:        75,
		},
		ScaleAction: &AzureMonitorAutoscaleSettingScaleAction{
			Direction: "Increase",
			Type:      "ChangeCount",
			Value:     int32Ptr(1),
			Cooldown:  "PT5M",
		},
	}
}

// validResource returns a minimal valid setting (one default profile)
// that individual cases mutate into the shape under test.
func validResource() *AzureMonitorAutoscaleSetting {
	return &AzureMonitorAutoscaleSetting{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureMonitorAutoscaleSetting",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-autoscale",
		},
		Spec: &AzureMonitorAutoscaleSettingSpec{
			ResourceGroup:    literal("app-rg"),
			Name:             "app-autoscale",
			Region:           "eastus",
			TargetResourceId: literal(testTargetId),
			Profiles:         []*AzureMonitorAutoscaleSettingProfile{defaultProfile()},
		},
	}
}

var _ = ginkgo.Describe("AzureMonitorAutoscaleSettingSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_monitor_autoscale_setting", func() {

			ginkgo.It("should not return a validation error for the minimal default-profile setting", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a profile carrying metric rules", func() {
				input := validResource()
				input.Spec.Profiles[0].Rules = []*AzureMonitorAutoscaleSettingRule{cpuRule()}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept every statistic token", func() {
				for _, statistic := range []string{"Average", "Max", "Min", "Sum"} {
					input := validResource()
					rule := cpuRule()
					rule.MetricTrigger.Statistic = statistic
					input.Spec.Profiles[0].Rules = []*AzureMonitorAutoscaleSettingRule{rule}
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil(), "statistic %q should be valid", statistic)
				}
			})

			ginkgo.It("should accept every time aggregation token", func() {
				for _, aggregation := range []string{"Average", "Count", "Maximum", "Minimum", "Total", "Last"} {
					input := validResource()
					rule := cpuRule()
					rule.MetricTrigger.TimeAggregation = aggregation
					input.Spec.Profiles[0].Rules = []*AzureMonitorAutoscaleSettingRule{rule}
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil(), "aggregation %q should be valid", aggregation)
				}
			})

			ginkgo.It("should accept every comparison operator token", func() {
				for _, operator := range []string{"Equals", "NotEquals", "GreaterThan", "GreaterThanOrEqual", "LessThan", "LessThanOrEqual"} {
					input := validResource()
					rule := cpuRule()
					rule.MetricTrigger.Operator = operator
					input.Spec.Profiles[0].Rules = []*AzureMonitorAutoscaleSettingRule{rule}
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil(), "operator %q should be valid", operator)
				}
			})

			ginkgo.It("should accept every scale action type in both directions", func() {
				for _, direction := range []string{"Increase", "Decrease"} {
					for _, actionType := range []string{"ChangeCount", "ExactCount", "PercentChangeCount", "ServiceAllowedNextValue"} {
						input := validResource()
						rule := cpuRule()
						rule.ScaleAction.Direction = direction
						rule.ScaleAction.Type = actionType
						input.Spec.Profiles[0].Rules = []*AzureMonitorAutoscaleSettingRule{rule}
						err := protovalidate.Validate(input)
						gomega.Expect(err).To(gomega.BeNil(), "direction %q type %q should be valid", direction, actionType)
					}
				}
			})

			ginkgo.It("should accept zero-valued capacity bounds and a zero scale-action value", func() {
				input := validResource()
				input.Spec.Profiles[0].Capacity.Minimum = int32Ptr(0)
				input.Spec.Profiles[0].Capacity.Default = int32Ptr(0)
				rule := cpuRule()
				rule.ScaleAction.Value = int32Ptr(0)
				input.Spec.Profiles[0].Rules = []*AzureMonitorAutoscaleSettingRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a metric trigger with dimensions and per-instance division", func() {
				input := validResource()
				rule := cpuRule()
				rule.MetricTrigger.DivideByInstanceCount = true
				rule.MetricTrigger.MetricNamespace = "Microsoft.Web/serverFarms"
				rule.MetricTrigger.Dimensions = []*AzureMonitorAutoscaleSettingDimension{
					{Name: "Instance", Operator: "Equals", Values: []string{"worker-0", "worker-1"}},
					{Name: "Status", Operator: "NotEquals", Values: []string{"Stopped"}},
				}
				input.Spec.Profiles[0].Rules = []*AzureMonitorAutoscaleSettingRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a recurrence profile alongside the default profile", func() {
				input := validResource()
				weekday := defaultProfile()
				weekday.Name = "weekday-business-hours"
				weekday.Recurrence = &AzureMonitorAutoscaleSettingRecurrence{
					Days:   []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"},
					Hour:   int32Ptr(8),
					Minute: int32Ptr(0),
				}
				input.Spec.Profiles = append(input.Spec.Profiles, weekday)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a recurrence timezone from the Azure vocabulary", func() {
				input := validResource()
				tz := "Eastern Standard Time"
				input.Spec.Profiles[0].Recurrence = &AzureMonitorAutoscaleSettingRecurrence{
					Timezone: &tz,
					Days:     []string{"Saturday", "Sunday"},
					Hour:     int32Ptr(0),
					Minute:   int32Ptr(30),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a fixed-date profile", func() {
				input := validResource()
				input.Spec.Profiles[0].FixedDate = &AzureMonitorAutoscaleSettingFixedDate{
					Start: "2026-11-27T00:00:00Z",
					End:   "2026-11-30T00:00:00Z",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept both predictive modes", func() {
				for _, mode := range []string{"Enabled", "ForecastOnly"} {
					input := validResource()
					input.Spec.Predictive = &AzureMonitorAutoscaleSettingPredictive{
						ScaleMode:     mode,
						LookAheadTime: "PT10M",
					}
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil(), "predictive mode %q should be valid", mode)
				}
			})

			ginkgo.It("should accept email-only, webhook-only, and combined notifications", func() {
				email := &AzureMonitorAutoscaleSettingNotificationEmail{
					CustomEmails: []string{"oncall@example.com"},
				}
				webhook := &AzureMonitorAutoscaleSettingNotificationWebhook{
					ServiceUri: "https://hooks.example.com/scale",
					Properties: map[string]string{"channel": "ops"},
				}
				for _, notification := range []*AzureMonitorAutoscaleSettingNotification{
					{Email: email},
					{Webhooks: []*AzureMonitorAutoscaleSettingNotificationWebhook{webhook}},
					{Email: email, Webhooks: []*AzureMonitorAutoscaleSettingNotificationWebhook{webhook}},
				} {
					input := validResource()
					input.Spec.Notification = notification
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("root fields", func() {

			ginkgo.It("should reject a missing resource group", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty name", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty region", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing target resource", func() {
				input := validResource()
				input.Spec.TargetResourceId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty profiles list", func() {
				input := validResource()
				input.Spec.Profiles = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject more than 20 profiles", func() {
				input := validResource()
				for i := 0; i < 20; i++ {
					extra := defaultProfile()
					extra.Name = "profile-" + strings.Repeat("x", i+1)
					input.Spec.Profiles = append(input.Spec.Profiles, extra)
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("profiles", func() {

			ginkgo.It("should reject a profile without capacity", func() {
				input := validResource()
				input.Spec.Profiles[0].Capacity = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject capacity bounds outside 0-1000", func() {
				for _, bad := range []int32{-1, 1001} {
					input := validResource()
					input.Spec.Profiles[0].Capacity.Maximum = int32Ptr(bad)
					err := protovalidate.Validate(input)
					gomega.Expect(err).NotTo(gomega.BeNil(), "maximum %d should be rejected", bad)
				}
			})

			ginkgo.It("should reject a capacity block missing its minimum", func() {
				input := validResource()
				input.Spec.Profiles[0].Capacity.Minimum = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject more than 10 rules in one profile", func() {
				input := validResource()
				for i := 0; i < 11; i++ {
					input.Spec.Profiles[0].Rules = append(input.Spec.Profiles[0].Rules, cpuRule())
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a profile carrying BOTH fixed_date and recurrence", func() {
				input := validResource()
				input.Spec.Profiles[0].FixedDate = &AzureMonitorAutoscaleSettingFixedDate{
					Start: "2026-11-27T00:00:00Z",
					End:   "2026-11-30T00:00:00Z",
				}
				input.Spec.Profiles[0].Recurrence = &AzureMonitorAutoscaleSettingRecurrence{
					Days:   []string{"Monday"},
					Hour:   int32Ptr(8),
					Minute: int32Ptr(0),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("at most one of fixed_date and recurrence"))
			})
		})

		ginkgo.Context("rules", func() {

			ginkgo.It("should reject an unknown statistic token", func() {
				input := validResource()
				rule := cpuRule()
				rule.MetricTrigger.Statistic = "Median"
				input.Spec.Profiles[0].Rules = []*AzureMonitorAutoscaleSettingRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown time aggregation token", func() {
				input := validResource()
				rule := cpuRule()
				rule.MetricTrigger.TimeAggregation = "Max"
				input.Spec.Profiles[0].Rules = []*AzureMonitorAutoscaleSettingRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown comparison operator token", func() {
				input := validResource()
				rule := cpuRule()
				rule.MetricTrigger.Operator = "Above"
				input.Spec.Profiles[0].Rules = []*AzureMonitorAutoscaleSettingRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a non-ISO-8601 time grain", func() {
				input := validResource()
				rule := cpuRule()
				rule.MetricTrigger.TimeGrain = "5m"
				input.Spec.Profiles[0].Rules = []*AzureMonitorAutoscaleSettingRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown scale direction and type", func() {
				input := validResource()
				rule := cpuRule()
				rule.ScaleAction.Direction = "Up"
				input.Spec.Profiles[0].Rules = []*AzureMonitorAutoscaleSettingRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())

				input = validResource()
				rule = cpuRule()
				rule.ScaleAction.Type = "StepCount"
				input.Spec.Profiles[0].Rules = []*AzureMonitorAutoscaleSettingRule{rule}
				err = protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a negative scale-action value and a bad cooldown", func() {
				input := validResource()
				rule := cpuRule()
				rule.ScaleAction.Value = int32Ptr(-1)
				input.Spec.Profiles[0].Rules = []*AzureMonitorAutoscaleSettingRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())

				input = validResource()
				rule = cpuRule()
				rule.ScaleAction.Cooldown = "5 minutes"
				input.Spec.Profiles[0].Rules = []*AzureMonitorAutoscaleSettingRule{rule}
				err = protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a dimension with no values or a bad operator", func() {
				input := validResource()
				rule := cpuRule()
				rule.MetricTrigger.Dimensions = []*AzureMonitorAutoscaleSettingDimension{
					{Name: "Instance", Operator: "Equals", Values: nil},
				}
				input.Spec.Profiles[0].Rules = []*AzureMonitorAutoscaleSettingRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())

				input = validResource()
				rule = cpuRule()
				rule.MetricTrigger.Dimensions = []*AzureMonitorAutoscaleSettingDimension{
					{Name: "Instance", Operator: "StartsWith", Values: []string{"worker"}},
				}
				input.Spec.Profiles[0].Rules = []*AzureMonitorAutoscaleSettingRule{rule}
				err = protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("schedules", func() {

			ginkgo.It("should reject a recurrence with no days or an unknown day token", func() {
				input := validResource()
				input.Spec.Profiles[0].Recurrence = &AzureMonitorAutoscaleSettingRecurrence{
					Days:   nil,
					Hour:   int32Ptr(8),
					Minute: int32Ptr(0),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())

				input = validResource()
				input.Spec.Profiles[0].Recurrence = &AzureMonitorAutoscaleSettingRecurrence{
					Days:   []string{"Weekdays"},
					Hour:   int32Ptr(8),
					Minute: int32Ptr(0),
				}
				err = protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject out-of-range recurrence hour and minute", func() {
				input := validResource()
				input.Spec.Profiles[0].Recurrence = &AzureMonitorAutoscaleSettingRecurrence{
					Days:   []string{"Monday"},
					Hour:   int32Ptr(24),
					Minute: int32Ptr(0),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())

				input = validResource()
				input.Spec.Profiles[0].Recurrence = &AzureMonitorAutoscaleSettingRecurrence{
					Days:   []string{"Monday"},
					Hour:   int32Ptr(8),
					Minute: int32Ptr(60),
				}
				err = protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a recurrence missing its hour", func() {
				input := validResource()
				input.Spec.Profiles[0].Recurrence = &AzureMonitorAutoscaleSettingRecurrence{
					Days:   []string{"Monday"},
					Minute: int32Ptr(0),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a timezone outside the Azure vocabulary", func() {
				input := validResource()
				tz := "America/New_York"
				input.Spec.Profiles[0].Recurrence = &AzureMonitorAutoscaleSettingRecurrence{
					Timezone: &tz,
					Days:     []string{"Monday"},
					Hour:     int32Ptr(8),
					Minute:   int32Ptr(0),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a non-RFC-3339 fixed-date start", func() {
				input := validResource()
				input.Spec.Profiles[0].FixedDate = &AzureMonitorAutoscaleSettingFixedDate{
					Start: "2026-11-27",
					End:   "2026-11-30T00:00:00Z",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("predictive and notification", func() {

			ginkgo.It("should reject an unknown predictive mode and a bad look-ahead", func() {
				input := validResource()
				input.Spec.Predictive = &AzureMonitorAutoscaleSettingPredictive{ScaleMode: "Disabled"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())

				input = validResource()
				input.Spec.Predictive = &AzureMonitorAutoscaleSettingPredictive{
					ScaleMode:     "Enabled",
					LookAheadTime: "10 minutes",
				}
				err = protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty notification block", func() {
				input := validResource()
				input.Spec.Notification = &AzureMonitorAutoscaleSettingNotification{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("at least one notification channel"))
			})

			ginkgo.It("should reject an email block without recipients", func() {
				// The retired admin flags were the only other email wiring;
				// an empty email block configures nothing ARM can deliver.
				input := validResource()
				input.Spec.Notification = &AzureMonitorAutoscaleSettingNotification{
					Email: &AzureMonitorAutoscaleSettingNotificationEmail{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("custom_emails"))
			})

			ginkgo.It("should reject a webhook without an http(s) scheme", func() {
				input := validResource()
				input.Spec.Notification = &AzureMonitorAutoscaleSettingNotification{
					Webhooks: []*AzureMonitorAutoscaleSettingNotificationWebhook{
						{ServiceUri: "ftp://hooks.example.com/scale"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
