package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/alloydb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func primaryInstance(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider, createdCluster *alloydb.Cluster) error {
	spec := locals.GcpAlloydbCluster.Spec
	instanceSpec := spec.PrimaryInstance

	args := &alloydb.InstanceArgs{
		Cluster:      createdCluster.Name,
		InstanceId:   pulumi.String(instanceSpec.InstanceId),
		InstanceType: pulumi.String("PRIMARY"),
		Labels:       pulumi.ToStringMap(locals.GcpLabels),
	}

	// Machine configuration: cpu_count or machine_type (mutually exclusive, validated by proto).
	if instanceSpec.CpuCount > 0 || instanceSpec.MachineType != "" {
		machineConfig := &alloydb.InstanceMachineConfigArgs{}
		if instanceSpec.CpuCount > 0 {
			machineConfig.CpuCount = pulumi.IntPtr(int(instanceSpec.CpuCount))
		}
		if instanceSpec.MachineType != "" {
			machineConfig.MachineType = pulumi.StringPtr(instanceSpec.MachineType)
		}
		args.MachineConfig = machineConfig
	}

	// Availability type.
	if instanceSpec.AvailabilityType != "" {
		args.AvailabilityType = pulumi.StringPtr(instanceSpec.AvailabilityType)
	}

	// Database flags.
	if len(instanceSpec.DatabaseFlags) > 0 {
		args.DatabaseFlags = pulumi.ToStringMap(instanceSpec.DatabaseFlags)
	}

	// Display name.
	if instanceSpec.DisplayName != "" {
		args.DisplayName = pulumi.StringPtr(instanceSpec.DisplayName)
	}

	// Query insights configuration.
	if instanceSpec.QueryInsightsConfig != nil {
		qiConfig := instanceSpec.QueryInsightsConfig
		qiArgs := &alloydb.InstanceQueryInsightsConfigArgs{}

		if qiConfig.QueryPlansPerMinute > 0 {
			qiArgs.QueryPlansPerMinute = pulumi.IntPtr(int(qiConfig.QueryPlansPerMinute))
		}
		if qiConfig.QueryStringLength > 0 {
			qiArgs.QueryStringLength = pulumi.IntPtr(int(qiConfig.QueryStringLength))
		}
		if qiConfig.RecordApplicationTags {
			qiArgs.RecordApplicationTags = pulumi.BoolPtr(true)
		}
		if qiConfig.RecordClientAddress {
			qiArgs.RecordClientAddress = pulumi.BoolPtr(true)
		}
		args.QueryInsightsConfig = qiArgs
	}

	// Client connection configuration (require_connectors, ssl_mode).
	if instanceSpec.RequireConnectors || instanceSpec.SslMode != "" {
		clientConfig := &alloydb.InstanceClientConnectionConfigArgs{}
		if instanceSpec.RequireConnectors {
			clientConfig.RequireConnectors = pulumi.BoolPtr(true)
		}
		if instanceSpec.SslMode != "" {
			clientConfig.SslConfig = &alloydb.InstanceClientConnectionConfigSslConfigArgs{
				SslMode: pulumi.StringPtr(instanceSpec.SslMode),
			}
		}
		args.ClientConnectionConfig = clientConfig
	}

	createdInstance, err := alloydb.NewInstance(ctx, "alloydb-primary-instance", args,
		pulumi.Provider(gcpProvider),
		pulumi.DependsOn([]pulumi.Resource{createdCluster}),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create alloydb primary instance")
	}

	// Export primary instance outputs.
	ctx.Export(OpPrimaryInstanceIp, createdInstance.IpAddress)
	ctx.Export(OpPrimaryInstanceName, createdInstance.Name)

	return nil
}
