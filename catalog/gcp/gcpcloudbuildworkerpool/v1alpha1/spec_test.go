package gcpcloudbuildworkerpoolv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpCloudBuildWorkerPoolSpec Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value}}
}

func reference(kind catalogkind.CatalogKind, name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Kind: kind, Name: name}}}
}

var _ = ginkgo.Describe("GcpCloudBuildWorkerPoolSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpCloudBuildWorkerPool {
		return &GcpCloudBuildWorkerPool{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpCloudBuildWorkerPool",
			Metadata:   &shared.CatalogObjectMetadata{Name: "private-builds"},
			Spec:       &GcpCloudBuildWorkerPoolSpec{Location: "us-central1"},
		}
	}

	ginkgo.It("should accept a default pool, a peered pool, and a PSC pool", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())

		peered := minimal()
		peered.Spec.ProjectId = reference(catalogkind.CatalogKind_GcpProject, "ci")
		peered.Spec.WorkerPoolId = "private-builds"
		peered.Spec.DisplayName = "Private builds"
		peered.Spec.Annotations = map[string]string{"owner": "platform"}
		peered.Spec.NetworkConfig = &GcpCloudBuildWorkerPoolNetworkConfig{
			PeeredNetwork:        reference(catalogkind.CatalogKind_GcpVpcNetwork, "ci-vpc"),
			PeeredNetworkIpRange: "/26",
		}
		peered.Spec.WorkerConfig = &GcpCloudBuildWorkerPoolWorkerConfig{
			MachineType:                "e2-standard-4",
			DiskSizeGb:                 200,
			NoExternalIp:               proto.Bool(true),
			EnableNestedVirtualization: proto.Bool(false),
		}
		peered.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(peered)).To(gomega.Succeed())

		psc := minimal()
		psc.Spec.PrivateServiceConnect = &GcpCloudBuildWorkerPoolPrivateServiceConnect{
			NetworkAttachment: "projects/ci/regions/us-central1/networkAttachments/builds",
			RouteAllTraffic:   true,
		}
		gomega.Expect(validator.Validate(psc)).To(gomega.Succeed())
	})

	ginkgo.It("should require a location", func() {
		msg := minimal()
		msg.Spec.Location = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should refuse both network arms at once", func() {
		msg := minimal()
		msg.Spec.NetworkConfig = &GcpCloudBuildWorkerPoolNetworkConfig{PeeredNetwork: literal("projects/123/global/networks/ci")}
		msg.Spec.PrivateServiceConnect = &GcpCloudBuildWorkerPoolPrivateServiceConnect{NetworkAttachment: "projects/ci/regions/us-central1/networkAttachments/builds"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require the peered network and a CIDR range", func() {
		msg := minimal()
		msg.Spec.NetworkConfig = &GcpCloudBuildWorkerPoolNetworkConfig{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		for _, cidr := range []string{"/24", "192.168.0.0/29"} {
			msg.Spec.NetworkConfig = &GcpCloudBuildWorkerPoolNetworkConfig{PeeredNetwork: literal("projects/123/global/networks/ci"), PeeredNetworkIpRange: cidr}
			gomega.Expect(validator.Validate(msg)).To(gomega.Succeed(), cidr)
		}
		for _, cidr := range []string{"24", "192.168.0.0", "192.168.0.0/"} {
			msg.Spec.NetworkConfig = &GcpCloudBuildWorkerPoolNetworkConfig{PeeredNetwork: literal("projects/123/global/networks/ci"), PeeredNetworkIpRange: cidr}
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), cidr)
		}
	})

	ginkgo.It("should require a full network attachment name", func() {
		msg := minimal()
		msg.Spec.PrivateServiceConnect = &GcpCloudBuildWorkerPoolPrivateServiceConnect{NetworkAttachment: "builds"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.PrivateServiceConnect = &GcpCloudBuildWorkerPoolPrivateServiceConnect{}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should bound the disk size and the display name", func() {
		msg := minimal()
		msg.Spec.WorkerConfig = &GcpCloudBuildWorkerPoolWorkerConfig{DiskSizeGb: 1001}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DisplayName = "a-display-name-that-is-far-longer-than-the-sixty-three-characters-google-allows"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a malformed pool ID and an unknown deletion policy", func() {
		for _, id := range []string{"Private", "1pool", "pool-", "pool_builds"} {
			msg := minimal()
			msg.Spec.WorkerPoolId = id
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), id)
		}
		msg := minimal()
		msg.Spec.DeletionPolicy = "FORCE"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
