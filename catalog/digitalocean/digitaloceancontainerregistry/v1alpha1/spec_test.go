package digitaloceancontainerregistryv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/catalog/digitalocean"
	"github.com/plantonhq/planton/shared"
)

func TestDigitalOceanContainerRegistrySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanContainerRegistrySpec Custom Validation Tests")
}

// int32Ptr returns a pointer for optional int32 fields.
func int32Ptr(i int32) *int32 {
	return &i
}

// registry returns a minimal valid registry the tests mutate per case.
func registry() *DigitalOceanContainerRegistry {
	return &DigitalOceanContainerRegistry{
		ApiVersion: "digital-ocean.planton.dev/v1alpha1",
		Kind:       "DigitalOceanContainerRegistry",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-registry",
		},
		Spec: &DigitalOceanContainerRegistrySpec{
			Name:             "acme-registry",
			SubscriptionTier: DigitalOceanContainerRegistryTier_starter,
		},
	}
}

var _ = ginkgo.Describe("DigitalOceanContainerRegistrySpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("accepts a minimal registry without a region (DigitalOcean chooses one)", func() {
			gomega.Expect(protovalidate.Validate(registry())).To(gomega.BeNil())
		})

		ginkgo.It("accepts a registry with an explicit region", func() {
			input := registry()
			input.Spec.Region = digitalocean.DigitalOceanRegion_nyc3
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts every subscription tier", func() {
			for _, tier := range []DigitalOceanContainerRegistryTier{
				DigitalOceanContainerRegistryTier_starter,
				DigitalOceanContainerRegistryTier_basic,
				DigitalOceanContainerRegistryTier_professional,
			} {
				input := registry()
				input.Spec.SubscriptionTier = tier
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("accepts docker credentials with defaults (read-only, provider-default expiry)", func() {
			input := registry()
			input.Spec.DockerCredentials = &DigitalOceanContainerRegistryDockerCredentials{}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts write credentials with an explicit expiry", func() {
			input := registry()
			input.Spec.DockerCredentials = &DigitalOceanContainerRegistryDockerCredentials{
				Write:         true,
				ExpirySeconds: int32Ptr(3600),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts the API-maximum expiry", func() {
			input := registry()
			input.Spec.DockerCredentials = &DigitalOceanContainerRegistryDockerCredentials{
				ExpirySeconds: int32Ptr(1576800000),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a single-character registry name", func() {
			input := registry()
			input.Spec.Name = "a"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("name validation", func() {

		ginkgo.It("rejects an empty name", func() {
			input := registry()
			input.Spec.Name = ""
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects uppercase letters", func() {
			input := registry()
			input.Spec.Name = "Acme-Registry"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a leading hyphen", func() {
			input := registry()
			input.Spec.Name = "-acme"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a trailing hyphen", func() {
			input := registry()
			input.Spec.Name = "acme-"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a name longer than 63 characters", func() {
			input := registry()
			input.Spec.Name = "a234567890123456789012345678901234567890123456789012345678901234"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("subscription_tier validation", func() {

		ginkgo.It("rejects an unspecified tier", func() {
			input := registry()
			input.Spec.SubscriptionTier = DigitalOceanContainerRegistryTier_digitalocean_container_registry_tier_unspecified
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("docker_credentials validation", func() {

		ginkgo.It("rejects a negative expiry", func() {
			input := registry()
			input.Spec.DockerCredentials = &DigitalOceanContainerRegistryDockerCredentials{
				ExpirySeconds: int32Ptr(-1),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an expiry above the API maximum", func() {
			input := registry()
			input.Spec.DockerCredentials = &DigitalOceanContainerRegistryDockerCredentials{
				ExpirySeconds: int32Ptr(1576800001),
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
