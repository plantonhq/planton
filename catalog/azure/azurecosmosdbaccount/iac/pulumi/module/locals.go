package module

import (
	"strings"

	azurecosmosdbaccountv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecosmosdbaccount/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureCosmosdbAccount *azurecosmosdbaccountv1alpha1.AzureCosmosdbAccount
	ResourceGroupName    string
	AzureTags            map[string]string
}

// kindStrings maps the spec's kind enum to ARM's wire values.
// Unspecified materializes GlobalDocumentDB in main.go -- the SQL API,
// azurerm's own default.
var kindStrings = map[azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountKind]string{
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountKind_GLOBAL_DOCUMENT_DB: "GlobalDocumentDB",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountKind_MONGO_DB:           "MongoDB",
}

// consistencyLevelStrings maps the consistency enum to ARM's wire values.
// Unspecified materializes Session -- Azure's recommended default.
var consistencyLevelStrings = map[azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountConsistencyLevel]string{
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountConsistencyLevel_STRONG:            "Strong",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountConsistencyLevel_BOUNDED_STALENESS: "BoundedStaleness",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountConsistencyLevel_SESSION:           "Session",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountConsistencyLevel_CONSISTENT_PREFIX: "ConsistentPrefix",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountConsistencyLevel_EVENTUAL:          "Eventual",
}

// capabilityStrings maps the capability enum to ARM's exact wire values --
// including the two that break the EnableX convention (MongoDBv3.4 and
// mongoEnableDocLevelTTL). The map is exhaustive by construction: a
// missing entry would silently drop a declared capability.
var capabilityStrings = map[azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountCapability]string{
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountCapability_ENABLE_SERVERLESS:                      "EnableServerless",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountCapability_ENABLE_CASSANDRA:                       "EnableCassandra",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountCapability_ENABLE_GREMLIN:                         "EnableGremlin",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountCapability_ENABLE_TABLE:                           "EnableTable",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountCapability_ENABLE_AGGREGATION_PIPELINE:            "EnableAggregationPipeline",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountCapability_ENABLE_MONGO:                           "EnableMongo",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountCapability_ENABLE_MONGO_16MB_DOCUMENT_SUPPORT:     "EnableMongo16MBDocumentSupport",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountCapability_MONGO_DB_V34:                           "MongoDBv3.4",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountCapability_MONGO_ENABLE_DOC_LEVEL_TTL:             "mongoEnableDocLevelTTL",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountCapability_DELETE_ALL_ITEMS_BY_PARTITION_KEY:      "DeleteAllItemsByPartitionKey",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountCapability_DISABLE_RATE_LIMITING_RESPONSES:        "DisableRateLimitingResponses",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountCapability_ALLOW_SELF_SERVE_UPGRADE_TO_MONGO36:    "AllowSelfServeUpgradeToMongo36",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountCapability_ENABLE_MONGO_RETRYABLE_WRITES:          "EnableMongoRetryableWrites",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountCapability_ENABLE_MONGO_ROLE_BASED_ACCESS_CONTROL: "EnableMongoRoleBasedAccessControl",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountCapability_ENABLE_UNIQUE_COMPOUND_NESTED_DOCS:     "EnableUniqueCompoundNestedDocs",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountCapability_ENABLE_NO_SQL_VECTOR_SEARCH:            "EnableNoSQLVectorSearch",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountCapability_ENABLE_NO_SQL_FULL_TEXT_SEARCH:         "EnableNoSQLFullTextSearch",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountCapability_ENABLE_TTL_ON_CUSTOM_PATH:              "EnableTtlOnCustomPath",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountCapability_ENABLE_PARTIAL_UNIQUE_INDEX:            "EnablePartialUniqueIndex",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountCapability_ENABLE_FABRIC_NETWORK_ACL_BYPASS:       "EnableFabricNetworkAclBypass",
}

// backupTypeStrings maps the backup-type enum to ARM's wire values.
var backupTypeStrings = map[azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountBackupType]string{
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountBackupType_PERIODIC:   "Periodic",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountBackupType_CONTINUOUS: "Continuous",
}

