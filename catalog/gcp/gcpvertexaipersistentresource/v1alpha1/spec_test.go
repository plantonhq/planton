package gcpvertexaipersistentresourcev1alpha1

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
	ginkgo.RunSpecs(t, "GcpVertexAiPersistentResourceSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func i64(v int64) *int64 { return &v }
func i32(v int32) *int32 { return &v }

var _ = ginkgo.Describe("GcpVertexAiPersistentResourceSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpVertexAiPersistentResource {
		return &GcpVertexAiPersistentResource{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpVertexAiPersistentResource",
			Metadata:   &shared.CloudResourceMetadata{Name: "training-pool"},
			Spec: &GcpVertexAiPersistentResourceSpec{
				Location: "us-central1",
				ResourcePools: []*GcpVertexAiPersistentResourceResourcePool{
					{
						MachineSpec:  &GcpVertexAiPersistentResourceMachineSpec{MachineType: "n1-standard-4"},
						ReplicaCount: i64(1),
					},
				},
			},
		}
	}

	ginkgo.It("should accept one fixed pool", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every field set", func() {
		msg := minimal()
		msg.Spec.ProjectId = litRef("ml-project")
		msg.Spec.PersistentResourceId = "gpu-pool"
		msg.Spec.DisplayName = "GPU pool"
		msg.Spec.Labels = map[string]string{"team": "ml"}
		msg.Spec.ResourcePools[0] = &GcpVertexAiPersistentResourceResourcePool{
			Id: "gpu-workers",
			MachineSpec: &GcpVertexAiPersistentResourceMachineSpec{
				MachineType:      "g2-standard-8",
				AcceleratorType:  "NVIDIA_L4",
				AcceleratorCount: 1,
			},
			ReplicaCount:    i64(2),
			AutoscalingSpec: &GcpVertexAiPersistentResourceAutoscalingSpec{MinReplicaCount: i64(1), MaxReplicaCount: i64(4)},
			DiskSpec:        &GcpVertexAiPersistentResourceDiskSpec{BootDiskSizeGb: i32(200), BootDiskType: "pd-ssd"},
		}
		msg.Spec.Network = litRef("projects/123456789012/global/networks/ml-vpc")
		msg.Spec.ReservedIpRanges = []string{"vertex-ai-range"}
		msg.Spec.EnableCustomServiceAccount = true
		msg.Spec.KmsKeyName = litRef("projects/p/locations/us-central1/keyRings/r/cryptoKeys/k")
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a PSC interface with DNS peering", func() {
		msg := minimal()
		msg.Spec.PscInterfaceConfig = &GcpVertexAiPersistentResourcePscInterfaceConfig{
			NetworkAttachment: "vertex-attachment",
			DnsPeeringConfigs: []*GcpVertexAiPersistentResourceDnsPeeringConfig{
				{Domain: "corp.example.com.", TargetProject: litRef("host-project"), TargetNetwork: litRef("corp-vpc")},
			},
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require at least one pool with a machine", func() {
		msg := minimal()
		msg.Spec.ResourcePools = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.ResourcePools[0].MachineSpec = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a zero replica floor, which Google rejects on a persistent resource", func() {
		msg := minimal()
		msg.Spec.ResourcePools[0].AutoscalingSpec = &GcpVertexAiPersistentResourceAutoscalingSpec{MinReplicaCount: i64(0), MaxReplicaCount: i64(2)}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.ResourcePools[0].ReplicaCount = i64(0)
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an accelerator Google does not list and a disk type it does not offer", func() {
		msg := minimal()
		msg.Spec.ResourcePools[0].MachineSpec.AcceleratorType = "NVIDIA_A10"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.ResourcePools[0].DiskSpec = &GcpVertexAiPersistentResourceDiskSpec{BootDiskType: "pd-balanced"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an id outside RFC 1035", func() {
		msg := minimal()
		msg.Spec.PersistentResourceId = "Training_Pool"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require a DNS peering domain ending with a dot", func() {
		msg := minimal()
		msg.Spec.PscInterfaceConfig = &GcpVertexAiPersistentResourcePscInterfaceConfig{
			DnsPeeringConfigs: []*GcpVertexAiPersistentResourceDnsPeeringConfig{
				{Domain: "corp.example.com", TargetProject: litRef("host-project"), TargetNetwork: litRef("corp-vpc")},
			},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
