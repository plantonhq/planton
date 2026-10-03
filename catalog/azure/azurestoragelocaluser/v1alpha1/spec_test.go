package azurestoragelocaluserv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureStorageLocalUserSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureStorageLocalUserSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const accountId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/plantonsftp"

// A well-formed OpenSSH ed25519 public key (public material -- safe in a fixture).
const ed25519PublicKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGJ3iXFCbaYbzUZzz2i2Cv6/L7Ohtq5rM9lZ74W2mBrz partner-pipeline"

// minimal valid spec: a password-authenticated user.
func minimalSpec() *AzureStorageLocalUser {
	return &AzureStorageLocalUser{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureStorageLocalUser",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-local-user",
		},
		Spec: &AzureStorageLocalUserSpec{
			StorageAccountId:   literal(accountId),
			UserName:           "partner01",
			SshPasswordEnabled: true,
		},
	}
}

var _ = ginkgo.Describe("AzureStorageLocalUserSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal password-authenticated user", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept a key-authenticated user with authorized keys", func() {
			input := minimalSpec()
			input.Spec.SshPasswordEnabled = false
			input.Spec.SshKeyEnabled = true
			input.Spec.SshAuthorizedKeys = []*AzureStorageLocalUserSshAuthorizedKey{
				{Key: ed25519PublicKey, Description: "partner pipeline key"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept both auth methods together", func() {
			input := minimalSpec()
			input.Spec.SshKeyEnabled = true
			input.Spec.SshAuthorizedKeys = []*AzureStorageLocalUserSshAuthorizedKey{
				{Key: ed25519PublicKey},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an RSA public key", func() {
			input := minimalSpec()
			input.Spec.SshPasswordEnabled = false
			input.Spec.SshKeyEnabled = true
			input.Spec.SshAuthorizedKeys = []*AzureStorageLocalUserSshAuthorizedKey{
				{Key: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a home directory and permission scopes", func() {
			input := minimalSpec()
			input.Spec.HomeDirectory = "inbound/partner01"
			input.Spec.PermissionScopes = []*AzureStorageLocalUserPermissionScope{
				{
					Service:      AzureStorageLocalUserPermissionService_BLOB,
					ResourceName: literal("inbound"),
					Read:         true,
					Write:        true,
					List:         true,
					Create:       true,
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a container reference for a scope's resource", func() {
			input := minimalSpec()
			input.Spec.PermissionScopes = []*AzureStorageLocalUserPermissionScope{
				{
					Service: AzureStorageLocalUserPermissionService_BLOB,
					ResourceName: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
							ValueFrom: &foreignkeyv1.ValueFromRef{
								Kind:      catalogkind.CatalogKind_AzureStorageContainer,
								Name:      "partner-inbound",
								FieldPath: "status.outputs.container_name",
							},
						},
					},
					Read: true,
					List: true,
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a FILE service scope", func() {
			input := minimalSpec()
			input.Spec.PermissionScopes = []*AzureStorageLocalUserPermissionScope{
				{
					Service:      AzureStorageLocalUserPermissionService_FILE,
					ResourceName: literal("team-share"),
					Read:         true,
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

		ginkgo.It("should reject names that break the user naming rules", func() {
			for _, name := range []string{"", "ab", "has-hyphen", "has_underscore", "UPPER", "user name"} {
				input := minimalSpec()
				input.Spec.UserName = name
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "name %q must be rejected", name)
			}
		})

		ginkgo.It("should reject a name longer than 64 characters", func() {
			input := minimalSpec()
			input.Spec.UserName = "a123456789b123456789c123456789d123456789e123456789f123456789g1234"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a user with no authentication method", func() {
			input := minimalSpec()
			input.Spec.SshPasswordEnabled = false
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject key auth without authorized keys", func() {
			input := minimalSpec()
			input.Spec.SshKeyEnabled = true
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject authorized keys when key auth is off", func() {
			input := minimalSpec()
			input.Spec.SshAuthorizedKeys = []*AzureStorageLocalUserSshAuthorizedKey{
				{Key: ed25519PublicKey},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a key that is not OpenSSH public-key material", func() {
			for _, key := range []string{"", "-----BEGIN OPENSSH PRIVATE KEY-----", "AAAAC3NzaC1lZDI1NTE5", "ssh-dss AAAA"} {
				input := minimalSpec()
				input.Spec.SshKeyEnabled = true
				input.Spec.SshAuthorizedKeys = []*AzureStorageLocalUserSshAuthorizedKey{
					{Key: key},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "key %q must be rejected", key)
			}
		})

		ginkgo.It("should reject a permission scope without a service", func() {
			input := minimalSpec()
			input.Spec.PermissionScopes = []*AzureStorageLocalUserPermissionScope{
				{ResourceName: literal("inbound"), Read: true},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a permission scope without a resource name", func() {
			input := minimalSpec()
			input.Spec.PermissionScopes = []*AzureStorageLocalUserPermissionScope{
				{Service: AzureStorageLocalUserPermissionService_BLOB, Read: true},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
