package azurestoragetablev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureStorageTableSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureStorageTableSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const accountId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/plantonapp"

// minimal valid spec: an entities table.
func minimalSpec() *AzureStorageTable {
	return &AzureStorageTable{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureStorageTable",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-table",
		},
		Spec: &AzureStorageTableSpec{
			StorageAccountId: literal(accountId),
			TableName:        "AppEntities",
		},
	}
}

// validAcl is a fully-populated table policy (table policies require the
// full validity window).
func validAcl(id string) *AzureStorageTableAcl {
	return &AzureStorageTableAcl{
		Id: id,
		AccessPolicies: []*AzureStorageTableAclAccessPolicy{{
			Permissions: "r",
			Start:       "2026-07-01T00:00:00Z",
			Expiry:      "2027-07-01T00:00:00Z",
		}},
	}
}

var _ = ginkgo.Describe("AzureStorageTableSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal table", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept alphanumeric names within the length bounds", func() {
			for _, name := range []string{"abc", "Devices", "auditTrail2026"} {
				input := minimalSpec()
				input.Spec.TableName = name
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "name %q must be accepted", name)
			}
		})

		ginkgo.It("should accept an ACL with the full window", func() {
			input := minimalSpec()
			input.Spec.Acls = []*AzureStorageTableAcl{validAcl("readers")}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept full-access permission letters", func() {
			input := minimalSpec()
			acl := validAcl("writers")
			acl.AccessPolicies[0].Permissions = "raud"
			input.Spec.Acls = []*AzureStorageTableAcl{acl}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a valueFrom reference for the account", func() {
			input := minimalSpec()
			input.Spec.StorageAccountId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{
						Kind:      catalogkind.CatalogKind_AzureStorageAccount,
						Name:      "app-storage",
						FieldPath: "status.outputs.storage_account_id",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing storage account reference", func() {
			input := minimalSpec()
			input.Spec.StorageAccountId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject names that break the table naming rules", func() {
			for _, name := range []string{"", "ab", "9starts", "has-hyphen", "has_underscore", "table"} {
				input := minimalSpec()
				input.Spec.TableName = name
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "name %q must be rejected", name)
			}
		})

		ginkgo.It("should reject a name longer than 63 characters", func() {
			input := minimalSpec()
			input.Spec.TableName = "a123456789b123456789c123456789d123456789e123456789f123456789g123"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject more than five ACLs", func() {
			input := minimalSpec()
			for i := 0; i < 6; i++ {
				input.Spec.Acls = append(input.Spec.Acls, validAcl("policy-"+string(rune('a'+i))))
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a policy without a start", func() {
			input := minimalSpec()
			acl := validAcl("readers")
			acl.AccessPolicies[0].Start = ""
			input.Spec.Acls = []*AzureStorageTableAcl{acl}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a policy without an expiry", func() {
			input := minimalSpec()
			acl := validAcl("readers")
			acl.AccessPolicies[0].Expiry = ""
			input.Spec.Acls = []*AzureStorageTableAcl{acl}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject permission letters out of order or unknown", func() {
			for _, permissions := range []string{"ar", "x", "rw", "R"} {
				input := minimalSpec()
				acl := validAcl("policy")
				acl.AccessPolicies[0].Permissions = permissions
				input.Spec.Acls = []*AzureStorageTableAcl{acl}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "permissions %q must be rejected", permissions)
			}
		})
	})
})
