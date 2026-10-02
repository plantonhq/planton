package gcptpuqueuedresourcev1alpha1

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
	ginkgo.RunSpecs(t, "GcpTpuQueuedResourceSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpTpuQueuedResourceSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpTpuQueuedResource {
		return &GcpTpuQueuedResource{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpTpuQueuedResource",
			Metadata:   &shared.CloudResourceMetadata{Name: "train-request"},
			Spec: &GcpTpuQueuedResourceSpec{
				Zone: "us-central1-a",
				NodeSpecs: []*GcpTpuQueuedResourceNodeSpec{
					{Node: &GcpTpuQueuedResourceNode{RuntimeVersion: "tpu-ubuntu2204-base"}},
				},
			},
		}
	}

	ginkgo.It("should accept the smallest request", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every field set", func() {
		msg := minimal()
		msg.Spec.ProjectId = litRef("ml-project")
		msg.Spec.QueuedResourceId = "train-request-01"
		msg.Spec.NodeSpecs = []*GcpTpuQueuedResourceNodeSpec{
			{NodeId: "worker-a", Node: &GcpTpuQueuedResourceNode{
				RuntimeVersion: "v2-alpha-tpuv5-lite", AcceleratorType: "v5litepod-8", Description: "slice a",
				NetworkConfig: &GcpTpuQueuedResourceNetworkConfig{
					Network:           litRef("projects/ml-project/global/networks/ml-vpc"),
					Subnetwork:        litRef("projects/ml-project/regions/us-central1/subnetworks/tpu"),
					EnableExternalIps: true, CanIpForward: false, QueueCount: 16,
				},
			}},
			{NodeId: "worker-b", Node: &GcpTpuQueuedResourceNode{RuntimeVersion: "v2-alpha-tpuv5-lite", AcceleratorType: "v5litepod-8"}},
		}
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require a zone and at least one node", func() {
		msg := minimal()
		msg.Spec.Zone = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.NodeSpecs = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require a node with a runtime version", func() {
		msg := minimal()
		msg.Spec.NodeSpecs[0].Node = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.NodeSpecs[0].Node.RuntimeVersion = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject duplicate node ids but allow several unnamed nodes", func() {
		msg := minimal()
		msg.Spec.NodeSpecs = []*GcpTpuQueuedResourceNodeSpec{
			{NodeId: "worker", Node: &GcpTpuQueuedResourceNode{RuntimeVersion: "r"}},
			{NodeId: "worker", Node: &GcpTpuQueuedResourceNode{RuntimeVersion: "r"}},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.NodeSpecs = []*GcpTpuQueuedResourceNodeSpec{
			{Node: &GcpTpuQueuedResourceNode{RuntimeVersion: "r"}},
			{Node: &GcpTpuQueuedResourceNode{RuntimeVersion: "r"}},
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject malformed ids and an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.QueuedResourceId = "Train"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.NodeSpecs[0].NodeId = "worker_a"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
