package gcpservicenetworkingconnectionv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpServiceNetworkingConnectionSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpServiceNetworkingConnectionSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpServiceNetworkingConnection {
		return &GcpServiceNetworkingConnection{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpServiceNetworkingConnection",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-psa-connection",
			},
			Spec: &GcpServiceNetworkingConnectionSpec{
				Network: litRef("projects/my-proj/global/networks/my-vpc"),
				ReservedPeeringRanges: []*foreignkeyv1.StringValueOrRef{
					litRef("psa-range"),
				},
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal spec (network + one reserved range)", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a project_id literal", func() {
		target := minimal()
		target.Spec.ProjectId = litRef("my-gcp-project-123")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a network by bare name", func() {
		target := minimal()
		target.Spec.Network = litRef("my-vpc")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept the default service explicitly", func() {
		target := minimal()
		target.Spec.Service = "servicenetworking.googleapis.com"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a third-party producer service name", func() {
		target := minimal()
		target.Spec.Service = "peering-service.example-producer.com"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept multiple reserved ranges", func() {
		target := minimal()
		target.Spec.ReservedPeeringRanges = []*foreignkeyv1.StringValueOrRef{
			litRef("psa-range-a"),
			litRef("psa-range-b"),
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept update_on_creation_fail", func() {
		target := minimal()
		target.Spec.UpdateOnCreationFail = true
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept reference-shaped network and ranges", func() {
		target := minimal()
		target.Spec.Network = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
				ValueFrom: &foreignkeyv1.ValueFromRef{Name: "my-vpc"},
			},
		}
		target.Spec.ReservedPeeringRanges = []*foreignkeyv1.StringValueOrRef{
			{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{Name: "my-psa-range"},
				},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing network", func() {
		target := minimal()
		target.Spec.Network = nil
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject empty reserved_peering_ranges", func() {
		target := minimal()
		target.Spec.ReservedPeeringRanges = nil
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("reserved_peering_ranges"))
	})

	ginkgo.It("should reject a service name without a domain suffix", func() {
		target := minimal()
		target.Spec.Service = "servicenetworking"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.ToLower(err.Error())).To(gomega.ContainSubstring("service"))
	})

	ginkgo.It("should reject a service name with uppercase characters", func() {
		target := minimal()
		target.Spec.Service = "ServiceNetworking.googleapis.com"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing spec", func() {
		target := minimal()
		target.Spec = nil
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing metadata", func() {
		target := minimal()
		target.Metadata = nil
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a wrong kind literal", func() {
		target := minimal()
		target.Kind = "GcpServiceNetworkingConnexion"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a wrong api_version literal", func() {
		target := minimal()
		target.ApiVersion = "gcp.planton.dev/v2"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept each deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
			target := minimal()
			target.Spec.DeletionPolicy = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "KEEP"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
