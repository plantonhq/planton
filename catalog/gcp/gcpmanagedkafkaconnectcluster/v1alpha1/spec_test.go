package gcpmanagedkafkaconnectclusterv1alpha1

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
	ginkgo.RunSpecs(t, "GcpManagedKafkaConnectClusterSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func valueFrom(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: name}},
	}
}

const gib = int64(1073741824)

var _ = ginkgo.Describe("GcpManagedKafkaConnectClusterSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpManagedKafkaConnectCluster {
		return &GcpManagedKafkaConnectCluster{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpManagedKafkaConnectCluster",
			Metadata:   &shared.CloudResourceMetadata{Name: "events-connect"},
			Spec: &GcpManagedKafkaConnectClusterSpec{
				Location:       "us-central1",
				KafkaCluster:   litRef("projects/p/locations/us-central1/clusters/events"),
				CapacityConfig: &GcpManagedKafkaConnectClusterCapacity{VcpuCount: 3, MemoryBytes: 3 * gib},
				NetworkConfigs: []*GcpManagedKafkaConnectClusterNetworkConfig{{PrimarySubnet: litRef("projects/p/regions/us-central1/subnetworks/connect")}},
			},
		}
	}

	ginkgo.It("should accept the smallest Connect cluster", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every field set and a Kafka cluster reference", func() {
		msg := minimal()
		msg.Spec.ProjectId = litRef("data-project")
		msg.Spec.ConnectClusterId = "events-connect-prod"
		msg.Spec.KafkaCluster = valueFrom("events")
		msg.Spec.CapacityConfig = &GcpManagedKafkaConnectClusterCapacity{VcpuCount: 6, MemoryBytes: 24 * gib}
		msg.Spec.NetworkConfigs[0].DnsDomainNames = []string{"source.us-east1.managedkafka.other-project.cloud.goog"}
		msg.Spec.Labels = map[string]string{"team": "data"}
		msg.Spec.DeletionPolicy = "ABANDON"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject a literal Kafka cluster that is not a full path", func() {
		msg := minimal()
		msg.Spec.KafkaCluster = litRef("events")
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.KafkaCluster = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject undersized capacity and a ratio outside 1:1 to 1:8", func() {
		msg := minimal()
		msg.Spec.CapacityConfig = &GcpManagedKafkaConnectClusterCapacity{VcpuCount: 2, MemoryBytes: 3 * gib}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.CapacityConfig = &GcpManagedKafkaConnectClusterCapacity{VcpuCount: 3, MemoryBytes: 2 * gib}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.CapacityConfig = &GcpManagedKafkaConnectClusterCapacity{VcpuCount: 4, MemoryBytes: 3 * gib}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.CapacityConfig = &GcpManagedKafkaConnectClusterCapacity{VcpuCount: 3, MemoryBytes: 25 * gib}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require a network and cap networks at 10", func() {
		msg := minimal()
		msg.Spec.NetworkConfigs = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		for i := 0; i < 10; i++ {
			msg.Spec.NetworkConfigs = append(msg.Spec.NetworkConfigs, &GcpManagedKafkaConnectClusterNetworkConfig{PrimarySubnet: litRef("s")})
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
