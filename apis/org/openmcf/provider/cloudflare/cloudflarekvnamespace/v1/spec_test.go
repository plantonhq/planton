package cloudflarekvnamespacev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestCloudflareKvNamespaceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareKvNamespaceSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CloudflareKvNamespaceSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("cloudflare_kv_namespace", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &CloudflareKvNamespace{
					ApiVersion: "cloudflare.openmcf.org/v1",
					Kind:       "CloudflareKvNamespace",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-kv-namespace",
					},
					Spec: &CloudflareKvNamespaceSpec{
						NamespaceName: "test-namespace",
						AccountId:     "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("cloudflare_kv_namespace", func() {

			ginkgo.It("should return a validation error when account_id is missing", func() {
				input := &CloudflareKvNamespace{
					ApiVersion: "cloudflare.openmcf.org/v1",
					Kind:       "CloudflareKvNamespace",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-kv-namespace",
					},
					Spec: &CloudflareKvNamespaceSpec{
						NamespaceName: "test-namespace",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when account_id is not 32 hex characters", func() {
				input := &CloudflareKvNamespace{
					ApiVersion: "cloudflare.openmcf.org/v1",
					Kind:       "CloudflareKvNamespace",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-kv-namespace",
					},
					Spec: &CloudflareKvNamespaceSpec{
						NamespaceName: "test-namespace",
						AccountId:     "not-a-valid-account-id",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
