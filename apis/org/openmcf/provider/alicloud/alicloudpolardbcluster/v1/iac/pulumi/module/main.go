package module

import (
	"github.com/pkg/errors"
	alicloudpolardbclusterv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/alicloud/alicloudpolardbcluster/v1"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/polardb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *alicloudpolardbclusterv1.AlicloudPolardbClusterStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	spec := locals.AlicloudPolardbCluster.Spec

	alicloudProvider, err := alicloud.NewProvider(ctx, "alicloud", &alicloud.ProviderArgs{
		Region: pulumi.String(spec.Region),
	})
	if err != nil {
		return errors.Wrap(err, "failed to create alicloud provider")
	}

	clusterArgs := &polardb.ClusterArgs{
		DbType:      pulumi.String(spec.DbType),
		DbVersion:   pulumi.String(spec.DbVersion),
		DbNodeClass: pulumi.String(spec.DbNodeClass),
		VswitchId:   pulumi.String(spec.VswitchId.GetValue()),
		DbNodeCount: pulumi.Int(dbNodeCount(spec)),
		Description: pulumi.String(clusterDescription(locals)),
		PayType:     pulumi.String(payType(spec)),
		Tags:        pulumi.ToStringMap(locals.Tags),
	}

	if spec.Period != nil {
		clusterArgs.Period = pulumi.Int(int(*spec.Period))
	}

	clusterArgs.RenewalStatus = optionalStringPtr(spec.RenewalStatus)

	if spec.AutoRenewPeriod != nil {
		clusterArgs.AutoRenewPeriod = pulumi.Int(int(*spec.AutoRenewPeriod))
	}

	if spec.ZoneId != "" {
		clusterArgs.ZoneId = pulumi.String(spec.ZoneId)
	}

	if len(spec.SecurityIps) > 0 {
		clusterArgs.SecurityIps = pulumi.ToStringArray(spec.SecurityIps)
	}

	if len(spec.SecurityGroupIds) > 0 {
		clusterArgs.SecurityGroupIds = pulumi.ToStringArray(spec.SecurityGroupIds)
	}

	if spec.MaintainTime != "" {
		clusterArgs.MaintainTime = pulumi.String(spec.MaintainTime)
	}

	if spec.ResourceGroupId != "" {
		clusterArgs.ResourceGroupId = pulumi.String(spec.ResourceGroupId)
	}

	clusterArgs.CreationCategory = optionalStringPtr(spec.CreationCategory)
	clusterArgs.SubCategory = optionalStringPtr(spec.SubCategory)
	clusterArgs.StorageType = optionalStringPtr(spec.StorageType)

	if spec.StorageSpace != nil {
		clusterArgs.StorageSpace = pulumi.Int(int(*spec.StorageSpace))
	}

	clusterArgs.TdeStatus = optionalStringPtr(spec.TdeStatus)

	if spec.EncryptionKey != "" {
		clusterArgs.EncryptionKey = pulumi.String(spec.EncryptionKey)
	}

	if spec.DeletionLock != nil {
		clusterArgs.DeletionLock = pulumi.Int(int(*spec.DeletionLock))
	}

	clusterArgs.CollectorStatus = optionalStringPtr(spec.CollectorStatus)
	clusterArgs.BackupRetentionPolicyOnClusterDeletion = optionalStringPtr(spec.BackupRetentionPolicyOnClusterDeletion)

	if len(spec.Parameters) > 0 {
		params := polardb.ClusterParameterArray{}
		for _, p := range spec.Parameters {
			params = append(params, polardb.ClusterParameterArgs{
				Name:  pulumi.String(p.Name),
				Value: pulumi.String(p.Value),
			})
		}
		clusterArgs.Parameters = params
	}

	cluster, err := polardb.NewCluster(ctx, locals.AlicloudPolardbCluster.Metadata.Name, clusterArgs,
		pulumi.Provider(alicloudProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create PolarDB cluster")
	}

	databaseIdMap := pulumi.StringMap{}

	for _, db := range spec.Databases {
		created, err := database(ctx, alicloudProvider, cluster, spec.DbType, db)
		if err != nil {
			return err
		}
		databaseIdMap[db.DbName] = created.ID()
	}

	for _, acct := range spec.Accounts {
		if err := account(ctx, alicloudProvider, cluster, acct); err != nil {
			return err
		}
	}

	ctx.Export(OpClusterId, cluster.ID())
	ctx.Export(OpConnectionString, cluster.ConnectionString)
	ctx.Export(OpPort, cluster.Port)
	ctx.Export(OpDatabaseIds, databaseIdMap)

	return nil
}
