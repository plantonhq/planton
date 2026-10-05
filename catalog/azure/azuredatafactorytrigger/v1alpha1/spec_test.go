package azuredatafactorytriggerv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureDataFactoryTriggerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureDataFactoryTriggerSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	testFactoryId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.DataFactory/factories/app-df"
	testStorageId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Storage/storageAccounts/appdata"
	testTopicId   = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.EventGrid/topics/app-events"
)

func pipelineRef(name string) *AzureDataFactoryTriggerPipelineReference {
	return &AzureDataFactoryTriggerPipelineReference{Name: literal(name)}
}

// validResource returns a valid schedule trigger that individual
// cases mutate into the shape under test.
func validResource() *AzureDataFactoryTrigger {
	return &AzureDataFactoryTrigger{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureDataFactoryTrigger",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-adf-trigger",
		},
		Spec: &AzureDataFactoryTriggerSpec{
			DataFactoryId: literal(testFactoryId),
			Name:          "nightly",
			Schedule: &AzureDataFactoryTriggerSchedule{
				Pipelines: []*AzureDataFactoryTriggerPipelineReference{pipelineRef("ingest-daily")},
			},
		},
	}
}

func validTumblingWindow() *AzureDataFactoryTriggerTumblingWindow {
	return &AzureDataFactoryTriggerTumblingWindow{
		Frequency: "Hour",
		Interval:  1,
		StartTime: "2026-09-01T00:00:00Z",
		Pipeline:  pipelineRef("ingest-daily"),
	}
}

func validBlobEvent() *AzureDataFactoryTriggerBlobEvent {
	return &AzureDataFactoryTriggerBlobEvent{
		StorageAccountId:   literal(testStorageId),
		Events:             []string{"Microsoft.Storage.BlobCreated"},
		BlobPathBeginsWith: "/landing/blobs/",
		Pipelines:          []*AzureDataFactoryTriggerPipelineReference{pipelineRef("ingest-daily")},
	}
}

func validCustomEvent() *AzureDataFactoryTriggerCustomEvent {
	return &AzureDataFactoryTriggerCustomEvent{
		EventgridTopicId: literal(testTopicId),
		Events:           []string{"orders.batch.ready"},
		Pipelines:        []*AzureDataFactoryTriggerPipelineReference{pipelineRef("ingest-daily")},
	}
}

