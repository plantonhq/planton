package module

import (
	"strings"

	"github.com/pkg/errors"
	ocidbsystemv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/oci/ocidbsystem/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/database"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func dbSystem(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) error {
	spec := locals.OciDbSystem.Spec

	args := &database.DbSystemArgs{
		AvailabilityDomain: pulumi.String(spec.AvailabilityDomain),
		CompartmentId:      pulumi.String(spec.CompartmentId.GetValue()),
		DbHome:             buildDbHome(spec.DbHome),
		Hostname:           pulumi.String(spec.Hostname),
		Shape:              pulumi.String(spec.Shape),
		SshPublicKeys:      pulumi.ToStringArray(spec.SshPublicKeys),
		SubnetId:           pulumi.String(spec.SubnetId.GetValue()),
		DisplayName:        pulumi.StringPtr(locals.DisplayName),
		FreeformTags:       pulumi.ToStringMap(locals.FreeformTags),
	}

	if spec.CpuCoreCount > 0 {
		args.CpuCoreCount = pulumi.IntPtr(int(spec.CpuCoreCount))
	}

	if spec.DatabaseEdition != ocidbsystemv1.OciDbSystemSpec_database_edition_unspecified {
		args.DatabaseEdition = pulumi.StringPtr(strings.ToUpper(spec.DatabaseEdition.String()))
	}

	if spec.LicenseModel != ocidbsystemv1.OciDbSystemSpec_license_model_unspecified {
		args.LicenseModel = pulumi.StringPtr(strings.ToUpper(spec.LicenseModel.String()))
	}

	if spec.DataStorageSizeInGb > 0 {
		args.DataStorageSizeInGb = pulumi.IntPtr(int(spec.DataStorageSizeInGb))
	}

	if spec.DataStoragePercentage > 0 {
		args.DataStoragePercentage = pulumi.IntPtr(int(spec.DataStoragePercentage))
	}

	if spec.DiskRedundancy != ocidbsystemv1.OciDbSystemSpec_disk_redundancy_unspecified {
		args.DiskRedundancy = pulumi.StringPtr(strings.ToUpper(spec.DiskRedundancy.String()))
	}

	if spec.NodeCount > 0 {
		args.NodeCount = pulumi.IntPtr(int(spec.NodeCount))
	}

	if spec.Domain != "" {
		args.Domain = pulumi.StringPtr(spec.Domain)
	}

	if spec.ClusterName != "" {
		args.ClusterName = pulumi.StringPtr(spec.ClusterName)
	}

	if len(spec.FaultDomains) > 0 {
		args.FaultDomains = pulumi.ToStringArray(spec.FaultDomains)
	}

	if len(spec.NsgIds) > 0 {
		nsgIds := make(pulumi.StringArray, len(spec.NsgIds))
		for i, nsg := range spec.NsgIds {
			nsgIds[i] = pulumi.String(nsg.GetValue())
		}
		args.NsgIds = nsgIds
	}

	if spec.BackupSubnetId != nil && spec.BackupSubnetId.GetValue() != "" {
		args.BackupSubnetId = pulumi.StringPtr(spec.BackupSubnetId.GetValue())
	}

	if len(spec.BackupNetworkNsgIds) > 0 {
		backupNsgIds := make(pulumi.StringArray, len(spec.BackupNetworkNsgIds))
		for i, nsg := range spec.BackupNetworkNsgIds {
			backupNsgIds[i] = pulumi.String(nsg.GetValue())
		}
		args.BackupNetworkNsgIds = backupNsgIds
	}

	if spec.KmsKeyId != nil && spec.KmsKeyId.GetValue() != "" {
		args.KmsKeyId = pulumi.StringPtr(spec.KmsKeyId.GetValue())
	}

	if spec.KmsKeyVersionId != "" {
		args.KmsKeyVersionId = pulumi.StringPtr(spec.KmsKeyVersionId)
	}

	if spec.TimeZone != "" {
		args.TimeZone = pulumi.StringPtr(spec.TimeZone)
	}

	if spec.SparseDiskgroup != nil {
		args.SparseDiskgroup = pulumi.BoolPtr(*spec.SparseDiskgroup)
	}

	if spec.StorageVolumePerformanceMode != ocidbsystemv1.OciDbSystemSpec_storage_volume_performance_mode_unspecified {
		args.StorageVolumePerformanceMode = pulumi.StringPtr(strings.ToUpper(spec.StorageVolumePerformanceMode.String()))
	}

	if spec.PrivateIp != "" {
		args.PrivateIp = pulumi.StringPtr(spec.PrivateIp)
	}

	if spec.DataCollectionOptions != nil {
		args.DataCollectionOptions = buildDataCollectionOptions(spec.DataCollectionOptions)
	}

	if spec.DbSystemOptions != nil {
		args.DbSystemOptions = buildDbSystemOptions(spec.DbSystemOptions)
	}

	if spec.MaintenanceWindowDetails != nil {
		args.MaintenanceWindowDetails = buildMaintenanceWindowDetails(spec.MaintenanceWindowDetails)
	}

	createdDbSystem, err := database.NewDbSystem(ctx, locals.DisplayName, args, pulumiOciOpt(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create oci database db system")
	}

	ctx.Export(OpDbSystemId, createdDbSystem.ID())
	ctx.Export(OpListenerPort, createdDbSystem.ListenerPort)

	ctx.Export(OpDbHomeId, createdDbSystem.DbHome.ApplyT(func(dbHome database.DbSystemDbHome) string {
		if dbHome.Id != nil {
			return *dbHome.Id
		}
		return ""
	}).(pulumi.StringOutput))

	ctx.Export(OpDatabaseId, createdDbSystem.DbHome.ApplyT(func(dbHome database.DbSystemDbHome) string {
		if dbHome.Database.Id != nil {
			return *dbHome.Database.Id
		}
		return ""
	}).(pulumi.StringOutput))

	return nil
}

func buildDbHome(dbHome *ocidbsystemv1.OciDbSystemSpec_DbHome) database.DbSystemDbHomeArgs {
	args := database.DbSystemDbHomeArgs{
		Database: buildDatabase(dbHome.Database),
	}

	if dbHome.DbVersion != "" {
		args.DbVersion = pulumi.StringPtr(dbHome.DbVersion)
	}

	if dbHome.DisplayName != "" {
		args.DisplayName = pulumi.StringPtr(dbHome.DisplayName)
	}

	if dbHome.DatabaseSoftwareImageId != nil && dbHome.DatabaseSoftwareImageId.GetValue() != "" {
		args.DatabaseSoftwareImageId = pulumi.StringPtr(dbHome.DatabaseSoftwareImageId.GetValue())
	}

	return args
}

func buildDatabase(db *ocidbsystemv1.OciDbSystemSpec_Database) database.DbSystemDbHomeDatabaseArgs {
	args := database.DbSystemDbHomeDatabaseArgs{
		AdminPassword: pulumi.String(db.AdminPassword),
		DbName:        pulumi.StringPtr(db.DbName),
	}

	if db.CharacterSet != "" {
		args.CharacterSet = pulumi.StringPtr(db.CharacterSet)
	}

	if db.NcharacterSet != "" {
		args.NcharacterSet = pulumi.StringPtr(db.NcharacterSet)
	}

	if db.PdbName != "" {
		args.PdbName = pulumi.StringPtr(db.PdbName)
	}

	if db.DbDomain != "" {
		args.DbDomain = pulumi.StringPtr(db.DbDomain)
	}

	if db.KmsKeyId != nil && db.KmsKeyId.GetValue() != "" {
		args.KmsKeyId = pulumi.StringPtr(db.KmsKeyId.GetValue())
	}

	if db.KmsKeyVersionId != "" {
		args.KmsKeyVersionId = pulumi.StringPtr(db.KmsKeyVersionId)
	}

	if db.VaultId != nil && db.VaultId.GetValue() != "" {
		args.VaultId = pulumi.StringPtr(db.VaultId.GetValue())
	}

	if db.DbBackupConfig != nil {
		args.DbBackupConfig = buildDbBackupConfig(db.DbBackupConfig)
	}

	return args
}

func buildDbBackupConfig(cfg *ocidbsystemv1.OciDbSystemSpec_DbBackupConfig) database.DbSystemDbHomeDatabaseDbBackupConfigPtrInput {
	args := &database.DbSystemDbHomeDatabaseDbBackupConfigArgs{}

	if cfg.AutoBackupEnabled != nil {
		args.AutoBackupEnabled = pulumi.BoolPtr(*cfg.AutoBackupEnabled)
	}

	if cfg.AutoBackupWindow != "" {
		args.AutoBackupWindow = pulumi.StringPtr(cfg.AutoBackupWindow)
	}

	if cfg.RecoveryWindowInDays > 0 {
		args.RecoveryWindowInDays = pulumi.IntPtr(int(cfg.RecoveryWindowInDays))
	}

	return args
}

func buildDataCollectionOptions(opts *ocidbsystemv1.OciDbSystemSpec_DataCollectionOptions) database.DbSystemDataCollectionOptionsPtrInput {
	args := &database.DbSystemDataCollectionOptionsArgs{}

	if opts.IsDiagnosticsEventsEnabled != nil {
		args.IsDiagnosticsEventsEnabled = pulumi.BoolPtr(*opts.IsDiagnosticsEventsEnabled)
	}

	if opts.IsHealthMonitoringEnabled != nil {
		args.IsHealthMonitoringEnabled = pulumi.BoolPtr(*opts.IsHealthMonitoringEnabled)
	}

	if opts.IsIncidentLogsEnabled != nil {
		args.IsIncidentLogsEnabled = pulumi.BoolPtr(*opts.IsIncidentLogsEnabled)
	}

	return args
}

func buildDbSystemOptions(opts *ocidbsystemv1.OciDbSystemSpec_DbSystemOptions) database.DbSystemDbSystemOptionsPtrInput {
	args := &database.DbSystemDbSystemOptionsArgs{}

	if opts.StorageManagement != ocidbsystemv1.OciDbSystemSpec_storage_management_unspecified {
		args.StorageManagement = pulumi.StringPtr(strings.ToUpper(opts.StorageManagement.String()))
	}

	return args
}

func buildMaintenanceWindowDetails(mw *ocidbsystemv1.OciDbSystemSpec_MaintenanceWindowDetails) database.DbSystemMaintenanceWindowDetailsPtrInput {
	args := &database.DbSystemMaintenanceWindowDetailsArgs{}

	if mw.Preference != ocidbsystemv1.OciDbSystemSpec_preference_unspecified {
		args.Preference = pulumi.StringPtr(strings.ToUpper(mw.Preference.String()))
	}

	if mw.PatchingMode != ocidbsystemv1.OciDbSystemSpec_patching_mode_unspecified {
		args.PatchingMode = pulumi.StringPtr(strings.ToUpper(mw.PatchingMode.String()))
	}

	if mw.LeadTimeInWeeks > 0 {
		args.LeadTimeInWeeks = pulumi.IntPtr(int(mw.LeadTimeInWeeks))
	}

	if len(mw.Months) > 0 {
		months := make(database.DbSystemMaintenanceWindowDetailsMonthArray, len(mw.Months))
		for i, m := range mw.Months {
			months[i] = &database.DbSystemMaintenanceWindowDetailsMonthArgs{
				Name: pulumi.String(m),
			}
		}
		args.Months = months
	}

	if len(mw.WeeksOfMonth) > 0 {
		weeks := make(pulumi.IntArray, len(mw.WeeksOfMonth))
		for i, w := range mw.WeeksOfMonth {
			weeks[i] = pulumi.Int(int(w))
		}
		args.WeeksOfMonths = weeks
	}

	if len(mw.DaysOfWeek) > 0 {
		days := make(database.DbSystemMaintenanceWindowDetailsDaysOfWeekArray, len(mw.DaysOfWeek))
		for i, d := range mw.DaysOfWeek {
			days[i] = &database.DbSystemMaintenanceWindowDetailsDaysOfWeekArgs{
				Name: pulumi.String(d),
			}
		}
		args.DaysOfWeeks = days
	}

	if len(mw.HoursOfDay) > 0 {
		hours := make(pulumi.IntArray, len(mw.HoursOfDay))
		for i, h := range mw.HoursOfDay {
			hours[i] = pulumi.Int(int(h))
		}
		args.HoursOfDays = hours
	}

	if mw.CustomActionTimeoutInMins > 0 {
		args.CustomActionTimeoutInMins = pulumi.IntPtr(int(mw.CustomActionTimeoutInMins))
	}

	if mw.IsCustomActionTimeoutEnabled != nil {
		args.IsCustomActionTimeoutEnabled = pulumi.BoolPtr(*mw.IsCustomActionTimeoutEnabled)
	}

	if mw.IsMonthlyPatchingEnabled != nil {
		args.IsMonthlyPatchingEnabled = pulumi.BoolPtr(*mw.IsMonthlyPatchingEnabled)
	}

	return args
}
