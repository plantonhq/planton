package gcpsharedvpchostv1alpha1

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
	ginkgo.RunSpecs(t, "GcpSharedVpcHostSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpSharedVpcHostSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpSharedVpcHost {
		return &GcpSharedVpcHost{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpSharedVpcHost",
			Metadata: &shared.CloudResourceMetadata{
				Name: "network-host",
			},
			Spec: &GcpSharedVpcHostSpec{},
		}
	}

	expectError := func(target *GcpSharedVpcHost, substring string) {
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring(substring))
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept an empty spec (the provider's default project becomes the host)", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept the host project by literal ID and by reference", func() {
		target := minimal()
		target.Spec.ProjectId = litRef("acme-network-host")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())

		target = minimal()
		target.Spec.ProjectId = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: "network-host"}},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept every deletion_policy value and the empty default", func() {
		for _, policy := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			target := minimal()
			target.Spec.DeletionPolicy = policy
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed(), policy)
		}
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject an unknown deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "RETAIN"
		expectError(target, "deletion_policy must be one of")
	})

	ginkgo.It("should reject a wrong kind or api_version", func() {
		target := minimal()
		target.Kind = "GcpProject"
		expectError(target, "kind")

		target = minimal()
		target.ApiVersion = "gcp.planton.dev/v1"
		expectError(target, "api_version")
	})
})
