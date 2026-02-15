package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/bigtable"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func bigtableInstance(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpBigtableInstance.Spec

	// Build cluster configurations from the spec.
	clusters := bigtable.InstanceClusterArray{}
	for _, c := range spec.Clusters {
		clusterArgs := bigtable.InstanceClusterArgs{
			ClusterId: pulumi.String(c.ClusterId),
			Zone:      pulumi.String(c.Zone),
		}

		// Storage type (optional, middleware applies default "SSD").
		if c.GetStorageType() != "" {
			clusterArgs.StorageType = pulumi.StringPtr(c.GetStorageType())
		}

		// Node scaling factor (optional, GCP defaults to 1X).
		if c.NodeScalingFactor != "" {
			clusterArgs.NodeScalingFactor = pulumi.StringPtr(c.NodeScalingFactor)
		}

		// CMEK encryption (optional).
		if c.KmsKeyName != nil && c.KmsKeyName.GetValue() != "" {
			clusterArgs.KmsKeyName = pulumi.StringPtr(c.KmsKeyName.GetValue())
		}

		// Scaling: either fixed num_nodes or autoscaling (mutually exclusive,
		// validated by proto CEL). If neither is set, Bigtable auto-allocates.
		if c.NumNodes > 0 {
			clusterArgs.NumNodes = pulumi.IntPtr(int(c.NumNodes))
		}
		if c.AutoscalingConfig != nil {
			autoscalingArgs := &bigtable.InstanceClusterAutoscalingConfigArgs{
				MinNodes:  pulumi.Int(int(c.AutoscalingConfig.MinNodes)),
				MaxNodes:  pulumi.Int(int(c.AutoscalingConfig.MaxNodes)),
				CpuTarget: pulumi.Int(int(c.AutoscalingConfig.CpuTarget)),
			}
			if c.AutoscalingConfig.StorageTarget > 0 {
				autoscalingArgs.StorageTarget = pulumi.IntPtr(int(c.AutoscalingConfig.StorageTarget))
			}
			clusterArgs.AutoscalingConfig = autoscalingArgs
		}

		clusters = append(clusters, clusterArgs)
	}

	args := &bigtable.InstanceArgs{
		Name:               pulumi.String(spec.InstanceName),
		Project:            pulumi.StringPtr(spec.ProjectId.GetValue()),
		Labels:             pulumi.ToStringMap(locals.GcpLabels),
		Clusters:           clusters,
		DeletionProtection: pulumi.BoolPtr(spec.GetDeletionProtection()),
		ForceDestroy:       pulumi.BoolPtr(spec.ForceDestroy),
	}

	// Display name (optional, GCP defaults to instance name).
	if spec.DisplayName != "" {
		args.DisplayName = pulumi.StringPtr(spec.DisplayName)
	}

	createdInstance, err := bigtable.NewInstance(ctx, "bigtable-instance", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create bigtable instance")
	}

	// Export outputs.
	ctx.Export(OpInstanceId, createdInstance.ID())
	ctx.Export(OpInstanceName, pulumi.String(spec.InstanceName))

	return nil
}
