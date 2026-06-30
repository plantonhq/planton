package gcpgkeworkloadidentitybindingv1

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"

	"buf.build/go/protovalidate"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestGcpGkeWorkloadIdentityBindingSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpGkeWorkloadIdentityBindingSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("GcpGkeWorkloadIdentityBindingSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("gcp_gke_workload_identity_binding", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &GcpGkeWorkloadIdentityBinding{
					ApiVersion: "gcp.planton.dev/v1",
					Kind:       "GcpGkeWorkloadIdentityBinding",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-workload-identity-binding",
					},
					Spec: &GcpGkeWorkloadIdentityBindingSpec{
						ProjectId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-project-123"},
						},
						ServiceAccountEmail: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "cert-manager@test-project-123.iam.gserviceaccount.com"},
						},
						KsaNamespace: "cert-manager",
						KsaName:      "cert-manager",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
