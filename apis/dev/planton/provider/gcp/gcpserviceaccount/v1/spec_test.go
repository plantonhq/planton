package gcpserviceaccountv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestGcpServiceAccountSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpServiceAccountSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("GcpServiceAccountSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("gcp_service_account", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &GcpServiceAccount{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpServiceAccount",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-service-account",
					},
					Spec: &GcpServiceAccountSpec{
						ServiceAccountId: "test-sa-123",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
