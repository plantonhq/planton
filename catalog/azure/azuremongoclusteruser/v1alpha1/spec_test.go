package azuremongoclusteruserv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureMongoClusterUserSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureMongoClusterUserSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// ref builds a StringValueOrRef carrying a valueFrom reference.
func ref(kind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{
				Name:      name,
				FieldPath: fieldPath,
			},
		},
	}
}

const testClusterId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.DocumentDB/mongoClusters/acme-orders-db"

const testObjectId = "11111111-2222-3333-4444-555555555555"

// validResource returns a valid user grant that individual cases
// mutate into the shape under test.
func validResource() *AzureMongoClusterUser {
	return &AzureMongoClusterUser{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureMongoClusterUser",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-mongou",
		},
		Spec: &AzureMongoClusterUserSpec{
			MongoClusterId: literal(testClusterId),
			ObjectId:       literal(testObjectId),
			PrincipalType:  "servicePrincipal",
			Roles: []*AzureMongoClusterUserRole{
				{Database: "admin", Role: "root"},
			},
		},
	}
}

var _ = ginkgo.Describe("AzureMongoClusterUserSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_mongo_cluster_user", func() {

			ginkgo.It("should not return a validation error for the minimal shape", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the user principal type", func() {
				input := validResource()
				input.Spec.PrincipalType = "user"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an object id supplied by reference", func() {
				input := validResource()
				input.Spec.ObjectId = ref("AzureUserAssignedIdentity", "app-uai", "status.outputs.principal_id")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_mongo_cluster_user", func() {

			ginkgo.It("should reject a missing cluster id", func() {
				input := validResource()
				input.Spec.MongoClusterId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing object id", func() {
				input := validResource()
				input.Spec.ObjectId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a literal object id that is not a UUID", func() {
				input := validResource()
				input.Spec.ObjectId = literal("app-uai")
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing or unknown principal type", func() {
				input := validResource()
				input.Spec.PrincipalType = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.PrincipalType = "group"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.PrincipalType = "ServicePrincipal"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty roles list", func() {
				input := validResource()
				input.Spec.Roles = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a role without a database or with an unknown role name", func() {
				input := validResource()
				input.Spec.Roles = []*AzureMongoClusterUserRole{{Database: "", Role: "root"}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Roles = []*AzureMongoClusterUserRole{{Database: "admin", Role: "dbOwner"}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})
