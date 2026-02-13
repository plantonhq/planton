package module

import (
	"github.com/pkg/errors"
	scaleway "github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/mongodb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// mongodbInstance provisions the Scaleway MongoDB instance and exports the
// core stack outputs (instance ID, endpoints, TLS certificate).
//
// The instance is created with:
//   - The admin user (user_name + password from spec).
//   - Optional Private Network attachment (IPAM-based).
//   - Optional public network endpoint (when both PN and public are wanted).
//   - Optional snapshot schedule configuration.
//   - Standard OpenMCF tags for resource identification.
//
// Networking logic:
//   - If private_network_id is set: attach PN endpoint.
//     If enable_public_network is also true: also attach public endpoint.
//     If enable_public_network is false (default): private-only.
//   - If private_network_id is not set: public endpoint by default
//     (Scaleway default behavior -- no explicit block needed).
func mongodbInstance(
	ctx *pulumi.Context,
	locals *Locals,
	scalewayProvider *scaleway.Provider,
) (*mongodb.Instance, error) {
	spec := locals.ScalewayMongodbInstance.Spec

	// Build the instance arguments.
	instanceArgs := &mongodb.InstanceArgs{
		Name:       pulumi.StringPtr(locals.ScalewayMongodbInstance.Metadata.Name),
		Version:    pulumi.StringPtr(spec.Version),
		NodeType:   pulumi.String(spec.NodeType),
		NodeNumber: pulumi.Int(int(spec.NodeNumber)),
		Region:     pulumi.StringPtr(spec.Region),
		Tags:       toPulumiStringArray(locals.ScalewayTags),

		// Admin user credentials (created with the instance).
		UserName: pulumi.StringPtr(spec.AdminUser),
		Password: pulumi.StringPtr(spec.AdminPassword),
	}

	// Volume configuration.
	if spec.VolumeType != "" {
		instanceArgs.VolumeType = pulumi.StringPtr(spec.VolumeType)
	}
	if spec.VolumeSizeInGb > 0 {
		instanceArgs.VolumeSizeInGb = pulumi.IntPtr(int(spec.VolumeSizeInGb))
	}

	// Snapshot schedule configuration.
	if spec.EnableSnapshotSchedule {
		instanceArgs.IsSnapshotScheduleEnabled = pulumi.BoolPtr(true)
		if spec.SnapshotScheduleFrequencyHours > 0 {
			instanceArgs.SnapshotScheduleFrequencyHours = pulumi.IntPtr(int(spec.SnapshotScheduleFrequencyHours))
		}
		if spec.SnapshotScheduleRetentionDays > 0 {
			instanceArgs.SnapshotScheduleRetentionDays = pulumi.IntPtr(int(spec.SnapshotScheduleRetentionDays))
		}
	}

	// Settings.
	if len(spec.Settings) > 0 {
		settingsMap := pulumi.StringMap{}
		for k, v := range spec.Settings {
			settingsMap[k] = pulumi.String(v)
		}
		instanceArgs.Settings = settingsMap
	}

	// Networking: Private Network attachment.
	if locals.PrivateNetworkId != "" {
		instanceArgs.PrivateNetwork = &mongodb.InstancePrivateNetworkArgs{
			PnId: pulumi.String(locals.PrivateNetworkId),
		}
	}

	// Networking: Public network endpoint.
	// Added when: (a) user explicitly wants both PN + public, or
	// (b) no PN is set (Scaleway creates public by default, but being
	//     explicit is cleaner than relying on implicit behavior).
	if locals.PrivateNetworkId != "" && spec.EnablePublicNetwork {
		instanceArgs.PublicNetwork = &mongodb.InstancePublicNetworkArgs{}
	}

	// Create the instance.
	createdInstance, err := mongodb.NewInstance(
		ctx,
		"instance",
		instanceArgs,
		pulumi.Provider(scalewayProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create mongodb instance")
	}

	// Export core output: instance ID.
	ctx.Export(OpInstanceId, createdInstance.ID())

	// Export TLS certificate (always available).
	ctx.Export(OpTlsCertificate, createdInstance.TlsCertificate)

	// Export public network endpoint (populated when public endpoint exists).
	ctx.Export(OpPublicDnsRecord, createdInstance.PublicNetwork.ApplyT(func(pn *mongodb.InstancePublicNetwork) string {
		if pn != nil && pn.DnsRecord != nil {
			return *pn.DnsRecord
		}
		return ""
	}).(pulumi.StringOutput))

	ctx.Export(OpPublicPort, createdInstance.PublicNetwork.ApplyT(func(pn *mongodb.InstancePublicNetwork) int {
		if pn != nil && pn.Port != nil {
			return *pn.Port
		}
		return 0
	}).(pulumi.IntOutput))

	// Export private network endpoint (populated when PN is attached).
	ctx.Export(OpPrivateDnsRecords, createdInstance.PrivateNetwork.ApplyT(func(pn *mongodb.InstancePrivateNetwork) []string {
		if pn != nil {
			return pn.DnsRecords
		}
		return nil
	}).(pulumi.StringArrayOutput))

	ctx.Export(OpPrivateIps, createdInstance.PrivateNetwork.ApplyT(func(pn *mongodb.InstancePrivateNetwork) []string {
		if pn != nil {
			return pn.Ips
		}
		return nil
	}).(pulumi.StringArrayOutput))

	ctx.Export(OpPrivatePort, createdInstance.PrivateNetwork.ApplyT(func(pn *mongodb.InstancePrivateNetwork) int {
		if pn != nil && pn.Port != nil {
			return *pn.Port
		}
		return 0
	}).(pulumi.IntOutput))

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
