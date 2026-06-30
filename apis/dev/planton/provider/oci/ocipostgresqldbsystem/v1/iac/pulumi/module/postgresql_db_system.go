package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	ocipostgresqldbsystemv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ocipostgresqldbsystem/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/psql"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func postgresqlDbSystem(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) error {
	spec := locals.OciPostgresqlDbSystem.Spec

	args := &psql.DbSystemArgs{
		CompartmentId:  pulumi.String(spec.CompartmentId.GetValue()),
		DbVersion:      pulumi.String(spec.DbVersion),
		DisplayName:    pulumi.String(locals.DisplayName),
		Shape:          pulumi.String(spec.Shape),
		NetworkDetails: buildNetworkDetails(spec.NetworkDetails),
		StorageDetails: buildStorageDetails(spec.StorageDetails),
		FreeformTags:   pulumi.ToStringMap(locals.FreeformTags),
	}

	if spec.InstanceOcpuCount > 0 {
		args.InstanceOcpuCount = pulumi.IntPtr(int(spec.InstanceOcpuCount))
	}

	if spec.InstanceMemorySizeInGbs > 0 {
		args.InstanceMemorySizeInGbs = pulumi.IntPtr(int(spec.InstanceMemorySizeInGbs))
	}

	if spec.InstanceCount > 0 {
		args.InstanceCount = pulumi.IntPtr(int(spec.InstanceCount))
	}

	if spec.Credentials != nil {
		args.Credentials = buildCredentials(spec.Credentials)
	}

	if spec.ManagementPolicy != nil {
		args.ManagementPolicy = buildManagementPolicy(spec.ManagementPolicy)
	}

	if spec.ConfigId != nil {
		args.ConfigId = pulumi.StringPtr(spec.ConfigId.GetValue())
	}

	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}

	if len(spec.InstancesDetails) > 0 {
		args.InstancesDetails = buildInstancesDetails(spec.InstancesDetails)
	}

	createdDb, err := psql.NewDbSystem(ctx, locals.DisplayName, args, pulumiOciOpt(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create oci postgresql db system")
	}

	ctx.Export(OpDbSystemId, createdDb.ID())
	ctx.Export(OpAdminUsername, createdDb.AdminUsername)

	ctx.Export(OpPrimaryDbEndpointPrivateIp, createdDb.NetworkDetails.PrimaryDbEndpointPrivateIp().ApplyT(func(ip *string) string {
		if ip != nil {
			return *ip
		}
		return ""
	}).(pulumi.StringOutput))

	return nil
}

func buildNetworkDetails(nd *ocipostgresqldbsystemv1.OciPostgresqlDbSystemSpec_NetworkDetails) psql.DbSystemNetworkDetailsInput {
	network := &psql.DbSystemNetworkDetailsArgs{
		SubnetId: pulumi.String(nd.SubnetId.GetValue()),
	}

	if len(nd.NsgIds) > 0 {
		nsgIds := make(pulumi.StringArray, len(nd.NsgIds))
		for i, nsg := range nd.NsgIds {
			nsgIds[i] = pulumi.String(nsg.GetValue())
		}
		network.NsgIds = nsgIds
	}

	if nd.IsReaderEndpointEnabled != nil {
		network.IsReaderEndpointEnabled = pulumi.BoolPtr(*nd.IsReaderEndpointEnabled)
	}

	if nd.PrimaryDbEndpointPrivateIp != "" {
		network.PrimaryDbEndpointPrivateIp = pulumi.StringPtr(nd.PrimaryDbEndpointPrivateIp)
	}

	return network
}

func buildStorageDetails(sd *ocipostgresqldbsystemv1.OciPostgresqlDbSystemSpec_StorageDetails) psql.DbSystemStorageDetailsInput {
	storage := &psql.DbSystemStorageDetailsArgs{
		IsRegionallyDurable: pulumi.Bool(sd.IsRegionallyDurable),
		SystemType:          pulumi.String("OCI_OPTIMIZED_STORAGE"),
	}

	if sd.AvailabilityDomain != "" {
		storage.AvailabilityDomain = pulumi.StringPtr(sd.AvailabilityDomain)
	}

	if sd.Iops > 0 {
		storage.Iops = pulumi.StringPtr(fmt.Sprintf("%d", sd.Iops))
	}

	return storage
}

