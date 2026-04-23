package module

import (
	"fmt"
	"strings"

	gcpcloudsqlv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/gcp/gcpcloudsql/v1"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/sql"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// databaseInstance creates a Cloud SQL database instance.
func databaseInstance(
	ctx *pulumi.Context,
	locals *Locals,
	gcpProvider *gcp.Provider,
) (*sql.DatabaseInstance, error) {

	spec := locals.GcpCloudSql.Spec

	// Convert database engine enum to string
	databaseVersion := spec.DatabaseVersion

	settings := &sql.DatabaseInstanceSettingsArgs{
		Tier:           pulumi.String(spec.Tier),
		DiskSize:       pulumi.Int(int(spec.StorageGb)),
		DiskType:       pulumi.String("PD_SSD"),
		DiskAutoresize: pulumi.BoolPtr(spec.DiskAutoResize),
		UserLabels:     pulumi.ToStringMap(locals.GcpLabels),
	}

	// Configure IP settings
	if spec.Network != nil {
		ipConfig := &sql.DatabaseInstanceSettingsIpConfigurationArgs{}

		if spec.Network.PrivateIpEnabled {
			ipConfig.PrivateNetwork = pulumi.String(spec.Network.VpcId.GetValue())
			ipConfig.Ipv4Enabled = pulumi.Bool(spec.Network.Ipv4Enabled)
		} else {
			ipConfig.Ipv4Enabled = pulumi.Bool(true)
		}

		// Add authorized networks if specified
		if len(spec.Network.AuthorizedNetworks) > 0 {
			authorizedNetworks := sql.DatabaseInstanceSettingsIpConfigurationAuthorizedNetworkArray{}
			for i, cidr := range spec.Network.AuthorizedNetworks {
				authorizedNetworks = append(authorizedNetworks, &sql.DatabaseInstanceSettingsIpConfigurationAuthorizedNetworkArgs{
					Name:  pulumi.String(fmt.Sprintf("authorized-network-%d", i)),
					Value: pulumi.String(cidr),
				})
			}
			ipConfig.AuthorizedNetworks = authorizedNetworks
		}

		settings.IpConfiguration = ipConfig
	}

	// Configure high availability
	if spec.HighAvailability != nil && spec.HighAvailability.Enabled {
		settings.AvailabilityType = pulumi.String("REGIONAL")
	} else {
		settings.AvailabilityType = pulumi.String("ZONAL")
	}

	// Configure backup
	if spec.Backup != nil && spec.Backup.Enabled {
		settings.BackupConfiguration = &sql.DatabaseInstanceSettingsBackupConfigurationArgs{
			Enabled:   pulumi.Bool(true),
			StartTime: pulumi.String(spec.Backup.StartTime),
			BackupRetentionSettings: &sql.DatabaseInstanceSettingsBackupConfigurationBackupRetentionSettingsArgs{
				RetainedBackups: pulumi.Int(int(spec.Backup.RetentionDays)),
			},
			PointInTimeRecoveryEnabled: pulumi.Bool(true),
		}
	}

	// Configure database flags if specified
	if len(spec.DatabaseFlags) > 0 {
		databaseFlags := sql.DatabaseInstanceSettingsDatabaseFlagArray{}
		for name, value := range spec.DatabaseFlags {
			databaseFlags = append(databaseFlags, &sql.DatabaseInstanceSettingsDatabaseFlagArgs{
				Name:  pulumi.String(name),
				Value: pulumi.String(value),
			})
		}
		settings.DatabaseFlags = databaseFlags
	}

	// Create the database instance
	instance, err := sql.NewDatabaseInstance(ctx,
		locals.GcpCloudSql.Metadata.Name,
		&sql.DatabaseInstanceArgs{
			Name:               pulumi.String(locals.GcpCloudSql.Metadata.Name),
			Project:            pulumi.String(spec.ProjectId.GetValue()),
			Region:             pulumi.String(spec.Region),
			DatabaseVersion:    pulumi.String(databaseVersion),
			Settings:           settings,
			RootPassword:       pulumi.String(spec.RootPassword),
			DeletionProtection: pulumi.Bool(spec.DeletionProtection),
		},
		pulumi.Provider(gcpProvider),
	)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

// getDatabaseEngineString converts the enum to the appropriate GCP database version string
func getDatabaseEngineString(engine gcpcloudsqlv1.GcpCloudSqlDatabaseEngine) string {
	switch engine {
	case gcpcloudsqlv1.GcpCloudSqlDatabaseEngine_MYSQL:
		return "MYSQL"
	case gcpcloudsqlv1.GcpCloudSqlDatabaseEngine_POSTGRESQL:
		return "POSTGRES"
	default:
		return strings.ToUpper(engine.String())
	}
}
