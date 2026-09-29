package module

import (
	"strings"

	azuremssqldatabasev1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremssqldatabase/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureMssqlDatabase *azuremssqldatabasev1alpha1.AzureMssqlDatabase
	AzureTags          map[string]string
}

// createModeStrings maps the spec's create-mode enum to ARM's values. An
// unspecified mode means a fresh (DEFAULT) database and is not sent at
// all, mirroring the Terraform module's null -- so the two engines
// produce the same ARM payload.
var createModeStrings = map[azuremssqldatabasev1alpha1.AzureMssqlDatabaseCreateMode]string{
	azuremssqldatabasev1alpha1.AzureMssqlDatabaseCreateMode_DEFAULT:                            "Default",
	azuremssqldatabasev1alpha1.AzureMssqlDatabaseCreateMode_COPY:                               "Copy",
	azuremssqldatabasev1alpha1.AzureMssqlDatabaseCreateMode_SECONDARY:                          "Secondary",
	azuremssqldatabasev1alpha1.AzureMssqlDatabaseCreateMode_ONLINE_SECONDARY:                   "OnlineSecondary",
	azuremssqldatabasev1alpha1.AzureMssqlDatabaseCreateMode_POINT_IN_TIME_RESTORE:              "PointInTimeRestore",
	azuremssqldatabasev1alpha1.AzureMssqlDatabaseCreateMode_RECOVERY:                           "Recovery",
	azuremssqldatabasev1alpha1.AzureMssqlDatabaseCreateMode_RESTORE:                            "Restore",
	azuremssqldatabasev1alpha1.AzureMssqlDatabaseCreateMode_RESTORE_LONG_TERM_RETENTION_BACKUP: "RestoreLongTermRetentionBackup",
}

// secondaryTypeStrings maps the spec's secondary-type enum to ARM's
// values.
var secondaryTypeStrings = map[azuremssqldatabasev1alpha1.AzureMssqlDatabaseSecondaryType]string{
	azuremssqldatabasev1alpha1.AzureMssqlDatabaseSecondaryType_GEO:   "Geo",
	azuremssqldatabasev1alpha1.AzureMssqlDatabaseSecondaryType_NAMED: "Named",
}

// licenseTypeStrings maps the Azure Hybrid Benefit enum to ARM's values.
var licenseTypeStrings = map[azuremssqldatabasev1alpha1.AzureMssqlDatabaseLicenseType]string{
	azuremssqldatabasev1alpha1.AzureMssqlDatabaseLicenseType_BASE_PRICE:       "BasePrice",
	azuremssqldatabasev1alpha1.AzureMssqlDatabaseLicenseType_LICENSE_INCLUDED: "LicenseIncluded",
}

// enclaveTypeStrings maps the confidential-computing enclave enum to ARM's
// values.
var enclaveTypeStrings = map[azuremssqldatabasev1alpha1.AzureMssqlDatabaseEnclaveType]string{
	azuremssqldatabasev1alpha1.AzureMssqlDatabaseEnclaveType_VBS:             "VBS",
	azuremssqldatabasev1alpha1.AzureMssqlDatabaseEnclaveType_DEFAULT_ENCLAVE: "Default",
}

// storageAccountTypeStrings maps the backup-redundancy enum to ARM's
// values.
var storageAccountTypeStrings = map[azuremssqldatabasev1alpha1.AzureMssqlDatabaseBackupStorageAccountType]string{
	azuremssqldatabasev1alpha1.AzureMssqlDatabaseBackupStorageAccountType_GEO_REDUNDANT:      "Geo",
	azuremssqldatabasev1alpha1.AzureMssqlDatabaseBackupStorageAccountType_GEO_ZONE_REDUNDANT: "GeoZone",
	azuremssqldatabasev1alpha1.AzureMssqlDatabaseBackupStorageAccountType_LOCALLY_REDUNDANT:  "Local",
	azuremssqldatabasev1alpha1.AzureMssqlDatabaseBackupStorageAccountType_ZONE_REDUNDANT:     "Zone",
}

// importStorageKeyTypeStrings maps the bacpac credential-kind enum to
// ARM's values.
var importStorageKeyTypeStrings = map[azuremssqldatabasev1alpha1.AzureMssqlDatabaseImportStorageKeyType]string{
	azuremssqldatabasev1alpha1.AzureMssqlDatabaseImportStorageKeyType_SHARED_ACCESS_KEY:  "SharedAccessKey",
	azuremssqldatabasev1alpha1.AzureMssqlDatabaseImportStorageKeyType_STORAGE_ACCESS_KEY: "StorageAccessKey",
}

// importAuthenticationTypeStrings maps the bacpac auth enum to ARM's
// values.
var importAuthenticationTypeStrings = map[azuremssqldatabasev1alpha1.AzureMssqlDatabaseImportAuthenticationType]string{
	azuremssqldatabasev1alpha1.AzureMssqlDatabaseImportAuthenticationType_SQL:         "Sql",
	azuremssqldatabasev1alpha1.AzureMssqlDatabaseImportAuthenticationType_AD_PASSWORD: "ADPassword",
}

// threatDetectionStateStrings maps the Defender policy-state enum to ARM's
// values.
var threatDetectionStateStrings = map[azuremssqldatabasev1alpha1.AzureMssqlDatabaseThreatDetectionState]string{
	azuremssqldatabasev1alpha1.AzureMssqlDatabaseThreatDetectionState_ENABLED:  "Enabled",
	azuremssqldatabasev1alpha1.AzureMssqlDatabaseThreatDetectionState_DISABLED: "Disabled",
}

func initializeLocals(ctx *pulumi.Context, stackInput *azuremssqldatabasev1alpha1.AzureMssqlDatabaseStackInput) *Locals {
	locals := &Locals{}

	locals.AzureMssqlDatabase = stackInput.Target
	target := stackInput.Target

	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzureMssqlDatabase.String()),
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

	// The user's spec tags merge over the metadata-derived tags -- user
	// tags deliberately win so an org's governance conventions can
	// override the derived values where they collide.
	for key, value := range target.Spec.Tags {
		locals.AzureTags[key] = value
	}

	return locals
}