func buildCredentials(creds *ocipostgresqldbsystemv1.OciPostgresqlDbSystemSpec_Credentials) psql.DbSystemCredentialsPtrInput {
	pd := creds.PasswordDetails
	pwdDetails := &psql.DbSystemCredentialsPasswordDetailsArgs{}

	if pd.PasswordType != ocipostgresqldbsystemv1.OciPostgresqlDbSystemSpec_password_type_unspecified {
		pwdDetails.PasswordType = pulumi.String(strings.ToUpper(pd.PasswordType.String()))
	}

	if pd.Password != "" {
		pwdDetails.Password = pulumi.StringPtr(pd.Password)
	}

	if pd.SecretId != nil {
		pwdDetails.SecretId = pulumi.StringPtr(pd.SecretId.GetValue())
	}

	if pd.SecretVersion != "" {
		pwdDetails.SecretVersion = pulumi.StringPtr(pd.SecretVersion)
	}

	return &psql.DbSystemCredentialsArgs{
		Username:        pulumi.String(creds.Username),
		PasswordDetails: pwdDetails,
	}
}

func buildManagementPolicy(mp *ocipostgresqldbsystemv1.OciPostgresqlDbSystemSpec_ManagementPolicy) psql.DbSystemManagementPolicyPtrInput {
	policy := &psql.DbSystemManagementPolicyArgs{}

	if mp.BackupPolicy != nil {
		policy.BackupPolicy = buildBackupPolicy(mp.BackupPolicy)
	}

	if mp.MaintenanceWindowStart != "" {
		policy.MaintenanceWindowStart = pulumi.StringPtr(mp.MaintenanceWindowStart)
	}

	return policy
}

func buildBackupPolicy(bp *ocipostgresqldbsystemv1.OciPostgresqlDbSystemSpec_BackupPolicy) psql.DbSystemManagementPolicyBackupPolicyPtrInput {
	backup := &psql.DbSystemManagementPolicyBackupPolicyArgs{}

	if bp.Kind != ocipostgresqldbsystemv1.OciPostgresqlDbSystemSpec_backup_kind_unspecified {
		backup.Kind = pulumi.StringPtr(strings.ToUpper(bp.Kind.String()))
	}

	if bp.BackupStart != "" {
		backup.BackupStart = pulumi.StringPtr(bp.BackupStart)
	}

	if bp.RetentionDays > 0 {
		backup.RetentionDays = pulumi.IntPtr(int(bp.RetentionDays))
	}

	if len(bp.DaysOfTheMonth) > 0 {
		days := make(pulumi.IntArray, len(bp.DaysOfTheMonth))
		for i, d := range bp.DaysOfTheMonth {
			days[i] = pulumi.Int(int(d))
		}
		backup.DaysOfTheMonths = days
	}

	if len(bp.DaysOfTheWeek) > 0 {
		backup.DaysOfTheWeeks = pulumi.ToStringArray(bp.DaysOfTheWeek)
	}

	return backup
}

func buildInstancesDetails(details []*ocipostgresqldbsystemv1.OciPostgresqlDbSystemSpec_InstanceDetails) psql.DbSystemInstancesDetailArrayInput {
	instances := make(psql.DbSystemInstancesDetailArray, len(details))
	for i, d := range details {
		instance := psql.DbSystemInstancesDetailArgs{}

		if d.DisplayName != "" {
			instance.DisplayName = pulumi.StringPtr(d.DisplayName)
		}

		if d.Description != "" {
			instance.Description = pulumi.StringPtr(d.Description)
		}

		if d.PrivateIp != "" {
			instance.PrivateIp = pulumi.StringPtr(d.PrivateIp)
		}

		instances[i] = instance
	}
	return instances
}
