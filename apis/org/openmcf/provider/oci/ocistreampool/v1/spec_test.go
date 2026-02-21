package ocistreampoolv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestOciStreamPoolSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciStreamPoolSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidStreamPool() *OciStreamPool {
	return &OciStreamPool{
		ApiVersion: "oci.openmcf.org/v1",
		Kind:       "OciStreamPool",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-pool",
		},
		Spec: &OciStreamPoolSpec{
			CompartmentId: newStringValueOrRef("ocid1.compartment.oc1..example"),
		},
	}
}

var _ = ginkgo.Describe("OciStreamPoolSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_stream_pool", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := minimalValidStreamPool()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with kafka settings", func() {
				input := minimalValidStreamPool()
				input.Spec.KafkaSettings = &OciStreamPoolSpec_KafkaSettings{
					AutoCreateTopicsEnable: proto.Bool(true),
					LogRetentionHours:      proto.Int32(72),
					NumPartitions:          proto.Int32(5),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with kms key", func() {
				input := minimalValidStreamPool()
				input.Spec.KmsKeyId = newStringValueOrRef("ocid1.key.oc1..example")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with private endpoint settings", func() {
				input := minimalValidStreamPool()
				input.Spec.PrivateEndpointSettings = &OciStreamPoolSpec_PrivateEndpointSettings{
					SubnetId: newStringValueOrRef("ocid1.subnet.oc1..example"),
					NsgIds: []*foreignkeyv1.StringValueOrRef{
						newStringValueOrRef("ocid1.nsg.oc1..example"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with a single stream", func() {
				input := minimalValidStreamPool()
				input.Spec.Streams = []*OciStreamPoolSpec_Stream{
					{Name: "events", Partitions: 3},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple streams", func() {
				input := minimalValidStreamPool()
				input.Spec.Streams = []*OciStreamPoolSpec_Stream{
					{Name: "events", Partitions: 3},
					{Name: "logs", Partitions: 1},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with explicit stream retention", func() {
				input := minimalValidStreamPool()
				input.Spec.Streams = []*OciStreamPoolSpec_Stream{
					{Name: "events", Partitions: 3, RetentionInHours: proto.Int32(48)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with max stream retention", func() {
				input := minimalValidStreamPool()
				input.Spec.Streams = []*OciStreamPoolSpec_Stream{
					{Name: "events", Partitions: 1, RetentionInHours: proto.Int32(168)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with compartment_id via valueFrom ref", func() {
				input := minimalValidStreamPool()
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

			ginkgo.It("should not return a validation error with kms key via valueFrom ref", func() {
				input := minimalValidStreamPool()
				input.Spec.KmsKeyId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-kms-key",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with private endpoint subnet via valueFrom ref", func() {
				input := minimalValidStreamPool()
				input.Spec.PrivateEndpointSettings = &OciStreamPoolSpec_PrivateEndpointSettings{
					SubnetId: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
							ValueFrom: &foreignkeyv1.ValueFromRef{
								Name: "my-subnet",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full configuration", func() {
				input := minimalValidStreamPool()
				input.Spec.KafkaSettings = &OciStreamPoolSpec_KafkaSettings{
					AutoCreateTopicsEnable: proto.Bool(true),
					LogRetentionHours:      proto.Int32(72),
					NumPartitions:          proto.Int32(5),
				}
				input.Spec.KmsKeyId = newStringValueOrRef("ocid1.key.oc1..example")
				input.Spec.PrivateEndpointSettings = &OciStreamPoolSpec_PrivateEndpointSettings{
					SubnetId:          newStringValueOrRef("ocid1.subnet.oc1..example"),
					NsgIds:            []*foreignkeyv1.StringValueOrRef{newStringValueOrRef("ocid1.nsg.oc1..example")},
					PrivateEndpointIp: "10.0.1.50",
				}
				input.Spec.Streams = []*OciStreamPoolSpec_Stream{
					{Name: "events", Partitions: 3, RetentionInHours: proto.Int32(48)},
					{Name: "logs", Partitions: 1},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_stream_pool", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidStreamPool()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidStreamPool()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidStreamPool()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciStreamPool{
					ApiVersion: "oci.openmcf.org/v1",
					Kind:       "OciStreamPool",
					Metadata:   &shared.CloudResourceMetadata{Name: "test-pool"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when compartment_id is missing", func() {
				input := minimalValidStreamPool()
				input.Spec.CompartmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when stream name is empty", func() {
				input := minimalValidStreamPool()
				input.Spec.Streams = []*OciStreamPoolSpec_Stream{
					{Name: "", Partitions: 1},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when stream partitions is zero", func() {
				input := minimalValidStreamPool()
				input.Spec.Streams = []*OciStreamPoolSpec_Stream{
					{Name: "events", Partitions: 0},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when stream retention is below 24", func() {
				input := minimalValidStreamPool()
				input.Spec.Streams = []*OciStreamPoolSpec_Stream{
					{Name: "events", Partitions: 1, RetentionInHours: proto.Int32(12)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when stream retention is above 168", func() {
				input := minimalValidStreamPool()
				input.Spec.Streams = []*OciStreamPoolSpec_Stream{
					{Name: "events", Partitions: 1, RetentionInHours: proto.Int32(200)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kafka log_retention_hours is below 24", func() {
				input := minimalValidStreamPool()
				input.Spec.KafkaSettings = &OciStreamPoolSpec_KafkaSettings{
					LogRetentionHours: proto.Int32(10),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kafka log_retention_hours is above 168", func() {
				input := minimalValidStreamPool()
				input.Spec.KafkaSettings = &OciStreamPoolSpec_KafkaSettings{
					LogRetentionHours: proto.Int32(300),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when private endpoint subnet_id is missing", func() {
				input := minimalValidStreamPool()
				input.Spec.PrivateEndpointSettings = &OciStreamPoolSpec_PrivateEndpointSettings{
					PrivateEndpointIp: "10.0.1.50",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

		})
	})
})