// continuousTierStrings maps the continuous-backup tier enum to ARM's
// wire values. Unspecified is never sent -- Azure defaults to 30 days.
var continuousTierStrings = map[azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountContinuousTier]string{
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountContinuousTier_CONTINUOUS_7_DAYS:  "Continuous7Days",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountContinuousTier_CONTINUOUS_30_DAYS: "Continuous30Days",
}

// backupStorageRedundancyStrings maps the periodic-backup redundancy enum
// to ARM's wire values. Unspecified is never sent -- Azure defaults Geo.
var backupStorageRedundancyStrings = map[azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountBackupStorageRedundancy]string{
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountBackupStorageRedundancy_GEO:   "Geo",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountBackupStorageRedundancy_LOCAL: "Local",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountBackupStorageRedundancy_ZONE:  "Zone",
}

// mongoServerVersionStrings maps the Mongo wire-protocol enum to ARM's
// version strings. Unspecified is never sent -- Azure picks its default.
var mongoServerVersionStrings = map[azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountMongoServerVersion]string{
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountMongoServerVersion_MONGO_3_2: "3.2",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountMongoServerVersion_MONGO_3_6: "3.6",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountMongoServerVersion_MONGO_4_0: "4.0",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountMongoServerVersion_MONGO_4_2: "4.2",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountMongoServerVersion_MONGO_5_0: "5.0",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountMongoServerVersion_MONGO_6_0: "6.0",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountMongoServerVersion_MONGO_7_0: "7.0",
}

// identityTypeStrings maps the identity-type enum to ARM's wire values.
var identityTypeStrings = map[azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountIdentityType]string{
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountIdentityType_SYSTEM_ASSIGNED:          "SystemAssigned",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountIdentityType_USER_ASSIGNED:            "UserAssigned",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountIdentityType_SYSTEM_AND_USER_ASSIGNED: "SystemAssigned, UserAssigned",
}

// analyticalSchemaTypeStrings maps the analytical-store schema enum to
// ARM's wire values.
var analyticalSchemaTypeStrings = map[azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountAnalyticalStorageSchemaType]string{
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountAnalyticalStorageSchemaType_WELL_DEFINED:  "WellDefined",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountAnalyticalStorageSchemaType_FULL_FIDELITY: "FullFidelity",
}

// createModeStrings maps the create-mode enum to ARM's wire values.
// Unspecified is never sent -- azurerm rejects create_mode on accounts
// without continuous backup, and the spec mirrors that contract.
var createModeStrings = map[azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountCreateMode]string{
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountCreateMode_DEFAULT: "Default",
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountCreateMode_RESTORE: "Restore",
}

// minimalTlsVersionStrings maps the TLS-floor enum to ARM's wire values.
// Unspecified materializes Tls12 in the module (Azure's own default
// since April 2023), matching the Terraform module. Tls12 is the only
// floor Azure still provisions; the retired Tls/Tls11 values no longer
// exist on the spec enum.
var minimalTlsVersionStrings = map[azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountMinimalTlsVersion]string{
	azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountMinimalTlsVersion_TLS_1_2: "Tls12",
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurecosmosdbaccountv1alpha1.AzureCosmosdbAccountStackInput) *Locals {
	locals := &Locals{}
	locals.AzureCosmosdbAccount = stackInput.Target
	target := stackInput.Target
	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Metadata-derived identity tags first, then the user's spec tags
	// merged over them: user tags deliberately win so an org's
	// governance conventions can override the derived values.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzureCosmosdbAccount.String()),
	}
	if target.Metadata.Id != "" {
		locals.AzureTags[azuretagkeys.ResourceId] = target.Metadata.Id
	}
	if target.Metadata.Org != "" {
		locals.AzureTags[azuretagkeys.Organization] = target.Metadata.Org
	}
	if target.Metadata.Env != "" {
		locals.AzureTags[azuretagkeys.Environment] = target.Metadata.Env
	}
	for key, value := range target.Spec.Tags {
		locals.AzureTags[key] = value
	}
	return locals
}
