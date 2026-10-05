package module

import (
	"github.com/pkg/errors"
	awseksnodegroupv1alpha1 "github.com/plantonhq/planton/catalog/aws/awseksnodegroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources provisions the managed node group. The cluster attaches by
// reference, the node IAM role is a referenced AwsIamRole that carries its
// own worker policies (this module never modifies a role it merely
// references), and launch mechanics come either from the inline knobs or
// from a referenced AwsLaunchTemplate -- the spec's CEL rules enforce
// AWS's mutual exclusions between the two styles.
func Resources(ctx *pulumi.Context, iacInput *awseksnodegroupv1alpha1.AwsEksNodeGroupIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which
	// resolves the right credential mechanism (static keys, keyless web identity,
	// or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.AwsEksNodeGroup.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	spec := locals.AwsEksNodeGroup.Spec

	subnetIds := make(pulumi.StringArray, 0, len(spec.SubnetIds))
	for _, subnet := range spec.SubnetIds {
		subnetIds = append(subnetIds, pulumi.String(subnet.GetValue()))
	}

	args := &eks.NodeGroupArgs{
		NodeGroupName: pulumi.StringPtr(locals.NodeGroupName),
		ClusterName:   pulumi.String(spec.ClusterName.GetValue()),
		NodeRoleArn:   pulumi.String(spec.NodeRoleArn.GetValue()),
		SubnetIds:     subnetIds,
		ScalingConfig: &eks.NodeGroupScalingConfigArgs{
			MinSize:     pulumi.Int(int(spec.Scaling.MinSize)),
			MaxSize:     pulumi.Int(int(spec.Scaling.MaxSize)),
			DesiredSize: pulumi.Int(int(spec.Scaling.DesiredSize)),
		},
		CapacityType: pulumi.StringPtr(capacityType(spec.CapacityType)),
		Tags:         pulumi.ToStringMap(locals.AwsTags),
	}

	// Inline launch mechanics vs a referenced launch template -- AWS rejects
	// mixing them ("...or the node group deployment will fail"), which the
	// spec's CEL rules enforce before anything reaches the API.
	if spec.LaunchTemplate != nil {
		args.LaunchTemplate = &eks.NodeGroupLaunchTemplateArgs{
			Id: pulumi.StringPtr(spec.LaunchTemplate.LaunchTemplateId.GetValue()),
			// The provider requires a version. "$Default" follows the
			// template's default version -- the setup that lets promoting a
			// template version roll the fleet. Pin a numeric version for
			// fully drift-free plans (AWS reads "$Default" back as the
			// resolved number).
			Version: pulumi.String(launchTemplateVersion(spec.LaunchTemplate)),
		}
	}
	if len(spec.InstanceTypes) > 0 {
		args.InstanceTypes = pulumi.ToStringArray(spec.InstanceTypes)
	}
	// 0 keeps the AWS default disk size (20 GiB Linux / 50 GiB Windows).
	if spec.DiskSizeGb > 0 {
		args.DiskSize = pulumi.IntPtr(int(spec.DiskSizeGb))
	}
	if spec.AmiType != "" {
		args.AmiType = pulumi.StringPtr(spec.AmiType)
	}
	if spec.RemoteAccess != nil {
		remoteAccess := &eks.NodeGroupRemoteAccessArgs{}
		if spec.RemoteAccess.Ec2SshKey != "" {
			remoteAccess.Ec2SshKey = pulumi.StringPtr(spec.RemoteAccess.Ec2SshKey)
		}
		if len(spec.RemoteAccess.SourceSecurityGroupIds) > 0 {
			securityGroupIds := make(pulumi.StringArray, 0, len(spec.RemoteAccess.SourceSecurityGroupIds))
			for _, securityGroup := range spec.RemoteAccess.SourceSecurityGroupIds {
				securityGroupIds = append(securityGroupIds, pulumi.String(securityGroup.GetValue()))
			}
			remoteAccess.SourceSecurityGroupIds = securityGroupIds
		}
		args.RemoteAccess = remoteAccess
	}

	if len(spec.Labels) > 0 {
		args.Labels = pulumi.ToStringMap(spec.Labels)
	}
	if len(spec.Taints) > 0 {
		taints := make(eks.NodeGroupTaintArray, 0, len(spec.Taints))
		for _, taint := range spec.Taints {
			taintArgs := &eks.NodeGroupTaintArgs{
				Key:    pulumi.String(taint.Key),
				Effect: pulumi.String(taint.Effect),
			}
			if taint.Value != "" {
				taintArgs.Value = pulumi.StringPtr(taint.Value)
			}
			taints = append(taints, taintArgs)
		}
		args.Taints = taints
	}

	if spec.UpdateConfig != nil {
		updateConfig := &eks.NodeGroupUpdateConfigArgs{}
		// Exactly one form is set (spec CEL enforces it), mirroring AWS's
		// ExactlyOneOf on the same fields.
		if spec.UpdateConfig.MaxUnavailable > 0 {
			updateConfig.MaxUnavailable = pulumi.IntPtr(int(spec.UpdateConfig.MaxUnavailable))
		}
		if spec.UpdateConfig.MaxUnavailablePercentage > 0 {
			updateConfig.MaxUnavailablePercentage = pulumi.IntPtr(int(spec.UpdateConfig.MaxUnavailablePercentage))
		}
		if spec.UpdateConfig.UpdateStrategy != "" {
			updateConfig.UpdateStrategy = pulumi.StringPtr(spec.UpdateConfig.UpdateStrategy)
		}
		args.UpdateConfig = updateConfig
	}

	if spec.NodeRepairConfig != nil {
		repairConfig := &eks.NodeGroupNodeRepairConfigArgs{
			Enabled: pulumi.BoolPtr(spec.NodeRepairConfig.Enabled),
		}
		if spec.NodeRepairConfig.MaxParallelNodesRepairedCount > 0 {
			repairConfig.MaxParallelNodesRepairedCount = pulumi.IntPtr(int(spec.NodeRepairConfig.MaxParallelNodesRepairedCount))
		}
		if spec.NodeRepairConfig.MaxParallelNodesRepairedPercentage > 0 {
			repairConfig.MaxParallelNodesRepairedPercentage = pulumi.IntPtr(int(spec.NodeRepairConfig.MaxParallelNodesRepairedPercentage))
		}
		if spec.NodeRepairConfig.MaxUnhealthyNodeThresholdCount > 0 {
			repairConfig.MaxUnhealthyNodeThresholdCount = pulumi.IntPtr(int(spec.NodeRepairConfig.MaxUnhealthyNodeThresholdCount))
		}
		if spec.NodeRepairConfig.MaxUnhealthyNodeThresholdPercentage > 0 {
			repairConfig.MaxUnhealthyNodeThresholdPercentage = pulumi.IntPtr(int(spec.NodeRepairConfig.MaxUnhealthyNodeThresholdPercentage))
		}
		if len(spec.NodeRepairConfig.Overrides) > 0 {
			overrides := make(eks.NodeGroupNodeRepairConfigNodeRepairConfigOverrideArray, 0, len(spec.NodeRepairConfig.Overrides))
			for _, override := range spec.NodeRepairConfig.Overrides {
				overrides = append(overrides, &eks.NodeGroupNodeRepairConfigNodeRepairConfigOverrideArgs{
					MinRepairWaitTimeMins:   pulumi.Int(int(override.MinRepairWaitTimeMins)),
					NodeMonitoringCondition: pulumi.String(override.NodeMonitoringCondition),
					NodeUnhealthyReason:     pulumi.String(override.NodeUnhealthyReason),
					RepairAction:            pulumi.String(override.RepairAction),
				})
			}
			repairConfig.NodeRepairConfigOverrides = overrides
		}
		args.NodeRepairConfig = repairConfig
	}

	// A pool of pre-initialized nodes that cuts scale-out from minutes to
	// seconds; updates in place. Stopped pools cost only EBS storage.
	if spec.WarmPoolConfig != nil {
		warmPoolConfig := &eks.NodeGroupWarmPoolConfigArgs{}
		if spec.WarmPoolConfig.PoolState != "" {
			warmPoolConfig.PoolState = pulumi.StringPtr(spec.WarmPoolConfig.PoolState)
		}
		if spec.WarmPoolConfig.MinSize > 0 {
			warmPoolConfig.MinSize = pulumi.IntPtr(int(spec.WarmPoolConfig.MinSize))
		}
		// Explicit 0 is meaningful (no prepared capacity beyond min_size),
		// so presence decides what is sent; unset keeps AWS's default
		// ceiling (max_size minus desired capacity).
		if spec.WarmPoolConfig.MaxGroupPreparedCapacity != nil {
			warmPoolConfig.MaxGroupPreparedCapacity = pulumi.IntPtr(int(*spec.WarmPoolConfig.MaxGroupPreparedCapacity))
		}
		if spec.WarmPoolConfig.ReuseOnScaleIn {
			warmPoolConfig.ReuseOnScaleIn = pulumi.BoolPtr(true)
		}
		args.WarmPoolConfig = warmPoolConfig
	}

	// Version changes roll the group node by node; release_version pins the
	// exact AMI release within the minor.
	if spec.Version != "" {
		args.Version = pulumi.StringPtr(spec.Version)
	}
	if spec.ReleaseVersion != "" {
		args.ReleaseVersion = pulumi.StringPtr(spec.ReleaseVersion)
	}
	if spec.ForceUpdateVersion {
		args.ForceUpdateVersion = pulumi.BoolPtr(true)
	}

	created, err := eks.NewNodeGroup(ctx, locals.NodeGroupName, args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create EKS node group")
	}

	ctx.Export(OpNodeGroupName, created.NodeGroupName)
	ctx.Export(OpNodeGroupArn, created.Arn)
	// AWS runs the fleet through an EC2 Auto Scaling group it manages;
	// surfacing its name gives ASG-level tooling (activity history, custom
	// metrics) a hook. The remote-access security group exists only when
	// AWS had to create one (SSH enabled without explicit source groups).
	ctx.Export(OpAsgName, created.Resources.Index(pulumi.Int(0)).AutoscalingGroups().Index(pulumi.Int(0)).Name().Elem())
	ctx.Export(OpRemoteAccessSgId, created.Resources.Index(pulumi.Int(0)).RemoteAccessSecurityGroupId().Elem())

	return nil
}

// capacityType converts the proto enum to the AWS API string.
func capacityType(value awseksnodegroupv1alpha1.AwsEksNodeGroupCapacityType) string {
	switch value {
	case awseksnodegroupv1alpha1.AwsEksNodeGroupCapacityType_spot:
		return "SPOT"
	case awseksnodegroupv1alpha1.AwsEksNodeGroupCapacityType_capacity_block:
		return "CAPACITY_BLOCK"
	default:
		return "ON_DEMAND"
	}
}

// launchTemplateVersion resolves the version string sent to AWS: the
// provider requires one, so an unset spec version means "$Default".
func launchTemplateVersion(launchTemplate *awseksnodegroupv1alpha1.AwsEksNodeGroupLaunchTemplate) string {
	if launchTemplate.Version != "" {
		return launchTemplate.Version
	}
	return "$Default"
}
