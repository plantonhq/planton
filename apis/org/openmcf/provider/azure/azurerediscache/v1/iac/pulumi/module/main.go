package module

import (
	"fmt"

	"github.com/pkg/errors"
	azurerediscachev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azurerediscache/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/redis"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azurerediscachev1.AzureRedisCacheStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	azureProviderConfig := stackInput.ProviderConfig

	// Create azure provider using the credentials from the input
	azureProvider, err := azure.NewProvider(ctx,
		"azure",
		&azure.ProviderArgs{
			ClientId:       pulumi.String(azureProviderConfig.ClientId),
			ClientSecret:   pulumi.String(azureProviderConfig.ClientSecret),
			SubscriptionId: pulumi.String(azureProviderConfig.SubscriptionId),
			TenantId:       pulumi.String(azureProviderConfig.TenantId),
		})
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureRedisCache.Spec

	// Build the Redis cache arguments.
	// The family ("C" or "P") is auto-derived from sku_name in locals.
	cacheArgs := &redis.CacheArgs{
		Name:                       pulumi.String(spec.Name),
		Location:                   pulumi.String(spec.Region),
		ResourceGroupName:          pulumi.String(locals.ResourceGroupName),
		SkuName:                    pulumi.String(spec.GetSkuName()),
		Family:                     pulumi.String(locals.Family),
		Capacity:                   pulumi.Int(int(spec.Capacity)),
		RedisVersion:               pulumi.StringPtr(spec.GetRedisVersion()),
		MinimumTlsVersion:          pulumi.StringPtr(spec.GetMinimumTlsVersion()),
		NonSslPortEnabled:          pulumi.BoolPtr(spec.GetNonSslPortEnabled()),
		PublicNetworkAccessEnabled: pulumi.BoolPtr(spec.GetPublicNetworkAccessEnabled()),
		Tags:                       pulumi.ToStringMap(locals.AzureTags),
		RedisConfiguration: &redis.CacheRedisConfigurationArgs{
			MaxmemoryPolicy: pulumi.StringPtr(spec.GetMaxmemoryPolicy()),
		},
	}

	// VNet injection (Premium SKU only).
	// When subnet_id is set, the cache is deployed inside the subnet with private IP.
	if spec.SubnetId != nil && spec.SubnetId.GetValue() != "" {
		cacheArgs.SubnetId = pulumi.StringPtr(spec.SubnetId.GetValue())
	}

	// Availability zones.
	if len(spec.Zones) > 0 {
		zoneArray := pulumi.StringArray{}
		for _, z := range spec.Zones {
			zoneArray = append(zoneArray, pulumi.String(z))
		}
		cacheArgs.Zones = zoneArray
	}

	// Redis Cluster sharding (Premium SKU only).
	if spec.ShardCount != nil {
		cacheArgs.ShardCount = pulumi.IntPtr(int(spec.GetShardCount()))
	}

	// Patch schedules for maintenance windows.
	if len(spec.PatchSchedules) > 0 {
		patchArray := redis.CachePatchScheduleArray{}
		for _, ps := range spec.PatchSchedules {
			patchArgs := &redis.CachePatchScheduleArgs{
				DayOfWeek: pulumi.String(ps.DayOfWeek),
			}
			if ps.StartHourUtc != nil {
				patchArgs.StartHourUtc = pulumi.IntPtr(int(ps.GetStartHourUtc()))
			}
			if ps.MaintenanceWindow != nil {
				patchArgs.MaintenanceWindow = pulumi.StringPtr(ps.GetMaintenanceWindow())
			}
			patchArray = append(patchArray, patchArgs)
		}
		cacheArgs.PatchSchedules = patchArray
	}

	// Create the Redis cache.
	cache, err := redis.NewCache(ctx,
		spec.Name,
		cacheArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create Redis cache %s", spec.Name)
	}

	// Create firewall rules.
	// Only effective when public access is enabled and cache is not VNet-injected.
	for _, rule := range spec.FirewallRules {
		_, err := redis.NewFirewallRule(ctx,
			fmt.Sprintf("%s-%s", spec.Name, rule.Name),
			&redis.FirewallRuleArgs{
				Name:              pulumi.String(rule.Name),
				RedisCacheName:    cache.Name,
				ResourceGroupName: pulumi.String(locals.ResourceGroupName),
				StartIp:           pulumi.String(rule.StartIp),
				EndIp:             pulumi.String(rule.EndIp),
			},
			pulumi.Provider(azureProvider),
			pulumi.DependsOn([]pulumi.Resource{cache}))
		if err != nil {
			return errors.Wrapf(err, "failed to create firewall rule %s", rule.Name)
		}
	}

	// Export stack outputs.
	ctx.Export(OpRedisId, cache.ID())
	ctx.Export(OpHostname, cache.Hostname)
	ctx.Export(OpSslPort, cache.SslPort)
	ctx.Export(OpPrimaryAccessKey, cache.PrimaryAccessKey)
	ctx.Export(OpPrimaryConnectionString, cache.PrimaryConnectionString)

	return nil
}
