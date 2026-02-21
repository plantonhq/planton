package ociqueuev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestOciQueueSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciQueueSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidQueue() *OciQueue {
	return &OciQueue{
		ApiVersion: "oci.openmcf.org/v1",
		Kind:       "OciQueue",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-queue",
		},
		Spec: &OciQueueSpec{
			CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
		},
	}
}

var _ = ginkgo.Describe("OciQueueSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_queue", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := minimalValidQueue()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with custom encryption key", func() {
				input := minimalValidQueue()
				input.Spec.CustomEncryptionKeyId = newStringValueOrRef("ocid1.key.oc1..example")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with dead letter queue configured", func() {
				input := minimalValidQueue()
				input.Spec.DeadLetterQueueDeliveryCount = proto.Int32(5)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with DLQ disabled (value 0)", func() {
				input := minimalValidQueue()
				input.Spec.DeadLetterQueueDeliveryCount = proto.Int32(0)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with retention configured", func() {
				input := minimalValidQueue()
				input.Spec.RetentionInSeconds = proto.Int32(86400)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with timeout and visibility", func() {
				input := minimalValidQueue()
				input.Spec.TimeoutInSeconds = proto.Int32(30)
				input.Spec.VisibilityInSeconds = proto.Int32(60)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with channel consumption limit", func() {
				input := minimalValidQueue()
				input.Spec.ChannelConsumptionLimit = proto.Int32(50)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with large messages enabled", func() {
				input := minimalValidQueue()
				input.Spec.IsLargeMessagesEnabled = proto.Bool(true)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with consumer group config", func() {
				input := minimalValidQueue()
				input.Spec.ConsumerGroupConfig = &OciQueueSpec_ConsumerGroupConfig{
					IsPrimaryEnabled:   proto.Bool(true),
					PrimaryDisplayName: "My Consumer Group",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with consumer group DLQ override", func() {
				input := minimalValidQueue()
				input.Spec.ConsumerGroupConfig = &OciQueueSpec_ConsumerGroupConfig{
					IsPrimaryEnabled:                    proto.Bool(true),
					PrimaryDeadLetterQueueDeliveryCount: proto.Int32(10),
					PrimaryDisplayName:                  "DLQ Group",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via valueFrom ref", func() {
				input := minimalValidQueue()
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

			ginkgo.It("should not return a validation error with encryption key via valueFrom ref", func() {
				input := minimalValidQueue()
				input.Spec.CustomEncryptionKeyId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-kms-key",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full configuration", func() {
				input := minimalValidQueue()
				input.Spec.CustomEncryptionKeyId = newStringValueOrRef("ocid1.key.oc1..example")
				input.Spec.DeadLetterQueueDeliveryCount = proto.Int32(5)
				input.Spec.RetentionInSeconds = proto.Int32(604800)
				input.Spec.TimeoutInSeconds = proto.Int32(30)
				input.Spec.VisibilityInSeconds = proto.Int32(120)
				input.Spec.ChannelConsumptionLimit = proto.Int32(50)
				input.Spec.IsLargeMessagesEnabled = proto.Bool(true)
				input.Spec.ConsumerGroupConfig = &OciQueueSpec_ConsumerGroupConfig{
					IsPrimaryEnabled:                    proto.Bool(true),
					PrimaryDeadLetterQueueDeliveryCount: proto.Int32(3),
					PrimaryDisplayName:                  "Primary Group",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_queue", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidQueue()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidQueue()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidQueue()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciQueue{
					ApiVersion: "oci.openmcf.org/v1",
					Kind:       "OciQueue",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-queue"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidQueue()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is empty", func() {
				input := minimalValidQueue()
				input.ApiVersion = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is empty", func() {
				input := minimalValidQueue()
				input.Kind = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

		})
	})
})
