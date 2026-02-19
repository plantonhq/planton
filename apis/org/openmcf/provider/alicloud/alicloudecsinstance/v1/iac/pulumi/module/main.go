package module

import (
	"github.com/pkg/errors"
	alicloudecsinstancev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudecsinstance/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudecsinstancev1.AlicloudEcsInstanceStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AlicloudEcsInstance.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	name := instanceName(locals)

	securityGroups := pulumi.StringArray{}
	for _, sgRef := range spec.SecurityGroupIds {
		securityGroups = append(securityGroups, pulumi.String(sgRef.GetValue()))
	}

	instanceArgs := &ecs.InstanceArgs{
		InstanceType:       pulumi.String(spec.InstanceType),
		ImageId:            pulumi.String(spec.ImageId),
		VswitchId:          pulumi.String(spec.VswitchId.GetValue()),
		SecurityGroups:     securityGroups,
		InstanceName:       pulumi.String(name),
		InstanceChargeType: pulumi.String(instanceChargeType(spec)),
		SystemDiskCategory: pulumi.String(systemDiskCategory(spec)),
		SystemDiskSize:     pulumi.Int(systemDiskSize(spec)),
		Tags:               pulumi.ToStringMap(locals.Tags),
	}

	if spec.HostName != "" {
		instanceArgs.HostName = pulumi.String(spec.HostName)
	}

	if spec.Description != "" {
		instanceArgs.Description = pulumi.String(spec.Description)
	}

	if spec.SystemDisk != nil {
		instanceArgs.SystemDiskPerformanceLevel = optionalStringPtr(spec.SystemDisk.PerformanceLevel)
		instanceArgs.SystemDiskEncrypted = optionalBool(spec.SystemDisk.Encrypted)

		if spec.SystemDisk.KmsKeyId != "" {
			instanceArgs.SystemDiskKmsKeyId = pulumi.String(spec.SystemDisk.KmsKeyId)
		}
	}

	if len(spec.DataDisks) > 0 {
		instanceArgs.DataDisks = buildDataDisks(spec.DataDisks)
	}

	if spec.KeyName != "" {
		instanceArgs.KeyName = pulumi.String(spec.KeyName)
	}

	if spec.Password != "" {
		instanceArgs.Password = pulumi.String(spec.Password)
	}

	instanceArgs.InternetMaxBandwidthOut = optionalInt(spec.InternetMaxBandwidthOut)
	instanceArgs.InternetChargeType = optionalStringPtr(spec.InternetChargeType)
	instanceArgs.Period = optionalInt(spec.Period)
	instanceArgs.PeriodUnit = optionalStringPtr(spec.PeriodUnit)
	instanceArgs.SpotStrategy = optionalStringPtr(spec.SpotStrategy)
	instanceArgs.SpotPriceLimit = optionalFloat64Ptr(spec.SpotPriceLimit)

	if spec.UserData != "" {
		instanceArgs.UserData = pulumi.String(spec.UserData)
	}

	if spec.RoleName != "" {
		instanceArgs.RoleName = pulumi.String(spec.RoleName)
	}

	instanceArgs.DeletionProtection = optionalBool(spec.DeletionProtection)
	instanceArgs.SecurityEnhancementStrategy = optionalStringPtr(spec.SecurityEnhancementStrategy)

	if spec.ResourceGroupId != "" {
		instanceArgs.ResourceGroupId = pulumi.String(spec.ResourceGroupId)
	}

	instance, err := ecs.NewInstance(ctx, name, instanceArgs, pulumi.Provider(alicloudProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create ECS instance %s", name)
	}

	ctx.Export(OpInstanceId, instance.ID())
	ctx.Export(OpPrivateIp, instance.PrivateIp)
	ctx.Export(OpPublicIp, instance.PublicIp)

	return nil
}

func buildDataDisks(disks []*alicloudecsinstancev1.AlicloudEcsDataDisk) ecs.InstanceDataDiskArray {
	result := ecs.InstanceDataDiskArray{}
	for _, disk := range disks {
		d := &ecs.InstanceDataDiskArgs{
			Size:             pulumi.Int(int(disk.Size)),
			Category:         pulumi.String(dataDiskCategory(disk)),
			DeleteWithInstance: pulumi.Bool(dataDiskDeleteWithInstance(disk)),
		}

		if disk.Name != "" {
			d.Name = pulumi.String(disk.Name)
		}

		if disk.PerformanceLevel != nil && *disk.PerformanceLevel != "" {
			d.PerformanceLevel = pulumi.String(*disk.PerformanceLevel)
		}

		if disk.Encrypted != nil {
			d.Encrypted = pulumi.Bool(*disk.Encrypted)
		}

		if disk.KmsKeyId != "" {
			d.KmsKeyId = pulumi.String(disk.KmsKeyId)
		}

		if disk.SnapshotId != "" {
			d.SnapshotId = pulumi.String(disk.SnapshotId)
		}

		if disk.Description != "" {
			d.Description = pulumi.String(disk.Description)
		}

		result = append(result, d)
	}
	return result
}
