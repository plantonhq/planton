package azurestorageaccountv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureStorageAccountSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureStorageAccountSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func int32Ptr(v int32) *int32 { return &v }

const (
	keyVaultKeyId = "https://vault.vault.azure.net/keys/cmk"
	identityArmId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai"
	subnetArmId   = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/app"
)

// minimal valid spec: a general-purpose v2 account on all defaults.
func minimalSpec() *AzureStorageAccount {
	return &AzureStorageAccount{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureStorageAccount",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-storage-account",
		},
		Spec: &AzureStorageAccountSpec{
			Region:        "eastus",
			ResourceGroup: literal("app-rg"),
			AccountName:   "plantonteststorage",
		},
	}
}

var _ = ginkgo.Describe("AzureStorageAccountSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal general-purpose v2 account", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept every replication type", func() {
			for _, repl := range []AzureStorageAccountReplicationType{
				AzureStorageAccountReplicationType_LRS,
				AzureStorageAccountReplicationType_ZRS,
				AzureStorageAccountReplicationType_GRS,
				AzureStorageAccountReplicationType_GZRS,
				AzureStorageAccountReplicationType_RA_GRS,
				AzureStorageAccountReplicationType_RA_GZRS,
			} {
				input := minimalSpec()
				input.Spec.ReplicationType = repl
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "replication %v must be accepted", repl)
			}
		})

		ginkgo.It("should accept a premium block-blob account", func() {
			input := minimalSpec()
			input.Spec.AccountKind = AzureStorageAccountKind_BLOCK_BLOB_STORAGE
			input.Spec.AccountTier = AzureStorageAccountTier_PREMIUM
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a premium file-storage account with provisioned v2 billing", func() {
			input := minimalSpec()
			input.Spec.AccountKind = AzureStorageAccountKind_FILE_STORAGE
			input.Spec.AccountTier = AzureStorageAccountTier_PREMIUM
			input.Spec.ProvisionedBillingModelVersion = "V2"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a legacy blob-storage account with an access tier", func() {
			input := minimalSpec()
			input.Spec.AccountKind = AzureStorageAccountKind_BLOB_STORAGE
			input.Spec.AccessTier = AzureStorageAccountAccessTier_COOL
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept the COLD access tier on storage v2", func() {
			input := minimalSpec()
			input.Spec.AccessTier = AzureStorageAccountAccessTier_COLD
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept the full security posture", func() {
			enabled := false
			input := minimalSpec()
			input.Spec.SharedAccessKeyEnabled = &enabled
			input.Spec.AllowNestedItemsToBePublic = &enabled
			input.Spec.PublicNetworkAccessEnabled = &enabled
			input.Spec.DefaultToOauthAuthentication = true
			input.Spec.MinTlsVersion = AzureStorageAccountMinTlsVersion_TLS1_2
			input.Spec.AllowedCopyScope = AzureStorageAccountAllowedCopyScope_AAD
			input.Spec.SasPolicy = &AzureStorageAccountSasPolicy{
				ExpirationPeriod: "90.00:00:00",
				ExpirationAction: AzureStorageAccountSasExpirationAction_BLOCK,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should reject the retired TLS 1.0/1.1 floors (reserved enum slots)", func() {
			for _, retired := range []AzureStorageAccountMinTlsVersion{1, 2} {
				input := minimalSpec()
				input.Spec.MinTlsVersion = retired
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "retired TLS floor %d must be rejected", retired)
			}
		})

		ginkgo.It("should accept a data-lake account with SFTP and NFSv3", func() {
			input := minimalSpec()
			input.Spec.IsHnsEnabled = true
			input.Spec.SftpEnabled = true
			input.Spec.Nfsv3Enabled = true
			input.Spec.ReplicationType = AzureStorageAccountReplicationType_LRS
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept NFSv3 on a premium block-blob account with RA_GRS", func() {
			input := minimalSpec()
			input.Spec.AccountKind = AzureStorageAccountKind_BLOCK_BLOB_STORAGE
			input.Spec.AccountTier = AzureStorageAccountTier_PREMIUM
			input.Spec.IsHnsEnabled = true
			input.Spec.Nfsv3Enabled = true
			input.Spec.ReplicationType = AzureStorageAccountReplicationType_RA_GRS
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept large file shares on a file-storage account", func() {
			input := minimalSpec()
			input.Spec.AccountKind = AzureStorageAccountKind_FILE_STORAGE
			input.Spec.AccountTier = AzureStorageAccountTier_PREMIUM
			input.Spec.LargeFileShareEnabled = true
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept partitioned DNS without a restore policy", func() {
			input := minimalSpec()
			input.Spec.DnsEndpointType = AzureStorageAccountDnsEndpointType_AZURE_DNS_ZONE
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept infrastructure encryption on storage v2 and premium file storage", func() {
			input := minimalSpec()
			input.Spec.InfrastructureEncryptionEnabled = true
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())

			input = minimalSpec()
			input.Spec.AccountKind = AzureStorageAccountKind_FILE_STORAGE
			input.Spec.AccountTier = AzureStorageAccountTier_PREMIUM
			input.Spec.InfrastructureEncryptionEnabled = true
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept account-scoped queue and table encryption on storage v2", func() {
			input := minimalSpec()
			input.Spec.QueueEncryptionKeyType = AzureStorageAccountEncryptionKeyType_ACCOUNT
			input.Spec.TableEncryptionKeyType = AzureStorageAccountEncryptionKeyType_ACCOUNT
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a system-assigned identity", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureStorageAccountIdentity{
				Type: AzureStorageAccountIdentityType_SYSTEM_ASSIGNED,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a customer-managed key with a user-assigned identity", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureStorageAccountIdentity{
				Type:        AzureStorageAccountIdentityType_USER_ASSIGNED,
				IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(identityArmId)},
			}
			input.Spec.CustomerManagedKey = &AzureStorageAccountCustomerManagedKey{
				KeyVaultKeyId:          literal(keyVaultKeyId),
				UserAssignedIdentityId: literal(identityArmId),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept locked-down network rules with all exception classes", func() {
			tenant := "00000000-0000-0000-0000-000000000000"
			input := minimalSpec()
			input.Spec.NetworkRules = &AzureStorageAccountNetworkRules{
				DefaultAction: AzureStorageAccountNetworkDefaultAction_DENY,
				Bypass: []AzureStorageAccountNetworkBypass{
					AzureStorageAccountNetworkBypass_AZURE_SERVICES,
					AzureStorageAccountNetworkBypass_METRICS,
				},
				IpRules:                 []string{"203.0.113.0/24", "198.51.100.42"},
				VirtualNetworkSubnetIds: []*foreignkeyv1.StringValueOrRef{literal(subnetArmId)},
				PrivateLinkAccess: []*AzureStorageAccountPrivateLinkAccess{
					{
						EndpointResourceId: "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Synapse/workspaces/ws",
						EndpointTenantId:   &tenant,
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept full blob properties with point-in-time restore", func() {
			input := minimalSpec()
			input.Spec.BlobProperties = &AzureStorageAccountBlobProperties{
				VersioningEnabled:              true,
				ChangeFeedEnabled:              true,
				ChangeFeedRetentionInDays:      int32Ptr(30),
				LastAccessTimeEnabled:          true,
				DefaultServiceVersion:          "2020-06-12",
				DeleteRetentionPolicy:          &AzureStorageAccountDeleteRetentionPolicy{Days: int32Ptr(30), PermanentDeleteEnabled: true},
				ContainerDeleteRetentionPolicy: &AzureStorageAccountContainerDeleteRetentionPolicy{Days: int32Ptr(14)},
				RestorePolicy:                  &AzureStorageAccountRestorePolicy{Days: 7},
				CorsRules: []*AzureStorageAccountCorsRule{
					{
						AllowedOrigins:  []string{"https://app.example.com"},
						AllowedMethods:  []string{"GET", "PUT", "PATCH"},
						AllowedHeaders:  []string{"*"},
						ExposedHeaders:  []string{"x-ms-meta-*"},
						MaxAgeInSeconds: 3600,
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept share properties with SMB multichannel on premium file storage", func() {
			input := minimalSpec()
			input.Spec.AccountKind = AzureStorageAccountKind_FILE_STORAGE
			input.Spec.AccountTier = AzureStorageAccountTier_PREMIUM
			input.Spec.ShareProperties = &AzureStorageAccountShareProperties{
				RetentionPolicy: &AzureStorageAccountShareRetentionPolicy{Days: int32Ptr(14)},
				Smb: &AzureStorageAccountSmbSettings{
					Versions:                     []string{"SMB3.1.1"},
					AuthenticationTypes:          []string{"Kerberos"},
					KerberosTicketEncryptionType: []string{"AES-256"},
					ChannelEncryptionType:        []string{"AES-256-GCM"},
					MultichannelEnabled:          true,
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept share properties on a standard storage v2 account", func() {
			input := minimalSpec()
			input.Spec.ShareProperties = &AzureStorageAccountShareProperties{
				RetentionPolicy: &AzureStorageAccountShareRetentionPolicy{Days: int32Ptr(7)},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a static website on storage v2 and block blob storage", func() {
			input := minimalSpec()
			input.Spec.StaticWebsite = &AzureStorageAccountStaticWebsite{
				IndexDocument:     "index.html",
				Error_404Document: "404.html",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())

			input = minimalSpec()
			input.Spec.AccountKind = AzureStorageAccountKind_BLOCK_BLOB_STORAGE
			input.Spec.AccountTier = AzureStorageAccountTier_PREMIUM
			input.Spec.StaticWebsite = &AzureStorageAccountStaticWebsite{IndexDocument: "index.html"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept routing and a custom domain", func() {
			input := minimalSpec()
			input.Spec.Routing = &AzureStorageAccountRouting{
				Choice:                   AzureStorageAccountRoutingChoice_INTERNET_ROUTING,
				PublishInternetEndpoints: true,
			}
			input.Spec.CustomDomain = &AzureStorageAccountCustomDomain{
				Name:         "assets.example.com",
				UseSubdomain: true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept Entra Kerberos files authentication without domain coordinates", func() {
			input := minimalSpec()
			input.Spec.AzureFilesAuthentication = &AzureStorageAccountAzureFilesAuthentication{
				DirectoryType:               AzureStorageAccountDirectoryServiceType_AADKERB,
				DefaultShareLevelPermission: AzureStorageAccountDefaultSharePermission_SHARE_PERMISSION_READER,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept on-premises AD files authentication with domain coordinates", func() {
			input := minimalSpec()
			input.Spec.AzureFilesAuthentication = &AzureStorageAccountAzureFilesAuthentication{
				DirectoryType: AzureStorageAccountDirectoryServiceType_AD,
				ActiveDirectory: &AzureStorageAccountActiveDirectory{
					DomainName:        "corp.example.com",
					DomainGuid:        "11111111-2222-3333-4444-555555555555",
					DomainSid:         "S-1-5-21-1111111111-2222222222-3333333333",
					StorageSid:        "S-1-5-21-1111111111-2222222222-3333333333-1001",
					ForestName:        "corp.example.com",
					NetbiosDomainName: "CORP",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an account-level immutability policy with versioning", func() {
			input := minimalSpec()
			input.Spec.BlobProperties = &AzureStorageAccountBlobProperties{VersioningEnabled: true}
			input.Spec.ImmutabilityPolicy = &AzureStorageAccountImmutabilityPolicy{
				State:                      AzureStorageAccountImmutabilityState_UNLOCKED,
				PeriodSinceCreationInDays:  30,
				AllowProtectedAppendWrites: true,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a lifecycle rule tiering and deleting base blobs", func() {
			input := minimalSpec()
			input.Spec.LifecycleRules = []*AzureStorageAccountLifecycleRule{
				{
					Name: "age-out-logs",
					Filters: &AzureStorageAccountLifecycleFilters{
						BlobTypes:   []AzureStorageAccountLifecycleBlobType{AzureStorageAccountLifecycleBlobType_BLOCK_BLOB},
						PrefixMatch: []string{"logs/"},
					},
					Actions: &AzureStorageAccountLifecycleActions{
						BaseBlob: &AzureStorageAccountLifecycleBaseBlobActions{
							TierToCoolAfterDaysSinceModificationGreaterThan:    int32Ptr(30),
							TierToArchiveAfterDaysSinceModificationGreaterThan: int32Ptr(90),
							DeleteAfterDaysSinceModificationGreaterThan:        int32Ptr(365),
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a lifecycle rule with snapshot and version actions", func() {
			input := minimalSpec()
			input.Spec.LifecycleRules = []*AzureStorageAccountLifecycleRule{
				{
					Name: "trim-history",
					Filters: &AzureStorageAccountLifecycleFilters{
						BlobTypes: []AzureStorageAccountLifecycleBlobType{AzureStorageAccountLifecycleBlobType_BLOCK_BLOB},
					},
					Actions: &AzureStorageAccountLifecycleActions{
						Snapshot: &AzureStorageAccountLifecycleSnapshotActions{
							DeleteAfterDaysSinceCreationGreaterThan: int32Ptr(90),
						},
						Version: &AzureStorageAccountLifecycleVersionActions{
							ChangeTierToCoolAfterDaysSinceCreation: int32Ptr(30),
							DeleteAfterDaysSinceCreation:           int32Ptr(180),
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a lifecycle rule filtering by index tags with base-blob actions only", func() {
			input := minimalSpec()
			input.Spec.LifecycleRules = []*AzureStorageAccountLifecycleRule{
				{
					Name: "expire-tagged",
					Filters: &AzureStorageAccountLifecycleFilters{
						BlobTypes: []AzureStorageAccountLifecycleBlobType{AzureStorageAccountLifecycleBlobType_BLOCK_BLOB},
						MatchBlobIndexTags: []*AzureStorageAccountLifecycleTagFilter{
							{Name: "retention", Value: "short"},
						},
					},
					Actions: &AzureStorageAccountLifecycleActions{
						BaseBlob: &AzureStorageAccountLifecycleBaseBlobActions{
							DeleteAfterDaysSinceCreationGreaterThan: int32Ptr(30),
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept user tags", func() {
			input := minimalSpec()
			input.Spec.Tags = map[string]string{"team": "platform", "cost-center": "eng"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing region", func() {
			input := minimalSpec()
			input.Spec.Region = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing resource group", func() {
			input := minimalSpec()
			input.Spec.ResourceGroup = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject account names that break the 3-24 lowercase-alphanumeric rule", func() {
			for _, name := range []string{"", "ab", "UPPERCASE", "has-hyphens", "has_underscore", "waytoolongaccountnamefortherule"} {
				input := minimalSpec()
				input.Spec.AccountName = name
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "account name %q must be rejected", name)
			}
		})

		ginkgo.It("should reject an unknown billing model version", func() {
			input := minimalSpec()
			input.Spec.ProvisionedBillingModelVersion = "V3"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an access tier on a block-blob account", func() {
			input := minimalSpec()
			input.Spec.AccountKind = AzureStorageAccountKind_BLOCK_BLOB_STORAGE
			input.Spec.AccountTier = AzureStorageAccountTier_PREMIUM
			input.Spec.AccessTier = AzureStorageAccountAccessTier_HOT
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject hierarchical namespace on a file-storage account", func() {
			input := minimalSpec()
			input.Spec.AccountKind = AzureStorageAccountKind_FILE_STORAGE
			input.Spec.AccountTier = AzureStorageAccountTier_PREMIUM
			input.Spec.IsHnsEnabled = true
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject SFTP without hierarchical namespace", func() {
			input := minimalSpec()
			input.Spec.SftpEnabled = true
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject NFSv3 without hierarchical namespace", func() {
			input := minimalSpec()
			input.Spec.Nfsv3Enabled = true
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject NFSv3 on a premium storage v2 pairing", func() {
			input := minimalSpec()
			input.Spec.AccountTier = AzureStorageAccountTier_PREMIUM
			input.Spec.IsHnsEnabled = true
			input.Spec.Nfsv3Enabled = true
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject NFSv3 with ZRS replication", func() {
			input := minimalSpec()
			input.Spec.IsHnsEnabled = true
			input.Spec.Nfsv3Enabled = true
			input.Spec.ReplicationType = AzureStorageAccountReplicationType_ZRS
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject ZRS on a blob-storage account", func() {
			input := minimalSpec()
			input.Spec.AccountKind = AzureStorageAccountKind_BLOB_STORAGE
			input.Spec.ReplicationType = AzureStorageAccountReplicationType_ZRS
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject blob versioning on a hierarchical-namespace account", func() {
			input := minimalSpec()
			input.Spec.IsHnsEnabled = true
			input.Spec.BlobProperties = &AzureStorageAccountBlobProperties{VersioningEnabled: true}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an immutability policy without blob versioning", func() {
			input := minimalSpec()
			input.Spec.ImmutabilityPolicy = &AzureStorageAccountImmutabilityPolicy{
				State:                     AzureStorageAccountImmutabilityState_UNLOCKED,
				PeriodSinceCreationInDays: 30,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject infrastructure encryption on a standard file-storage pairing", func() {
			input := minimalSpec()
			input.Spec.AccountKind = AzureStorageAccountKind_BLOB_STORAGE
			input.Spec.InfrastructureEncryptionEnabled = true
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a restore policy with partitioned DNS", func() {
			input := minimalSpec()
			input.Spec.DnsEndpointType = AzureStorageAccountDnsEndpointType_AZURE_DNS_ZONE
			input.Spec.BlobProperties = &AzureStorageAccountBlobProperties{
				VersioningEnabled:     true,
				ChangeFeedEnabled:     true,
				DeleteRetentionPolicy: &AzureStorageAccountDeleteRetentionPolicy{Days: int32Ptr(14)},
				RestorePolicy:         &AzureStorageAccountRestorePolicy{Days: 7},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject SMB multichannel on a standard-tier account", func() {
			input := minimalSpec()
			input.Spec.ShareProperties = &AzureStorageAccountShareProperties{
				Smb: &AzureStorageAccountSmbSettings{MultichannelEnabled: true},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject blob properties on a file-storage account", func() {
			input := minimalSpec()
			input.Spec.AccountKind = AzureStorageAccountKind_FILE_STORAGE
			input.Spec.AccountTier = AzureStorageAccountTier_PREMIUM
			input.Spec.BlobProperties = &AzureStorageAccountBlobProperties{VersioningEnabled: true}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject share properties on a premium storage v2 account", func() {
			input := minimalSpec()
			input.Spec.AccountTier = AzureStorageAccountTier_PREMIUM
			input.Spec.ShareProperties = &AzureStorageAccountShareProperties{
				RetentionPolicy: &AzureStorageAccountShareRetentionPolicy{Days: int32Ptr(7)},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a static website on a blob-storage account", func() {
			input := minimalSpec()
			input.Spec.AccountKind = AzureStorageAccountKind_BLOB_STORAGE
			input.Spec.StaticWebsite = &AzureStorageAccountStaticWebsite{IndexDocument: "index.html"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a static website with no documents", func() {
			input := minimalSpec()
			input.Spec.StaticWebsite = &AzureStorageAccountStaticWebsite{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject large file shares on a block-blob account", func() {
			input := minimalSpec()
			input.Spec.AccountKind = AzureStorageAccountKind_BLOCK_BLOB_STORAGE
			input.Spec.AccountTier = AzureStorageAccountTier_PREMIUM
			input.Spec.LargeFileShareEnabled = true
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject account-scoped queue/table encryption on the legacy storage kind", func() {
			input := minimalSpec()
			input.Spec.AccountKind = AzureStorageAccountKind_STORAGE
			input.Spec.QueueEncryptionKeyType = AzureStorageAccountEncryptionKeyType_ACCOUNT
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

			input = minimalSpec()
			input.Spec.AccountKind = AzureStorageAccountKind_STORAGE
			input.Spec.TableEncryptionKeyType = AzureStorageAccountEncryptionKeyType_ACCOUNT
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a customer-managed key without an identity", func() {
			input := minimalSpec()
			input.Spec.CustomerManagedKey = &AzureStorageAccountCustomerManagedKey{
				KeyVaultKeyId:          literal(keyVaultKeyId),
				UserAssignedIdentityId: literal(identityArmId),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a customer-managed key with a system-assigned-only identity", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureStorageAccountIdentity{
				Type: AzureStorageAccountIdentityType_SYSTEM_ASSIGNED,
			}
			input.Spec.CustomerManagedKey = &AzureStorageAccountCustomerManagedKey{
				KeyVaultKeyId:          literal(keyVaultKeyId),
				UserAssignedIdentityId: literal(identityArmId),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a customer-managed key missing its required references", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureStorageAccountIdentity{
				Type:        AzureStorageAccountIdentityType_USER_ASSIGNED,
				IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(identityArmId)},
			}
			input.Spec.CustomerManagedKey = &AzureStorageAccountCustomerManagedKey{
				KeyVaultKeyId: literal(keyVaultKeyId),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a user-assigned identity without identity ids", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureStorageAccountIdentity{
				Type: AzureStorageAccountIdentityType_USER_ASSIGNED,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a system-assigned identity carrying identity ids", func() {
			input := minimalSpec()
			input.Spec.Identity = &AzureStorageAccountIdentity{
				Type:        AzureStorageAccountIdentityType_SYSTEM_ASSIGNED,
				IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(identityArmId)},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject network rules without a default action", func() {
			input := minimalSpec()
			input.Spec.NetworkRules = &AzureStorageAccountNetworkRules{
				IpRules: []string{"203.0.113.0/24"},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a private-link tenant that is not a UUID", func() {
			tenant := "not-a-uuid"
			input := minimalSpec()
			input.Spec.NetworkRules = &AzureStorageAccountNetworkRules{
				DefaultAction: AzureStorageAccountNetworkDefaultAction_DENY,
				PrivateLinkAccess: []*AzureStorageAccountPrivateLinkAccess{
					{EndpointResourceId: "/subscriptions/s/rg", EndpointTenantId: &tenant},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a restore policy without versioning, change feed, or soft delete", func() {
			// missing versioning
			input := minimalSpec()
			input.Spec.BlobProperties = &AzureStorageAccountBlobProperties{
				ChangeFeedEnabled:     true,
				DeleteRetentionPolicy: &AzureStorageAccountDeleteRetentionPolicy{Days: int32Ptr(14)},
				RestorePolicy:         &AzureStorageAccountRestorePolicy{Days: 7},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

			// missing change feed
			input = minimalSpec()
			input.Spec.BlobProperties = &AzureStorageAccountBlobProperties{
				VersioningEnabled:     true,
				DeleteRetentionPolicy: &AzureStorageAccountDeleteRetentionPolicy{Days: int32Ptr(14)},
				RestorePolicy:         &AzureStorageAccountRestorePolicy{Days: 7},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

			// missing blob soft delete
			input = minimalSpec()
			input.Spec.BlobProperties = &AzureStorageAccountBlobProperties{
				VersioningEnabled: true,
				ChangeFeedEnabled: true,
				RestorePolicy:     &AzureStorageAccountRestorePolicy{Days: 7},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject out-of-range retention windows", func() {
			input := minimalSpec()
			input.Spec.BlobProperties = &AzureStorageAccountBlobProperties{
				DeleteRetentionPolicy: &AzureStorageAccountDeleteRetentionPolicy{Days: int32Ptr(366)},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

			input = minimalSpec()
			input.Spec.BlobProperties = &AzureStorageAccountBlobProperties{
				ChangeFeedRetentionInDays: int32Ptr(0),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a CORS rule with an unknown method", func() {
			input := minimalSpec()
			input.Spec.BlobProperties = &AzureStorageAccountBlobProperties{
				CorsRules: []*AzureStorageAccountCorsRule{
					{
						AllowedOrigins:  []string{"*"},
						AllowedMethods:  []string{"TRACE"},
						AllowedHeaders:  []string{"*"},
						ExposedHeaders:  []string{"*"},
						MaxAgeInSeconds: 60,
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown SMB version", func() {
			input := minimalSpec()
			input.Spec.ShareProperties = &AzureStorageAccountShareProperties{
				Smb: &AzureStorageAccountSmbSettings{Versions: []string{"SMB1.0"}},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed SAS expiration period", func() {
			for _, period := range []string{"90 days", "90.0:0:0", ""} {
				input := minimalSpec()
				input.Spec.SasPolicy = &AzureStorageAccountSasPolicy{ExpirationPeriod: period}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "period %q must be rejected", period)
			}
		})

		ginkgo.It("should reject an immutability policy without a state or with a zero period", func() {
			input := minimalSpec()
			input.Spec.BlobProperties = &AzureStorageAccountBlobProperties{VersioningEnabled: true}
			input.Spec.ImmutabilityPolicy = &AzureStorageAccountImmutabilityPolicy{
				PeriodSinceCreationInDays: 30,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

			input = minimalSpec()
			input.Spec.BlobProperties = &AzureStorageAccountBlobProperties{VersioningEnabled: true}
			input.Spec.ImmutabilityPolicy = &AzureStorageAccountImmutabilityPolicy{
				State: AzureStorageAccountImmutabilityState_UNLOCKED,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a lifecycle rule with no actions", func() {
			input := minimalSpec()
			input.Spec.LifecycleRules = []*AzureStorageAccountLifecycleRule{
				{
					Name: "no-op",
					Filters: &AzureStorageAccountLifecycleFilters{
						BlobTypes: []AzureStorageAccountLifecycleBlobType{AzureStorageAccountLifecycleBlobType_BLOCK_BLOB},
					},
					Actions: &AzureStorageAccountLifecycleActions{},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a lifecycle rule with no blob types", func() {
			input := minimalSpec()
			input.Spec.LifecycleRules = []*AzureStorageAccountLifecycleRule{
				{
					Name:    "untyped",
					Filters: &AzureStorageAccountLifecycleFilters{},
					Actions: &AzureStorageAccountLifecycleActions{
						BaseBlob: &AzureStorageAccountLifecycleBaseBlobActions{
							DeleteAfterDaysSinceCreationGreaterThan: int32Ptr(30),
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject two aging bases on the same lifecycle destination", func() {
			input := minimalSpec()
			input.Spec.LifecycleRules = []*AzureStorageAccountLifecycleRule{
				{
					Name: "conflicting",
					Filters: &AzureStorageAccountLifecycleFilters{
						BlobTypes: []AzureStorageAccountLifecycleBlobType{AzureStorageAccountLifecycleBlobType_BLOCK_BLOB},
					},
					Actions: &AzureStorageAccountLifecycleActions{
						BaseBlob: &AzureStorageAccountLifecycleBaseBlobActions{
							TierToCoolAfterDaysSinceModificationGreaterThan: int32Ptr(30),
							TierToCoolAfterDaysSinceCreationGreaterThan:     int32Ptr(30),
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())

			input = minimalSpec()
			input.Spec.LifecycleRules = []*AzureStorageAccountLifecycleRule{
				{
					Name: "conflicting-delete",
					Filters: &AzureStorageAccountLifecycleFilters{
						BlobTypes: []AzureStorageAccountLifecycleBlobType{AzureStorageAccountLifecycleBlobType_BLOCK_BLOB},
					},
					Actions: &AzureStorageAccountLifecycleActions{
						BaseBlob: &AzureStorageAccountLifecycleBaseBlobActions{
							DeleteAfterDaysSinceModificationGreaterThan:   int32Ptr(365),
							DeleteAfterDaysSinceLastAccessTimeGreaterThan: int32Ptr(365),
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject auto-tier-to-hot without the last-access basis", func() {
			input := minimalSpec()
			input.Spec.LifecycleRules = []*AzureStorageAccountLifecycleRule{
				{
					Name: "auto-hot",
					Filters: &AzureStorageAccountLifecycleFilters{
						BlobTypes: []AzureStorageAccountLifecycleBlobType{AzureStorageAccountLifecycleBlobType_BLOCK_BLOB},
					},
					Actions: &AzureStorageAccountLifecycleActions{
						BaseBlob: &AzureStorageAccountLifecycleBaseBlobActions{
							AutoTierToHotFromCoolEnabled:                    true,
							TierToCoolAfterDaysSinceModificationGreaterThan: int32Ptr(30),
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an index-tag filter combined with snapshot or version actions", func() {
			input := minimalSpec()
			input.Spec.LifecycleRules = []*AzureStorageAccountLifecycleRule{
				{
					Name: "tags-with-snapshots",
					Filters: &AzureStorageAccountLifecycleFilters{
						BlobTypes: []AzureStorageAccountLifecycleBlobType{AzureStorageAccountLifecycleBlobType_BLOCK_BLOB},
						MatchBlobIndexTags: []*AzureStorageAccountLifecycleTagFilter{
							{Name: "retention", Value: "short"},
						},
					},
					Actions: &AzureStorageAccountLifecycleActions{
						Snapshot: &AzureStorageAccountLifecycleSnapshotActions{
							DeleteAfterDaysSinceCreationGreaterThan: int32Ptr(30),
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an index-tag operation other than equality", func() {
			op := "!="
			input := minimalSpec()
			input.Spec.LifecycleRules = []*AzureStorageAccountLifecycleRule{
				{
					Name: "bad-op",
					Filters: &AzureStorageAccountLifecycleFilters{
						BlobTypes: []AzureStorageAccountLifecycleBlobType{AzureStorageAccountLifecycleBlobType_BLOCK_BLOB},
						MatchBlobIndexTags: []*AzureStorageAccountLifecycleTagFilter{
							{Name: "retention", Operation: &op, Value: "short"},
						},
					},
					Actions: &AzureStorageAccountLifecycleActions{
						BaseBlob: &AzureStorageAccountLifecycleBaseBlobActions{
							DeleteAfterDaysSinceCreationGreaterThan: int32Ptr(30),
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject AD files authentication without domain coordinates", func() {
			input := minimalSpec()
			input.Spec.AzureFilesAuthentication = &AzureStorageAccountAzureFilesAuthentication{
				DirectoryType: AzureStorageAccountDirectoryServiceType_AD,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a domain GUID that is not a UUID", func() {
			input := minimalSpec()
			input.Spec.AzureFilesAuthentication = &AzureStorageAccountAzureFilesAuthentication{
				DirectoryType: AzureStorageAccountDirectoryServiceType_AADKERB,
				ActiveDirectory: &AzureStorageAccountActiveDirectory{
					DomainName: "corp.example.com",
					DomainGuid: "not-a-guid",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
