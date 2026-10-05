package azurecomputegalleryv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureComputeGallerySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureComputeGallerySpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// validResource returns a valid gallery that individual cases mutate
// into the shape under test.
func validResource() *AzureComputeGallery {
	return &AzureComputeGallery{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureComputeGallery",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-gallery",
		},
		Spec: &AzureComputeGallerySpec{
			ResourceGroup: literal("app-rg"),
			Name:          "platform.images",
			Region:        "eastus",
		},
	}
}

// validCommunitySharing returns a complete community-sharing block.
func validCommunitySharing() *AzureComputeGallerySharing {
	return &AzureComputeGallerySharing{
		Permission: "Community",
		CommunityGallery: &AzureComputeGalleryCommunitySharing{
			Eula:           "https://example.com/eula",
			Prefix:         "acmeimages",
			PublisherEmail: "images@acme.example",
			PublisherUri:   "https://acme.example",
		},
	}
}

var _ = ginkgo.Describe("AzureComputeGallerySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_compute_gallery", func() {

			ginkgo.It("should not return a validation error for the minimal shape", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a description and tags", func() {
				input := validResource()
				input.Spec.Description = "The platform team's golden images"
				input.Spec.Tags = map[string]string{"costCenter": "platform"}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept names with dots and underscores up to 80 characters", func() {
				input := validResource()
				input.Spec.Name = "platform_images.prod"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				input.Spec.Name = strings.Repeat("a", 80)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept Private and Groups sharing without a community block", func() {
				input := validResource()
				input.Spec.Sharing = &AzureComputeGallerySharing{Permission: "Private"}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				input.Spec.Sharing = &AzureComputeGallerySharing{Permission: "Groups"}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept Community sharing with a complete community block", func() {
				input := validResource()
				input.Spec.Sharing = validCommunitySharing()
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept community prefixes at both length bounds", func() {
				input := validResource()
				input.Spec.Sharing = validCommunitySharing()
				input.Spec.Sharing.CommunityGallery.Prefix = "acme5"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				input.Spec.Sharing.CommunityGallery.Prefix = strings.Repeat("a", 16)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_compute_gallery", func() {

			ginkgo.It("should reject a missing resource group", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing region", func() {
				input := validResource()
				input.Spec.Region = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject names with dashes (galleries forbid them)", func() {
				input := validResource()
				input.Spec.Name = "platform-images"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject names over 80 characters and empty names", func() {
				input := validResource()
				input.Spec.Name = strings.Repeat("a", 81)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Name = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown sharing permission", func() {
				input := validResource()
				input.Spec.Sharing = &AzureComputeGallerySharing{Permission: "Public"}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject Community sharing without a community block", func() {
				input := validResource()
				input.Spec.Sharing = &AzureComputeGallerySharing{Permission: "Community"}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject community prefixes outside 5-16 alphanumerics", func() {
				input := validResource()
				input.Spec.Sharing = validCommunitySharing()
				input.Spec.Sharing.CommunityGallery.Prefix = "acme"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Sharing.CommunityGallery.Prefix = strings.Repeat("a", 17)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Sharing.CommunityGallery.Prefix = "acme-img"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a community block missing its required fields", func() {
				input := validResource()
				input.Spec.Sharing = validCommunitySharing()
				input.Spec.Sharing.CommunityGallery.Eula = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Sharing = validCommunitySharing()
				input.Spec.Sharing.CommunityGallery.PublisherEmail = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Sharing = validCommunitySharing()
				input.Spec.Sharing.CommunityGallery.PublisherUri = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})
