package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/spanner"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func spannerInstance(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpSpannerInstance.Spec

	args := &spanner.InstanceArgs{
		Name:        pulumi.StringPtr(spec.InstanceName),
		Config:      pulumi.String(spec.Config),
		DisplayName: pulumi.String(spec.DisplayName),
		Project:     pulumi.StringPtr(spec.ProjectId.GetValue()),
		Labels:      pulumi.ToStringMap(locals.GcpLabels),
	}

	// Capacity: exactly one of num_nodes, processing_units, or autoscaling_config.
	if spec.NumNodes > 0 {
		args.NumNodes = pulumi.IntPtr(int(spec.NumNodes))
	}
	if spec.ProcessingUnits > 0 {
		args.ProcessingUnits = pulumi.IntPtr(int(spec.ProcessingUnits))
	}
	if spec.AutoscalingConfig != nil {
		autoscalingArgs := &spanner.InstanceAutoscalingConfigArgs{}

		if spec.AutoscalingConfig.AutoscalingLimits != nil {
			limits := spec.AutoscalingConfig.AutoscalingLimits
			limitsArgs := &spanner.InstanceAutoscalingConfigAutoscalingLimitsArgs{}

			if limits.MinNodes > 0 {
				limitsArgs.MinNodes = pulumi.IntPtr(int(limits.MinNodes))
			}
			if limits.MaxNodes > 0 {
				limitsArgs.MaxNodes = pulumi.IntPtr(int(limits.MaxNodes))
			}
			if limits.MinProcessingUnits > 0 {
				limitsArgs.MinProcessingUnits = pulumi.IntPtr(int(limits.MinProcessingUnits))
			}
			if limits.MaxProcessingUnits > 0 {
				limitsArgs.MaxProcessingUnits = pulumi.IntPtr(int(limits.MaxProcessingUnits))
			}
			autoscalingArgs.AutoscalingLimits = limitsArgs
		}

		if spec.AutoscalingConfig.AutoscalingTargets != nil {
			targets := spec.AutoscalingConfig.AutoscalingTargets
			targetsArgs := &spanner.InstanceAutoscalingConfigAutoscalingTargetsArgs{}

			if targets.HighPriorityCpuUtilizationPercent > 0 {
				targetsArgs.HighPriorityCpuUtilizationPercent = pulumi.IntPtr(int(targets.HighPriorityCpuUtilizationPercent))
			}
			if targets.StorageUtilizationPercent > 0 {
				targetsArgs.StorageUtilizationPercent = pulumi.IntPtr(int(targets.StorageUtilizationPercent))
			}
			autoscalingArgs.AutoscalingTargets = targetsArgs
		}

		args.AutoscalingConfig = autoscalingArgs
	}

	// Instance type.
	if spec.InstanceType != "" {
		args.InstanceType = pulumi.StringPtr(spec.InstanceType)
	}

	// Edition.
	if spec.Edition != "" {
		args.Edition = pulumi.StringPtr(spec.Edition)
	}

	// Default backup schedule type.
	if spec.DefaultBackupScheduleType != "" {
		args.DefaultBackupScheduleType = pulumi.StringPtr(spec.DefaultBackupScheduleType)
	}

	// Force destroy.
	if spec.ForceDestroy {
		args.ForceDestroy = pulumi.BoolPtr(true)
	}

	createdInstance, err := spanner.NewInstance(ctx, "spanner-instance", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create spanner instance")
	}

	// Export outputs.
	// instance_id: fully qualified path projects/{project}/instances/{name}.
	ctx.Export(OpInstanceId, pulumi.Sprintf(
		"projects/%s/instances/%s",
		pulumi.String(spec.ProjectId.GetValue()),
		createdInstance.Name,
	))
	ctx.Export(OpInstanceName, createdInstance.Name)
	ctx.Export(OpState, createdInstance.State)

	// Log the instance for debugging.
	_ = fmt.Sprintf("Spanner instance %s created", spec.InstanceName)

	return nil
}
