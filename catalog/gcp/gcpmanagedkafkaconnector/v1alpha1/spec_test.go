package gcpmanagedkafkaconnectorv1alpha1

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
	ginkgo.RunSpecs(t, "GcpManagedKafkaConnectorSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpManagedKafkaConnectorSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpManagedKafkaConnector {
		return &GcpManagedKafkaConnector{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpManagedKafkaConnector",
			Metadata:   &shared.CloudResourceMetadata{Name: "orders-to-pubsub"},
			Spec: &GcpManagedKafkaConnectorSpec{
				Location:       "us-central1",
				ConnectCluster: litRef("events-connect"),
			},
		}
	}

	ginkgo.It("should accept a connector with only its Connect cluster", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every field set", func() {
		msg := minimal()
		msg.Spec.ProjectId = litRef("data-project")
		msg.Spec.ConnectCluster = litRef("projects/p/locations/us-central1/connectClusters/events-connect")
		msg.Spec.ConnectorId = "orders-to-pubsub"
		msg.Spec.Configs = map[string]string{
			"connector.class": "com.google.pubsub.kafka.sink.CloudPubSubSinkConnector",
			"tasks.max":       "3",
			"topics":          "orders",
		}
		msg.Spec.TaskRestartPolicy = &GcpManagedKafkaConnectorTaskRestartPolicy{MinimumBackoff: "60s", MaximumBackoff: "1800.5s"}
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require a Connect cluster and a location", func() {
		msg := minimal()
		msg.Spec.ConnectCluster = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Location = "us central"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject backoffs that are not second durations", func() {
		for _, d := range []string{"60", "1m", "1.1234567890s", "-5s"} {
			msg := minimal()
			msg.Spec.TaskRestartPolicy = &GcpManagedKafkaConnectorTaskRestartPolicy{MinimumBackoff: d}
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), d)
		}
		msg := minimal()
		msg.Spec.TaskRestartPolicy = &GcpManagedKafkaConnectorTaskRestartPolicy{MaximumBackoff: "30m"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
