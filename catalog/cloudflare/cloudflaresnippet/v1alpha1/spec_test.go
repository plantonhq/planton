package cloudflaresnippetv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestCloudflareSnippetSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareSnippetSpec Custom Validation Tests")
}

func zoneRef() *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "0da42c8d2132a9ddaf714f9e7c920711"},
	}
}

func validSnippet(spec *CloudflareSnippetSpec) *CloudflareSnippet {
	return &CloudflareSnippet{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareSnippet",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-snippet",
		},
		Spec: spec,
	}
}

const mainJs = "export default { async fetch(request) { return fetch(request); } };"

var _ = ginkgo.Describe("CloudflareSnippetSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a single-file snippet", func() {
			input := validSnippet(&CloudflareSnippetSpec{
				ZoneId:      zoneRef(),
				SnippetName: "redirect_legacy_urls",
				Files: []*CloudflareSnippetFile{
					{Name: "main.js", Content: mainJs},
				},
				MainModule: "main.js",
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a multi-file snippet whose main module is in the list", func() {
			input := validSnippet(&CloudflareSnippetSpec{
				ZoneId:      zoneRef(),
				SnippetName: "header_rewrites",
				Files: []*CloudflareSnippetFile{
					{Name: "main.js", Content: mainJs},
					{Name: "helpers.js", Content: "export const x = 1;"},
				},
				MainModule: "main.js",
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a main_module that names no file", func() {
			input := validSnippet(&CloudflareSnippetSpec{
				ZoneId:      zoneRef(),
				SnippetName: "redirect_legacy_urls",
				Files: []*CloudflareSnippetFile{
					{Name: "main.js", Content: mainJs},
				},
				MainModule: "index.js",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a snippet name with hyphens", func() {
			input := validSnippet(&CloudflareSnippetSpec{
				ZoneId:      zoneRef(),
				SnippetName: "redirect-legacy-urls",
				Files: []*CloudflareSnippetFile{
					{Name: "main.js", Content: mainJs},
				},
				MainModule: "main.js",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an empty file list", func() {
			input := validSnippet(&CloudflareSnippetSpec{
				ZoneId:      zoneRef(),
				SnippetName: "redirect_legacy_urls",
				MainModule:  "main.js",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a file without content", func() {
			input := validSnippet(&CloudflareSnippetSpec{
				ZoneId:      zoneRef(),
				SnippetName: "redirect_legacy_urls",
				Files: []*CloudflareSnippetFile{
					{Name: "main.js"},
				},
				MainModule: "main.js",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing zone_id", func() {
			input := validSnippet(&CloudflareSnippetSpec{
				SnippetName: "redirect_legacy_urls",
				Files: []*CloudflareSnippetFile{
					{Name: "main.js", Content: mainJs},
				},
				MainModule: "main.js",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
