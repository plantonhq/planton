package module

import (
	"github.com/pkg/errors"
	alicloudmongodbinstancev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudmongodbinstance/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/mongodb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudmongodbinstancev1.AlicloudMongodbInstanceStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AlicloudMongodbInstance.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	name := instanceName(locals)

	instanceArgs := &mongodb.InstanceArgs{
		EngineVersion:      pulumi.String(spec.EngineVersion),
		DbInstanceClass:    pulumi.String(spec.DbInstanceClass),
		DbInstanceStorage:  pulumi.Int(int(spec.DbInstanceStorage)),
		AccountPassword:    pulumi.String(spec.AccountPassword),
		VswitchId:          pulumi.String(spec.VswitchId.GetValue()),
		Name:               pulumi.String(name),
		ReplicationFactor:  pulumi.Int(replicationFactor(spec)),
		StorageEngine:      pulumi.String(storageEngine(spec)),
		InstanceChargeType: pulumi.String(instanceChargeType(spec)),
		Tags:               pulumi.ToStringMap(locals.Tags),
	}

	if spec.ZoneId != "" {
		instanceArgs.ZoneId = pulumi.String(spec.ZoneId)
	}

	if spec.SecondaryZoneId != "" {
		instanceArgs.SecondaryZoneId = pulumi.String(spec.SecondaryZoneId)
	}

	if spec.HiddenZoneId != "" {
		instanceArgs.HiddenZoneId = pulumi.String(spec.HiddenZoneId)
	}

	instanceArgs.ReadonlyReplicas = optionalInt(spec.ReadonlyReplicas)
	instanceArgs.StorageType = optionalStringPtr(spec.StorageType)
	instanceArgs.ProvisionedIops = optionalInt(spec.ProvisionedIops)

	if len(spec.SecurityIpList) > 0 {
		instanceArgs.SecurityIpLists = pulumi.ToStringArray(spec.SecurityIpList)
	}

	if spec.SecurityGroupId != "" {
		instanceArgs.SecurityGroupId = pulumi.String(spec.SecurityGroupId)
	}

	if spec.ResourceGroupId != "" {
		instanceArgs.ResourceGroupId = pulumi.String(spec.ResourceGroupId)
	}

	instanceArgs.SslAction = optionalStringPtr(spec.SslAction)
	instanceArgs.TdeStatus = optionalStringPtr(spec.TdeStatus)

	if spec.EncryptionKey != "" {
		instanceArgs.EncryptionKey = pulumi.String(spec.EncryptionKey)
	}

	instanceArgs.Encrypted = optionalBool(spec.Encrypted)

	if spec.CloudDiskEncryptionKey != "" {
		instanceArgs.CloudDiskEncryptionKey = pulumi.String(spec.CloudDiskEncryptionKey)
	}

	if spec.MaintainStartTime != "" {
		instanceArgs.MaintainStartTime = pulumi.String(spec.MaintainStartTime)
	}

	if spec.MaintainEndTime != "" {
		instanceArgs.MaintainEndTime = pulumi.String(spec.MaintainEndTime)
	}

	if spec.BackupTime != "" {
		instanceArgs.BackupTime = pulumi.String(spec.BackupTime)
	}

	if len(spec.BackupPeriod) > 0 {
		instanceArgs.BackupPeriods = pulumi.ToStringArray(spec.BackupPeriod)
	}

	if len(spec.Parameters) > 0 {
		params := mongodb.InstanceParameterArray{}
		for k, v := range spec.Parameters {
			params = append(params, &mongodb.InstanceParameterArgs{
				Name:  pulumi.String(k),
				Value: pulumi.String(v),
			})
		}
		instanceArgs.Parameters = params
	}

	instanceArgs.DbInstanceReleaseProtection = optionalBool(spec.DbInstanceReleaseProtection)
	instanceArgs.Period = optionalInt(spec.Period)
	instanceArgs.AutoRenew = optionalBool(spec.AutoRenew)
	instanceArgs.AutoRenewDuration = optionalInt(spec.AutoRenewDuration)

	instance, err := mongodb.NewInstance(ctx, name, instanceArgs, pulumi.Provider(alicloudProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create MongoDB instance %s", name)
	}

	ctx.Export(OpInstanceId, instance.ID())
	ctx.Export(OpReplicaSetName, instance.ReplicaSetName)

	return nil
}
