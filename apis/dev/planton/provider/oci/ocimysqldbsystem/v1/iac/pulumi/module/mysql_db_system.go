package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	ocimysqldbsystemv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ocimysqldbsystem/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/mysql"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func mysqlDbSystem(ctx *pulumi.Context, locals *Locals, provider *oci.Provider) error {
	spec := locals.OciMysqlDbSystem.Spec

	args := &mysql.MysqlDbSystemArgs{
		CompartmentId:      pulumi.String(spec.CompartmentId.GetValue()),
		AvailabilityDomain: pulumi.String(spec.AvailabilityDomain),
		ShapeName:          pulumi.String(spec.ShapeName),
		SubnetId:           pulumi.String(spec.SubnetId.GetValue()),
		DisplayName:        pulumi.StringPtr(locals.DisplayName),
		FreeformTags:       pulumi.ToStringMap(locals.FreeformTags),
	}

	if spec.AdminUsername != "" {
		args.AdminUsername = pulumi.StringPtr(spec.AdminUsername)
	}

	if spec.AdminPassword != "" {
		args.AdminPassword = pulumi.StringPtr(spec.AdminPassword)
	}

	if spec.MysqlVersion != "" {
		args.MysqlVersion = pulumi.StringPtr(spec.MysqlVersion)
	}

	if spec.ConfigurationId != nil {
		args.ConfigurationId = pulumi.StringPtr(spec.ConfigurationId.GetValue())
	}

	if spec.IsHighlyAvailable != nil {
		args.IsHighlyAvailable = pulumi.BoolPtr(*spec.IsHighlyAvailable)
	}

	if spec.HostnameLabel != "" {
		args.HostnameLabel = pulumi.StringPtr(spec.HostnameLabel)
	}

	if spec.IpAddress != "" {
		args.IpAddress = pulumi.StringPtr(spec.IpAddress)
	}

	if spec.FaultDomain != "" {
		args.FaultDomain = pulumi.StringPtr(spec.FaultDomain)
	}

	if spec.Port > 0 {
		args.Port = pulumi.IntPtr(int(spec.Port))
	}

	if spec.PortX > 0 {
		args.PortX = pulumi.IntPtr(int(spec.PortX))
	}

	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}

	if spec.CrashRecovery != "" {
		args.CrashRecovery = pulumi.StringPtr(spec.CrashRecovery)
	}

	if spec.DatabaseManagement != "" {
		args.DatabaseManagement = pulumi.StringPtr(spec.DatabaseManagement)
	}

	if len(spec.NsgIds) > 0 {
		nsgIds := make(pulumi.StringArray, len(spec.NsgIds))
		for i, nsg := range spec.NsgIds {
			nsgIds[i] = pulumi.String(nsg.GetValue())
		}
		args.NsgIds = nsgIds
	}

	if spec.DataStorage != nil {
		args.DataStorage = buildDataStorage(spec.DataStorage)
	}

	if spec.BackupPolicy != nil {
		args.BackupPolicy = buildBackupPolicy(spec.BackupPolicy)
	}

	if spec.Maintenance != nil {
		args.Maintenance = buildMaintenance(spec.Maintenance)
	}

	if spec.DeletionPolicy != nil {
		args.DeletionPolicies = buildDeletionPolicy(spec.DeletionPolicy)
	}

	if spec.EncryptData != nil {
		args.EncryptData = buildEncryptData(spec.EncryptData)
	}

	if spec.SecureConnections != nil {
		args.SecureConnections = buildSecureConnections(spec.SecureConnections)
	}

	if len(spec.CustomerContacts) > 0 {
		contacts := make(mysql.MysqlDbSystemCustomerContactArray, len(spec.CustomerContacts))
		for i, cc := range spec.CustomerContacts {
			contacts[i] = mysql.MysqlDbSystemCustomerContactArgs{
				Email: pulumi.String(cc.Email),
			}
		}
		args.CustomerContacts = contacts
	}

	if spec.ReadEndpoint != nil {
		args.ReadEndpoint = buildReadEndpoint(spec.ReadEndpoint)
	}

	if spec.DatabaseConsole != nil {
		args.DatabaseConsole = buildDatabaseConsole(spec.DatabaseConsole)
	}

	if spec.Rest != nil {
		args.Rest = buildRest(spec.Rest)
	}

	createdDb, err := mysql.NewMysqlDbSystem(ctx, locals.DisplayName, args, pulumiOciOpt(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create oci mysql db system")
	}

	ctx.Export(OpDbSystemId, createdDb.ID())

	ctx.Export(OpEndpointHostname, createdDb.Endpoints.ApplyT(func(eps []mysql.MysqlDbSystemEndpoint) string {
		if len(eps) > 0 && eps[0].Hostname != nil {
			return *eps[0].Hostname
		}
		return ""
	}).(pulumi.StringOutput))

	ctx.Export(OpEndpointIpAddress, createdDb.Endpoints.ApplyT(func(eps []mysql.MysqlDbSystemEndpoint) string {
		if len(eps) > 0 && eps[0].IpAddress != nil {
			return *eps[0].IpAddress
		}
		return ""
	}).(pulumi.StringOutput))

	ctx.Export(OpEndpointPort, createdDb.Endpoints.ApplyT(func(eps []mysql.MysqlDbSystemEndpoint) string {
		if len(eps) > 0 && eps[0].Port != nil {
			return fmt.Sprintf("%d", *eps[0].Port)
		}
		return ""
	}).(pulumi.StringOutput))

	return nil
}

