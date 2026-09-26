package module

import (
	"fmt"

	"github.com/pkg/errors"
	azurerediscachev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurerediscache/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/redis"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azurerediscachev1alpha1.AzureRedisCacheStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the Azure provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureRedisCache.Spec

	// The cache itself. Azure sizes it as {family}{capacity} within the
	// tier; the family letter is derived from the tier in locals so the
	// spec never spells the same fact twice. Provisioning is the slowest
	// in the Azure catalog -- 15-40 minutes is normal.
	cacheArgs := &redis.CacheArgs{
		Name:              pulumi.String(spec.CacheName),
		Location:          pulumi.String(spec.Region),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		SkuName:           pulumi.String(locals.SkuName),
		Family:            pulumi.String(locals.Family),
		Capacity:          pulumi.Int(int(spec.Capacity)),
		// The plaintext port sends commands AND keys unencrypted -- off
		// unless a legacy client genuinely cannot speak TLS.
		NonSslPortEnabled: pulumi.BoolPtr(spec.NonSslPortEnabled),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// Presence-guarded proto defaults: stack inputs never materialize
	// them, so an unset field must deploy the spec's documented default,
	// not the Go zero value.
	redisVersion := "6"
	if spec.RedisVersion != nil {
		redisVersion = spec.GetRedisVersion()
	}
	cacheArgs.RedisVersion = pulumi.StringPtr(redisVersion)

	publicNetworkAccess := true
	if spec.PublicNetworkAccessEnabled != nil {
		publicNetworkAccess = spec.GetPublicNetworkAccessEnabled()
	}
	// false forces all traffic through Private Link (AzurePrivateEndpoint).
	cacheArgs.PublicNetworkAccessEnabled = pulumi.BoolPtr(publicNetworkAccess)

	// The keyless posture: keys can only be turned off once Entra auth is
	// on (spec-enforced, mirroring ARM's own contract); identities then
	// connect with tokens under access-policy assignments.
	accessKeysAuth := true
	if spec.AccessKeysAuthenticationEnabled != nil {
		accessKeysAuth = spec.GetAccessKeysAuthenticationEnabled()
	}
	cacheArgs.AccessKeysAuthenticationEnabled = pulumi.BoolPtr(accessKeysAuth)

	// VNet injection (Premium): the cache gets a private IP inside a
	// subnet dedicated to Redis. The legacy isolation mechanism --
	// Private Link is the recommendation for new designs; both are
	// modeled. Fixed at creation.
	if spec.SubnetId.GetValue() != "" {
		cacheArgs.SubnetId = pulumi.StringPtr(spec.SubnetId.GetValue())
	}
	if spec.PrivateStaticIpAddress != nil {
		cacheArgs.PrivateStaticIpAddress = pulumi.StringPtr(spec.GetPrivateStaticIpAddress())
	}

	// Zone pinning for datacenter-failure resilience. Fixed at creation.
	if len(spec.Zones) > 0 {
		cacheArgs.Zones = pulumi.ToStringArray(spec.Zones)
	}

	// Clustering (Premium): each shard is a primary/replica pair, so
	// memory and throughput scale with the shard count. Mutually
	// exclusive with extra replicas (spec-enforced).
	if spec.ShardCount != nil {
		cacheArgs.ShardCount = pulumi.IntPtr(int(spec.GetShardCount()))
	}

	// Extra read replicas per primary (Premium). Only ARM's modern name
	// is modeled; the legacy replicas_per_master alias mirrors it
	// server-side.
	if spec.ReplicasPerPrimary != nil {
		cacheArgs.ReplicasPerPrimary = pulumi.IntPtr(int(spec.GetReplicasPerPrimary()))
	}

	// Tenant-level platform settings -- distinct from redis_configuration
	// (the Redis engine's own settings); used by support scenarios.
	if len(spec.TenantSettings) > 0 {
		cacheArgs.TenantSettings = pulumi.ToStringMap(spec.TenantSettings)
	}

	// The managed identity, used for keyless persistence-storage access
	// (data_persistence_authentication_method = MANAGED_IDENTITY).
	if spec.Identity != nil {
		identityArgs := &redis.CacheIdentityArgs{
			Type: pulumi.String(identityTypeStrings[spec.Identity.Type]),
		}
		if len(spec.Identity.UserAssignedIdentityIds) > 0 {
			identityIds := pulumi.StringArray{}
			for _, id := range spec.Identity.UserAssignedIdentityIds {
				identityIds = append(identityIds, pulumi.String(id.GetValue()))
			}
			identityArgs.IdentityIds = identityIds
		}
		cacheArgs.Identity = identityArgs
	}

	// Redis engine and platform behavior. The block is emitted only when
	// the spec carries it, so an omitted block deploys Azure's defaults
	// identically on both engines. Unset attributes inside the block are
	// not sent -- Azure sizes the memory dials from total memory.
	if spec.RedisConfiguration != nil {
		configuration := spec.RedisConfiguration
		configArgs := &redis.CacheRedisConfigurationArgs{
			ActiveDirectoryAuthenticationEnabled: pulumi.BoolPtr(configuration.ActiveDirectoryAuthenticationEnabled),
		}

		// Presence-guarded proto default: unset deploys volatile-lru.
		maxmemoryPolicy := "volatile-lru"
		if configuration.MaxmemoryPolicy != nil {
			maxmemoryPolicy = configuration.GetMaxmemoryPolicy()
		}
		configArgs.MaxmemoryPolicy = pulumi.StringPtr(maxmemoryPolicy)

		if configuration.MaxmemoryReserved != nil {
			configArgs.MaxmemoryReserved = pulumi.IntPtr(int(configuration.GetMaxmemoryReserved()))
		}
		if configuration.MaxmemoryDelta != nil {
			configArgs.MaxmemoryDelta = pulumi.IntPtr(int(configuration.GetMaxmemoryDelta()))
		}
		if configuration.MaxfragmentationmemoryReserved != nil {
			configArgs.MaxfragmentationmemoryReserved = pulumi.IntPtr(int(configuration.GetMaxfragmentationmemoryReserved()))
		}
		if configuration.NotifyKeyspaceEvents != "" {
			configArgs.NotifyKeyspaceEvents = pulumi.StringPtr(configuration.NotifyKeyspaceEvents)
		}

		// Presence-guarded proto default: unset deploys true. false is
		// only legal inside a VNet-injected cache (spec-enforced); the
		// provider only transmits the setting when subnet_id is set.
		authenticationEnabled := true
		if configuration.AuthenticationEnabled != nil {
			authenticationEnabled = configuration.GetAuthenticationEnabled()
		}
		configArgs.AuthenticationEnabled = pulumi.BoolPtr(authenticationEnabled)

		if configuration.DataPersistenceAuthenticationMethod != azurerediscachev1alpha1.AzureRedisCachePersistenceAuthMethod_azure_redis_cache_persistence_auth_method_unspecified {
			configArgs.DataPersistenceAuthenticationMethod = pulumi.StringPtr(persistenceAuthStrings[configuration.DataPersistenceAuthenticationMethod])
		}

		// RDB snapshots (Premium): periodic full dumps to a storage
		// account. The connection string is secret-bearing and never
		// echoed back by ARM.
		configArgs.RdbBackupEnabled = pulumi.BoolPtr(configuration.RdbBackupEnabled)
		if configuration.RdbBackupFrequency != nil {
			configArgs.RdbBackupFrequency = pulumi.IntPtr(int(configuration.GetRdbBackupFrequency()))
		}
		if configuration.RdbBackupMaxSnapshotCount != nil {
			configArgs.RdbBackupMaxSnapshotCount = pulumi.IntPtr(int(configuration.GetRdbBackupMaxSnapshotCount()))
		}
		if configuration.RdbStorageConnectionString != "" {
			configArgs.RdbStorageConnectionString = pulumi.StringPtr(configuration.RdbStorageConnectionString)
		}

		// AOF log (Premium): near-synchronous write logging for tight
		// recovery points; Azure alternates between the two accounts
		// during storage maintenance.
		configArgs.AofBackupEnabled = pulumi.BoolPtr(configuration.AofBackupEnabled)
		if configuration.AofStorageConnectionString_0 != "" {
			configArgs.AofStorageConnectionString0 = pulumi.StringPtr(configuration.AofStorageConnectionString_0)
		}
		if configuration.AofStorageConnectionString_1 != "" {
			configArgs.AofStorageConnectionString1 = pulumi.StringPtr(configuration.AofStorageConnectionString_1)
		}

		if configuration.StorageAccountSubscriptionId != nil {
			configArgs.StorageAccountSubscriptionId = pulumi.StringPtr(configuration.GetStorageAccountSubscriptionId())
		}

		cacheArgs.RedisConfiguration = configArgs
	}

	// Weekly maintenance windows during which Azure may patch the Redis
	// engine and platform.
	if len(spec.PatchSchedules) > 0 {
		patchArray := redis.CachePatchScheduleArray{}
		for _, schedule := range spec.PatchSchedules {
			patchArgs := &redis.CachePatchScheduleArgs{
				DayOfWeek: pulumi.String(dayOfWeekStrings[schedule.DayOfWeek]),
			}
			// Presence-guarded proto defaults (0 and PT5H).
			startHour := 0
			if schedule.StartHourUtc != nil {
				startHour = int(schedule.GetStartHourUtc())
			}
			patchArgs.StartHourUtc = pulumi.IntPtr(startHour)

			window := "PT5H"
			if schedule.MaintenanceWindow != nil {
				window = schedule.GetMaintenanceWindow()
			}
			patchArgs.MaintenanceWindow = pulumi.StringPtr(window)

			patchArray = append(patchArray, patchArgs)
		}
		cacheArgs.PatchSchedules = patchArray
	}

	createdCache, err := redis.NewCache(ctx,
		spec.CacheName,
		cacheArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create redis cache %s", spec.CacheName)
	}

	// Public-endpoint IP allow-list. One ARM sub-resource per rule; only
	// effective while public network access is on and the cache is not
	// VNet-injected. ARM rejects hyphens in rule names (spec-enforced).
	for _, rule := range spec.FirewallRules {
		_, err := redis.NewFirewallRule(ctx,
			fmt.Sprintf("%s-%s", spec.CacheName, rule.Name),
			&redis.FirewallRuleArgs{
				Name:              pulumi.String(rule.Name),
				RedisCacheName:    createdCache.Name,
				ResourceGroupName: pulumi.String(locals.ResourceGroupName),
				StartIp:           pulumi.String(rule.StartIp),
				EndIp:             pulumi.String(rule.EndIp),
			},
			pulumi.Provider(azureProvider),
			pulumi.DependsOn([]pulumi.Resource{createdCache}))
		if err != nil {
			return errors.Wrapf(err, "failed to create firewall rule %s", rule.Name)
		}
	}

	// The system-assigned identity's principal, when one exists -- what
	// RBAC grants target (e.g. Storage Blob Data Contributor for
	// managed-identity persistence).
	identityPrincipalId := createdCache.Identity.PrincipalId().ApplyT(func(principalId *string) string {
		if principalId == nil {
			return ""
		}
		return *principalId
	}).(pulumi.StringOutput)

	// Export stack outputs. The keys and connection strings are
	// secret-bearing; region is the linked-server location seam.
	ctx.Export(OpRedisCacheId, createdCache.ID())
	ctx.Export(OpRedisCacheName, createdCache.Name)
	ctx.Export(OpRegion, createdCache.Location)
	ctx.Export(OpResourceGroupName, createdCache.ResourceGroupName)
	ctx.Export(OpHostname, createdCache.Hostname)
	ctx.Export(OpPort, createdCache.Port)
	ctx.Export(OpSslPort, createdCache.SslPort)
	ctx.Export(OpPrimaryAccessKey, pulumi.ToSecret(createdCache.PrimaryAccessKey))
	ctx.Export(OpSecondaryAccessKey, pulumi.ToSecret(createdCache.SecondaryAccessKey))
	ctx.Export(OpPrimaryConnectionString, pulumi.ToSecret(createdCache.PrimaryConnectionString))
	ctx.Export(OpSecondaryConnectionString, pulumi.ToSecret(createdCache.SecondaryConnectionString))
	ctx.Export(OpIdentityPrincipalId, identityPrincipalId)

	return nil
}
