package awscloudwatchcompositealarmv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	fkv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsCloudwatchCompositeAlarmSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsCloudwatchCompositeAlarmSpec Validation Suite")
}

var _ = ginkgo.Describe("AwsCloudwatchCompositeAlarmSpec validations", func() {

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal composite alarm (region + alarm_rule)", func() {
		input := &AwsCloudwatchCompositeAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchCompositeAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "shared-cause",
			},
			Spec: &AwsCloudwatchCompositeAlarmSpec{
				Region:    "us-west-2",
				AlarmRule: `ALARM("cpu-high") AND ALARM("error-rate-high")`,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a composite alarm with OR/NOT and parentheses in the rule", func() {
		input := &AwsCloudwatchCompositeAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchCompositeAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "dependency-aware",
			},
			Spec: &AwsCloudwatchCompositeAlarmSpec{
				Region:    "us-west-2",
				AlarmRule: `(ALARM("api-5xx") OR ALARM("api-latency")) AND NOT ALARM("upstream-db")`,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a fully configured composite alarm (actions + suppressor)", func() {
		snsArn := &fkv1.StringValueOrRef{
			LiteralOrRef: &fkv1.StringValueOrRef_Value{
				Value: "arn:aws:sns:us-west-2:123456789012:oncall",
			},
		}
		input := &AwsCloudwatchCompositeAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchCompositeAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "prod-outage",
			},
			Spec: &AwsCloudwatchCompositeAlarmSpec{
				Region:           "us-west-2",
				AlarmRule:        `ALARM("cpu-high") AND ALARM("error-rate-high")`,
				AlarmDescription: "Shared-cause outage: both CPU and error rate are breaching. Page on-call once.",
				ActionsEnabled:   boolPtr(true),
				AlarmActions:     []*fkv1.StringValueOrRef{snsArn},
				OkActions:        []*fkv1.StringValueOrRef{snsArn},
				ActionsSuppressor: &AwsCloudwatchCompositeAlarmActionsSuppressor{
					Alarm: &fkv1.StringValueOrRef{
						LiteralOrRef: &fkv1.StringValueOrRef_ValueFrom{
							ValueFrom: &fkv1.ValueFromRef{
								Kind: catalogkind.CatalogKind_AwsCloudwatchAlarm,
								Name: "maintenance-window",
							},
						},
					},
					WaitPeriod:      60,
					ExtensionPeriod: 120,
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts actions_enabled = false (evaluation without actions)", func() {
		input := &AwsCloudwatchCompositeAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchCompositeAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "silent-composite",
			},
			Spec: &AwsCloudwatchCompositeAlarmSpec{
				Region:         "us-west-2",
				AlarmRule:      `ALARM("cpu-high")`,
				ActionsEnabled: boolPtr(false),
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Rejections
	// -------------------------------------------------------------------------

	ginkgo.It("fails when alarm_rule is missing", func() {
		input := &AwsCloudwatchCompositeAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchCompositeAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "no-rule",
			},
			Spec: &AwsCloudwatchCompositeAlarmSpec{
				Region: "us-west-2",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when region is missing", func() {
		input := &AwsCloudwatchCompositeAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchCompositeAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "no-region",
			},
			Spec: &AwsCloudwatchCompositeAlarmSpec{
				AlarmRule: `ALARM("cpu-high")`,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when more than 5 alarm_actions are provided", func() {
		actions := make([]*fkv1.StringValueOrRef, 6)
		for i := range actions {
			actions[i] = &fkv1.StringValueOrRef{
				LiteralOrRef: &fkv1.StringValueOrRef_Value{
					Value: "arn:aws:sns:us-west-2:123456789012:topic",
				},
			}
		}
		input := &AwsCloudwatchCompositeAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchCompositeAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "too-many-actions",
			},
			Spec: &AwsCloudwatchCompositeAlarmSpec{
				Region:       "us-west-2",
				AlarmRule:    `ALARM("cpu-high")`,
				AlarmActions: actions,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when more than 5 ok_actions are provided", func() {
		actions := make([]*fkv1.StringValueOrRef, 6)
		for i := range actions {
			actions[i] = &fkv1.StringValueOrRef{
				LiteralOrRef: &fkv1.StringValueOrRef_Value{
					Value: "arn:aws:sns:us-west-2:123456789012:topic",
				},
			}
		}
		input := &AwsCloudwatchCompositeAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchCompositeAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "too-many-ok-actions",
			},
			Spec: &AwsCloudwatchCompositeAlarmSpec{
				Region:    "us-west-2",
				AlarmRule: `ALARM("cpu-high")`,
				OkActions: actions,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when more than 5 insufficient_data_actions are provided", func() {
		actions := make([]*fkv1.StringValueOrRef, 6)
		for i := range actions {
			actions[i] = &fkv1.StringValueOrRef{
				LiteralOrRef: &fkv1.StringValueOrRef_Value{
					Value: "arn:aws:sns:us-west-2:123456789012:topic",
				},
			}
		}
		input := &AwsCloudwatchCompositeAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchCompositeAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "too-many-insufficient-actions",
			},
			Spec: &AwsCloudwatchCompositeAlarmSpec{
				Region:                  "us-west-2",
				AlarmRule:               `ALARM("cpu-high")`,
				InsufficientDataActions: actions,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when the actions suppressor has no alarm", func() {
		input := &AwsCloudwatchCompositeAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchCompositeAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "suppressor-no-alarm",
			},
			Spec: &AwsCloudwatchCompositeAlarmSpec{
				Region:    "us-west-2",
				AlarmRule: `ALARM("cpu-high")`,
				ActionsSuppressor: &AwsCloudwatchCompositeAlarmActionsSuppressor{
					WaitPeriod:      60,
					ExtensionPeriod: 120,
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when api_version is wrong", func() {
		input := &AwsCloudwatchCompositeAlarm{
			ApiVersion: "wrong.planton.dev/v1",
			Kind:       "AwsCloudwatchCompositeAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "wrong-api-version",
			},
			Spec: &AwsCloudwatchCompositeAlarmSpec{
				Region:    "us-west-2",
				AlarmRule: `ALARM("cpu-high")`,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when kind is wrong", func() {
		input := &AwsCloudwatchCompositeAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCompositeAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "wrong-kind",
			},
			Spec: &AwsCloudwatchCompositeAlarmSpec{
				Region:    "us-west-2",
				AlarmRule: `ALARM("cpu-high")`,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when metadata is missing", func() {
		input := &AwsCloudwatchCompositeAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchCompositeAlarm",
			Spec: &AwsCloudwatchCompositeAlarmSpec{
				Region:    "us-west-2",
				AlarmRule: `ALARM("cpu-high")`,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when spec is missing", func() {
		input := &AwsCloudwatchCompositeAlarm{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCloudwatchCompositeAlarm",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "no-spec",
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})

func boolPtr(b bool) *bool { return &b }