func buildDataStorage(ds *ocimysqldbsystemv1.OciMysqlDbSystemSpec_DataStorage) mysql.MysqlDbSystemDataStoragePtrInput {
	storage := &mysql.MysqlDbSystemDataStorageArgs{}

	if ds.DataStorageSizeInGb > 0 {
		storage.DataStorageSizeInGb = pulumi.IntPtr(int(ds.DataStorageSizeInGb))
	}

	if ds.IsAutoExpandStorageEnabled != nil {
		storage.IsAutoExpandStorageEnabled = pulumi.BoolPtr(*ds.IsAutoExpandStorageEnabled)
	}

	if ds.MaxStorageSizeInGbs > 0 {
		storage.MaxStorageSizeInGbs = pulumi.IntPtr(int(ds.MaxStorageSizeInGbs))
	}

	return storage
}

func buildBackupPolicy(bp *ocimysqldbsystemv1.OciMysqlDbSystemSpec_BackupPolicy) mysql.MysqlDbSystemBackupPolicyPtrInput {
	policy := &mysql.MysqlDbSystemBackupPolicyArgs{}

	if bp.IsEnabled != nil {
		policy.IsEnabled = pulumi.BoolPtr(*bp.IsEnabled)
	}

	if bp.RetentionInDays > 0 {
		policy.RetentionInDays = pulumi.IntPtr(int(bp.RetentionInDays))
	}

	if bp.WindowStartTime != "" {
		policy.WindowStartTime = pulumi.StringPtr(bp.WindowStartTime)
	}

	if bp.PitrPolicy != nil {
		pitr := &mysql.MysqlDbSystemBackupPolicyPitrPolicyArgs{}
		if bp.PitrPolicy.IsEnabled != nil {
			pitr.IsEnabled = pulumi.BoolPtr(*bp.PitrPolicy.IsEnabled)
		}
		policy.PitrPolicy = pitr
	}

	return policy
}

func buildMaintenance(m *ocimysqldbsystemv1.OciMysqlDbSystemSpec_Maintenance) mysql.MysqlDbSystemMaintenancePtrInput {
	maint := &mysql.MysqlDbSystemMaintenanceArgs{
		WindowStartTime: pulumi.String(m.WindowStartTime),
	}

	if m.MaintenanceScheduleType != ocimysqldbsystemv1.OciMysqlDbSystemSpec_maintenance_schedule_type_unspecified {
		maint.MaintenanceScheduleType = pulumi.StringPtr(strings.ToUpper(m.MaintenanceScheduleType.String()))
	}

	if m.VersionPreference != ocimysqldbsystemv1.OciMysqlDbSystemSpec_version_preference_unspecified {
		maint.VersionPreference = pulumi.StringPtr(strings.ToUpper(m.VersionPreference.String()))
	}

	if m.VersionTrackPreference != ocimysqldbsystemv1.OciMysqlDbSystemSpec_version_track_preference_unspecified {
		maint.VersionTrackPreference = pulumi.StringPtr(strings.ToUpper(m.VersionTrackPreference.String()))
	}

	return maint
}

