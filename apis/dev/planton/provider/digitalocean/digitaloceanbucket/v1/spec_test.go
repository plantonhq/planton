package digitaloceanbucketv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestDigitalOceanBucketSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanBucketSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("DigitalOceanBucketSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("digitalocean_bucket", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &DigitalOceanBucket{
					ApiVersion: "digital-ocean.planton.dev/v1",
					Kind:       "DigitalOceanBucket",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-bucket",
					},
					Spec: &DigitalOceanBucketSpec{
						BucketName: "test-bucket",
						Region:     digitalocean.DigitalOceanRegion_nyc3,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
