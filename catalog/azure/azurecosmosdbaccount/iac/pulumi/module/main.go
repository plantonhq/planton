package module

import (
	"github.com/pkg/errors"
	azurecosmosdbaccountv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecosmosdbaccount/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/cosmosdb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureCosmosdbAccount.Spec

	// Unspecified kind materializes GlobalDocumentDB -- the SQL API,
	// azurerm's own default. Fixed at creation: the wire protocol
	// shapes how every byte is stored.
	kind := "GlobalDocumentDB"
	if spec.Kind != azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountKind_azure_cosmosdb_account_kind_unspecified {
		kind = kindStrings[spec.Kind]
	}

	// Unspecified consistency materializes Session -- Azure's
	// recommended default. The staleness dials carry the proto defaults
	// (5 / 100) when unset: IaC inputs built from a manifest do NOT
	// materialize proto defaults, so bare getters would send zeros the
	// API rejects.
	consistencyLevel := "Session"
	if spec.ConsistencyPolicy.ConsistencyLevel != azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountConsistencyLevel_azure_cosmosdb_account_consistency_level_unspecified {
		consistencyLevel = consistencyLevelStrings[spec.ConsistencyPolicy.ConsistencyLevel]
	}
	maxIntervalInSeconds := 5
	if spec.ConsistencyPolicy.MaxIntervalInSeconds != nil {
		maxIntervalInSeconds = int(spec.ConsistencyPolicy.GetMaxIntervalInSeconds())
	}
	maxStalenessPrefix := 100
	if spec.ConsistencyPolicy.MaxStalenessPrefix != nil {
		maxStalenessPrefix = int(spec.ConsistencyPolicy.GetMaxStalenessPrefix())
	}

	// The replicated regions: the priority-0 entry is the write region;
	// adding/removing regions is an in-place update.
	geoLocations := cosmosdb.AccountGeoLocationArray{}
	for _, geo := range spec.GeoLocations {
		geoLocations = append(geoLocations, cosmosdb.AccountGeoLocationArgs{
			Location:         pulumi.String(geo.Location),
			FailoverPriority: pulumi.Int(int(geo.FailoverPriority)),
			ZoneRedundant:    pulumi.Bool(geo.GetZoneRedundant()),
		})
	}

	accountArgs := &cosmosdb.AccountArgs{
		Name:              pulumi.String(spec.AccountName),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		// Azure's only offer type -- nothing to choose, so not modeled
		// in the spec.
		OfferType: pulumi.String("Standard"),
		Kind:      pulumi.String(kind),
		ConsistencyPolicy: cosmosdb.AccountConsistencyPolicyArgs{
			ConsistencyLevel:     pulumi.String(consistencyLevel),
			MaxIntervalInSeconds: pulumi.Int(maxIntervalInSeconds),
			MaxStalenessPrefix:   pulumi.Int(maxStalenessPrefix),
		},
		GeoLocations: geoLocations,
		// Presence-guarded to the proto defaults: bare getters on unset
		// optional bools would silently flip Azure's own defaults.
		FreeTierEnabled:               pulumi.Bool(spec.GetFreeTierEnabled()),
		AutomaticFailoverEnabled:      pulumi.Bool(spec.GetAutomaticFailoverEnabled()),
		MultipleWriteLocationsEnabled: pulumi.Bool(spec.GetMultipleWriteLocationsEnabled()),
		PublicNetworkAccessEnabled:    pulumi.Bool(publicNetworkAccessEnabled(spec)),
		IsVirtualNetworkFilterEnabled: pulumi.Bool(spec.GetIsVirtualNetworkFilterEnabled()),
		// Key- and metadata-write posture: disabling local auth forces
		// every data-plane caller through Entra ID (the exported account
		// keys stop authenticating).
		//
		// PARITY-EXCEPTION: pulumi-azure v6.38 has bridged only the
		// deprecated localAuthenticationDisabled input, so this module
		// sends the negation of the spec's local_authentication_enabled;
		// the Terraform module uses azurerm's non-deprecated
		// local_authentication_enabled directly. Both set the same
		// DisableLocalAuth wire property -- the created account is
		// identical. Re-align here when the bridge exposes
		// localAuthenticationEnabled.
		AccessKeyMetadataWritesEnabled:   pulumi.Bool(accessKeyMetadataWritesEnabled(spec)),
		LocalAuthenticationDisabled:      pulumi.Bool(!localAuthenticationEnabled(spec)),
		NetworkAclBypassForAzureServices: pulumi.Bool(spec.GetNetworkAclBypassForAzureServices()),
		AnalyticalStorageEnabled:         pulumi.Bool(spec.GetAnalyticalStorageEnabled()),
		BurstCapacityEnabled:             pulumi.Bool(spec.GetBurstCapacityEnabled()),
		PartitionMergeEnabled:            pulumi.Bool(spec.GetPartitionMergeEnabled()),
		Tags:                             pulumi.ToStringMap(locals.AzureTags),
	}

	// Capabilities are exactly what the spec declares -- never injected
	// silently (a MONGO_DB account declares ENABLE_MONGO itself; presets
	// and docs teach this). Most capability changes recreate the account.
	if len(spec.Capabilities) > 0 {
		capabilityArray := cosmosdb.AccountCapabilityArray{}
		for _, capability := range spec.Capabilities {
			capabilityArray = append(capabilityArray, cosmosdb.AccountCapabilityArgs{
				Name: pulumi.String(capabilityStrings[capability]),
			})
		}
		accountArgs.Capabilities = capabilityArray
	}

	// Network posture: the virtual-network filter admits the declared
	// subnets; the IP filter admits the declared addresses.
	if len(spec.VirtualNetworkRules) > 0 {
		ruleArray := cosmosdb.AccountVirtualNetworkRuleArray{}
		for _, rule := range spec.VirtualNetworkRules {
			ruleArray = append(ruleArray, cosmosdb.AccountVirtualNetworkRuleArgs{
				Id:                               pulumi.String(rule.SubnetId.GetValue()),
				IgnoreMissingVnetServiceEndpoint: pulumi.Bool(rule.GetIgnoreMissingVnetServiceEndpoint()),
			})
		}
		accountArgs.VirtualNetworkRules = ruleArray
	}
	if len(spec.IpRangeFilter) > 0 {
		accountArgs.IpRangeFilters = pulumi.ToStringArray(spec.IpRangeFilter)
	}
	if len(spec.NetworkAclBypassIds) > 0 {
		accountArgs.NetworkAclBypassIds = pulumi.ToStringArray(spec.NetworkAclBypassIds)
	}

	// Backup: PERIODIC -> CONTINUOUS upgrades in place; the reverse
	// recreates the account. Per-mode field pairings are enforced by the
	// spec's validation rules; unset dials are omitted so Azure's own
	// defaults apply.
	if spec.Backup != nil {
		backupArgs := cosmosdb.AccountBackupArgs{
			Type: pulumi.String(backupTypeStrings[spec.Backup.Type]),
		}
		if spec.Backup.Tier != azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountContinuousTier_azure_cosmosdb_account_continuous_tier_unspecified {
			backupArgs.Tier = pulumi.String(continuousTierStrings[spec.Backup.Tier])
		}
		if spec.Backup.IntervalInMinutes != nil {
			backupArgs.IntervalInMinutes = pulumi.Int(int(spec.Backup.GetIntervalInMinutes()))
		}
		if spec.Backup.RetentionInHours != nil {
			backupArgs.RetentionInHours = pulumi.Int(int(spec.Backup.GetRetentionInHours()))
		}
		if spec.Backup.StorageRedundancy != azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountBackupStorageRedundancy_azure_cosmosdb_account_backup_storage_redundancy_unspecified {
			backupArgs.StorageRedundancy = pulumi.String(backupStorageRedundancyStrings[spec.Backup.StorageRedundancy])
		}
		accountArgs.Backup = backupArgs
	}

	if spec.MongoServerVersion != azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountMongoServerVersion_azure_cosmosdb_account_mongo_server_version_unspecified {
		accountArgs.MongoServerVersion = pulumi.String(mongoServerVersionStrings[spec.MongoServerVersion])
	}

	// The TLS floor for every endpoint. Unset materializes Tls12 --
	// Azure's own default since April 2023 -- so both engines send the
	// same value; 1.0/1.1 exist only for legacy-client migrations.
	minimalTlsVersion := "Tls12"
	if spec.MinimalTlsVersion != azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountMinimalTlsVersion_azure_cosmosdb_account_minimal_tls_version_unspecified {
		minimalTlsVersion = minimalTlsVersionStrings[spec.MinimalTlsVersion]
	}
	accountArgs.MinimalTlsVersion = pulumi.String(minimalTlsVersion)

	// The managed identity, and which identity the account acts AS by
	// default against other services (CMK unwrapping): the composed
	// "UserAssignedIdentity=<id>" form makes CMK ride an identity that
	// exists -- and holds Key Vault grants -- before the account does.
	if spec.Identity != nil {
		identityArgs := cosmosdb.AccountIdentityArgs{
			Type: pulumi.String(identityTypeStrings[spec.Identity.Type]),
		}
		if len(spec.Identity.IdentityIds) > 0 {
			identityIds := pulumi.StringArray{}
			for _, identityId := range spec.Identity.IdentityIds {
				identityIds = append(identityIds, pulumi.String(identityId.GetValue()))
			}
			identityArgs.IdentityIds = identityIds
		}
		accountArgs.Identity = identityArgs
	}
	if spec.DefaultIdentity != nil {
		switch spec.DefaultIdentity.Type {
		case azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountDefaultIdentityType_FIRST_PARTY:
			accountArgs.DefaultIdentityType = pulumi.String("FirstPartyIdentity")
		case azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountDefaultIdentityType_SYSTEM_ASSIGNED_DEFAULT:
			accountArgs.DefaultIdentityType = pulumi.String("SystemAssignedIdentity")
		case azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountDefaultIdentityType_USER_ASSIGNED_DEFAULT:
			accountArgs.DefaultIdentityType = pulumi.String("UserAssignedIdentity=" + spec.DefaultIdentity.UserAssignedIdentityId.GetValue())
		}
	}

	// CMK encryption rides the key's VERSIONLESS Key Vault identifier so
	// rotation propagates without touching the account. Fixed at creation.
	if spec.KeyVaultKeyId.GetValue() != "" {
		accountArgs.KeyVaultKeyId = pulumi.String(spec.KeyVaultKeyId.GetValue())
	}

	if spec.AnalyticalStorage != nil {
		accountArgs.AnalyticalStorage = cosmosdb.AccountAnalyticalStorageArgs{
			SchemaType: pulumi.String(analyticalSchemaTypeStrings[spec.AnalyticalStorage.SchemaType]),
		}
	}

	// The account-wide provisioned-throughput cap (-1 = unlimited) --
	// the guardrail against runaway RU provisioning cost.
	if spec.Capacity != nil {
		accountArgs.Capacity = cosmosdb.AccountCapacityArgs{
			TotalThroughputLimit: pulumi.Int(int(spec.Capacity.TotalThroughputLimit)),
		}
	}

	if spec.CorsRule != nil {
		corsArgs := cosmosdb.AccountCorsRuleArgs{
			AllowedOrigins: pulumi.ToStringArray(spec.CorsRule.AllowedOrigins),
			AllowedMethods: pulumi.ToStringArray(spec.CorsRule.AllowedMethods),
			AllowedHeaders: pulumi.ToStringArray(spec.CorsRule.AllowedHeaders),
			ExposedHeaders: pulumi.ToStringArray(spec.CorsRule.ExposedHeaders),
		}
		if spec.CorsRule.MaxAgeInSeconds != nil {
			corsArgs.MaxAgeInSeconds = pulumi.Int(int(spec.CorsRule.GetMaxAgeInSeconds()))
		}
		accountArgs.CorsRule = corsArgs
	}

	// RESTORE creates the account FROM a continuous-backup restore
	// point; sent only when the spec sets it (azurerm rejects
	// create_mode on accounts without continuous backup).
	if spec.CreateMode != azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountCreateMode_azure_cosmosdb_account_create_mode_unspecified {
		accountArgs.CreateMode = pulumi.String(createModeStrings[spec.CreateMode])
	}
	if spec.Restore != nil {
		restoreArgs := cosmosdb.AccountRestoreArgs{
			SourceCosmosdbAccountId: pulumi.String(spec.Restore.SourceCosmosdbAccountId),
			RestoreTimestampInUtc:   pulumi.String(spec.Restore.RestoreTimestampInUtc),
		}
		if len(spec.Restore.Databases) > 0 {
			databaseArray := cosmosdb.AccountRestoreDatabaseArray{}
			for _, database := range spec.Restore.Databases {
				databaseArgs := cosmosdb.AccountRestoreDatabaseArgs{
					Name: pulumi.String(database.Name),
				}
				if len(database.CollectionNames) > 0 {
					databaseArgs.CollectionNames = pulumi.ToStringArray(database.CollectionNames)
				}
				databaseArray = append(databaseArray, databaseArgs)
			}
			restoreArgs.Databases = databaseArray
		}
		if len(spec.Restore.GremlinDatabases) > 0 {
			gremlinArray := cosmosdb.AccountRestoreGremlinDatabaseArray{}
			for _, gremlinDatabase := range spec.Restore.GremlinDatabases {
				gremlinArgs := cosmosdb.AccountRestoreGremlinDatabaseArgs{
					Name: pulumi.String(gremlinDatabase.Name),
				}
				if len(gremlinDatabase.GraphNames) > 0 {
					gremlinArgs.GraphNames = pulumi.ToStringArray(gremlinDatabase.GraphNames)
				}
				gremlinArray = append(gremlinArray, gremlinArgs)
			}
			restoreArgs.GremlinDatabases = gremlinArray
		}
		if len(spec.Restore.TablesToRestore) > 0 {
			restoreArgs.TablesToRestores = pulumi.ToStringArray(spec.Restore.TablesToRestore)
		}
		accountArgs.Restore = restoreArgs
	}

	createdAccount, err := cosmosdb.NewAccount(ctx,
		spec.AccountName,
		accountArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create cosmosdb account %s", spec.AccountName)
	}

	// Export outputs. The keys and connection strings are the
	// credential surface (secret-bearing); databases and containers are
	// their own kinds, so no database ids are exported here.
	ctx.Export(OpCosmosdbAccountId, createdAccount.ID())
	ctx.Export(OpCosmosdbAccountName, createdAccount.Name)
	ctx.Export(OpEndpoint, createdAccount.Endpoint)
	ctx.Export(OpReadEndpoints, createdAccount.ReadEndpoints)
	ctx.Export(OpWriteEndpoints, createdAccount.WriteEndpoints)
	ctx.Export(OpPrimaryKey, pulumi.ToSecret(createdAccount.PrimaryKey))
	ctx.Export(OpSecondaryKey, pulumi.ToSecret(createdAccount.SecondaryKey))
	ctx.Export(OpPrimaryReadonlyKey, pulumi.ToSecret(createdAccount.PrimaryReadonlyKey))
	ctx.Export(OpSecondaryReadonlyKey, pulumi.ToSecret(createdAccount.SecondaryReadonlyKey))
	ctx.Export(OpPrimarySqlConnectionString, pulumi.ToSecret(createdAccount.PrimarySqlConnectionString))
	ctx.Export(OpSecondarySqlConnectionString, pulumi.ToSecret(createdAccount.SecondarySqlConnectionString))
	ctx.Export(OpPrimaryReadonlySqlConnectionString, pulumi.ToSecret(createdAccount.PrimaryReadonlySqlConnectionString))
	ctx.Export(OpSecondaryReadonlySqlConnectionString, pulumi.ToSecret(createdAccount.SecondaryReadonlySqlConnectionString))
	ctx.Export(OpPrimaryMongodbConnectionString, pulumi.ToSecret(createdAccount.PrimaryMongodbConnectionString))
	ctx.Export(OpSecondaryMongodbConnectionString, pulumi.ToSecret(createdAccount.SecondaryMongodbConnectionString))
	ctx.Export(OpPrimaryReadonlyMongodbConnectionString, pulumi.ToSecret(createdAccount.PrimaryReadonlyMongodbConnectionString))
	ctx.Export(OpSecondaryReadonlyMongodbConnectionString, pulumi.ToSecret(createdAccount.SecondaryReadonlyMongodbConnectionString))
	// The system-assigned principal id, empty when no identity (or a
	// user-assigned-only identity) is requested.
	ctx.Export(OpIdentityPrincipalId, createdAccount.Identity.ApplyT(func(identity *cosmosdb.AccountIdentity) string {
		if identity == nil || identity.PrincipalId == nil {
			return ""
		}
		return *identity.PrincipalId
	}).(pulumi.StringOutput))

	return nil
}

// publicNetworkAccessEnabled returns the spec value presence-guarded to
// the proto default (true).
func publicNetworkAccessEnabled(spec *azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountSpec) bool {
	if spec.PublicNetworkAccessEnabled == nil {
		return true
	}
	return spec.GetPublicNetworkAccessEnabled()
}

// accessKeyMetadataWritesEnabled returns the spec value presence-guarded
// to the proto default (true).
func accessKeyMetadataWritesEnabled(spec *azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountSpec) bool {
	if spec.AccessKeyMetadataWritesEnabled == nil {
		return true
	}
	return spec.GetAccessKeyMetadataWritesEnabled()
}

// localAuthenticationEnabled returns the spec value presence-guarded to
// the proto default (true).
func localAuthenticationEnabled(spec *azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountSpec) bool {
	if spec.LocalAuthenticationEnabled == nil {
		return true
	}
	return spec.GetLocalAuthenticationEnabled()
}
