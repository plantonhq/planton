package awssesconfigurationsetv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	fkv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsSesConfigurationSetSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsSesConfigurationSetSpec Validation Suite")
}

func boolPtr(b bool) *bool       { return &b }
func int32Ptr(i int32) *int32    { return &i }
func stringPtr(s string) *string { return &s }

func minimalConfigSet() *AwsSesConfigurationSet {
	return &AwsSesConfigurationSet{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsSesConfigurationSet",
		Metadata: &shared.CloudResourceMetadata{
			Name: "txn-set",
		},
		Spec: &AwsSesConfigurationSetSpec{Region: "us-west-2"},
	}
}

func cloudWatchDestination(name string) *AwsSesConfigurationSetEventDestination {
	return &AwsSesConfigurationSetEventDestination{
		Name: name,
		MatchingEventTypes: []string{
			"SEND", "BOUNCE", "COMPLAINT",
		},
		CloudWatch: &AwsSesConfigurationSetEventDestinationCloudWatch{
			Dimensions: []*AwsSesConfigurationSetEventDestinationCloudWatchDimension{
				{
					Name:         "campaign",
					ValueSource:  "MESSAGE_TAG",
					DefaultValue: "none",
				},
			},
		},
	}
}

var _ = ginkgo.Describe("AwsSesConfigurationSetSpec validations", func() {

	ginkgo.It("accepts a minimal configuration set", func() {
		err := protovalidate.Validate(minimalConfigSet())
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts transactional posture with TLS REQUIRE and suppression reasons", func() {
		input := minimalConfigSet()
		input.Spec.DeliveryOptions = &AwsSesConfigurationSetDeliveryOptions{
			TlsPolicy: stringPtr("REQUIRE"),
		}
		input.Spec.SuppressedReasons = []string{"BOUNCE", "COMPLAINT"}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a VDM override that states one dial off and leaves the other unset", func() {
		input := minimalConfigSet()
		input.Spec.VdmOptions = &AwsSesConfigurationSetVdmOptions{
			EngagementMetricsEnabled: boolPtr(false),
		}
		gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		gomega.Expect(input.Spec.VdmOptions.OptimizedSharedDeliveryEnabled).To(gomega.BeNil(), "the unset dial stays unset at the schema layer; the platform fills its declared default before the module reads it")
	})

	ginkgo.It("accepts a CloudWatch event destination", func() {
		input := minimalConfigSet()
		input.Spec.EventDestinations = []*AwsSesConfigurationSetEventDestination{
			cloudWatchDestination("bounce-metrics"),
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.Context("CEL: event_destination_names_unique", func() {
		ginkgo.It("fails when event destination names are duplicated", func() {
			input := minimalConfigSet()
			input.Spec.EventDestinations = []*AwsSesConfigurationSetEventDestination{
				cloudWatchDestination("dup"),
				cloudWatchDestination("dup"),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("CEL: exactly_one_destination", func() {
		ginkgo.It("fails when no destination arm is set", func() {
			input := minimalConfigSet()
			input.Spec.EventDestinations = []*AwsSesConfigurationSetEventDestination{
				{
					Name:               "nowhere",
					MatchingEventTypes: []string{"SEND"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when multiple destination arms are set", func() {
			input := minimalConfigSet()
			dest := cloudWatchDestination("both")
			dest.SnsTopic = &fkv1.StringValueOrRef{
				LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "arn:aws:sns:us-west-2:123456789012:topic"},
			}
			input.Spec.EventDestinations = []*AwsSesConfigurationSetEventDestination{dest}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("tls_policy values", func() {
		ginkgo.It("accepts REQUIRE and OPTIONAL", func() {
			for _, policy := range []string{"REQUIRE", "OPTIONAL"} {
				input := minimalConfigSet()
				input.Spec.DeliveryOptions = &AwsSesConfigurationSetDeliveryOptions{
					TlsPolicy: stringPtr(policy),
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("fails on an invalid tls_policy", func() {
			input := minimalConfigSet()
			input.Spec.DeliveryOptions = &AwsSesConfigurationSetDeliveryOptions{
				TlsPolicy: stringPtr("MANDATORY"),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("max_delivery_seconds bounds", func() {
		ginkgo.It("accepts 300 and 50400", func() {
			for _, seconds := range []int32{300, 50400} {
				input := minimalConfigSet()
				input.Spec.DeliveryOptions = &AwsSesConfigurationSetDeliveryOptions{
					MaxDeliverySeconds: int32Ptr(seconds),
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("fails below 300", func() {
			input := minimalConfigSet()
			input.Spec.DeliveryOptions = &AwsSesConfigurationSetDeliveryOptions{
				MaxDeliverySeconds: int32Ptr(299),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails above 50400", func() {
			input := minimalConfigSet()
			input.Spec.DeliveryOptions = &AwsSesConfigurationSetDeliveryOptions{
				MaxDeliverySeconds: int32Ptr(50401),
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("suppressed_reasons values", func() {
		ginkgo.It("accepts BOUNCE and COMPLAINT", func() {
			input := minimalConfigSet()
			input.Spec.SuppressedReasons = []string{"BOUNCE", "COMPLAINT"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("fails on an invalid suppressed reason", func() {
			input := minimalConfigSet()
			input.Spec.SuppressedReasons = []string{"SPAM"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails when suppressed reasons are duplicated", func() {
			input := minimalConfigSet()
			input.Spec.SuppressedReasons = []string{"BOUNCE", "BOUNCE"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("matching_event_types values", func() {
		ginkgo.It("fails when matching_event_types is empty", func() {
			input := minimalConfigSet()
			input.Spec.EventDestinations = []*AwsSesConfigurationSetEventDestination{
				{
					Name:               "empty-types",
					MatchingEventTypes: []string{},
					CloudWatch: &AwsSesConfigurationSetEventDestinationCloudWatch{
						Dimensions: []*AwsSesConfigurationSetEventDestinationCloudWatchDimension{
							{Name: "x", ValueSource: "MESSAGE_TAG", DefaultValue: "none"},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("fails on an invalid matching_event_type", func() {
			dest := cloudWatchDestination("bad-type")
			dest.MatchingEventTypes = []string{"SENT"}
			input := minimalConfigSet()
			input.Spec.EventDestinations = []*AwsSesConfigurationSetEventDestination{dest}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("cloud_watch dimensions required", func() {
		ginkgo.It("fails when cloud_watch has no dimensions", func() {
			input := minimalConfigSet()
			input.Spec.EventDestinations = []*AwsSesConfigurationSetEventDestination{
				{
					Name:               "no-dims",
					MatchingEventTypes: []string{"SEND"},
					CloudWatch:         &AwsSesConfigurationSetEventDestinationCloudWatch{},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})
