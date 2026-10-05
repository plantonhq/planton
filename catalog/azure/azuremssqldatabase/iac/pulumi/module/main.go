package module

import (
	"github.com/pkg/errors"
	azuremssqldatabasev1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremssqldatabase/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/mssql"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azuremssqldatabasev1alpha1.AzureMssqlDatabaseIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureMssqlDatabase.Spec

	databaseArgs := &mssql.DatabaseArgs{
		Name:     pulumi.String(spec.DatabaseName),
		ServerId: pulumi.String(spec.ServerId.GetValue()),
		Tags:     pulumi.ToStringMap(locals.AzureTags),
	}

	// The SKU (or "ElasticPool" when pooled). Unset is not sent, letting
	// Azure compute its default (serverless GP_S_Gen5_2) -- mirroring the
	// Terraform module's null.
	if spec.SkuName != "" {
		databaseArgs.SkuName = pulumi.String(spec.SkuName)
	}
	if spec.ElasticPoolId.GetValue() != "" {
		databaseArgs.ElasticPoolId = pulumi.String(spec.ElasticPoolId.GetValue())
	}

	// Fractional sizes are legal ARM values (Basic tops out at 2 GB, S0
	// at 250 GB); Hyperscale grows elastically and ignores the ceiling.
	if spec.MaxSizeGb != nil {
		databaseArgs.MaxSizeGb = pulumi.Float64(spec.GetMaxSizeGb())
	}

	// Collation is presence-guarded to the spec default -- IaC inputs
	// built from a manifest do NOT materialize proto defaults.
	if spec.Collation != nil {
		databaseArgs.Collation = pulumi.String(spec.GetCollation())
	} else {
		databaseArgs.Collation = pulumi.String("SQL_Latin1_General_CP1_CI_AS")
	}

	if spec.LicenseType != azuremssqldatabasev1alpha1.AzureMssqlDatabaseLicenseType_azure_mssql_database_license_type_unspecified {
		databaseArgs.LicenseType = pulumi.String(licenseTypeStrings[spec.LicenseType])
	}

	// Serverless dials -- spec validation gates them to GP_S_/HS_S_ skus
	// before the deploy ever runs.
	if spec.AutoPauseDelayInMinutes != nil {
		databaseArgs.AutoPauseDelayInMinutes = pulumi.Int(int(spec.GetAutoPauseDelayInMinutes()))
	}
	if spec.MinCapacity != nil {
		databaseArgs.MinCapacity = pulumi.Float64(spec.GetMinCapacity())
	}

	// Availability: Hyperscale readable replicas, Premium/BC read
	// scale-out, and zone spreading.
	if spec.ReadReplicaCount != nil {
		databaseArgs.ReadReplicaCount = pulumi.Int(int(spec.GetReadReplicaCount()))
	}
	databaseArgs.ReadScale = pulumi.Bool(spec.ReadScale)
	databaseArgs.ZoneRedundant = pulumi.Bool(spec.ZoneRedundant)

	// Integrity and confidential computing. Changing enclave_type
	// replaces the database (ARM's contract).
	databaseArgs.LedgerEnabled = pulumi.Bool(spec.LedgerEnabled)
	if spec.EnclaveType != azuremssqldatabasev1alpha1.AzureMssqlDatabaseEnclaveType_azure_mssql_database_enclave_type_unspecified {
		databaseArgs.EnclaveType = pulumi.String(enclaveTypeStrings[spec.EnclaveType])
	}

	// Pooled databases inherit the pool's window -- spec validation
	// forces this empty when elastic_pool_id is set.
	if spec.MaintenanceConfigurationName != "" {
		databaseArgs.MaintenanceConfigurationName = pulumi.String(spec.MaintenanceConfigurationName)
	}

	// Lifecycle: how the database comes into existence. An unspecified
	// mode means a fresh (DEFAULT) database and is not sent at all. Each
	// mode consumes its matching source; the pairings are spec-validated.
	if spec.CreateMode != azuremssqldatabasev1alpha1.AzureMssqlDatabaseCreateMode_azure_mssql_database_create_mode_unspecified {
		databaseArgs.CreateMode = pulumi.String(createModeStrings[spec.CreateMode])
	}
	if spec.CreationSourceDatabaseId.GetValue() != "" {
		databaseArgs.CreationSourceDatabaseId = pulumi.String(spec.CreationSourceDatabaseId.GetValue())
	}
	if spec.SecondaryType != azuremssqldatabasev1alpha1.AzureMssqlDatabaseSecondaryType_azure_mssql_database_secondary_type_unspecified {
		databaseArgs.SecondaryType = pulumi.String(secondaryTypeStrings[spec.SecondaryType])
	}
	if spec.RestorePointInTime != "" {
		databaseArgs.RestorePointInTime = pulumi.String(spec.RestorePointInTime)
	}
	if spec.RecoverDatabaseId != "" {
		databaseArgs.RecoverDatabaseId = pulumi.String(spec.RecoverDatabaseId)
	}
	if spec.RecoveryPointId != "" {
		databaseArgs.RecoveryPointId = pulumi.String(spec.RecoveryPointId)
	}
	if spec.RestoreDroppedDatabaseId != "" {
		databaseArgs.RestoreDroppedDatabaseId = pulumi.String(spec.RestoreDroppedDatabaseId)
	}
	if spec.RestoreLongTermRetentionBackupId != "" {
		databaseArgs.RestoreLongTermRetentionBackupId = pulumi.String(spec.RestoreLongTermRetentionBackupId)
	}

	// Backup redundancy and the DW-only geo-backup dial (presence-guarded
	// to Azure's default of true).
	if spec.StorageAccountType != azuremssqldatabasev1alpha1.AzureMssqlDatabaseBackupStorageAccountType_azure_mssql_database_backup_storage_account_type_unspecified {
		databaseArgs.StorageAccountType = pulumi.String(storageAccountTypeStrings[spec.StorageAccountType])
	}
	if spec.GeoBackupEnabled != nil {
		databaseArgs.GeoBackupEnabled = pulumi.Bool(spec.GetGeoBackupEnabled())
	} else {
		databaseArgs.GeoBackupEnabled = pulumi.Bool(true)
	}

	if spec.SampleName != "" {
		databaseArgs.SampleName = pulumi.String(spec.SampleName)
	}

	// Database-scoped identities for the database-scoped CMK.
	if len(spec.UserAssignedIdentityIds) > 0 {
		identityIds := make([]string, 0, len(spec.UserAssignedIdentityIds))
		for _, identityId := range spec.UserAssignedIdentityIds {
			identityIds = append(identityIds, identityId.GetValue())
		}
		databaseArgs.Identity = &mssql.DatabaseIdentityArgs{
			Type:        pulumi.String("UserAssigned"),
			IdentityIds: pulumi.ToStringArray(identityIds),
		}
	}

	// Transparent data encryption: the database-scoped CMK overrides the
	// server's key for this database; rotation re-encrypts automatically
	// when enabled. The on/off dial is presence-guarded to Azure's
	// default of true.
	if spec.TransparentDataEncryptionEnabled != nil {
		databaseArgs.TransparentDataEncryptionEnabled = pulumi.Bool(spec.GetTransparentDataEncryptionEnabled())
	} else {
		databaseArgs.TransparentDataEncryptionEnabled = pulumi.Bool(true)
	}
	// The rotation flag rides WITH the key: the provider requires the two
	// set together (sending the flag alone -- even false -- is rejected),
	// so it is only forwarded when a CMK key exists.
	if spec.TransparentDataEncryptionKeyVaultKeyId.GetValue() != "" {
		databaseArgs.TransparentDataEncryptionKeyVaultKeyId = pulumi.String(spec.TransparentDataEncryptionKeyVaultKeyId.GetValue())
		databaseArgs.TransparentDataEncryptionKeyAutomaticRotationEnabled = pulumi.Bool(spec.TransparentDataEncryptionKeyAutomaticRotationEnabled)
	}

	// A bacpac import applied right after creation (fresh databases only
	// -- spec-validated).
	if spec.Import != nil {
		importArgs := &mssql.DatabaseImportArgs{
			StorageUri:                 pulumi.String(spec.Import.StorageUri),
			StorageKey:                 pulumi.String(spec.Import.StorageKey.GetValue()),
			StorageKeyType:             pulumi.String(importStorageKeyTypeStrings[spec.Import.StorageKeyType]),
			AdministratorLogin:         pulumi.String(spec.Import.AdministratorLogin),
			AdministratorLoginPassword: pulumi.String(spec.Import.AdministratorLoginPassword.GetValue()),
			AuthenticationType:         pulumi.String(importAuthenticationTypeStrings[spec.Import.AuthenticationType]),
		}
		if spec.Import.StorageAccountId != "" {
			importArgs.StorageAccountId = pulumi.String(spec.Import.StorageAccountId)
		}
		databaseArgs.Import = importArgs
	}

	// The point-in-time-restore horizon and differential-backup cadence.
	if spec.ShortTermRetentionPolicy != nil {
		strArgs := &mssql.DatabaseShortTermRetentionPolicyArgs{
			RetentionDays: pulumi.Int(int(spec.ShortTermRetentionPolicy.RetentionDays)),
		}
		if spec.ShortTermRetentionPolicy.BackupIntervalInHours != nil {
			strArgs.BackupIntervalInHours = pulumi.Int(int(spec.ShortTermRetentionPolicy.GetBackupIntervalInHours()))
		}
		databaseArgs.ShortTermRetentionPolicy = strArgs
	}

	// Long-term full-backup retention. ARM's "PT0S" means a horizon keeps
	// nothing, so unset horizons are simply not sent.
	if spec.LongTermRetentionPolicy != nil {
		ltr := spec.LongTermRetentionPolicy
		ltrArgs := &mssql.DatabaseLongTermRetentionPolicyArgs{}
		if ltr.WeeklyRetention != "" {
			ltrArgs.WeeklyRetention = pulumi.String(ltr.WeeklyRetention)
		}
		if ltr.MonthlyRetention != "" {
			ltrArgs.MonthlyRetention = pulumi.String(ltr.MonthlyRetention)
		}
		if ltr.YearlyRetention != "" {
			ltrArgs.YearlyRetention = pulumi.String(ltr.YearlyRetention)
		}
		if ltr.WeekOfYear != nil {
			ltrArgs.WeekOfYear = pulumi.Int(int(ltr.GetWeekOfYear()))
		}
		databaseArgs.LongTermRetentionPolicy = ltrArgs
	}

	// Database-scoped Microsoft Defender threat detection (overrides the
	// server-scope policy for this database). ARM models
	// email_account_admins as an Enabled/Disabled string here -- the
	// bool maps to that wire vocabulary.
	if spec.ThreatDetectionPolicy != nil {
		policy := spec.ThreatDetectionPolicy
		emailAdmins := "Disabled"
		if policy.EmailAccountAdmins {
			emailAdmins = "Enabled"
		}
		tdArgs := &mssql.DatabaseThreatDetectionPolicyArgs{
			State:              pulumi.String(threatDetectionStateStrings[policy.State]),
			EmailAccountAdmins: pulumi.String(emailAdmins),
		}
		if len(policy.DisabledAlerts) > 0 {
			tdArgs.DisabledAlerts = pulumi.ToStringArray(policy.DisabledAlerts)
		}
		if len(policy.EmailAddresses) > 0 {
			tdArgs.EmailAddresses = pulumi.ToStringArray(policy.EmailAddresses)
		}
		if policy.RetentionDays != nil {
			tdArgs.RetentionDays = pulumi.Int(int(policy.GetRetentionDays()))
		}
		if policy.StorageEndpoint != "" {
			tdArgs.StorageEndpoint = pulumi.String(policy.StorageEndpoint)
		}
		if policy.StorageAccountAccessKey.GetValue() != "" {
			tdArgs.StorageAccountAccessKey = pulumi.String(policy.StorageAccountAccessKey.GetValue())
		}
		databaseArgs.ThreatDetectionPolicy = tdArgs
	}

	database, err := mssql.NewDatabase(ctx,
		spec.DatabaseName,
		databaseArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create mssql database %s", spec.DatabaseName)
	}

	// Export outputs from the created resource.
	ctx.Export(OpDatabaseId, database.ID())
	ctx.Export(OpDatabaseName, database.Name)

	return nil
}
