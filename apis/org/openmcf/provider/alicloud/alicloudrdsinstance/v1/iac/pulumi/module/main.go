package module

import (
	"github.com/pkg/errors"
	alicloudrdsinstancev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudrdsinstance/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/rds"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudrdsinstancev1.AlicloudRdsInstanceStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AlicloudRdsInstance.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	name := instanceName(locals)

	instanceArgs := &rds.InstanceArgs{
		Engine:             pulumi.String(spec.Engine),
		EngineVersion:      pulumi.String(spec.EngineVersion),
		InstanceType:       pulumi.String(spec.InstanceType),
		InstanceStorage:    pulumi.Int(int(spec.InstanceStorage)),
		VswitchId:          pulumi.String(spec.VswitchId.GetValue()),
		InstanceName:       pulumi.String(name),
		InstanceChargeType: pulumi.String(instanceChargeType(spec)),
		Category:           pulumi.String(category(spec)),
		Tags:               pulumi.ToStringMap(locals.Tags),
	}

	if spec.DbInstanceStorageType != nil && *spec.DbInstanceStorageType != "" {
		instanceArgs.DbInstanceStorageType = pulumi.String(*spec.DbInstanceStorageType)
	}

	if spec.ZoneId != "" {
		instanceArgs.ZoneId = pulumi.String(spec.ZoneId)
	}

	if spec.ZoneIdSlaveA != "" {
		instanceArgs.ZoneIdSlaveA = pulumi.String(spec.ZoneIdSlaveA)
	}

	if len(spec.SecurityIps) > 0 {
		instanceArgs.SecurityIps = pulumi.ToStringArray(spec.SecurityIps)
	}

	if len(spec.SecurityGroupIds) > 0 {
		instanceArgs.SecurityGroupIds = pulumi.ToStringArray(spec.SecurityGroupIds)
	}

	instanceArgs.MonitoringPeriod = optionalInt(spec.MonitoringPeriod)

	if spec.MaintainTime != "" {
		instanceArgs.MaintainTime = pulumi.String(spec.MaintainTime)
	}

	instanceArgs.DeletionProtection = optionalBool(spec.DeletionProtection)
	instanceArgs.SslAction = optionalString(sslAction(spec))
	instanceArgs.TdeStatus = optionalString(tdeAction(spec))

	if spec.EncryptionKey != "" {
		instanceArgs.EncryptionKey = pulumi.String(spec.EncryptionKey)
	}

	instanceArgs.AutoRenew = optionalBool(spec.AutoRenew)
	instanceArgs.AutoRenewPeriod = optionalInt(spec.AutoRenewPeriod)
	instanceArgs.Period = optionalInt(spec.Period)

	if spec.ResourceGroupId != "" {
		instanceArgs.ResourceGroupId = pulumi.String(spec.ResourceGroupId)
	}

	if len(spec.Parameters) > 0 {
		params := rds.InstanceParameterArray{}
		for _, p := range spec.Parameters {
			params = append(params, rds.InstanceParameterArgs{
				Name:  pulumi.String(p.Name),
				Value: pulumi.String(p.Value),
			})
		}
		instanceArgs.Parameters = params
	}

	instance, err := rds.NewInstance(ctx, name, instanceArgs, pulumi.Provider(alicloudProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create RDS instance %s", name)
	}

	databaseIdMap := pulumi.StringMap{}

	for _, db := range spec.Databases {
		created, err := database(ctx, alicloudProvider, instance, spec.Engine, db)
		if err != nil {
			return err
		}
		databaseIdMap[db.Name] = created.ID()
	}

	for _, acct := range spec.Accounts {
		if err := account(ctx, alicloudProvider, instance, acct); err != nil {
			return err
		}
	}

	ctx.Export(OpInstanceId, instance.ID())
	ctx.Export(OpConnectionString, instance.ConnectionString)
	ctx.Export(OpPort, instance.Port)
	ctx.Export(OpDatabaseIds, databaseIdMap)

	return nil
}

func sslAction(spec *alicloudrdsinstancev1.AlicloudRdsInstanceSpec) string {
	if spec.SslAction != nil {
		return *spec.SslAction
	}
	return ""
}

func tdeAction(spec *alicloudrdsinstancev1.AlicloudRdsInstanceSpec) string {
	if spec.TdeStatus != nil {
		return *spec.TdeStatus
	}
	return ""
}
