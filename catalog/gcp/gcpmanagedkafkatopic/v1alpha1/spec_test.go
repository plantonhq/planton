package gcpmanagedkafkatopicv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpManagedKafkaTopicSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpManagedKafkaTopicSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpManagedKafkaTopic {
		return &GcpManagedKafkaTopic{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpManagedKafkaTopic",
			Metadata:   &shared.CloudResourceMetadata{Name: "orders"},
			Spec: &GcpManagedKafkaTopicSpec{
				Location:          "us-central1",
				Cluster:           litRef("events"),
				ReplicationFactor: 3,
			},
		}
	}

	ginkgo.It("should accept a topic with only its cluster and replication factor", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every field set, with a full cluster path and a dotted topic name", func() {
		msg := minimal()
		msg.Spec.ProjectId = litRef("data-project")
		msg.Spec.Cluster = litRef("projects/data-project/locations/us-central1/clusters/events")
		msg.Spec.TopicId = "orders.v1"
		msg.Spec.PartitionCount = 12
		msg.Spec.Configs = map[string]string{"cleanup.policy": "compact"}
		msg.Spec.DeletionPolicy = "ABANDON"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require a cluster, a location, and a replication factor", func() {
		msg := minimal()
		msg.Spec.Cluster = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Location = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.ReplicationFactor = 0
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a topic name outside Kafka's rule and a negative partition count", func() {
		msg := minimal()
		msg.Spec.TopicId = "orders/v1"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.PartitionCount = -1
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