func buildDeletionPolicy(dp *ocimysqldbsystemv1.OciMysqlDbSystemSpec_DeletionPolicy) mysql.MysqlDbSystemDeletionPolicyArrayInput {
	policy := mysql.MysqlDbSystemDeletionPolicyArgs{}

	if dp.AutomaticBackupRetention != "" {
		policy.AutomaticBackupRetention = pulumi.StringPtr(dp.AutomaticBackupRetention)
	}

	if dp.FinalBackup != "" {
		policy.FinalBackup = pulumi.StringPtr(dp.FinalBackup)
	}

	if dp.IsDeleteProtected != nil {
		policy.IsDeleteProtected = pulumi.BoolPtr(*dp.IsDeleteProtected)
	}

	return mysql.MysqlDbSystemDeletionPolicyArray{policy}
}

func buildEncryptData(ed *ocimysqldbsystemv1.OciMysqlDbSystemSpec_EncryptData) mysql.MysqlDbSystemEncryptDataPtrInput {
	encrypt := &mysql.MysqlDbSystemEncryptDataArgs{}

	if ed.KeyGenerationType != ocimysqldbsystemv1.OciMysqlDbSystemSpec_key_generation_type_unspecified {
		encrypt.KeyGenerationType = pulumi.String(strings.ToUpper(ed.KeyGenerationType.String()))
	}

	if ed.KeyId != nil {
		encrypt.KeyId = pulumi.StringPtr(ed.KeyId.GetValue())
	}

	return encrypt
}

func buildSecureConnections(sc *ocimysqldbsystemv1.OciMysqlDbSystemSpec_SecureConnections) mysql.MysqlDbSystemSecureConnectionsPtrInput {
	secure := &mysql.MysqlDbSystemSecureConnectionsArgs{}

	if sc.CertificateGenerationType != ocimysqldbsystemv1.OciMysqlDbSystemSpec_certificate_generation_type_unspecified {
		certType := sc.CertificateGenerationType.String()
		if certType == "system_cert" {
			certType = "SYSTEM"
		} else {
			certType = strings.ToUpper(certType)
		}
		secure.CertificateGenerationType = pulumi.String(certType)
	}

	if sc.CertificateId != nil {
		secure.CertificateId = pulumi.StringPtr(sc.CertificateId.GetValue())
	}

	return secure
}

func buildReadEndpoint(re *ocimysqldbsystemv1.OciMysqlDbSystemSpec_ReadEndpoint) mysql.MysqlDbSystemReadEndpointPtrInput {
	endpoint := &mysql.MysqlDbSystemReadEndpointArgs{}

	if re.IsEnabled != nil {
		endpoint.IsEnabled = pulumi.BoolPtr(*re.IsEnabled)
	}

	if len(re.ExcludeIps) > 0 {
		endpoint.ExcludeIps = pulumi.ToStringArray(re.ExcludeIps)
	}

	if re.ReadEndpointHostnameLabel != "" {
		endpoint.ReadEndpointHostnameLabel = pulumi.StringPtr(re.ReadEndpointHostnameLabel)
	}

	if re.ReadEndpointIpAddress != "" {
		endpoint.ReadEndpointIpAddress = pulumi.StringPtr(re.ReadEndpointIpAddress)
	}

	return endpoint
}

func buildDatabaseConsole(dc *ocimysqldbsystemv1.OciMysqlDbSystemSpec_DatabaseConsole) mysql.MysqlDbSystemDatabaseConsolePtrInput {
	console := &mysql.MysqlDbSystemDatabaseConsoleArgs{}

	if dc.Status != ocimysqldbsystemv1.OciMysqlDbSystemSpec_database_console_status_unspecified {
		console.Status = pulumi.String(strings.ToUpper(dc.Status.String()))
	}

	if dc.Port > 0 {
		console.Port = pulumi.IntPtr(int(dc.Port))
	}

	return console
}

func buildRest(r *ocimysqldbsystemv1.OciMysqlDbSystemSpec_Rest) mysql.MysqlDbSystemRestPtrInput {
	rest := &mysql.MysqlDbSystemRestArgs{}

	if r.Configuration != "" {
		rest.Configuration = pulumi.String(r.Configuration)
	}

	if r.Port > 0 {
		rest.Port = pulumi.IntPtr(int(r.Port))
	}

	return rest
}
