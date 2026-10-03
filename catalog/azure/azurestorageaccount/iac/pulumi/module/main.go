package module

import (
	"fmt"

	"github.com/pkg/errors"
	azurestorageaccountv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurestorageaccount/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurestorageaccountv1alpha1.AzureStorageAccountIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureStorageAccount.Spec

	accountArgs := &storage.AccountArgs{
		Name:              pulumi.String(spec.AccountName),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// The SKU trio. Unspecified enums materialize the spec's documented
	// defaults here (StorageV2 / Standard / LRS) -- IaC inputs built
	// from a manifest do NOT materialize proto defaults, and azurerm
	// REQUIRES tier and replication. Kind and tier are fixed shapes;
	// replication may move within its zonal/non-zonal family in place.
	accountKind := "StorageV2"
	if spec.AccountKind != azurestorageaccountv1alpha1.AzureStorageAccountKind_azure_storage_account_kind_unspecified {
		accountKind = accountKindStrings[spec.AccountKind]
	}
	accountArgs.AccountKind = pulumi.String(accountKind)

	accountTier := "Standard"
	if spec.AccountTier != azurestorageaccountv1alpha1.AzureStorageAccountTier_azure_storage_account_tier_unspecified {
		accountTier = accountTierStrings[spec.AccountTier]
	}
	accountArgs.AccountTier = pulumi.String(accountTier)

	replicationType := "LRS"
	if spec.ReplicationType != azurestorageaccountv1alpha1.AzureStorageAccountReplicationType_azure_storage_account_replication_type_unspecified {
		replicationType = replicationTypeStrings[spec.ReplicationType]
	}
	accountArgs.AccountReplicationType = pulumi.String(replicationType)

	// Unspecified access tier is not sent at all -- Azure computes Hot on
	// the kinds that support tiers, mirroring the Terraform module's null.
	if spec.AccessTier != azurestorageaccountv1alpha1.AzureStorageAccountAccessTier_azure_storage_account_access_tier_unspecified {
		accountArgs.AccessTier = pulumi.String(accessTierStrings[spec.AccessTier])
	}

	if spec.ProvisionedBillingModelVersion != "" {
		accountArgs.ProvisionedBillingModelVersion = pulumi.String(spec.ProvisionedBillingModelVersion)
	}
	if spec.EdgeZone != "" {
		accountArgs.EdgeZone = pulumi.String(spec.EdgeZone)
	}

	// Transport and authorization posture. The true-default optional
	// bools are presence-guarded to their proto defaults; the TLS floor
	// materializes TLS1_2 (Azure's default and the compliance floor).
	if spec.HttpsTrafficOnlyEnabled != nil {
		accountArgs.HttpsTrafficOnlyEnabled = pulumi.Bool(spec.GetHttpsTrafficOnlyEnabled())
	} else {
		accountArgs.HttpsTrafficOnlyEnabled = pulumi.Bool(true)
	}

	minTlsVersion := "TLS1_2"
	if spec.MinTlsVersion != azurestorageaccountv1alpha1.AzureStorageAccountMinTlsVersion_azure_storage_account_min_tls_version_unspecified {
		minTlsVersion = minTlsVersionStrings[spec.MinTlsVersion]
	}
	accountArgs.MinTlsVersion = pulumi.String(minTlsVersion)

	if spec.SharedAccessKeyEnabled != nil {
		accountArgs.SharedAccessKeyEnabled = pulumi.Bool(spec.GetSharedAccessKeyEnabled())
	} else {
		accountArgs.SharedAccessKeyEnabled = pulumi.Bool(true)
	}

	accountArgs.DefaultToOauthAuthentication = pulumi.Bool(spec.DefaultToOauthAuthentication)

	if spec.AllowNestedItemsToBePublic != nil {
		accountArgs.AllowNestedItemsToBePublic = pulumi.Bool(spec.GetAllowNestedItemsToBePublic())
	} else {
		accountArgs.AllowNestedItemsToBePublic = pulumi.Bool(true)
	}

	if spec.PublicNetworkAccessEnabled != nil {
		accountArgs.PublicNetworkAccessEnabled = pulumi.Bool(spec.GetPublicNetworkAccessEnabled())
	} else {
		accountArgs.PublicNetworkAccessEnabled = pulumi.Bool(true)
	}

	// Unspecified copy scope is not sent -- copy stays unrestricted.
	if spec.AllowedCopyScope != azurestorageaccountv1alpha1.AzureStorageAccountAllowedCopyScope_azure_storage_account_allowed_copy_scope_unspecified {
		accountArgs.AllowedCopyScope = pulumi.String(allowedCopyScopeStrings[spec.AllowedCopyScope])
	}

	// The account-wide SAS lifetime policy: violations are logged (Log)
	// or tokens rejected outright (Block).
	if spec.SasPolicy != nil {
		expirationAction := "Log"
		if spec.SasPolicy.ExpirationAction != azurestorageaccountv1alpha1.AzureStorageAccountSasExpirationAction_azure_storage_account_sas_expiration_action_unspecified {
			expirationAction = sasExpirationActionStrings[spec.SasPolicy.ExpirationAction]
		}
		accountArgs.SasPolicy = &storage.AccountSasPolicyArgs{
			ExpirationPeriod: pulumi.String(spec.SasPolicy.ExpirationPeriod),
			ExpirationAction: pulumi.String(expirationAction),
		}
	}

	if spec.LocalUserEnabled != nil {
		accountArgs.LocalUserEnabled = pulumi.Bool(spec.GetLocalUserEnabled())
	} else {
		accountArgs.LocalUserEnabled = pulumi.Bool(true)
	}
	accountArgs.SftpEnabled = pulumi.Bool(spec.SftpEnabled)
	accountArgs.CrossTenantReplicationEnabled = pulumi.Bool(spec.CrossTenantReplicationEnabled)

	// Data-lake / protocol switches -- all create-time architectural
	// choices (HNS and NFSv3 are ForceNew; large file shares are one-way).
	accountArgs.IsHnsEnabled = pulumi.Bool(spec.IsHnsEnabled)
	accountArgs.Nfsv3Enabled = pulumi.Bool(spec.Nfsv3Enabled)
	// Sent only when true (mirroring the Terraform module's null): the
	// flag is one-way and Computed -- premium FileStorage accounts have
	// it on inherently, so an explicit false would fight Azure. False
	// means "leave it to Azure", never "disable".
	if spec.LargeFileShareEnabled {
		accountArgs.LargeFileShareEnabled = pulumi.Bool(true)
	}
	if spec.DnsEndpointType != azurestorageaccountv1alpha1.AzureStorageAccountDnsEndpointType_azure_storage_account_dns_endpoint_type_unspecified {
		accountArgs.DnsEndpointType = pulumi.String(dnsEndpointTypeStrings[spec.DnsEndpointType])
	}

	// Encryption depth: infrastructure encryption double-encrypts at
	// rest; the key-type fields move queue/table data under the account
	// key scope so the CMK below covers them too. All fixed at creation.
	accountArgs.InfrastructureEncryptionEnabled = pulumi.Bool(spec.InfrastructureEncryptionEnabled)
	if spec.QueueEncryptionKeyType != azurestorageaccountv1alpha1.AzureStorageAccountEncryptionKeyType_azure_storage_account_encryption_key_type_unspecified {
		accountArgs.QueueEncryptionKeyType = pulumi.String(encryptionKeyTypeStrings[spec.QueueEncryptionKeyType])
	}
	if spec.TableEncryptionKeyType != azurestorageaccountv1alpha1.AzureStorageAccountEncryptionKeyType_azure_storage_account_encryption_key_type_unspecified {
		accountArgs.TableEncryptionKeyType = pulumi.String(encryptionKeyTypeStrings[spec.TableEncryptionKeyType])
	}

	// The account's managed identity. A user-assigned identity must be
	// attached here for customer_managed_key to unwrap its key.
	if spec.Identity != nil {
		identityIds := make([]string, 0, len(spec.Identity.IdentityIds))
		for _, identityId := range spec.Identity.IdentityIds {
			identityIds = append(identityIds, identityId.GetValue())
		}
		identityArgs := &storage.AccountIdentityArgs{
			Type: pulumi.String(identityTypeStrings[spec.Identity.Type]),
		}
		if len(identityIds) > 0 {
			identityArgs.IdentityIds = pulumi.ToStringArray(identityIds)
		}
		accountArgs.Identity = identityArgs
	}

	// Customer-managed-key encryption. The key ID references the Key
	// Vault data plane (versionless IDs track rotations automatically);
	// the unwrapping identity must have wrap/unwrap access on the key's
	// vault BEFORE this account is created -- compose the grant in the
	// same manifest set.
	if spec.CustomerManagedKey != nil {
		accountArgs.CustomerManagedKey = &storage.AccountCustomerManagedKeyArgs{
			KeyVaultKeyId:          pulumi.String(spec.CustomerManagedKey.KeyVaultKeyId.GetValue()),
			UserAssignedIdentityId: pulumi.String(spec.CustomerManagedKey.UserAssignedIdentityId.GetValue()),
		}
	}

	// Data-plane firewall. ARM (control-plane) operations are never
	// subject to these rules; unset bypass lets Azure default to
	// AzureServices.
	if spec.NetworkRules != nil {
		networkRulesArgs := &storage.AccountNetworkRulesTypeArgs{
			DefaultAction: pulumi.String(networkDefaultActionStrings[spec.NetworkRules.DefaultAction]),
		}
		if len(spec.NetworkRules.Bypass) > 0 {
			bypasses := make([]string, 0, len(spec.NetworkRules.Bypass))
			for _, bypass := range spec.NetworkRules.Bypass {
				bypasses = append(bypasses, networkBypassStrings[bypass])
			}
			networkRulesArgs.Bypasses = pulumi.ToStringArray(bypasses)
		}
		if len(spec.NetworkRules.IpRules) > 0 {
			networkRulesArgs.IpRules = pulumi.ToStringArray(spec.NetworkRules.IpRules)
		}
		if len(spec.NetworkRules.VirtualNetworkSubnetIds) > 0 {
			subnetIds := make([]string, 0, len(spec.NetworkRules.VirtualNetworkSubnetIds))
			for _, subnetId := range spec.NetworkRules.VirtualNetworkSubnetIds {
				subnetIds = append(subnetIds, subnetId.GetValue())
			}
			networkRulesArgs.VirtualNetworkSubnetIds = pulumi.ToStringArray(subnetIds)
		}
		if len(spec.NetworkRules.PrivateLinkAccess) > 0 {
			privateLinkAccesses := storage.AccountNetworkRulesPrivateLinkAccessArray{}
			for _, privateLinkAccess := range spec.NetworkRules.PrivateLinkAccess {
				accessArgs := &storage.AccountNetworkRulesPrivateLinkAccessArgs{
					EndpointResourceId: pulumi.String(privateLinkAccess.EndpointResourceId),
				}
				if privateLinkAccess.EndpointTenantId != nil && privateLinkAccess.GetEndpointTenantId() != "" {
					accessArgs.EndpointTenantId = pulumi.String(privateLinkAccess.GetEndpointTenantId())
				}
				privateLinkAccesses = append(privateLinkAccesses, accessArgs)
			}
			networkRulesArgs.PrivateLinkAccesses = privateLinkAccesses
		}
		accountArgs.NetworkRules = networkRulesArgs
	}

	// Blob service settings. Versioning is the foundation the restore and
	// immutability features build on; the retention blocks are Azure's
	// recycle bin for blobs and containers. Retention days are
	// presence-guarded to Azure's default of 7.
	if spec.BlobProperties != nil {
		blobPropertiesArgs := &storage.AccountBlobPropertiesArgs{
			VersioningEnabled:     pulumi.Bool(spec.BlobProperties.VersioningEnabled),
			ChangeFeedEnabled:     pulumi.Bool(spec.BlobProperties.ChangeFeedEnabled),
			LastAccessTimeEnabled: pulumi.Bool(spec.BlobProperties.LastAccessTimeEnabled),
		}
		if spec.BlobProperties.ChangeFeedRetentionInDays != nil {
			blobPropertiesArgs.ChangeFeedRetentionInDays = pulumi.Int(int(spec.BlobProperties.GetChangeFeedRetentionInDays()))
		}
		if spec.BlobProperties.DefaultServiceVersion != "" {
			blobPropertiesArgs.DefaultServiceVersion = pulumi.String(spec.BlobProperties.DefaultServiceVersion)
		}
		if spec.BlobProperties.DeleteRetentionPolicy != nil {
			days := 7
			if spec.BlobProperties.DeleteRetentionPolicy.Days != nil {
				days = int(spec.BlobProperties.DeleteRetentionPolicy.GetDays())
			}
			blobPropertiesArgs.DeleteRetentionPolicy = &storage.AccountBlobPropertiesDeleteRetentionPolicyArgs{
				Days:                   pulumi.Int(days),
				PermanentDeleteEnabled: pulumi.Bool(spec.BlobProperties.DeleteRetentionPolicy.PermanentDeleteEnabled),
			}
		}
		if spec.BlobProperties.ContainerDeleteRetentionPolicy != nil {
			days := 7
			if spec.BlobProperties.ContainerDeleteRetentionPolicy.Days != nil {
				days = int(spec.BlobProperties.ContainerDeleteRetentionPolicy.GetDays())
			}
			blobPropertiesArgs.ContainerDeleteRetentionPolicy = &storage.AccountBlobPropertiesContainerDeleteRetentionPolicyArgs{
				Days: pulumi.Int(days),
			}
		}
		if spec.BlobProperties.RestorePolicy != nil {
			blobPropertiesArgs.RestorePolicy = &storage.AccountBlobPropertiesRestorePolicyArgs{
				Days: pulumi.Int(int(spec.BlobProperties.RestorePolicy.Days)),
			}
		}
		if len(spec.BlobProperties.CorsRules) > 0 {
			corsRules := storage.AccountBlobPropertiesCorsRuleArray{}
			for _, corsRule := range spec.BlobProperties.CorsRules {
				corsRules = append(corsRules, &storage.AccountBlobPropertiesCorsRuleArgs{
					AllowedOrigins:  pulumi.ToStringArray(corsRule.AllowedOrigins),
					AllowedMethods:  pulumi.ToStringArray(corsRule.AllowedMethods),
					AllowedHeaders:  pulumi.ToStringArray(corsRule.AllowedHeaders),
					ExposedHeaders:  pulumi.ToStringArray(corsRule.ExposedHeaders),
					MaxAgeInSeconds: pulumi.Int(int(corsRule.MaxAgeInSeconds)),
				})
			}
			blobPropertiesArgs.CorsRules = corsRules
		}
		accountArgs.BlobProperties = blobPropertiesArgs
	}

	// File service settings: the share recycle bin and the SMB protocol
	// dials (multichannel is premium-only -- enforced by spec validation
	// before the deploy ever runs).
	if spec.ShareProperties != nil {
		sharePropertiesArgs := &storage.AccountSharePropertiesArgs{}
		if spec.ShareProperties.RetentionPolicy != nil {
			days := 7
			if spec.ShareProperties.RetentionPolicy.Days != nil {
				days = int(spec.ShareProperties.RetentionPolicy.GetDays())
			}
			sharePropertiesArgs.RetentionPolicy = &storage.AccountSharePropertiesRetentionPolicyArgs{
				Days: pulumi.Int(days),
			}
		}
		if spec.ShareProperties.Smb != nil {
			smbArgs := &storage.AccountSharePropertiesSmbArgs{
				MultichannelEnabled: pulumi.Bool(spec.ShareProperties.Smb.MultichannelEnabled),
			}
			if len(spec.ShareProperties.Smb.Versions) > 0 {
				smbArgs.Versions = pulumi.ToStringArray(spec.ShareProperties.Smb.Versions)
			}
			if len(spec.ShareProperties.Smb.AuthenticationTypes) > 0 {
				smbArgs.AuthenticationTypes = pulumi.ToStringArray(spec.ShareProperties.Smb.AuthenticationTypes)
			}
			if len(spec.ShareProperties.Smb.KerberosTicketEncryptionType) > 0 {
				smbArgs.KerberosTicketEncryptionTypes = pulumi.ToStringArray(spec.ShareProperties.Smb.KerberosTicketEncryptionType)
			}
			if len(spec.ShareProperties.Smb.ChannelEncryptionType) > 0 {
				smbArgs.ChannelEncryptionTypes = pulumi.ToStringArray(spec.ShareProperties.Smb.ChannelEncryptionType)
			}
			sharePropertiesArgs.Smb = smbArgs
		}
		if len(spec.ShareProperties.CorsRules) > 0 {
			corsRules := storage.AccountSharePropertiesCorsRuleArray{}
			for _, corsRule := range spec.ShareProperties.CorsRules {
				corsRules = append(corsRules, &storage.AccountSharePropertiesCorsRuleArgs{
					AllowedOrigins:  pulumi.ToStringArray(corsRule.AllowedOrigins),
					AllowedMethods:  pulumi.ToStringArray(corsRule.AllowedMethods),
					AllowedHeaders:  pulumi.ToStringArray(corsRule.AllowedHeaders),
					ExposedHeaders:  pulumi.ToStringArray(corsRule.ExposedHeaders),
					MaxAgeInSeconds: pulumi.Int(int(corsRule.MaxAgeInSeconds)),
				})
			}
			sharePropertiesArgs.CorsRules = corsRules
		}
		accountArgs.ShareProperties = sharePropertiesArgs
	}

	// Traffic routing preference and the optional routing-specific
	// endpoint publication.
	if spec.Routing != nil {
		routingChoice := "MicrosoftRouting"
		if spec.Routing.Choice != azurestorageaccountv1alpha1.AzureStorageAccountRoutingChoice_azure_storage_account_routing_choice_unspecified {
			routingChoice = routingChoiceStrings[spec.Routing.Choice]
		}
		accountArgs.Routing = &storage.AccountRoutingArgs{
			Choice:                    pulumi.String(routingChoice),
			PublishInternetEndpoints:  pulumi.Bool(spec.Routing.PublishInternetEndpoints),
			PublishMicrosoftEndpoints: pulumi.Bool(spec.Routing.PublishMicrosoftEndpoints),
		}
	}

	// Custom domain for the blob endpoint; use_subdomain validates
	// ownership via the asverify CNAME to avoid downtime on live domains.
	if spec.CustomDomain != nil {
		accountArgs.CustomDomain = &storage.AccountCustomDomainArgs{
			Name:         pulumi.String(spec.CustomDomain.Name),
			UseSubdomain: pulumi.Bool(spec.CustomDomain.UseSubdomain),
		}
	}

	// Identity-based SMB authentication for Azure Files.
	if spec.AzureFilesAuthentication != nil {
		filesAuthArgs := &storage.AccountAzureFilesAuthenticationArgs{
			DirectoryType: pulumi.String(directoryTypeStrings[spec.AzureFilesAuthentication.DirectoryType]),
		}
		if spec.AzureFilesAuthentication.DefaultShareLevelPermission != azurestorageaccountv1alpha1.AzureStorageAccountDefaultSharePermission_azure_storage_account_default_share_permission_unspecified {
			filesAuthArgs.DefaultShareLevelPermission = pulumi.String(defaultSharePermissionStrings[spec.AzureFilesAuthentication.DefaultShareLevelPermission])
		}
		if spec.AzureFilesAuthentication.ActiveDirectory != nil {
			activeDirectory := spec.AzureFilesAuthentication.ActiveDirectory
			activeDirectoryArgs := &storage.AccountAzureFilesAuthenticationActiveDirectoryArgs{
				DomainName: pulumi.String(activeDirectory.DomainName),
				DomainGuid: pulumi.String(activeDirectory.DomainGuid),
			}
			if activeDirectory.DomainSid != "" {
				activeDirectoryArgs.DomainSid = pulumi.String(activeDirectory.DomainSid)
			}
			if activeDirectory.StorageSid != "" {
				activeDirectoryArgs.StorageSid = pulumi.String(activeDirectory.StorageSid)
			}
			if activeDirectory.ForestName != "" {
				activeDirectoryArgs.ForestName = pulumi.String(activeDirectory.ForestName)
			}
			if activeDirectory.NetbiosDomainName != "" {
				activeDirectoryArgs.NetbiosDomainName = pulumi.String(activeDirectory.NetbiosDomainName)
			}
			filesAuthArgs.ActiveDirectory = activeDirectoryArgs
		}
		accountArgs.AzureFilesAuthentication = filesAuthArgs
	}

	// Account-level WORM policy. LOCKED is irreversible -- Azure itself
	// cannot shorten a locked retention window; the block is ForceNew.
	if spec.ImmutabilityPolicy != nil {
		accountArgs.ImmutabilityPolicy = &storage.AccountImmutabilityPolicyArgs{
			State:                      pulumi.String(immutabilityStateStrings[spec.ImmutabilityPolicy.State]),
			PeriodSinceCreationInDays:  pulumi.Int(int(spec.ImmutabilityPolicy.PeriodSinceCreationInDays)),
			AllowProtectedAppendWrites: pulumi.Bool(spec.ImmutabilityPolicy.AllowProtectedAppendWrites),
		}
	}

	createdAccount, err := storage.NewAccount(ctx,
		spec.AccountName,
		accountArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create storage account %s", spec.AccountName)
	}

	// Blob lifecycle management: ARM models this as ONE policy document
	// per account (name hardcoded "default" server-side), which is why
	// the rules fold into the account spec instead of being their own
	// kind. Absent aging thresholds are simply not sent -- the provider's
	// -1 sentinel is an HCL-ergonomics artifact this module never needs.
	if len(spec.LifecycleRules) > 0 {
		rules := storage.ManagementPolicyRuleArray{}
		for _, lifecycleRule := range spec.LifecycleRules {
			enabled := true
			if lifecycleRule.Enabled != nil {
				enabled = lifecycleRule.GetEnabled()
			}

			blobTypes := make([]string, 0, len(lifecycleRule.Filters.BlobTypes))
			for _, blobType := range lifecycleRule.Filters.BlobTypes {
				blobTypes = append(blobTypes, lifecycleBlobTypeStrings[blobType])
			}
			filtersArgs := &storage.ManagementPolicyRuleFiltersArgs{
				BlobTypes: pulumi.ToStringArray(blobTypes),
			}
			if len(lifecycleRule.Filters.PrefixMatch) > 0 {
				filtersArgs.PrefixMatches = pulumi.ToStringArray(lifecycleRule.Filters.PrefixMatch)
			}
			if len(lifecycleRule.Filters.MatchBlobIndexTags) > 0 {
				tagFilters := storage.ManagementPolicyRuleFiltersMatchBlobIndexTagArray{}
				for _, tagFilter := range lifecycleRule.Filters.MatchBlobIndexTags {
					operation := "=="
					if tagFilter.Operation != nil {
						operation = tagFilter.GetOperation()
					}
					tagFilters = append(tagFilters, &storage.ManagementPolicyRuleFiltersMatchBlobIndexTagArgs{
						Name:      pulumi.String(tagFilter.Name),
						Operation: pulumi.String(operation),
						Value:     pulumi.String(tagFilter.Value),
					})
				}
				filtersArgs.MatchBlobIndexTags = tagFilters
			}

			actionsArgs := &storage.ManagementPolicyRuleActionsArgs{}
			if baseBlob := lifecycleRule.Actions.BaseBlob; baseBlob != nil {
				baseBlobArgs := &storage.ManagementPolicyRuleActionsBaseBlobArgs{}
				if baseBlob.TierToCoolAfterDaysSinceModificationGreaterThan != nil {
					baseBlobArgs.TierToCoolAfterDaysSinceModificationGreaterThan = pulumi.Int(int(baseBlob.GetTierToCoolAfterDaysSinceModificationGreaterThan()))
				}
				if baseBlob.TierToCoolAfterDaysSinceLastAccessTimeGreaterThan != nil {
					baseBlobArgs.TierToCoolAfterDaysSinceLastAccessTimeGreaterThan = pulumi.Int(int(baseBlob.GetTierToCoolAfterDaysSinceLastAccessTimeGreaterThan()))
				}
				if baseBlob.TierToCoolAfterDaysSinceCreationGreaterThan != nil {
					baseBlobArgs.TierToCoolAfterDaysSinceCreationGreaterThan = pulumi.Int(int(baseBlob.GetTierToCoolAfterDaysSinceCreationGreaterThan()))
				}
				if baseBlob.AutoTierToHotFromCoolEnabled {
					baseBlobArgs.AutoTierToHotFromCoolEnabled = pulumi.Bool(true)
				}
				if baseBlob.TierToColdAfterDaysSinceModificationGreaterThan != nil {
					baseBlobArgs.TierToColdAfterDaysSinceModificationGreaterThan = pulumi.Int(int(baseBlob.GetTierToColdAfterDaysSinceModificationGreaterThan()))
				}
				if baseBlob.TierToColdAfterDaysSinceLastAccessTimeGreaterThan != nil {
					baseBlobArgs.TierToColdAfterDaysSinceLastAccessTimeGreaterThan = pulumi.Int(int(baseBlob.GetTierToColdAfterDaysSinceLastAccessTimeGreaterThan()))
				}
				if baseBlob.TierToColdAfterDaysSinceCreationGreaterThan != nil {
					baseBlobArgs.TierToColdAfterDaysSinceCreationGreaterThan = pulumi.Int(int(baseBlob.GetTierToColdAfterDaysSinceCreationGreaterThan()))
				}
				if baseBlob.TierToArchiveAfterDaysSinceModificationGreaterThan != nil {
					baseBlobArgs.TierToArchiveAfterDaysSinceModificationGreaterThan = pulumi.Int(int(baseBlob.GetTierToArchiveAfterDaysSinceModificationGreaterThan()))
				}
				if baseBlob.TierToArchiveAfterDaysSinceLastAccessTimeGreaterThan != nil {
					baseBlobArgs.TierToArchiveAfterDaysSinceLastAccessTimeGreaterThan = pulumi.Int(int(baseBlob.GetTierToArchiveAfterDaysSinceLastAccessTimeGreaterThan()))
				}
				if baseBlob.TierToArchiveAfterDaysSinceCreationGreaterThan != nil {
					baseBlobArgs.TierToArchiveAfterDaysSinceCreationGreaterThan = pulumi.Int(int(baseBlob.GetTierToArchiveAfterDaysSinceCreationGreaterThan()))
				}
				if baseBlob.TierToArchiveAfterDaysSinceLastTierChangeGreaterThan != nil {
					baseBlobArgs.TierToArchiveAfterDaysSinceLastTierChangeGreaterThan = pulumi.Int(int(baseBlob.GetTierToArchiveAfterDaysSinceLastTierChangeGreaterThan()))
				}
				if baseBlob.DeleteAfterDaysSinceModificationGreaterThan != nil {
					baseBlobArgs.DeleteAfterDaysSinceModificationGreaterThan = pulumi.Int(int(baseBlob.GetDeleteAfterDaysSinceModificationGreaterThan()))
				}
				if baseBlob.DeleteAfterDaysSinceLastAccessTimeGreaterThan != nil {
					baseBlobArgs.DeleteAfterDaysSinceLastAccessTimeGreaterThan = pulumi.Int(int(baseBlob.GetDeleteAfterDaysSinceLastAccessTimeGreaterThan()))
				}
				if baseBlob.DeleteAfterDaysSinceCreationGreaterThan != nil {
					baseBlobArgs.DeleteAfterDaysSinceCreationGreaterThan = pulumi.Int(int(baseBlob.GetDeleteAfterDaysSinceCreationGreaterThan()))
				}
				actionsArgs.BaseBlob = baseBlobArgs
			}
			if snapshot := lifecycleRule.Actions.Snapshot; snapshot != nil {
				snapshotArgs := &storage.ManagementPolicyRuleActionsSnapshotArgs{}
				if snapshot.ChangeTierToCoolAfterDaysSinceCreation != nil {
					snapshotArgs.ChangeTierToCoolAfterDaysSinceCreation = pulumi.Int(int(snapshot.GetChangeTierToCoolAfterDaysSinceCreation()))
				}
				if snapshot.TierToColdAfterDaysSinceCreationGreaterThan != nil {
					snapshotArgs.TierToColdAfterDaysSinceCreationGreaterThan = pulumi.Int(int(snapshot.GetTierToColdAfterDaysSinceCreationGreaterThan()))
				}
				if snapshot.ChangeTierToArchiveAfterDaysSinceCreation != nil {
					snapshotArgs.ChangeTierToArchiveAfterDaysSinceCreation = pulumi.Int(int(snapshot.GetChangeTierToArchiveAfterDaysSinceCreation()))
				}
				if snapshot.TierToArchiveAfterDaysSinceLastTierChangeGreaterThan != nil {
					snapshotArgs.TierToArchiveAfterDaysSinceLastTierChangeGreaterThan = pulumi.Int(int(snapshot.GetTierToArchiveAfterDaysSinceLastTierChangeGreaterThan()))
				}
				if snapshot.DeleteAfterDaysSinceCreationGreaterThan != nil {
					snapshotArgs.DeleteAfterDaysSinceCreationGreaterThan = pulumi.Int(int(snapshot.GetDeleteAfterDaysSinceCreationGreaterThan()))
				}
				actionsArgs.Snapshot = snapshotArgs
			}
			if version := lifecycleRule.Actions.Version; version != nil {
				versionArgs := &storage.ManagementPolicyRuleActionsVersionArgs{}
				if version.ChangeTierToCoolAfterDaysSinceCreation != nil {
					versionArgs.ChangeTierToCoolAfterDaysSinceCreation = pulumi.Int(int(version.GetChangeTierToCoolAfterDaysSinceCreation()))
				}
				if version.TierToColdAfterDaysSinceCreationGreaterThan != nil {
					versionArgs.TierToColdAfterDaysSinceCreationGreaterThan = pulumi.Int(int(version.GetTierToColdAfterDaysSinceCreationGreaterThan()))
				}
				if version.ChangeTierToArchiveAfterDaysSinceCreation != nil {
					versionArgs.ChangeTierToArchiveAfterDaysSinceCreation = pulumi.Int(int(version.GetChangeTierToArchiveAfterDaysSinceCreation()))
				}
				if version.TierToArchiveAfterDaysSinceLastTierChangeGreaterThan != nil {
					versionArgs.TierToArchiveAfterDaysSinceLastTierChangeGreaterThan = pulumi.Int(int(version.GetTierToArchiveAfterDaysSinceLastTierChangeGreaterThan()))
				}
				if version.DeleteAfterDaysSinceCreation != nil {
					versionArgs.DeleteAfterDaysSinceCreation = pulumi.Int(int(version.GetDeleteAfterDaysSinceCreation()))
				}
				actionsArgs.Version = versionArgs
			}

			rules = append(rules, &storage.ManagementPolicyRuleArgs{
				Name:    pulumi.String(lifecycleRule.Name),
				Enabled: pulumi.Bool(enabled),
				Filters: filtersArgs,
				Actions: actionsArgs,
			})
		}

		if _, err := storage.NewManagementPolicy(ctx,
			fmt.Sprintf("%s-lifecycle", spec.AccountName),
			&storage.ManagementPolicyArgs{
				StorageAccountId: createdAccount.ID(),
				Rules:            rules,
			},
			pulumi.Provider(azureProvider),
			pulumi.Parent(createdAccount)); err != nil {
			return errors.Wrap(err, "failed to create lifecycle management policy")
		}
	}

	// Static website hosting, via the standalone resource (azurerm
	// removed the inline static_website block in v5). The service
	// auto-creates the $web container; upload site content there.
	if spec.StaticWebsite != nil {
		staticWebsiteArgs := &storage.AccountStaticWebsiteArgs{
			StorageAccountId: createdAccount.ID(),
		}
		if spec.StaticWebsite.IndexDocument != "" {
			staticWebsiteArgs.IndexDocument = pulumi.String(spec.StaticWebsite.IndexDocument)
		}
		if spec.StaticWebsite.Error_404Document != "" {
			staticWebsiteArgs.Error404Document = pulumi.String(spec.StaticWebsite.Error_404Document)
		}
		if _, err := storage.NewAccountStaticWebsite(ctx,
			fmt.Sprintf("%s-static-website", spec.AccountName),
			staticWebsiteArgs,
			pulumi.Provider(azureProvider),
			pulumi.Parent(createdAccount)); err != nil {
			return errors.Wrap(err, "failed to configure static website")
		}
	}

	// Export outputs from the created account. The secondary
	// endpoints resolve to empty strings on non-read-access replication
	// types; the identity principal is empty unless the type includes
	// SystemAssigned.
	ctx.Export(OpStorageAccountId, createdAccount.ID())
	ctx.Export(OpStorageAccountName, createdAccount.Name)
	ctx.Export(OpResourceGroupName, createdAccount.ResourceGroupName)
	ctx.Export(OpPrimaryBlobEndpoint, createdAccount.PrimaryBlobEndpoint)
	ctx.Export(OpPrimaryBlobHost, createdAccount.PrimaryBlobHost)
	ctx.Export(OpPrimaryQueueEndpoint, createdAccount.PrimaryQueueEndpoint)
	ctx.Export(OpPrimaryTableEndpoint, createdAccount.PrimaryTableEndpoint)
	ctx.Export(OpPrimaryFileEndpoint, createdAccount.PrimaryFileEndpoint)
	ctx.Export(OpPrimaryDfsEndpoint, createdAccount.PrimaryDfsEndpoint)
	ctx.Export(OpPrimaryWebEndpoint, createdAccount.PrimaryWebEndpoint)
	ctx.Export(OpPrimaryWebHost, createdAccount.PrimaryWebHost)
	ctx.Export(OpSecondaryBlobEndpoint, createdAccount.SecondaryBlobEndpoint)
	ctx.Export(OpSecondaryQueueEndpoint, createdAccount.SecondaryQueueEndpoint)
	ctx.Export(OpSecondaryTableEndpoint, createdAccount.SecondaryTableEndpoint)
	ctx.Export(OpSecondaryFileEndpoint, createdAccount.SecondaryFileEndpoint)
	ctx.Export(OpSecondaryDfsEndpoint, createdAccount.SecondaryDfsEndpoint)
	ctx.Export(OpSecondaryWebEndpoint, createdAccount.SecondaryWebEndpoint)
	ctx.Export(OpPrimaryAccessKey, pulumi.ToSecret(createdAccount.PrimaryAccessKey))
	ctx.Export(OpSecondaryAccessKey, pulumi.ToSecret(createdAccount.SecondaryAccessKey))
	ctx.Export(OpPrimaryConnectionString, pulumi.ToSecret(createdAccount.PrimaryConnectionString))
	ctx.Export(OpSecondaryConnectionString, pulumi.ToSecret(createdAccount.SecondaryConnectionString))
	ctx.Export(OpPrimaryBlobConnectionString, pulumi.ToSecret(createdAccount.PrimaryBlobConnectionString))
	ctx.Export(OpSecondaryBlobConnectionString, pulumi.ToSecret(createdAccount.SecondaryBlobConnectionString))
	ctx.Export(OpIdentityPrincipalId, createdAccount.Identity.ApplyT(func(identity *storage.AccountIdentity) string {
		if identity == nil || identity.PrincipalId == nil {
			return ""
		}
		return *identity.PrincipalId
	}).(pulumi.StringOutput))

	// The per-service diagnostic-settings scopes, constructed from the
	// account ID (identically on both engines). Data-access telemetry
	// (StorageRead/StorageWrite/StorageDelete logs) lives on these
	// implicit service sub-resources, not the account itself -- ARM
	// materializes them with the account, so there is nothing to read
	// back.
	ctx.Export(OpBlobServiceId, pulumi.Sprintf("%s/blobServices/default", createdAccount.ID()))
	ctx.Export(OpFileServiceId, pulumi.Sprintf("%s/fileServices/default", createdAccount.ID()))
	ctx.Export(OpQueueServiceId, pulumi.Sprintf("%s/queueServices/default", createdAccount.ID()))
	ctx.Export(OpTableServiceId, pulumi.Sprintf("%s/tableServices/default", createdAccount.ID()))

	return nil
}
