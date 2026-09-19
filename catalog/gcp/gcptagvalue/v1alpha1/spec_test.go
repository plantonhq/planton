package gcptagvaluev1alpha1

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
	ginkgo.RunSpecs(t, "GcpTagValueSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpTagValueSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpTagValue {
		return &GcpTagValue{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpTagValue",
			Metadata: &shared.CloudResourceMetadata{
				Name: "prod",
			},
			Spec: &GcpTagValueSpec{
				TagKey: litRef("tagKeys/281475647562788"),
			},
		}
	}

	expectError := func(target *GcpTagValue, substring string) {
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring(substring))
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a value under a key by literal name, named from metadata", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a key by reference, an explicit short name, and a description", func() {
		target := minimal()
		target.Spec.TagKey = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: "environment"}},
		}
		target.Spec.ShortName = "production"
		target.Spec.Description = "Customer-facing workloads"
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

	ginkgo.It("should reject a value without a key", func() {
		target := minimal()
		target.Spec.TagKey = nil
		expectError(target, "tag_key")
	})

	ginkgo.It("should reject a key literal that is not tagKeys/{id}", func() {
		for _, key := range []string{"environment", "281475647562788", "tagKeys/environment"} {
			target := minimal()
			target.Spec.TagKey = litRef(key)
			expectError(target, "tagKeys/{numeric_id}")
		}
	})

	ginkgo.It("should reject short names with the characters Google forbids", func() {
		for _, name := range []string{"prod/eu", `prod\eu`, "prod'eu", `prod"eu`} {
			target := minimal()
			target.Spec.ShortName = name
			expectError(target, "must not contain a slash, backslash, or quote")
		}
	})

	ginkgo.It("should reject an over-long description and an unknown deletion_policy", func() {
		target := minimal()
		target.Spec.Description = string(make([]byte, 257))
		expectError(target, "description")

		target = minimal()
		target.Spec.DeletionPolicy = "RETAIN"
		expectError(target, "deletion_policy must be one of")
	})
})
