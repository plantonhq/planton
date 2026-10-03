package awsmemorydbaclv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsMemorydbAclSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsMemorydbAclSpec Validation Tests")
}

// applicationAcl is the common production shape: an ACL granting a set of
// per-application users access to the clusters that attach it.
func applicationAcl() *AwsMemorydbAcl {
	return &AwsMemorydbAcl{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsMemorydbAcl",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "payments-env-acl",
		},
		Spec: &AwsMemorydbAclSpec{
			Region: "us-west-2",
			UserNames: []*foreignkeyv1.StringValueOrRef{
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "orders-service"}},
				{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{
					Name: "analytics-service",
				}}},
			},
		},
	}
}

var _ = ginkgo.Describe("AwsMemorydbAclSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_memorydb_acl", func() {

			ginkgo.It("should accept an ACL with literal and referenced members", func() {
				err := protovalidate.Validate(applicationAcl())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an empty ACL (MemoryDB has no mandatory member)", func() {
				input := applicationAcl()
				input.Spec.UserNames = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("required fields", func() {

			ginkgo.It("should reject a missing region", func() {
				input := applicationAcl()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing spec", func() {
				input := applicationAcl()
				input.Spec = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a wrong kind constant", func() {
				input := applicationAcl()
				input.Kind = "AwsMemorydbAclx"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
