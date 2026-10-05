package azurestoragedatalakegen2filesystemv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureStorageDataLakeGen2FilesystemSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureStorageDataLakeGen2FilesystemSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const accountId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/plantonlake"

const objectId = "11111111-2222-3333-4444-555555555555"

// minimal valid spec: a bare filesystem on an HNS account.
func minimalSpec() *AzureStorageDataLakeGen2Filesystem {
	return &AzureStorageDataLakeGen2Filesystem{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureStorageDataLakeGen2Filesystem",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-filesystem",
		},
		Spec: &AzureStorageDataLakeGen2FilesystemSpec{
			StorageAccountId: literal(accountId),
			FilesystemName:   "raw-zone",
		},
	}
}

var _ = ginkgo.Describe("AzureStorageDataLakeGen2FilesystemSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal filesystem", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept the special $root name", func() {
			input := minimalSpec()
			input.Spec.FilesystemName = "$root"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a valueFrom reference for the account", func() {
			input := minimalSpec()
			input.Spec.StorageAccountId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{
						Kind:      catalogkind.CatalogKind_AzureStorageAccount,
						Name:      "lake-storage",
						FieldPath: "status.outputs.storage_account_id",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a default encryption scope reference", func() {
			input := minimalSpec()
			input.Spec.DefaultEncryptionScope = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{
						Kind:      catalogkind.CatalogKind_AzureStorageEncryptionScope,
						Name:      "regulated-scope",
						FieldPath: "status.outputs.encryption_scope_name",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept owner and group as object IDs or $superuser", func() {
			input := minimalSpec()
			input.Spec.Owner = literal(objectId)
			input.Spec.Group = literal("$superuser")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept owner by reference to a user-assigned identity", func() {
			input := minimalSpec()
			input.Spec.Owner = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{
						Kind:      catalogkind.CatalogKind_AzureUserAssignedIdentity,
						Name:      "lake-engineering",
						FieldPath: "status.outputs.principal_id",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a full root ACL", func() {
			input := minimalSpec()
			input.Spec.Aces = []*AzureStorageDataLakeGen2FilesystemAce{
				{Type: AzureStorageDataLakeGen2FilesystemAceType_USER, Permissions: "rwx"},
				{Type: AzureStorageDataLakeGen2FilesystemAceType_USER, ObjectId: literal(objectId), Permissions: "--x"},
				{Type: AzureStorageDataLakeGen2FilesystemAceType_GROUP, Permissions: "r-x"},
				{Type: AzureStorageDataLakeGen2FilesystemAceType_MASK, Permissions: "r-x"},
				{Type: AzureStorageDataLakeGen2FilesystemAceType_OTHER, Permissions: "---"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an ACL entry naming its principal by reference", func() {
			input := minimalSpec()
			input.Spec.Aces = []*AzureStorageDataLakeGen2FilesystemAce{
				{
					Type: AzureStorageDataLakeGen2FilesystemAceType_USER,
					ObjectId: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
							ValueFrom: &foreignkeyv1.ValueFromRef{
								Kind:      catalogkind.CatalogKind_AzureUserAssignedIdentity,
								Name:      "lake-engineering",
								FieldPath: "status.outputs.principal_id",
							},
						},
					},
					Permissions: "rwx",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a DEFAULT-scope inheritance entry", func() {
			input := minimalSpec()
			input.Spec.Aces = []*AzureStorageDataLakeGen2FilesystemAce{
				{
					Scope:       AzureStorageDataLakeGen2FilesystemAceScope_DEFAULT,
					Type:        AzureStorageDataLakeGen2FilesystemAceType_GROUP,
					ObjectId:    literal(objectId),
					Permissions: "r-x",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept properties metadata", func() {
			input := minimalSpec()
			input.Spec.Properties = map[string]string{"environment": "cHJvZHVjdGlvbg=="}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing storage account reference", func() {
			input := minimalSpec()
			input.Spec.StorageAccountId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject names that break the filesystem naming rules", func() {
			for _, name := range []string{"", "ab", "-leading-hyphen", "UPPER", "has_underscore", "$web"} {
				input := minimalSpec()
				input.Spec.FilesystemName = name
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "name %q must be rejected", name)
			}
		})

		ginkgo.It("should reject a name longer than 63 characters", func() {
			input := minimalSpec()
			input.Spec.FilesystemName = "a123456789b123456789c123456789d123456789e123456789f123456789g123"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an ACL entry without a type", func() {
			input := minimalSpec()
			input.Spec.Aces = []*AzureStorageDataLakeGen2FilesystemAce{
				{Permissions: "rwx"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject ACL permissions outside the POSIX short form", func() {
			for _, permissions := range []string{"", "rw", "rwxr", "wrx", "RWX", "r+x"} {
				input := minimalSpec()
				input.Spec.Aces = []*AzureStorageDataLakeGen2FilesystemAce{
					{Type: AzureStorageDataLakeGen2FilesystemAceType_USER, Permissions: permissions},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "permissions %q must be rejected", permissions)
			}
		})

		ginkgo.It("should reject an object ID on a MASK entry", func() {
			input := minimalSpec()
			input.Spec.Aces = []*AzureStorageDataLakeGen2FilesystemAce{
				{Type: AzureStorageDataLakeGen2FilesystemAceType_MASK, ObjectId: literal(objectId), Permissions: "r-x"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an object ID on an OTHER entry", func() {
			input := minimalSpec()
			input.Spec.Aces = []*AzureStorageDataLakeGen2FilesystemAce{
				{Type: AzureStorageDataLakeGen2FilesystemAceType_OTHER, ObjectId: literal(objectId), Permissions: "---"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
