package cloudflaresnippetrulesv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestCloudflareSnippetRulesSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareSnippetRulesSpec Custom Validation Tests")
}

func zoneRef() *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "0da42c8d2132a9ddaf714f9e7c920711"},
	}
}

func snippetRef(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: name},
	}
}

func validRules(spec *CloudflareSnippetRulesSpec) *CloudflareSnippetRules {
	return &CloudflareSnippetRules{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareSnippetRules",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-snippet-rules",
		},
		Spec: spec,
	}
}

var _ = ginkgo.Describe("CloudflareSnippetRulesSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a single-rule table", func() {
			input := validRules(&CloudflareSnippetRulesSpec{
				ZoneId: zoneRef(),
				Rules: []*CloudflareSnippetRule{
					{
						Expression:  `starts_with(http.request.uri.path, "/legacy")`,
						SnippetName: snippetRef("redirect_legacy_urls"),
					},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a multi-rule table with a staged disabled rule", func() {
			input := validRules(&CloudflareSnippetRulesSpec{
				ZoneId: zoneRef(),
				Rules: []*CloudflareSnippetRule{
					{
						Expression:  `starts_with(http.request.uri.path, "/legacy")`,
						SnippetName: snippetRef("redirect_legacy_urls"),
						Description: "send legacy URLs to the redirect snippet",
					},
					{
						Expression:  `http.host eq "beta.example.com"`,
						SnippetName: snippetRef("header_rewrites"),
						Enabled:     proto.Bool(false),
					},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject an empty rules list", func() {
			input := validRules(&CloudflareSnippetRulesSpec{
				ZoneId: zoneRef(),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a rule without an expression", func() {
			input := validRules(&CloudflareSnippetRulesSpec{
				ZoneId: zoneRef(),
				Rules: []*CloudflareSnippetRule{
					{SnippetName: snippetRef("redirect_legacy_urls")},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a rule without a snippet reference", func() {
			input := validRules(&CloudflareSnippetRulesSpec{
				ZoneId: zoneRef(),
				Rules: []*CloudflareSnippetRule{
					{Expression: `starts_with(http.request.uri.path, "/legacy")`},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing zone_id", func() {
			input := validRules(&CloudflareSnippetRulesSpec{
				Rules: []*CloudflareSnippetRule{
					{
						Expression:  `starts_with(http.request.uri.path, "/legacy")`,
						SnippetName: snippetRef("redirect_legacy_urls"),
					},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
