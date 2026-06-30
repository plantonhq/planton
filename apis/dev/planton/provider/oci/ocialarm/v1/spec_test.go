package ocialarmv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestOciAlarmSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciAlarmSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidAlarm() *OciAlarm {
	return &OciAlarm{
		ApiVersion: "oci.planton.dev/v1",
		Kind:       "OciAlarm",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-alarm",
		},
		Spec: &OciAlarmSpec{
			CompartmentId:       newStringValueOrRef("ocid1.compartment.oc1..example"),
			MetricCompartmentId: newStringValueOrRef("ocid1.compartment.oc1..metrics"),
			Namespace:           "oci_computeagent",
			Query:               "CpuUtilization[5m].mean() > 80",
			Severity:            OciAlarmSpec_critical,
			Destinations:        []string{"ocid1.onstopic.oc1..example"},
			IsEnabled:           true,
		},
	}
}

var _ = ginkgo.Describe("OciAlarmSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_alarm", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := minimalValidAlarm()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with alarm disabled", func() {
				input := minimalValidAlarm()
				input.Spec.IsEnabled = false
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all severity values", func() {
				for _, sev := range []OciAlarmSpec_Severity{
					OciAlarmSpec_critical,
					OciAlarmSpec_error,
					OciAlarmSpec_warning,
					OciAlarmSpec_info,
				} {
					input := minimalValidAlarm()
					input.Spec.Severity = sev
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should not return a validation error with notification body", func() {
				input := minimalValidAlarm()
				input.Spec.Body = "CPU utilization exceeded threshold on {{resourceId}}"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with alarm summary", func() {
				input := minimalValidAlarm()
				input.Spec.AlarmSummary = "High CPU on compute instances"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with notification title", func() {
				input := minimalValidAlarm()
				input.Spec.NotificationTitle = "ALERT: High CPU Utilization"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with pending duration", func() {
				input := minimalValidAlarm()
				input.Spec.PendingDuration = "PT5M"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with evaluation slack duration", func() {
				input := minimalValidAlarm()
				input.Spec.EvaluationSlackDuration = "PT10M"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with repeat notification duration", func() {
				input := minimalValidAlarm()
				input.Spec.RepeatNotificationDuration = "PT30M"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with pretty_json message format", func() {
				input := minimalValidAlarm()
				input.Spec.MessageFormat = OciAlarmSpec_pretty_json
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with ons_optimized message format", func() {
				input := minimalValidAlarm()
				input.Spec.MessageFormat = OciAlarmSpec_ons_optimized
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with subtree monitoring enabled", func() {
				input := minimalValidAlarm()
				subtree := true
				input.Spec.MetricCompartmentIdInSubtree = &subtree
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with per-dimension notifications enabled", func() {
				input := minimalValidAlarm()
				perDim := true
				input.Spec.IsNotificationsPerMetricDimensionEnabled = &perDim
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with resource group", func() {
				input := minimalValidAlarm()
				input.Spec.ResourceGroup = "prod-instances"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with overrides", func() {
				input := minimalValidAlarm()
				input.Spec.RuleName = "BASE"
				input.Spec.Overrides = []*OciAlarmSpec_AlarmOverride{
					{
						RuleName: "critical-override",
						Query:    "CpuUtilization[5m].mean() > 95",
						Severity: OciAlarmSpec_critical,
					},
					{
						RuleName: "warning-override",
						Query:    "CpuUtilization[5m].mean() > 70",
						Severity: OciAlarmSpec_warning,
						Body:     "CPU approaching threshold",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple destinations", func() {
				input := minimalValidAlarm()
				input.Spec.Destinations = []string{
					"ocid1.onstopic.oc1..topic1",
					"ocid1.stream.oc1..stream1",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via valueFrom ref", func() {
				input := minimalValidAlarm()
				input.Spec.CompartmentId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-compartment",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with metric_compartment_id via valueFrom ref", func() {
				input := minimalValidAlarm()
				input.Spec.MetricCompartmentId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "metrics-compartment",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full configuration", func() {
				input := minimalValidAlarm()
				input.Spec.Body = "CPU alert: {{severity}} on {{resourceId}}"
				input.Spec.AlarmSummary = "High CPU utilization detected"
				input.Spec.NotificationTitle = "OCI Alert"
				input.Spec.PendingDuration = "PT5M"
				input.Spec.EvaluationSlackDuration = "PT10M"
				input.Spec.RepeatNotificationDuration = "PT1H"
				input.Spec.MessageFormat = OciAlarmSpec_pretty_json
				subtree := false
				input.Spec.MetricCompartmentIdInSubtree = &subtree
				perDim := true
				input.Spec.IsNotificationsPerMetricDimensionEnabled = &perDim
				input.Spec.ResourceGroup = "prod"
				input.Spec.NotificationVersion = "1.X"
				input.Spec.RuleName = "BASE"
				input.Spec.Overrides = []*OciAlarmSpec_AlarmOverride{
					{
						RuleName:        "crit",
						Query:           "CpuUtilization[5m].mean() > 95",
						Severity:        OciAlarmSpec_critical,
						Body:            "Critical CPU",
						PendingDuration: "PT1M",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_alarm", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidAlarm()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidAlarm()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidAlarm()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciAlarm{
					ApiVersion: "oci.planton.dev/v1",
					Kind:       "OciAlarm",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-alarm"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidAlarm()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metric_compartment_id is missing", func() {
				input := minimalValidAlarm()
				input.Spec.MetricCompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when namespace is empty", func() {
				input := minimalValidAlarm()
				input.Spec.Namespace = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when query is empty", func() {
				input := minimalValidAlarm()
				input.Spec.Query = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when severity is unspecified", func() {
				input := minimalValidAlarm()
				input.Spec.Severity = OciAlarmSpec_unspecified
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when destinations is empty", func() {
				input := minimalValidAlarm()
				input.Spec.Destinations = []string{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when override rule_name is empty", func() {
				input := minimalValidAlarm()
				input.Spec.Overrides = []*OciAlarmSpec_AlarmOverride{
					{
						RuleName: "",
						Query:    "CpuUtilization[5m].mean() > 95",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is empty", func() {
				input := minimalValidAlarm()
				input.ApiVersion = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is empty", func() {
				input := minimalValidAlarm()
				input.Kind = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

		})
	})
})
