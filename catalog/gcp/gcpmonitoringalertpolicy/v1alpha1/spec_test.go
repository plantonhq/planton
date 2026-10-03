package gcpmonitoringalertpolicyv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpMonitoringAlertPolicySpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

// thresholdCondition builds the workhorse condition shape used across cases.
func thresholdCondition() *GcpMonitoringAlertPolicyCondition {
	return &GcpMonitoringAlertPolicyCondition{
		DisplayName: "cpu above 80%",
		ConditionThreshold: &GcpMonitoringAlertPolicyConditionThreshold{
			Filter:         `metric.type="compute.googleapis.com/instance/cpu/utilization" AND resource.type="gce_instance"`,
			Comparison:     "COMPARISON_GT",
			ThresholdValue: 0.8,
			Duration:       "300s",
			Aggregations: []*GcpMonitoringAlertPolicyAggregation{{
				AlignmentPeriod:  "60s",
				PerSeriesAligner: "ALIGN_MEAN",
			}},
		},
	}
}

var _ = ginkgo.Describe("GcpMonitoringAlertPolicySpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpMonitoringAlertPolicy {
		return &GcpMonitoringAlertPolicy{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpMonitoringAlertPolicy",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-alert-policy",
			},
			Spec: &GcpMonitoringAlertPolicySpec{
				Combiner:   "OR",
				Conditions: []*GcpMonitoringAlertPolicyCondition{thresholdCondition()},
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal threshold policy", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept each combiner", func() {
		for _, v := range []string{"AND", "OR", "AND_WITH_MATCHING_RESOURCE"} {
			target := minimal()
			target.Spec.Combiner = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should accept each severity", func() {
		for _, v := range []string{"CRITICAL", "ERROR", "WARNING"} {
			target := minimal()
			target.Spec.Severity = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should accept notification channel references", func() {
		target := minimal()
		target.Spec.NotificationChannels = []*foreignkeyv1.StringValueOrRef{
			litRef("projects/p/notificationChannels/123"),
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an absence condition", func() {
		target := minimal()
		target.Spec.Conditions = []*GcpMonitoringAlertPolicyCondition{{
			DisplayName: "heartbeat gone quiet",
			ConditionAbsent: &GcpMonitoringAlertPolicyConditionAbsent{
				Filter:   `metric.type="custom.googleapis.com/heartbeat"`,
				Duration: "300s",
			},
		}}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a log-match condition with rate limit", func() {
		target := minimal()
		target.Spec.Conditions = []*GcpMonitoringAlertPolicyCondition{{
			DisplayName: "panic logged",
			ConditionMatchedLog: &GcpMonitoringAlertPolicyConditionMatchedLog{
				Filter:          `severity>=ERROR AND textPayload:"panic"`,
				LabelExtractors: map[string]string{"vm": "EXTRACT(resource.labels.instance_id)"},
			},
		}}
		target.Spec.AlertStrategy = &GcpMonitoringAlertPolicyAlertStrategy{
			NotificationRateLimit: &GcpMonitoringAlertPolicyNotificationRateLimit{Period: "300s"},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a PromQL condition", func() {
		target := minimal()
		target.Spec.Conditions = []*GcpMonitoringAlertPolicyCondition{{
			DisplayName: "5xx rate",
			ConditionPrometheusQueryLanguage: &GcpMonitoringAlertPolicyConditionPromql{
				Query:              `rate(http_requests_total{code=~"5.."}[5m]) > 0.1`,
				Duration:           "300s",
				EvaluationInterval: "30s",
			},
		}}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a SQL condition with one schedule and one test", func() {
		target := minimal()
		target.Spec.Conditions = []*GcpMonitoringAlertPolicyCondition{{
			DisplayName: "error spike in logs",
			ConditionSql: &GcpMonitoringAlertPolicyConditionSql{
				Query:   "SELECT COUNT(*) FROM my_project.global._Default._AllLogs WHERE severity = 'ERROR'",
				Minutes: &GcpMonitoringAlertPolicySqlMinutes{Periodicity: 15},
				RowCountTest: &GcpMonitoringAlertPolicySqlRowCountTest{
					Comparison: "COMPARISON_GT",
					Threshold:  100,
				},
			},
		}}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a ratio threshold with forecast and trigger", func() {
		target := minimal()
		cond := thresholdCondition()
		cond.ConditionThreshold.DenominatorFilter = `metric.type="serviceruntime.googleapis.com/api/request_count"`
		cond.ConditionThreshold.DenominatorAggregations = []*GcpMonitoringAlertPolicyAggregation{{
			AlignmentPeriod:  "60s",
			PerSeriesAligner: "ALIGN_MEAN",
		}}
		cond.ConditionThreshold.ForecastOptions = &GcpMonitoringAlertPolicyForecastOptions{ForecastHorizon: "3600s"}
		cond.ConditionThreshold.Trigger = &GcpMonitoringAlertPolicyTrigger{Count: 2}
		cond.ConditionThreshold.EvaluationMissingData = "EVALUATION_MISSING_DATA_INACTIVE"
		target.Spec.Conditions = []*GcpMonitoringAlertPolicyCondition{cond}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept alert strategy, documentation, and labels", func() {
		target := minimal()
		target.Spec.AlertStrategy = &GcpMonitoringAlertPolicyAlertStrategy{
			AutoClose:           "1800s",
			NotificationPrompts: []string{"OPENED", "CLOSED"},
			NotificationChannelStrategy: []*GcpMonitoringAlertPolicyNotificationChannelStrategy{{
				NotificationChannelNames: []*foreignkeyv1.StringValueOrRef{litRef("projects/p/notificationChannels/123")},
				RenotifyInterval:         "3600s",
			}},
		}
		target.Spec.Documentation = &GcpMonitoringAlertPolicyDocumentation{
			Content:  "1. Check the dashboard. 2. Roll back the last deploy.",
			MimeType: "text/markdown",
			Subject:  "CPU saturation on ${resource.label.instance_id}",
			Links: []*GcpMonitoringAlertPolicyDocumentationLink{
				{DisplayName: "Runbook", Url: "https://wiki.example.com/runbooks/cpu"},
			},
		}
		target.Spec.Labels = map[string]string{"team": "platform"}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept each deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
			target := minimal()
			target.Spec.DeletionPolicy = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing combiner", func() {
		target := minimal()
		target.Spec.Combiner = ""
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an invalid combiner", func() {
		target := minimal()
		target.Spec.Combiner = "XOR"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "AND_WITH_MATCHING_RESOURCE")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject zero conditions and more than six", func() {
		none := minimal()
		none.Spec.Conditions = nil
		gomega.Expect(validator.Validate(none)).ToNot(gomega.Succeed())

		seven := minimal()
		seven.Spec.Conditions = nil
		for i := 0; i < 7; i++ {
			seven.Spec.Conditions = append(seven.Spec.Conditions, thresholdCondition())
		}
		gomega.Expect(validator.Validate(seven)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a condition without any type", func() {
		target := minimal()
		target.Spec.Conditions = []*GcpMonitoringAlertPolicyCondition{{DisplayName: "empty"}}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "exactly one")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a condition with two types", func() {
		target := minimal()
		cond := thresholdCondition()
		cond.ConditionAbsent = &GcpMonitoringAlertPolicyConditionAbsent{
			Filter:   "metric.type=\"x\"",
			Duration: "300s",
		}
		target.Spec.Conditions = []*GcpMonitoringAlertPolicyCondition{cond}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an invalid comparison", func() {
		target := minimal()
		target.Spec.Conditions[0].ConditionThreshold.Comparison = "GREATER"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an invalid aligner and reducer", func() {
		badAligner := minimal()
		badAligner.Spec.Conditions[0].ConditionThreshold.Aggregations[0].PerSeriesAligner = "ALIGN_AVG"
		gomega.Expect(validator.Validate(badAligner)).ToNot(gomega.Succeed())

		badReducer := minimal()
		badReducer.Spec.Conditions[0].ConditionThreshold.Aggregations[0].CrossSeriesReducer = "REDUCE_AVG"
		gomega.Expect(validator.Validate(badReducer)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a SQL condition with two schedules or two tests", func() {
		twoSchedules := minimal()
		twoSchedules.Spec.Conditions = []*GcpMonitoringAlertPolicyCondition{{
			DisplayName: "bad",
			ConditionSql: &GcpMonitoringAlertPolicyConditionSql{
				Query:        "SELECT 1",
				Minutes:      &GcpMonitoringAlertPolicySqlMinutes{Periodicity: 15},
				Hourly:       &GcpMonitoringAlertPolicySqlHourly{Periodicity: 1},
				RowCountTest: &GcpMonitoringAlertPolicySqlRowCountTest{Comparison: "COMPARISON_GT", Threshold: 1},
			},
		}}
		gomega.Expect(validator.Validate(twoSchedules)).ToNot(gomega.Succeed())

		twoTests := minimal()
		twoTests.Spec.Conditions = []*GcpMonitoringAlertPolicyCondition{{
			DisplayName: "bad",
			ConditionSql: &GcpMonitoringAlertPolicyConditionSql{
				Query:        "SELECT 1",
				Minutes:      &GcpMonitoringAlertPolicySqlMinutes{Periodicity: 15},
				RowCountTest: &GcpMonitoringAlertPolicySqlRowCountTest{Comparison: "COMPARISON_GT", Threshold: 1},
				BooleanTest:  &GcpMonitoringAlertPolicySqlBooleanTest{Column: "is_bad"},
			},
		}}
		gomega.Expect(validator.Validate(twoTests)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a trigger setting both count and percent", func() {
		target := minimal()
		target.Spec.Conditions[0].ConditionThreshold.Trigger = &GcpMonitoringAlertPolicyTrigger{
			Count:   2,
			Percent: 50,
		}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an invalid notification prompt", func() {
		target := minimal()
		target.Spec.AlertStrategy = &GcpMonitoringAlertPolicyAlertStrategy{
			NotificationPrompts: []string{"REOPENED"},
		}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a non-markdown documentation mime type", func() {
		target := minimal()
		target.Spec.Documentation = &GcpMonitoringAlertPolicyDocumentation{
			Content:  "runbook",
			MimeType: "text/plain",
		}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject more than three documentation links", func() {
		target := minimal()
		links := make([]*GcpMonitoringAlertPolicyDocumentationLink, 4)
		for i := range links {
			links[i] = &GcpMonitoringAlertPolicyDocumentationLink{Url: "https://example.com"}
		}
		target.Spec.Documentation = &GcpMonitoringAlertPolicyDocumentation{Links: links}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})
})
