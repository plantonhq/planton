package gcpbigquerycapacitycommitmentv1alpha1

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
	ginkgo.RunSpecs(t, "GcpBigQueryCapacityCommitmentSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpBigQueryCapacityCommitmentSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpBigQueryCapacityCommitment {
		return &GcpBigQueryCapacityCommitment{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpBigQueryCapacityCommitment",
			Metadata:   &shared.CloudResourceMetadata{Name: "annual-100"},
			Spec:       &GcpBigQueryCapacityCommitmentSpec{SlotCount: 100, Plan: "ANNUAL"},
		}
	}

	ginkgo.It("should accept a commitment with slots and a plan", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every field set", func() {
		msg := minimal()
		msg.Spec.ProjectId = litRef("bq-admin")
		msg.Spec.Location = "US"
		msg.Spec.CapacityCommitmentId = "annual-100-2026"
		msg.Spec.RenewalPlan = "ANNUAL"
		msg.Spec.Edition = "ENTERPRISE"
		msg.Spec.EnforceSingleAdminProjectPerOrg = true
		msg.Spec.DeletionPolicy = "ABANDON"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require slots and a plan", func() {
		msg := minimal()
		msg.Spec.SlotCount = 0
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.Plan = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a malformed ID, an unknown edition, and an unknown deletion policy", func() {
		for _, id := range []string{"Annual", "-annual", "annual-", "annual_100"} {
			msg := minimal()
			msg.Spec.CapacityCommitmentId = id
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), id)
		}
		msg := minimal()
		msg.Spec.Edition = "FLAT_RATE"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
