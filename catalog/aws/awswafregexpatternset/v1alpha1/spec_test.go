package awswafregexpatternsetv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestAwsWafRegexPatternSetSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsWafRegexPatternSetSpec Validation Suite")
}

// helper to create a minimal valid AwsWafRegexPatternSet wrapper.
func minimalPatternSet(spec *AwsWafRegexPatternSetSpec) *AwsWafRegexPatternSet {
	return &AwsWafRegexPatternSet{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsWafRegexPatternSet",
		Metadata:   &shared.CatalogObjectMetadata{Name: "test-pattern-set"},
		Spec:       spec,
	}
}

// helper to create a minimal valid spec.
func minimalSpec() *AwsWafRegexPatternSetSpec {
	return &AwsWafRegexPatternSetSpec{
		Region:             "us-west-2",
		Scope:              "REGIONAL",
		RegularExpressions: []string{"^/wp-admin/.*"},
	}
}

var _ = ginkgo.Describe("AwsWafRegexPatternSetSpec validations", func() {

	ginkgo.It("accepts a minimal REGIONAL pattern set", func() {
		gomega.Expect(protovalidate.Validate(minimalPatternSet(minimalSpec()))).To(gomega.BeNil())
	})

	ginkgo.It("accepts a CLOUDFRONT pattern set in us-east-1", func() {
		spec := minimalSpec()
		spec.Scope = "CLOUDFRONT"
		spec.Region = "us-east-1"
		spec.Description = "Blocked admin-path probes"
		gomega.Expect(protovalidate.Validate(minimalPatternSet(spec))).To(gomega.BeNil())
	})

	ginkgo.It("rejects CLOUDFRONT scope outside us-east-1", func() {
		spec := minimalSpec()
		spec.Scope = "CLOUDFRONT"
		gomega.Expect(protovalidate.Validate(minimalPatternSet(spec))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an invalid scope", func() {
		spec := minimalSpec()
		spec.Scope = "GLOBAL"
		gomega.Expect(protovalidate.Validate(minimalPatternSet(spec))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an empty expression list", func() {
		spec := minimalSpec()
		spec.RegularExpressions = nil
		gomega.Expect(protovalidate.Validate(minimalPatternSet(spec))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects an expression over 200 characters", func() {
		spec := minimalSpec()
		long := ""
		for i := 0; i < 21; i++ {
			long += "0123456789"
		}
		spec.RegularExpressions = []string{long}
		gomega.Expect(protovalidate.Validate(minimalPatternSet(spec))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a missing region", func() {
		spec := minimalSpec()
		spec.Region = ""
		gomega.Expect(protovalidate.Validate(minimalPatternSet(spec))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a description over 256 characters", func() {
		spec := minimalSpec()
		long := ""
		for i := 0; i < 26; i++ {
			long += "0123456789"
		}
		spec.Description = long
		gomega.Expect(protovalidate.Validate(minimalPatternSet(spec))).NotTo(gomega.BeNil())
	})

	ginkgo.It("rejects a description with characters outside AWS's WAF charset", func() {
		spec := minimalSpec()
		spec.Description = "Probe paths (E2E)"
		gomega.Expect(protovalidate.Validate(minimalPatternSet(spec))).NotTo(gomega.BeNil())
	})
})
