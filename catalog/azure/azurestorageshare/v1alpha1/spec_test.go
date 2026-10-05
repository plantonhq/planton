package azurestoragesharev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureStorageShareSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureStorageShareSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const accountId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/plantonapp"

// minimal valid spec: an SMB share with a quota.
func minimalSpec() *AzureStorageShare {
	return &AzureStorageShare{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureStorageShare",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-share",
		},
		Spec: &AzureStorageShareSpec{
			StorageAccountId: literal(accountId),
			ShareName:        "team-data",
			QuotaGb:          100,
		},
	}
}

var _ = ginkgo.Describe("AzureStorageShareSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal SMB share", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept both protocols", func() {
			for _, protocol := range []AzureStorageShareProtocol{
				AzureStorageShareProtocol_SMB,
				AzureStorageShareProtocol_NFS,
			} {
				input := minimalSpec()
				input.Spec.EnabledProtocol = protocol
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "protocol %v must be accepted", protocol)
			}
		})

		ginkgo.It("should accept every access tier", func() {
			for _, tier := range []AzureStorageShareAccessTier{
				AzureStorageShareAccessTier_TRANSACTION_OPTIMIZED,
				AzureStorageShareAccessTier_HOT,
				AzureStorageShareAccessTier_COOL,
				AzureStorageShareAccessTier_PREMIUM,
			} {
				input := minimalSpec()
				input.Spec.AccessTier = tier
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "tier %v must be accepted", tier)
			}
		})

		ginkgo.It("should accept the quota bounds", func() {
			for _, quota := range []int32{1, 5120, 102400} {
				input := minimalSpec()
				input.Spec.QuotaGb = quota
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "quota %d must be accepted", quota)
			}
		})

		ginkgo.It("should accept an ACL with a full access policy", func() {
			input := minimalSpec()
			input.Spec.Acls = []*AzureStorageShareAcl{{
				Id: "readers",
				AccessPolicies: []*AzureStorageShareAclAccessPolicy{{
					Permissions: "rl",
					Start:       "2026-07-01T00:00:00Z",
					Expiry:      "2027-07-01T00:00:00Z",
				}},
			}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an ACL policy without a window (SAS-side window)", func() {
			input := minimalSpec()
			input.Spec.Acls = []*AzureStorageShareAcl{{
				Id: "full-access",
				AccessPolicies: []*AzureStorageShareAclAccessPolicy{{
					Permissions: "rwdl",
				}},
			}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept share metadata", func() {
			input := minimalSpec()
			input.Spec.Metadata = map[string]string{"purpose": "team-files", "owner": "platform"}
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

		ginkgo.It("should reject names that break the share naming rules", func() {
			for _, name := range []string{"", "ab", "UPPER", "-leading", "trailing-", "double--hyphen", "has_underscore"} {
				input := minimalSpec()
				input.Spec.ShareName = name
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "name %q must be rejected", name)
			}
		})

		ginkgo.It("should reject a name longer than 63 characters", func() {
			input := minimalSpec()
			input.Spec.ShareName = "a123456789b123456789c123456789d123456789e123456789f123456789g123"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing quota", func() {
			input := minimalSpec()
			input.Spec.QuotaGb = 0
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject quotas outside 1-102400", func() {
			for _, quota := range []int32{-1, 102401} {
				input := minimalSpec()
				input.Spec.QuotaGb = quota
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "quota %d must be rejected", quota)
			}
		})

		ginkgo.It("should reject more than five ACLs", func() {
			input := minimalSpec()
			for i := 0; i < 6; i++ {
				input.Spec.Acls = append(input.Spec.Acls, &AzureStorageShareAcl{
					Id: "policy-" + string(rune('a'+i)),
					AccessPolicies: []*AzureStorageShareAclAccessPolicy{{
						Permissions: "r",
					}},
				})
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an ACL without an id", func() {
			input := minimalSpec()
			input.Spec.Acls = []*AzureStorageShareAcl{{
				AccessPolicies: []*AzureStorageShareAclAccessPolicy{{Permissions: "r"}},
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject permission letters out of order or unknown", func() {
			for _, permissions := range []string{"lr", "x", "rwx", "R"} {
				input := minimalSpec()
				input.Spec.Acls = []*AzureStorageShareAcl{{
					Id:             "policy",
					AccessPolicies: []*AzureStorageShareAclAccessPolicy{{Permissions: permissions}},
				}}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "permissions %q must be rejected", permissions)
			}
		})

		ginkgo.It("should reject an undefined protocol enum value", func() {
			input := minimalSpec()
			input.Spec.EnabledProtocol = AzureStorageShareProtocol(99)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an undefined access tier enum value", func() {
			input := minimalSpec()
			input.Spec.AccessTier = AzureStorageShareAccessTier(99)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
