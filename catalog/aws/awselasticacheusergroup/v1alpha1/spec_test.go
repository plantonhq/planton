package awselasticacheusergroupv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsElasticacheUserGroupSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsElasticacheUserGroupSpec Validation Tests")
}

func literalRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// minimalGroup carries the mandatory default user plus one application user.
func minimalGroup() *AwsElasticacheUserGroup {
	return &AwsElasticacheUserGroup{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsElasticacheUserGroup",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "orders-rbac",
		},
		Spec: &AwsElasticacheUserGroupSpec{
			Region: "us-west-2",
			Engine: "redis",
			UserIds: []*foreignkeyv1.StringValueOrRef{
				literalRef("rbac-default-user"),
				literalRef("orders-service"),
			},
		},
	}
}

var _ = ginkgo.Describe("AwsElasticacheUserGroupSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("aws_elasticache_user_group", func() {

			ginkgo.It("should accept a group with a default and an application user", func() {
				err := protovalidate.Validate(minimalGroup())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a single-member group", func() {
				input := minimalGroup()
				input.Spec.UserIds = input.Spec.UserIds[:1]
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the valkey engine", func() {
				input := minimalGroup()
				input.Spec.Engine = "valkey"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("aws_elasticache_user_group", func() {

			ginkgo.It("should reject a missing region", func() {
				input := minimalGroup()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unsupported engine", func() {
				input := minimalGroup()
				input.Spec.Engine = "memcached"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty membership list", func() {
				input := minimalGroup()
				input.Spec.UserIds = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
