package gcpvpcnetworkv1alpha1

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"

	"buf.build/go/protovalidate"
	"github.com/plantonhq/planton/shared"
)

func TestGcpVpcNetworkSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpVpcNetworkSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("GcpVpcNetworkSpec Custom Validation Tests", func() {

	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpVpcNetwork {
		return &GcpVpcNetwork{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpVpcNetwork",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-gcp-vpc-network",
			},
			Spec: &GcpVpcNetworkSpec{
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
				},
				NetworkName: "test-vpc-network",
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept an empty project_id (provider default)", func() {
		target := minimal()
		target.Spec.ProjectId = nil
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept GLOBAL routing mode", func() {
		target := minimal()
		mode := GcpVpcNetworkRoutingMode_GLOBAL
		target.Spec.RoutingMode = &mode
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a description", func() {
		target := minimal()
		target.Spec.Description = "Production VPC for application workloads"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept mtu at both bounds", func() {
		for _, mtu := range []int32{1300, 8896} {
			target := minimal()
			target.Spec.Mtu = &mtu
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should accept ULA internal IPv6 with an explicit range", func() {
		target := minimal()
		target.Spec.EnableUlaInternalIpv6 = true
		target.Spec.InternalIpv6Range = "fd20:1:2::/48"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept each firewall policy enforcement order", func() {
		for _, order := range []string{"BEFORE_CLASSIC_FIREWALL", "AFTER_CLASSIC_FIREWALL"} {
			target := minimal()
			target.Spec.NetworkFirewallPolicyEnforcementOrder = order
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should accept a full BGP best-path selection block", func() {
		target := minimal()
		target.Spec.BgpBestPathSelection = &GcpVpcNetworkBgpBestPathSelection{
			Mode:             "STANDARD",
			AlwaysCompareMed: true,
			InterRegionCost:  "ADD_COST_TO_MED",
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept delete_default_routes_on_create", func() {
		target := minimal()
		target.Spec.DeleteDefaultRoutesOnCreate = true
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept resource manager tags", func() {
		target := minimal()
		target.Spec.ResourceManagerTags = map[string]string{"tagKeys/123456": "tagValues/789012"}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept each deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
			target := minimal()
			target.Spec.DeletionPolicy = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing network_name", func() {
		target := minimal()
		target.Spec.NetworkName = ""
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an invalid network_name (uppercase)", func() {
		target := minimal()
		target.Spec.NetworkName = "INVALID-NAME"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an mtu below 1300", func() {
		target := minimal()
		mtu := int32(1299)
		target.Spec.Mtu = &mtu
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an mtu above 8896", func() {
		target := minimal()
		mtu := int32(8897)
		target.Spec.Mtu = &mtu
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an invalid firewall policy enforcement order", func() {
		target := minimal()
		target.Spec.NetworkFirewallPolicyEnforcementOrder = "SOMETIME_LATER"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an invalid BGP mode", func() {
		target := minimal()
		target.Spec.BgpBestPathSelection = &GcpVpcNetworkBgpBestPathSelection{Mode: "FASTEST"}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an invalid BGP inter-region cost", func() {
		target := minimal()
		target.Spec.BgpBestPathSelection = &GcpVpcNetworkBgpBestPathSelection{InterRegionCost: "FREE"}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an over-long description", func() {
		target := minimal()
		long := make([]byte, 2049)
		for i := range long {
			long[i] = 'a'
		}
		target.Spec.Description = string(long)
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})
})
