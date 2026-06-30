package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	scaleway "github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/databases"
)

// rdbInstance provisions the Scaleway RDB instance and exports the core
// stack outputs (instance ID, endpoints, certificate).
//
// The instance is created with:
//   - The admin user (user_name + password from spec).
//   - Optional Private Network attachment (IPAM-based).
//   - Optional HA, backup, encryption, and volume configuration.
//   - Standard Planton tags for resource identification.
func rdbInstance(
	ctx *pulumi.Context,
	locals *Locals,
	scalewayProvider *scaleway.Provider,
) (*databases.Instance, error) {
	spec := locals.ScalewayRdbInstance.Spec

	// Build the instance arguments.
	instanceArgs := &databases.InstanceArgs{
		Name:     pulumi.String(locals.ScalewayRdbInstance.Metadata.Name),
		Engine:   pulumi.String(spec.Engine),
		NodeType: pulumi.String(spec.NodeType),
		Region:   pulumi.StringPtr(spec.Region),
		Tags:     toPulumiStringArray(locals.ScalewayTags),

		// Admin user credentials (created with the instance).
		UserName: pulumi.StringPtr(spec.AdminUser),
		Password: pulumi.StringPtr(spec.AdminPassword),

		// High availability.
		IsHaCluster: pulumi.BoolPtr(spec.IsHaCluster),
	}

	// Volume configuration.
	if spec.VolumeType != "" {
		instanceArgs.VolumeType = pulumi.StringPtr(spec.VolumeType)
	}
	if spec.VolumeSizeInGb > 0 {
		instanceArgs.VolumeSizeInGb = pulumi.IntPtr(int(spec.VolumeSizeInGb))
	}

	// Backup configuration.
	instanceArgs.DisableBackup = pulumi.BoolPtr(spec.DisableBackup)
	if spec.BackupScheduleFrequencyHours > 0 {
		instanceArgs.BackupScheduleFrequency = pulumi.IntPtr(int(spec.BackupScheduleFrequencyHours))
	}
	if spec.BackupScheduleRetentionDays > 0 {
		instanceArgs.BackupScheduleRetention = pulumi.IntPtr(int(spec.BackupScheduleRetentionDays))
	}

	// Encryption at rest.
	if spec.EncryptionAtRest {
		instanceArgs.EncryptionAtRest = pulumi.BoolPtr(true)
	}

	// Optional Private Network attachment with IPAM.
	if locals.PrivateNetworkId != "" {
		instanceArgs.PrivateNetwork = &databases.InstancePrivateNetworkArgs{
			PnId:       pulumi.String(locals.PrivateNetworkId),
			EnableIpam: pulumi.BoolPtr(true),
		}
	}

	// Engine settings.
	if len(spec.Settings) > 0 {
		settingsMap := pulumi.StringMap{}
		for k, v := range spec.Settings {
			settingsMap[k] = pulumi.String(v)
		}
		instanceArgs.Settings = settingsMap
	}
	if len(spec.InitSettings) > 0 {
		initSettingsMap := pulumi.StringMap{}
		for k, v := range spec.InitSettings {
			initSettingsMap[k] = pulumi.String(v)
		}
		instanceArgs.InitSettings = initSettingsMap
	}

	// Create the instance.
	createdInstance, err := databases.NewInstance(
		ctx,
		"instance",
		instanceArgs,
		pulumi.Provider(scalewayProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create rdb instance")
	}

	// Export core outputs.
	ctx.Export(OpInstanceId, createdInstance.ID())
	ctx.Export(OpEndpointIp, createdInstance.EndpointIp)
	ctx.Export(OpEndpointPort, createdInstance.EndpointPort)
	ctx.Export(OpCertificate, createdInstance.Certificate)

	// Export Private Network endpoint if attached.
	if locals.PrivateNetworkId != "" {
		ctx.Export(OpPrivateEndpointIp, createdInstance.PrivateNetwork.ApplyT(func(pn *databases.InstancePrivateNetwork) string {
			if pn != nil && pn.Ip != nil {
				return *pn.Ip
			}
			return ""
		}).(pulumi.StringOutput))
		ctx.Export(OpPrivateEndpointPort, createdInstance.PrivateNetwork.ApplyT(func(pn *databases.InstancePrivateNetwork) int {
			if pn != nil && pn.Port != nil {
				return *pn.Port
			}
			return 0
		}).(pulumi.IntOutput))
	}

	return createdInstance, nil
}

// toPulumiStringArray converts a Go string slice to a Pulumi StringArray.
func toPulumiStringArray(tags []string) pulumi.StringArray {
	result := make(pulumi.StringArray, len(tags))
	for i, tag := range tags {
		result[i] = pulumi.String(tag)
	}
	return result
}
