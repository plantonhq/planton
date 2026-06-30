package digitaloceanappplatformservicev1

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"

	"buf.build/go/protovalidate"
	"github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestDigitalOceanAppPlatformServiceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanAppPlatformServiceSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("DigitalOceanAppPlatformServiceSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("digitalocean_app_platform_service", func() {

			ginkgo.It("should not return a validation error for minimal valid fields with git source", func() {
				input := &DigitalOceanAppPlatformService{
					ApiVersion: "digital-ocean.planton.dev/v1",
					Kind:       "DigitalOceanAppPlatformService",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-app-service",
					},
					Spec: &DigitalOceanAppPlatformServiceSpec{
						ServiceName:      "test-service",
						Region:           digitalocean.DigitalOceanRegion_nyc3,
						ServiceType:      DigitalOceanAppPlatformServiceType_web_service,
						InstanceSizeSlug: DigitalOceanAppPlatformInstanceSize_basic_xxs,
						Source: &DigitalOceanAppPlatformServiceSpec_GitSource{
							GitSource: &DigitalOceanAppPlatformGitSource{
								RepoUrl: "https://github.com/example/repo.git",
								Branch:  "main",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for minimal valid fields with image source", func() {
				input := &DigitalOceanAppPlatformService{
					ApiVersion: "digital-ocean.planton.dev/v1",
					Kind:       "DigitalOceanAppPlatformService",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-app-service",
					},
					Spec: &DigitalOceanAppPlatformServiceSpec{
						ServiceName:      "test-service",
						Region:           digitalocean.DigitalOceanRegion_nyc3,
						ServiceType:      DigitalOceanAppPlatformServiceType_web_service,
						InstanceSizeSlug: DigitalOceanAppPlatformInstanceSize_basic_xxs,
						Source: &DigitalOceanAppPlatformServiceSpec_ImageSource{
							ImageSource: &DigitalOceanAppPlatformRegistrySource{
								Registry: &foreignkeyv1.StringValueOrRef{
									LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "registry.digitalocean.com/myregistry"},
								},
								Repository: "myapp",
								Tag:        "latest",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error when autoscale is properly configured", func() {
				input := &DigitalOceanAppPlatformService{
					ApiVersion: "digital-ocean.planton.dev/v1",
					Kind:       "DigitalOceanAppPlatformService",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-app-service",
					},
					Spec: &DigitalOceanAppPlatformServiceSpec{
						ServiceName:      "test-service",
						Region:           digitalocean.DigitalOceanRegion_nyc3,
						ServiceType:      DigitalOceanAppPlatformServiceType_web_service,
						InstanceSizeSlug: DigitalOceanAppPlatformInstanceSize_basic_xxs,
						EnableAutoscale:  true,
						MinInstanceCount: 2,
						MaxInstanceCount: 5,
						Source: &DigitalOceanAppPlatformServiceSpec_GitSource{
							GitSource: &DigitalOceanAppPlatformGitSource{
								RepoUrl: "https://github.com/example/repo.git",
								Branch:  "main",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
