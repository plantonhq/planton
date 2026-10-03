package gcpcloudbuildrepositoryv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpCloudBuildRepositorySpec Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value}}
}

func reference(kind catalogkind.CatalogKind, name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Kind: kind, Name: name}}}
}

var _ = ginkgo.Describe("GcpCloudBuildRepositorySpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpCloudBuildRepository {
		return &GcpCloudBuildRepository{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpCloudBuildRepository",
			Metadata:   &shared.CatalogObjectMetadata{Name: "orders"},
			Spec: &GcpCloudBuildRepositorySpec{
				ParentConnection: reference(catalogkind.CatalogKind_GcpCloudBuildConnection, "github"),
				RemoteUri:        "https://github.com/acme/orders.git",
			},
		}
	}

	ginkgo.It("should accept a referenced and a literal parent connection", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())

		msg := minimal()
		msg.Spec.ParentConnection = literal("projects/ci/locations/us-central1/connections/github")
		msg.Spec.RepositoryId = "acme-orders"
		msg.Spec.Annotations = map[string]string{"team": "orders"}
		msg.Spec.DeletionPolicy = "ABANDON"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require the parent connection and the remote URI", func() {
		msg := minimal()
		msg.Spec.ParentConnection = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		msg = minimal()
		msg.Spec.RemoteUri = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should refuse a short connection name and a non-HTTPS URI", func() {
		msg := minimal()
		msg.Spec.ParentConnection = literal("github")
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		msg = minimal()
		msg.Spec.RemoteUri = "git@github.com:acme/orders.git"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a bad repository ID and an unknown deletion policy", func() {
		msg := minimal()
		msg.Spec.RepositoryId = "acme/orders"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		msg = minimal()
		msg.Spec.DeletionPolicy = "FORCE"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
