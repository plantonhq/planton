package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/filestore"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func filestoreInstance(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpFilestoreInstance.Spec

	// Build NFS export options for the file share.
	var nfsExportOptions filestore.InstanceFileSharesNfsExportOptionArray
	for _, opt := range spec.FileShare.NfsExportOptions {
		exportOpt := &filestore.InstanceFileSharesNfsExportOptionArgs{}

		if len(opt.IpRanges) > 0 {
			exportOpt.IpRanges = pulumi.ToStringArray(opt.IpRanges)
		}
		if opt.AccessMode != "" {
			exportOpt.AccessMode = pulumi.StringPtr(opt.AccessMode)
		}
		if opt.SquashMode != "" {
			exportOpt.SquashMode = pulumi.StringPtr(opt.SquashMode)
		}
		if opt.AnonUid != nil {
			exportOpt.AnonUid = pulumi.IntPtr(int(*opt.AnonUid))
		}
		if opt.AnonGid != nil {
			exportOpt.AnonGid = pulumi.IntPtr(int(*opt.AnonGid))
		}

		nfsExportOptions = append(nfsExportOptions, exportOpt)
	}

	// Build the file share configuration (singular -- one per instance).
	fileShareArgs := &filestore.InstanceFileSharesArgs{
		Name:       pulumi.String(spec.FileShare.Name),
		CapacityGb: pulumi.Int(int(spec.FileShare.CapacityGb)),
	}
	if len(nfsExportOptions) > 0 {
		fileShareArgs.NfsExportOptions = nfsExportOptions
	}

	// Build the network configuration (singular -- one per instance).
	networkArgs := &filestore.InstanceNetworkArgs{
		Network: pulumi.String(spec.NetworkConfig.Network.GetValue()),
		Modes:   pulumi.StringArray{pulumi.String("MODE_IPV4")},
	}
	if spec.NetworkConfig.ConnectMode != "" {
		networkArgs.ConnectMode = pulumi.StringPtr(spec.NetworkConfig.ConnectMode)
	}
	if spec.NetworkConfig.ReservedIpRange != "" {
		networkArgs.ReservedIpRange = pulumi.StringPtr(spec.NetworkConfig.ReservedIpRange)
	}

	args := &filestore.InstanceArgs{
		Name:       pulumi.String(spec.InstanceName),
		Project:    pulumi.StringPtr(spec.ProjectId.GetValue()),
		Location:   pulumi.StringPtr(spec.Location),
		Tier:       pulumi.String(spec.Tier),
		FileShares: fileShareArgs,
		Networks:   filestore.InstanceNetworkArray{networkArgs},
		Labels:     pulumi.ToStringMap(locals.GcpLabels),
	}

	// Description.
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}

	// NFS protocol version.
	if spec.Protocol != "" {
		args.Protocol = pulumi.StringPtr(spec.Protocol)
	}

	// CMEK encryption.
	if spec.KmsKeyName != nil && spec.KmsKeyName.GetValue() != "" {
		args.KmsKeyName = pulumi.StringPtr(spec.KmsKeyName.GetValue())
	}

	// Deletion protection.
	if spec.DeletionProtectionEnabled {
		args.DeletionProtectionEnabled = pulumi.BoolPtr(true)
	}
	if spec.DeletionProtectionReason != "" {
		args.DeletionProtectionReason = pulumi.StringPtr(spec.DeletionProtectionReason)
	}

	// Performance configuration.
	if spec.PerformanceConfig != nil {
		perfArgs := &filestore.InstancePerformanceConfigArgs{}
		if spec.PerformanceConfig.FixedIops != nil {
			perfArgs.FixedIops = &filestore.InstancePerformanceConfigFixedIopsArgs{
				MaxIops: pulumi.IntPtr(int(spec.PerformanceConfig.FixedIops.MaxIops)),
			}
		}
		if spec.PerformanceConfig.IopsPerTb != nil {
			perfArgs.IopsPerTb = &filestore.InstancePerformanceConfigIopsPerTbArgs{
				MaxIopsPerTb: pulumi.IntPtr(int(spec.PerformanceConfig.IopsPerTb.MaxIopsPerTb)),
			}
		}
		args.PerformanceConfig = perfArgs
	}

	createdInstance, err := filestore.NewInstance(ctx, "filestore-instance", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create filestore instance")
	}

	// Export outputs.
	ctx.Export(OpInstanceId, createdInstance.ID())
	ctx.Export(OpInstanceName, createdInstance.Name)
	ctx.Export(OpFileShareName, pulumi.String(spec.FileShare.Name))
	ctx.Export(OpCreateTime, createdInstance.CreateTime)

	// Extract IP addresses from the first (only) network.
	ipAddresses := createdInstance.Networks.ApplyT(func(networks []filestore.InstanceNetwork) []string {
		if len(networks) > 0 && len(networks[0].IpAddresses) > 0 {
			return networks[0].IpAddresses
		}
		return []string{}
	}).(pulumi.StringArrayOutput)
	ctx.Export(OpIpAddresses, ipAddresses)

	return nil
}
