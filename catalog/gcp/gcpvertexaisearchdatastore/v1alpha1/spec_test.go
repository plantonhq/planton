package gcpvertexaisearchdatastorev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpVertexAiSearchDataStoreSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpVertexAiSearchDataStoreSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// A structured store enrolled in search -- Google's own test shape.
	minimal := func() *GcpVertexAiSearchDataStore {
		return &GcpVertexAiSearchDataStore{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpVertexAiSearchDataStore",
			Metadata:   &shared.CloudResourceMetadata{Name: "product-catalog"},
			Spec: &GcpVertexAiSearchDataStoreSpec{
				Location:         "global",
				IndustryVertical: "GENERIC",
				ContentConfig:    "NO_CONTENT",
				SolutionTypes:    []string{"SOLUTION_TYPE_SEARCH"},
			},
		}
	}

	website := func() *GcpVertexAiSearchDataStore {
		msg := minimal()
		msg.Spec.ContentConfig = "PUBLIC_WEBSITE"
		msg.Spec.TargetSites = []*GcpVertexAiSearchDataStoreTargetSite{{
			ProvidedUriPattern: "cloud.google.com/docs/*",
			Type:               "INCLUDE",
		}}
		return msg
	}

	ginkgo.It("should accept a structured store", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a website store with target sites", func() {
		gomega.Expect(validator.Validate(website())).To(gomega.Succeed())
	})

	ginkgo.It("should accept advanced site search with sitemaps and its config", func() {
		msg := website()
		msg.Spec.CreateAdvancedSiteSearch = true
		msg.Spec.AdvancedSiteSearchConfig = &GcpVertexAiSearchDataStoreAdvancedSiteSearchConfig{DisableInitialIndex: true}
		msg.Spec.SitemapUris = []string{"https://www.example.com/sitemap.xml"}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a custom schema, every parsing lever, CMEK, and an explicit id", func() {
		msg := minimal()
		msg.Spec.ProjectId = litRef("ai-project")
		msg.Spec.DataStoreId = "product-catalog-v2"
		msg.Spec.DisplayName = "Product catalog"
		msg.Spec.ContentConfig = "CONTENT_REQUIRED"
		msg.Spec.SolutionTypes = []string{"SOLUTION_TYPE_SEARCH", "SOLUTION_TYPE_CHAT"}
		msg.Spec.AclEnabled = true
		msg.Spec.SkipDefaultSchemaCreation = true
		msg.Spec.Schema = &GcpVertexAiSearchDataStoreSchema{
			SchemaId:   "product-schema",
			JsonSchema: `{"$schema":"https://json-schema.org/draft/2020-12/schema","type":"object","properties":{"title":{"type":"string","keyPropertyMapping":"title"}}}`,
		}
		msg.Spec.KmsKeyName = litRef("projects/ai-project/locations/us/keyRings/ai/cryptoKeys/search")
		msg.Spec.DocumentProcessingConfig = &GcpVertexAiSearchDataStoreDocumentProcessingConfig{
			ChunkingConfig: &GcpVertexAiSearchDataStoreChunkingConfig{
				ChunkSize:               proto.Int32(400),
				IncludeAncestorHeadings: true,
			},
			DefaultParsingConfig: &GcpVertexAiSearchDataStoreParsingConfig{
				LayoutParsingConfig: &GcpVertexAiSearchDataStoreLayoutParsingConfig{
					EnableTableAnnotation: true,
					ExcludeHtmlElements:   []string{"nav", "footer"},
				},
			},
			ParsingConfigOverrides: []*GcpVertexAiSearchDataStoreParsingConfigOverride{{
				FileType: "pdf",
				ParsingConfig: &GcpVertexAiSearchDataStoreParsingConfig{
					OcrParsingConfig: &GcpVertexAiSearchDataStoreOcrParsingConfig{UseNativeText: true},
				},
			}, {
				FileType:      "html",
				ParsingConfig: &GcpVertexAiSearchDataStoreParsingConfig{DigitalParsing: true},
			}},
		}
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require location and industry_vertical from Google's lists", func() {
		msg := minimal()
		msg.Spec.Location = "us-central1"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.IndustryVertical = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.IndustryVertical = "RETAIL"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a data store id that is not RFC 1034 and unknown enum values", func() {
		msg := minimal()
		msg.Spec.DataStoreId = "Product_Catalog"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.ContentConfig = "WEBSITE"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.SolutionTypes = []string{"SOLUTION_TYPE_AGENT"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should wall target sites to website stores", func() {
		msg := minimal()
		msg.Spec.TargetSites = []*GcpVertexAiSearchDataStoreTargetSite{{ProvidedUriPattern: "example.com"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = website()
		msg.Spec.TargetSites[0].Type = "ALLOW"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = website()
		msg.Spec.TargetSites[0].ProvidedUriPattern = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should wall sitemaps and the advanced config to advanced site search", func() {
		msg := website()
		msg.Spec.SitemapUris = []string{"https://www.example.com/sitemap.xml"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = website()
		msg.Spec.AdvancedSiteSearchConfig = &GcpVertexAiSearchDataStoreAdvancedSiteSearchConfig{DisableAutomaticRefresh: true}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.CreateAdvancedSiteSearch = true
		msg.Spec.SitemapUris = []string{"https://www.example.com/sitemap.xml"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require the default schema skipped before a custom schema", func() {
		msg := minimal()
		msg.Spec.Schema = &GcpVertexAiSearchDataStoreSchema{SchemaId: "s", JsonSchema: "{}"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.SkipDefaultSchemaCreation = true
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.Schema.JsonSchema = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should allow at most one parser per parsing config and a known file type", func() {
		msg := minimal()
		msg.Spec.DocumentProcessingConfig = &GcpVertexAiSearchDataStoreDocumentProcessingConfig{
			DefaultParsingConfig: &GcpVertexAiSearchDataStoreParsingConfig{
				DigitalParsing:   true,
				OcrParsingConfig: &GcpVertexAiSearchDataStoreOcrParsingConfig{},
			},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DocumentProcessingConfig = &GcpVertexAiSearchDataStoreDocumentProcessingConfig{
			ParsingConfigOverrides: []*GcpVertexAiSearchDataStoreParsingConfigOverride{{
				FileType:      "txt",
				ParsingConfig: &GcpVertexAiSearchDataStoreParsingConfig{DigitalParsing: true},
			}},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.DocumentProcessingConfig = &GcpVertexAiSearchDataStoreDocumentProcessingConfig{
			ChunkingConfig: &GcpVertexAiSearchDataStoreChunkingConfig{ChunkSize: proto.Int32(50)},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
