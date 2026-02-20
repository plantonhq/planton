package module

import (
	"fmt"

	"github.com/pkg/errors"
	alicloudkubernetesnodepoolv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudkubernetesnodepool/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/cs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudkubernetesnodepoolv1.AlicloudKubernetesNodePoolStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AlicloudKubernetesNodePool.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	args := &cs.NodePoolArgs{
		ClusterId:     pulumi.String(locals.ClusterId),
		NodePoolName:  pulumi.String(spec.Name),
		VswitchIds:    pulumi.ToStringArray(locals.VswitchIds),
		InstanceTypes: pulumi.ToStringArray(spec.InstanceTypes),
		ImageType:     pulumi.String(imageType(spec)),
		Tags:          pulumi.ToStringMap(locals.Tags),

		SystemDiskCategory: pulumi.String(systemDiskCategory(spec.SystemDisk)),
		SystemDiskSize:     pulumi.Int(systemDiskSize(spec.SystemDisk)),

		InstanceChargeType:  pulumi.String(instanceChargeType(spec)),
		InstallCloudMonitor: pulumi.Bool(installCloudMonitor(spec)),
	}

	if spec.DesiredSize != nil {
		args.DesiredSize = pulumi.String(fmt.Sprintf("%d", *spec.DesiredSize))
	}

	if spec.SystemDisk != nil {
		if spec.SystemDisk.PerformanceLevel != "" {
			args.SystemDiskPerformanceLevel = pulumi.String(spec.SystemDisk.PerformanceLevel)
		}
		if spec.SystemDisk.Encrypted != nil && *spec.SystemDisk.Encrypted {
			args.SystemDiskEncrypted = pulumi.Bool(true)
		}
		if spec.SystemDisk.KmsKeyId != "" {
			args.SystemDiskKmsKey = pulumi.String(spec.SystemDisk.KmsKeyId)
		}
	}

	if len(spec.DataDisks) > 0 {
		args.DataDisks = dataDisks(spec.DataDisks)
	}

	if len(locals.SecurityGroupIds) > 0 {
		args.SecurityGroupIds = pulumi.ToStringArray(locals.SecurityGroupIds)
	}

	if spec.InternetMaxBandwidthOut != nil {
		args.InternetMaxBandwidthOut = pulumi.Int(int(*spec.InternetMaxBandwidthOut))
	}

	if spec.InternetChargeType != nil {
		args.InternetChargeType = pulumi.String(*spec.InternetChargeType)
	}

	if spec.KeyName != "" {
		args.KeyName = pulumi.String(spec.KeyName)
	}

	if spec.Password != "" {
		args.Password = pulumi.String(spec.Password)
	}

	if len(spec.Labels) > 0 {
		args.Labels = nodeLabels(spec.Labels)
	}

	if len(spec.Taints) > 0 {
		args.Taints = nodeTaints(spec.Taints)
	}

	if spec.CpuPolicy != nil {
		args.CpuPolicy = pulumi.String(*spec.CpuPolicy)
	}

	if spec.RuntimeName != "" {
		args.RuntimeName = pulumi.String(spec.RuntimeName)
	}

	if spec.RuntimeVersion != "" {
		args.RuntimeVersion = pulumi.String(spec.RuntimeVersion)
	}

	if spec.Unschedulable != nil {
		args.Unschedulable = pulumi.Bool(*spec.Unschedulable)
	}

	if spec.UserData != "" {
		args.UserData = pulumi.String(spec.UserData)
	}

	if spec.ScalingConfig != nil {
		args.ScalingConfig = scalingConfig(spec.ScalingConfig)
	}

	if spec.MultiAzPolicy != nil {
		args.MultiAzPolicy = pulumi.String(*spec.MultiAzPolicy)
	}

	if spec.Management != nil {
		args.Management = managementConfig(spec.Management)
	}

	if spec.SpotStrategy != nil {
		args.SpotStrategy = pulumi.String(*spec.SpotStrategy)
	}

	if len(spec.SpotPriceLimits) > 0 {
		args.SpotPriceLimits = spotPriceLimits(spec.SpotPriceLimits)
	}

	if spec.Period != nil {
		args.Period = pulumi.Int(int(*spec.Period))
	}

	if spec.AutoRenew != nil {
		args.AutoRenew = pulumi.Bool(*spec.AutoRenew)
	}

	if spec.AutoRenewPeriod != nil {
		args.AutoRenewPeriod = pulumi.Int(int(*spec.AutoRenewPeriod))
	}

	if spec.ResourceGroupId != "" {
		args.ResourceGroupId = pulumi.String(spec.ResourceGroupId)
	}

	if spec.RamRoleName != "" {
		args.RamRoleName = pulumi.String(spec.RamRoleName)
	}

	nodePool, err := cs.NewNodePool(ctx, spec.Name, args, pulumi.Provider(alicloudProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create ACK node pool %s", spec.Name)
	}

	ctx.Export(OpNodePoolId, nodePool.NodePoolId)
	ctx.Export(OpScalingGroupId, nodePool.ScalingGroupId)

	return nil
}

func dataDisks(disks []*alicloudkubernetesnodepoolv1.AlicloudKubernetesNodePoolDataDisk) cs.NodePoolDataDiskArray {
	result := cs.NodePoolDataDiskArray{}
	for _, d := range disks {
		disk := cs.NodePoolDataDiskArgs{
			Size: pulumi.Int(int(d.Size)),
		}
		category := "cloud_essd"
		if d.Category != nil {
			category = *d.Category
		}
		disk.Category = pulumi.String(category)
		if d.Name != "" {
			disk.Name = pulumi.String(d.Name)
		}
		if d.PerformanceLevel != "" {
			disk.PerformanceLevel = pulumi.String(d.PerformanceLevel)
		}
		if d.Encrypted != "" {
			disk.Encrypted = pulumi.String(d.Encrypted)
		}
		if d.KmsKeyId != "" {
			disk.KmsKeyId = pulumi.String(d.KmsKeyId)
		}
		result = append(result, disk)
	}
	return result
}

func nodeLabels(labels map[string]string) cs.NodePoolLabelArray {
	result := cs.NodePoolLabelArray{}
	for k, v := range labels {
		result = append(result, cs.NodePoolLabelArgs{
			Key:   pulumi.String(k),
			Value: pulumi.String(v),
		})
	}
	return result
}

func nodeTaints(taints []*alicloudkubernetesnodepoolv1.AlicloudKubernetesNodePoolTaint) cs.NodePoolTaintArray {
	result := cs.NodePoolTaintArray{}
	for _, t := range taints {
		taint := cs.NodePoolTaintArgs{
			Key: pulumi.String(t.Key),
		}
		if t.Value != "" {
			taint.Value = pulumi.String(t.Value)
		}
		if t.Effect != "" {
			taint.Effect = pulumi.String(t.Effect)
		}
		result = append(result, taint)
	}
	return result
}

func scalingConfig(sc *alicloudkubernetesnodepoolv1.AlicloudKubernetesNodePoolScalingConfig) cs.NodePoolScalingConfigPtrInput {
	enable := true
	if sc.Enable != nil {
		enable = *sc.Enable
	}
	args := cs.NodePoolScalingConfigArgs{
		Enable:  pulumi.Bool(enable),
		MinSize: pulumi.Int(int(sc.MinSize)),
		MaxSize: pulumi.Int(int(sc.MaxSize)),
	}
	if sc.Type != nil {
		args.Type = pulumi.String(*sc.Type)
	}
	return args.ToNodePoolScalingConfigPtrOutput()
}

func managementConfig(mgmt *alicloudkubernetesnodepoolv1.AlicloudKubernetesNodePoolManagement) cs.NodePoolManagementPtrInput {
	enable := true
	if mgmt.Enable != nil {
		enable = *mgmt.Enable
	}
	args := cs.NodePoolManagementArgs{
		Enable: pulumi.Bool(enable),
	}
	if mgmt.AutoRepair != nil {
		args.AutoRepair = pulumi.Bool(*mgmt.AutoRepair)
	}
	if mgmt.AutoUpgrade != nil {
		args.AutoUpgrade = pulumi.Bool(*mgmt.AutoUpgrade)
	}
	if mgmt.MaxUnavailable != nil {
		args.MaxUnavailable = pulumi.Int(int(*mgmt.MaxUnavailable))
	}
	return args.ToNodePoolManagementPtrOutput()
}

func spotPriceLimits(limits []*alicloudkubernetesnodepoolv1.AlicloudKubernetesNodePoolSpotPriceLimit) cs.NodePoolSpotPriceLimitArray {
	result := cs.NodePoolSpotPriceLimitArray{}
	for _, l := range limits {
		result = append(result, cs.NodePoolSpotPriceLimitArgs{
			InstanceType: pulumi.String(l.InstanceType),
			PriceLimit:   pulumi.String(l.PriceLimit),
		})
	}
	return result
}
