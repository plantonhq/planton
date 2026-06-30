package module

import (
	"github.com/pkg/errors"
	alicloudredisinstancev1 "github.com/plantonhq/planton/apis/dev/planton/provider/alicloud/alicloudredisinstance/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/kvstore"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudredisinstancev1.AliCloudRedisInstanceStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AliCloudRedisInstance.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	name := instanceName(locals)

	instanceArgs := &kvstore.InstanceArgs{
		InstanceClass:  pulumi.String(spec.InstanceClass),
		Password:       pulumi.String(spec.Password),
		EngineVersion:  pulumi.String(engineVersion(spec)),
		InstanceType:   pulumi.String(instanceType(spec)),
		DbInstanceName: pulumi.String(name),
		PaymentType:    pulumi.String(paymentType(spec)),
		VswitchId:      pulumi.String(spec.VswitchId.GetValue()),
		VpcAuthMode:    pulumi.String(vpcAuthMode(spec)),
		Tags:           pulumi.ToStringMap(locals.Tags),
	}

	if spec.ZoneId != "" {
		instanceArgs.ZoneId = pulumi.String(spec.ZoneId)
	}

	if spec.SecondaryZoneId != "" {
		instanceArgs.SecondaryZoneId = pulumi.String(spec.SecondaryZoneId)
	}

	if len(spec.SecurityIps) > 0 {
		instanceArgs.SecurityIps = pulumi.ToStringArray(spec.SecurityIps)
	}

	if spec.SecurityGroupId != "" {
		instanceArgs.SecurityGroupId = pulumi.String(spec.SecurityGroupId)
	}

	if spec.ResourceGroupId != "" {
		instanceArgs.ResourceGroupId = pulumi.String(spec.ResourceGroupId)
	}

	instanceArgs.ShardCount = optionalInt(spec.ShardCount)
	instanceArgs.ReadOnlyCount = optionalInt(spec.ReadOnlyCount)
	instanceArgs.SslEnable = optionalStringPtr(spec.SslEnable)
	instanceArgs.TdeStatus = optionalStringPtr(spec.TdeStatus)

	if spec.EncryptionKey != "" {
		instanceArgs.EncryptionKey = pulumi.String(spec.EncryptionKey)
	}

	if len(spec.Config) > 0 {
		instanceArgs.Config = pulumi.ToStringMap(spec.Config)
	}

	instanceArgs.InstanceReleaseProtection = optionalBool(spec.InstanceReleaseProtection)

	if spec.MaintainStartTime != "" {
		instanceArgs.MaintainStartTime = pulumi.String(spec.MaintainStartTime)
	}

	if spec.MaintainEndTime != "" {
		instanceArgs.MaintainEndTime = pulumi.String(spec.MaintainEndTime)
	}

	if len(spec.BackupPeriod) > 0 {
		instanceArgs.BackupPeriods = pulumi.ToStringArray(spec.BackupPeriod)
	}

	if spec.BackupTime != "" {
		instanceArgs.BackupTime = pulumi.String(spec.BackupTime)
	}

	if spec.PrivateConnectionPrefix != "" {
		instanceArgs.PrivateConnectionPrefix = pulumi.String(spec.PrivateConnectionPrefix)
	}

	instanceArgs.AutoRenew = optionalBool(spec.AutoRenew)
	instanceArgs.AutoRenewPeriod = optionalInt(spec.AutoRenewPeriod)
	instanceArgs.Period = optionalStringPtr(spec.Period)

	instance, err := kvstore.NewInstance(ctx, name, instanceArgs, pulumi.Provider(alicloudProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Redis instance %s", name)
	}

	ctx.Export(OpInstanceId, instance.ID())
	ctx.Export(OpConnectionDomain, instance.ConnectionDomain)
	ctx.Export(OpPrivateConnectionPort, instance.PrivateConnectionPort)
	ctx.Export(OpPrivateIp, instance.PrivateIp)

	return nil
}
