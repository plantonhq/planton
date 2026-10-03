package cloudflarecachesettingsv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestCloudflareCacheSettingsSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareCacheSettingsSpec Custom Validation Tests")
}

func boolPtr(b bool) *bool { return &b }

func zoneRef() *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "023e105f4ecef8ad9ca31a8372d0c353"}}
}

func validCacheSettings(spec *CloudflareCacheSettingsSpec) *CloudflareCacheSettings {
	return &CloudflareCacheSettings{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareCacheSettings",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-cache-settings",
		},
		Spec: spec,
	}
}

var _ = ginkgo.Describe("CloudflareCacheSettingsSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal spec managing one toggle", func() {
			input := validCacheSettings(&CloudflareCacheSettingsSpec{
				ZoneId:           zoneRef(),
				SmartTieredCache: boolPtr(true),
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a spec managing only cache variants", func() {
			input := validCacheSettings(&CloudflareCacheSettingsSpec{
				ZoneId: zoneRef(),
				CacheVariants: &CloudflareCacheSettingsVariants{
					Jpg:  []string{"image/webp", "image/avif"},
					Jpeg: []string{"image/webp"},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an explicit false (managed-off) toggle", func() {
			input := validCacheSettings(&CloudflareCacheSettingsSpec{
				ZoneId:           zoneRef(),
				ArgoSmartRouting: boolPtr(false),
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a spec without zone_id", func() {
			input := validCacheSettings(&CloudflareCacheSettingsSpec{
				SmartTieredCache: boolPtr(true),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a spec that manages no settings at all", func() {
			input := validCacheSettings(&CloudflareCacheSettingsSpec{
				ZoneId: zoneRef(),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
