package module

import (
	"strings"

	"github.com/pkg/errors"
	ociautonomousdatabasev1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ociautonomousdatabase/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/database"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func autonomousDatabase(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) error {
	spec := locals.OciAutonomousDatabase.Spec

	args := &database.AutonomousDatabaseArgs{
		CompartmentId: pulumi.String(spec.CompartmentId.GetValue()),
		DbName:        pulumi.String(spec.DbName),
		DisplayName:   pulumi.StringPtr(locals.DisplayName),
		FreeformTags:  pulumi.ToStringMap(locals.FreeformTags),
	}

	if spec.DbWorkload != ociautonomousdatabasev1.OciAutonomousDatabaseSpec_db_workload_unspecified {
		args.DbWorkload = pulumi.StringPtr(strings.ToUpper(spec.DbWorkload.String()))
	}

	if spec.DbVersion != "" {
		args.DbVersion = pulumi.StringPtr(spec.DbVersion)
	}

	if spec.DatabaseEdition != ociautonomousdatabasev1.OciAutonomousDatabaseSpec_database_edition_unspecified {
		args.DatabaseEdition = pulumi.StringPtr(strings.ToUpper(spec.DatabaseEdition.String()))
	}

	if spec.LicenseModel != ociautonomousdatabasev1.OciAutonomousDatabaseSpec_license_model_unspecified {
		args.LicenseModel = pulumi.StringPtr(strings.ToUpper(spec.LicenseModel.String()))
	}

	if spec.CharacterSet != "" {
		args.CharacterSet = pulumi.StringPtr(spec.CharacterSet)
	}

	if spec.NcharacterSet != "" {
		args.NcharacterSet = pulumi.StringPtr(spec.NcharacterSet)
	}

	if spec.ComputeModel != ociautonomousdatabasev1.OciAutonomousDatabaseSpec_compute_model_unspecified {
		args.ComputeModel = pulumi.StringPtr(strings.ToUpper(spec.ComputeModel.String()))
	}

	if spec.ComputeCount != nil {
		args.ComputeCount = pulumi.Float64Ptr(float64(*spec.ComputeCount))
	}

	if spec.DataStorageSizeInTbs > 0 {
		args.DataStorageSizeInTbs = pulumi.IntPtr(int(spec.DataStorageSizeInTbs))
	}

	if spec.DataStorageSizeInGb > 0 {
		args.DataStorageSizeInGb = pulumi.IntPtr(int(spec.DataStorageSizeInGb))
	}

	if spec.IsAutoScalingEnabled != nil {
		args.IsAutoScalingEnabled = pulumi.BoolPtr(*spec.IsAutoScalingEnabled)
	}

	if spec.IsAutoScalingForStorageEnabled != nil {
		args.IsAutoScalingForStorageEnabled = pulumi.BoolPtr(*spec.IsAutoScalingForStorageEnabled)
	}

	if spec.AdminPassword != "" {
		args.AdminPassword = pulumi.StringPtr(spec.AdminPassword)
	}

	if spec.SecretId != nil {
		args.SecretId = pulumi.StringPtr(spec.SecretId.GetValue())
	}

	if spec.SecretVersionNumber > 0 {
		args.SecretVersionNumber = pulumi.IntPtr(int(spec.SecretVersionNumber))
	}

	if spec.SubnetId != nil {
		args.SubnetId = pulumi.StringPtr(spec.SubnetId.GetValue())
	}

	if len(spec.NsgIds) > 0 {
		nsgIds := make(pulumi.StringArray, len(spec.NsgIds))
		for i, nsg := range spec.NsgIds {
			nsgIds[i] = pulumi.String(nsg.GetValue())
		}
		args.NsgIds = nsgIds
	}

	if spec.PrivateEndpointLabel != "" {
		args.PrivateEndpointLabel = pulumi.StringPtr(spec.PrivateEndpointLabel)
	}

	if spec.PrivateEndpointIp != "" {
		args.PrivateEndpointIp = pulumi.StringPtr(spec.PrivateEndpointIp)
	}

	if len(spec.WhitelistedIps) > 0 {
		args.WhitelistedIps = pulumi.ToStringArray(spec.WhitelistedIps)
	}

	if spec.IsMtlsConnectionRequired != nil {
		args.IsMtlsConnectionRequired = pulumi.BoolPtr(*spec.IsMtlsConnectionRequired)
	}

	if spec.IsAccessControlEnabled != nil {
		args.IsAccessControlEnabled = pulumi.BoolPtr(*spec.IsAccessControlEnabled)
	}

	if spec.KmsKeyId != nil {
		args.KmsKeyId = pulumi.StringPtr(spec.KmsKeyId.GetValue())
	}

	if spec.VaultId != nil {
		args.VaultId = pulumi.StringPtr(spec.VaultId.GetValue())
	}

	if spec.IsDedicated != nil {
		args.IsDedicated = pulumi.BoolPtr(*spec.IsDedicated)
	}

	if spec.IsFreeTier != nil {
		args.IsFreeTier = pulumi.BoolPtr(*spec.IsFreeTier)
	}

	if spec.IsDevTier != nil {
		args.IsDevTier = pulumi.BoolPtr(*spec.IsDevTier)
	}

	if spec.AutonomousContainerDatabaseId != nil {
		args.AutonomousContainerDatabaseId = pulumi.StringPtr(spec.AutonomousContainerDatabaseId.GetValue())
	}

	if spec.BackupRetentionPeriodInDays > 0 {
		args.BackupRetentionPeriodInDays = pulumi.IntPtr(int(spec.BackupRetentionPeriodInDays))
	}

	if spec.IsLocalDataGuardEnabled != nil {
		args.IsLocalDataGuardEnabled = pulumi.BoolPtr(*spec.IsLocalDataGuardEnabled)
	}

	if spec.AutonomousMaintenanceScheduleType != ociautonomousdatabasev1.OciAutonomousDatabaseSpec_maintenance_schedule_type_unspecified {
		args.AutonomousMaintenanceScheduleType = pulumi.StringPtr(strings.ToUpper(spec.AutonomousMaintenanceScheduleType.String()))
	}

	if len(spec.CustomerContacts) > 0 {
		contacts := make(database.AutonomousDatabaseCustomerContactArray, len(spec.CustomerContacts))
		for i, cc := range spec.CustomerContacts {
			contacts[i] = database.AutonomousDatabaseCustomerContactArgs{
				Email: pulumi.StringPtr(cc.Email),
			}
		}
		args.CustomerContacts = contacts
	}

	createdAdb, err := database.NewAutonomousDatabase(ctx, locals.DisplayName, args, pulumiOciOpt(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create oci autonomous database")
	}

	ctx.Export(OpAutonomousDatabaseId, createdAdb.ID())
	ctx.Export(OpServiceConsoleUrl, createdAdb.ServiceConsoleUrl)
	ctx.Export(OpPrivateEndpoint, createdAdb.PrivateEndpoint)

	ctx.Export(OpConnectionStringHigh, createdAdb.ConnectionStrings.ApplyT(func(cs []database.AutonomousDatabaseConnectionString) string {
		if len(cs) > 0 && cs[0].High != nil {
			return *cs[0].High
		}
		return ""
	}).(pulumi.StringOutput))

	ctx.Export(OpConnectionStringMedium, createdAdb.ConnectionStrings.ApplyT(func(cs []database.AutonomousDatabaseConnectionString) string {
		if len(cs) > 0 && cs[0].Medium != nil {
			return *cs[0].Medium
		}
		return ""
	}).(pulumi.StringOutput))

	ctx.Export(OpConnectionStringLow, createdAdb.ConnectionStrings.ApplyT(func(cs []database.AutonomousDatabaseConnectionString) string {
		if len(cs) > 0 && cs[0].Low != nil {
			return *cs[0].Low
		}
		return ""
	}).(pulumi.StringOutput))

	return nil
}
