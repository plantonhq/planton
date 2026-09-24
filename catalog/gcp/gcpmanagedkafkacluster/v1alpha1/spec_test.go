package gcpmanagedkafkaclusterv1alpha1

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
	ginkgo.RunSpecs(t, "GcpManagedKafkaClusterSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

const gib = int64(1073741824)

var _ = ginkgo.Describe("GcpManagedKafkaClusterSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpManagedKafkaCluster {
		return &GcpManagedKafkaCluster{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpManagedKafkaCluster",
			Metadata:   &shared.CloudResourceMetadata{Name: "events"},
			Spec: &GcpManagedKafkaClusterSpec{
				Location:       "us-central1",
				CapacityConfig: &GcpManagedKafkaClusterCapacity{VcpuCount: 3, MemoryBytes: 3 * gib},
				NetworkConfigs: []*GcpManagedKafkaClusterNetworkConfig{{Subnet: litRef("projects/p/regions/us-central1/subnetworks/kafka")}},
			},
		}
	}

	ginkgo.It("should accept the smallest cluster", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every field set", func() {
		msg := minimal()
		msg.Spec.ProjectId = litRef("data-project")
		msg.Spec.ClusterId = "events-prod"
		msg.Spec.CapacityConfig = &GcpManagedKafkaClusterCapacity{VcpuCount: 12, MemoryBytes: 48 * gib}
		msg.Spec.BrokerDiskSizeGib = 500
		msg.Spec.KmsKey = litRef("projects/p/locations/us-central1/keyRings/r/cryptoKeys/k")
		msg.Spec.RebalanceMode = "AUTO_REBALANCE_ON_SCALE_UP"
		msg.Spec.TlsConfig = &GcpManagedKafkaClusterTlsConfig{
			SslPrincipalMappingRules: "RULE:^CN=(.*?),.*$/$1/,DEFAULT",
			CaPools:                  []string{"projects/pki/locations/us-central1/caPools/clients"},
		}
		msg.Spec.Labels = map[string]string{"team": "data"}
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an empty TLS config (the clearing form)", func() {
		msg := minimal()
		msg.Spec.TlsConfig = &GcpManagedKafkaClusterTlsConfig{}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require a location, capacity, and at least one network", func() {
		msg := minimal()
		msg.Spec.Location = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.CapacityConfig = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.NetworkConfigs = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject fewer than 3 vCPUs and memory outside 1-8 GiB per vCPU", func() {
		msg := minimal()
		msg.Spec.CapacityConfig = &GcpManagedKafkaClusterCapacity{VcpuCount: 2, MemoryBytes: 3 * gib}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.CapacityConfig = &GcpManagedKafkaClusterCapacity{VcpuCount: 4, MemoryBytes: 3 * gib}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.CapacityConfig = &GcpManagedKafkaClusterCapacity{VcpuCount: 3, MemoryBytes: 25 * gib}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.CapacityConfig = &GcpManagedKafkaClusterCapacity{VcpuCount: 3, MemoryBytes: 24 * gib}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject a small broker disk, too many networks, and a malformed CA pool", func() {
		msg := minimal()
		msg.Spec.BrokerDiskSizeGib = 50
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		for i := 0; i < 10; i++ {
			msg.Spec.NetworkConfigs = append(msg.Spec.NetworkConfigs, &GcpManagedKafkaClusterNetworkConfig{Subnet: litRef("s")})
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.TlsConfig = &GcpManagedKafkaClusterTlsConfig{CaPools: []string{"clients"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a malformed cluster ID, an unknown rebalance mode, and an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.ClusterId = "Events_Prod"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.RebalanceMode = "MODE_UNSPECIFIED"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
