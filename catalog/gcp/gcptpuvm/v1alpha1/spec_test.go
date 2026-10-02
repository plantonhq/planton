package gcptpuvmv1alpha1

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
	ginkgo.RunSpecs(t, "GcpTpuVmSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpTpuVmSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpTpuVm {
		return &GcpTpuVm{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpTpuVm",
			Metadata:   &shared.CloudResourceMetadata{Name: "train-v5e"},
			Spec: &GcpTpuVmSpec{
				Zone:           "us-central1-a",
				RuntimeVersion: "v2-alpha-tpuv5-lite",
			},
		}
	}

	ginkgo.It("should accept the smallest TPU", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every field set", func() {
		msg := minimal()
		msg.Spec.ProjectId = litRef("ml-project")
		msg.Spec.NodeId = "train-v5e-01"
		msg.Spec.AcceleratorType = "v5litepod-8"
		msg.Spec.Description = "Fine-tuning"
		msg.Spec.CidrBlock = "10.100.0.0/29"
		msg.Spec.NetworkConfig = &GcpTpuVmNetworkConfig{
			Network:           litRef("projects/ml-project/global/networks/ml-vpc"),
			Subnetwork:        litRef("projects/ml-project/regions/us-central1/subnetworks/tpu"),
			EnableExternalIps: false,
			CanIpForward:      true,
			QueueCount:        32,
		}
		msg.Spec.ServiceAccount = &GcpTpuVmServiceAccount{Email: litRef("tpu@ml-project.iam.gserviceaccount.com"), Scopes: []string{"https://www.googleapis.com/auth/cloud-platform"}}
		msg.Spec.SchedulingConfig = &GcpTpuVmSchedulingConfig{Spot: true}
		msg.Spec.DataDisks = []*GcpTpuVmDataDisk{{SourceDisk: litRef("projects/ml-project/zones/us-central1-a/disks/data"), Mode: "READ_ONLY"}}
		msg.Spec.EnableSecureBoot = true
		msg.Spec.Labels = map[string]string{"team": "research"}
		msg.Spec.Metadata = map[string]string{"startup-script": "pip install -U jax"}
		msg.Spec.Tags = []string{"tpu"}
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a topology-shaped slice and several network interfaces", func() {
		msg := minimal()
		msg.Spec.AcceleratorConfig = &GcpTpuVmAcceleratorConfig{Type: "V5LITE_POD", Topology: "2x4"}
		msg.Spec.NetworkConfigs = []*GcpTpuVmNetworkConfig{{Network: litRef("projects/p/global/networks/a")}, {Network: litRef("projects/p/global/networks/b")}}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require a zone and a runtime version", func() {
		msg := minimal()
		msg.Spec.Zone = "us-central1"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.RuntimeVersion = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject both accelerator forms", func() {
		msg := minimal()
		msg.Spec.AcceleratorType = "v2-8"
		msg.Spec.AcceleratorConfig = &GcpTpuVmAcceleratorConfig{Type: "V2", Topology: "2x2"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject both network forms", func() {
		msg := minimal()
		msg.Spec.NetworkConfig = &GcpTpuVmNetworkConfig{}
		msg.Spec.NetworkConfigs = []*GcpTpuVmNetworkConfig{{}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a malformed topology and generation", func() {
		msg := minimal()
		msg.Spec.AcceleratorConfig = &GcpTpuVmAcceleratorConfig{Type: "v5e", Topology: "2x2"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.AcceleratorConfig = &GcpTpuVmAcceleratorConfig{Type: "V5P", Topology: "2-2"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should accept only a /29 CIDR block", func() {
		msg := minimal()
		msg.Spec.CidrBlock = "10.0.0.0/28"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require a source disk and a known mode on a data disk", func() {
		msg := minimal()
		msg.Spec.DataDisks = []*GcpTpuVmDataDisk{{Mode: "READ_WRITE"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.DataDisks = []*GcpTpuVmDataDisk{{SourceDisk: litRef("projects/p/zones/z/disks/d"), Mode: "RW"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a node id starting with a digit and an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.NodeId = "1tpu"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