var _ = ginkgo.Describe("AzureDataFactoryTriggerSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.Context("variant selection", func() {

			ginkgo.It("should accept a minimal schedule trigger", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a minimal tumbling window trigger", func() {
				input := validResource()
				input.Spec.Schedule = nil
				input.Spec.TumblingWindow = validTumblingWindow()
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a minimal blob event trigger", func() {
				input := validResource()
				input.Spec.Schedule = nil
				input.Spec.BlobEvent = validBlobEvent()
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a minimal custom event trigger", func() {
				input := validResource()
				input.Spec.Schedule = nil
				input.Spec.CustomEvent = validCustomEvent()
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("schedule variant", func() {

			ginkgo.It("should accept the full recurrence surface", func() {
				input := validResource()
				input.Spec.Description = "fires the nightly ingest"
				input.Spec.Annotations = []string{"team:data"}
				input.Spec.Activated = proto.Bool(false)
				input.Spec.Schedule.Frequency = proto.String("Week")
				input.Spec.Schedule.Interval = proto.Int32(2)
				input.Spec.Schedule.StartTime = "2026-09-01T00:00:00Z"
				input.Spec.Schedule.EndTime = "2026-12-31T00:00:00Z"
				input.Spec.Schedule.TimeZone = "UTC"
				input.Spec.Schedule.RecurrenceSchedule = &AzureDataFactoryTriggerRecurrenceSchedule{
					DaysOfMonth: []int32{1, 15, -1},
					DaysOfWeek:  []string{"Monday", "Friday"},
					Hours:       []int32{0, 12},
					Minutes:     []int32{0, 30},
					Monthly: []*AzureDataFactoryTriggerMonthlyOccurrence{
						{Weekday: "Friday", Week: proto.Int32(2)},
						{Weekday: "Monday", Week: proto.Int32(-1)},
					},
				}
				input.Spec.Schedule.Pipelines[0].Parameters = map[string]string{"window": "P1D"}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept multiple pipelines", func() {
				input := validResource()
				input.Spec.Schedule.Pipelines = append(input.Spec.Schedule.Pipelines, pipelineRef("publish-daily"))
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept the provider's own boundary hour and minute values", func() {
				input := validResource()
				input.Spec.Schedule.RecurrenceSchedule = &AzureDataFactoryTriggerRecurrenceSchedule{
					Hours:   []int32{24},
					Minutes: []int32{60},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})

		ginkgo.Context("tumbling window variant", func() {

			ginkgo.It("should accept the full window surface", func() {
				input := validResource()
				input.Spec.Schedule = nil
				tw := validTumblingWindow()
				tw.EndTime = "2026-12-31T00:00:00Z"
				tw.Delay = "00:10:00"
				tw.MaxConcurrency = proto.Int32(10)
				tw.Retry = &AzureDataFactoryTriggerRetryPolicy{Count: 3, Interval: proto.Int32(60)}
				tw.Dependencies = []*AzureDataFactoryTriggerDependency{
					{TriggerName: literal("upstream-window"), Offset: "-24:00:00", Size: "24:00:00"},
					{Offset: "-24:00:00"},
				}
				tw.AdditionalProperties = map[string]string{"custom": "value"}
				tw.Pipeline.Parameters = map[string]string{"windowStart": "@trigger().outputs.windowStartTime"}
				input.Spec.TumblingWindow = tw
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a day-prefixed TimeSpan delay", func() {
				input := validResource()
				input.Spec.Schedule = nil
				tw := validTumblingWindow()
				tw.Delay = "1.06:00:00"
				input.Spec.TumblingWindow = tw
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.Context("variant selection", func() {

			ginkgo.It("should reject a trigger with no variant", func() {
				input := validResource()
				input.Spec.Schedule = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a trigger with two variants", func() {
				input := validResource()
				input.Spec.TumblingWindow = validTumblingWindow()
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("shared fields", func() {

			ginkgo.It("should reject a missing data factory id", func() {
				input := validResource()
				input.Spec.DataFactoryId = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty name", func() {
				input := validResource()
				input.Spec.Name = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject names containing Azure's forbidden characters", func() {
				input := validResource()
				for _, name := range []string{"-nightly", "night<ly", "night.ly", "night\\ly", "night/ly"} {
					input.Spec.Name = name
					gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "expected %q to be rejected", name)
				}
			})
		})

		ginkgo.Context("schedule variant", func() {

			ginkgo.It("should reject an empty pipelines list", func() {
				input := validResource()
				input.Spec.Schedule.Pipelines = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a pipeline reference without a name", func() {
				input := validResource()
				input.Spec.Schedule.Pipelines[0].Name = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a frequency outside the vocabulary", func() {
				input := validResource()
				input.Spec.Schedule.Frequency = proto.String("Fortnight")
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a zero interval", func() {
				input := validResource()
				input.Spec.Schedule.Interval = proto.Int32(0)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a malformed start_time", func() {
				input := validResource()
				input.Spec.Schedule.StartTime = "tomorrow"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject out-of-range recurrence values", func() {
				input := validResource()
				input.Spec.Schedule.RecurrenceSchedule = &AzureDataFactoryTriggerRecurrenceSchedule{Hours: []int32{25}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Schedule.RecurrenceSchedule = &AzureDataFactoryTriggerRecurrenceSchedule{Minutes: []int32{61}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Schedule.RecurrenceSchedule = &AzureDataFactoryTriggerRecurrenceSchedule{DaysOfMonth: []int32{0}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Schedule.RecurrenceSchedule = &AzureDataFactoryTriggerRecurrenceSchedule{DaysOfMonth: []int32{32}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown weekday", func() {
				input := validResource()
				input.Spec.Schedule.RecurrenceSchedule = &AzureDataFactoryTriggerRecurrenceSchedule{DaysOfWeek: []string{"Funday"}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject more than seven days_of_week entries", func() {
				input := validResource()
				input.Spec.Schedule.RecurrenceSchedule = &AzureDataFactoryTriggerRecurrenceSchedule{
					DaysOfWeek: []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday", "Monday"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a zero monthly week", func() {
				input := validResource()
				input.Spec.Schedule.RecurrenceSchedule = &AzureDataFactoryTriggerRecurrenceSchedule{
					Monthly: []*AzureDataFactoryTriggerMonthlyOccurrence{{Weekday: "Friday", Week: proto.Int32(0)}},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("tumbling window variant", func() {

			ginkgo.It("should reject a frequency outside the tumbling vocabulary", func() {
				input := validResource()
				input.Spec.Schedule = nil
				tw := validTumblingWindow()
				tw.Frequency = "Day"
				input.Spec.TumblingWindow = tw
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a zero interval", func() {
				input := validResource()
				input.Spec.Schedule = nil
				tw := validTumblingWindow()
				tw.Interval = 0
				input.Spec.TumblingWindow = tw
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing start_time", func() {
				input := validResource()
				input.Spec.Schedule = nil
				tw := validTumblingWindow()
				tw.StartTime = ""
				input.Spec.TumblingWindow = tw
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing pipeline", func() {
				input := validResource()
				input.Spec.Schedule = nil
				tw := validTumblingWindow()
				tw.Pipeline = nil
				input.Spec.TumblingWindow = tw
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a malformed delay", func() {
				input := validResource()
				input.Spec.Schedule = nil
				tw := validTumblingWindow()
				tw.Delay = "ten minutes"
				input.Spec.TumblingWindow = tw
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject max_concurrency outside 1-50", func() {
				input := validResource()
				input.Spec.Schedule = nil
				tw := validTumblingWindow()
				tw.MaxConcurrency = proto.Int32(51)
				input.Spec.TumblingWindow = tw
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a retry policy with zero count", func() {
				input := validResource()
				input.Spec.Schedule = nil
				tw := validTumblingWindow()
				tw.Retry = &AzureDataFactoryTriggerRetryPolicy{Count: 0}
				input.Spec.TumblingWindow = tw
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a malformed dependency offset", func() {
				input := validResource()
				input.Spec.Schedule = nil
				tw := validTumblingWindow()
				tw.Dependencies = []*AzureDataFactoryTriggerDependency{{Offset: "yesterday"}}
				input.Spec.TumblingWindow = tw
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("blob event variant", func() {

			ginkgo.It("should reject a missing storage account id", func() {
				input := validResource()
				input.Spec.Schedule = nil
				be := validBlobEvent()
				be.StorageAccountId = nil
				input.Spec.BlobEvent = be
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty events list", func() {
				input := validResource()
				input.Spec.Schedule = nil
				be := validBlobEvent()
				be.Events = nil
				input.Spec.BlobEvent = be
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an event outside the blob vocabulary", func() {
				input := validResource()
				input.Spec.Schedule = nil
				be := validBlobEvent()
				be.Events = []string{"Microsoft.Storage.BlobRenamed"}
				input.Spec.BlobEvent = be
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a trigger without any path filter", func() {
				input := validResource()
				input.Spec.Schedule = nil
				be := validBlobEvent()
				be.BlobPathBeginsWith = ""
				be.BlobPathEndsWith = ""
				input.Spec.BlobEvent = be
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty pipelines list", func() {
				input := validResource()
				input.Spec.Schedule = nil
				be := validBlobEvent()
				be.Pipelines = nil
				input.Spec.BlobEvent = be
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("custom event variant", func() {

			ginkgo.It("should reject a missing event grid topic id", func() {
				input := validResource()
				input.Spec.Schedule = nil
				ce := validCustomEvent()
				ce.EventgridTopicId = nil
				input.Spec.CustomEvent = ce
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty events list", func() {
				input := validResource()
				input.Spec.Schedule = nil
				ce := validCustomEvent()
				ce.Events = nil
				input.Spec.CustomEvent = ce
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty event type string", func() {
				input := validResource()
				input.Spec.Schedule = nil
				ce := validCustomEvent()
				ce.Events = []string{""}
				input.Spec.CustomEvent = ce
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty pipelines list", func() {
				input := validResource()
				input.Spec.Schedule = nil
				ce := validCustomEvent()
				ce.Pipelines = nil
				input.Spec.CustomEvent = ce
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})
