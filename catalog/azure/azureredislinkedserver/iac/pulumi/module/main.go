package module

import (
	"regexp"

	"github.com/pkg/errors"
	azureredislinkedserverv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureredislinkedserver/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/redis"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// redisCacheIdPattern parses a Redis cache ARM ID into its resource group
// and cache name. The type segments are matched case-insensitively: ARM
// has emitted both .../Microsoft.Cache/Redis/{name} and .../redis/{name}
// over the API's life, and ARM ID comparison is case-insensitive on type
// segments (matching the Terraform module's regex semantics).
var redisCacheIdPattern = regexp.MustCompile(`(?i)/resourcegroups/([^/]+)/.*/redis/([^/]+)$`)

func Resources(ctx *pulumi.Context, iacInput *azureredislinkedserverv1alpha1.AzureRedisLinkedServerIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureRedisLinkedServer.Spec

	// The primary cache's name and resource group, parsed from its ARM ID
	// -- the link is that cache's child, so neither is ever spelled
	// twice. A malformed id fails loudly here instead of deploying into
	// the wrong scope.
	matches := redisCacheIdPattern.FindStringSubmatch(locals.TargetRedisCacheId)
	if matches == nil {
		return errors.Errorf("target_redis_cache_id %q is not a redis cache ARM id", locals.TargetRedisCacheId)
	}
	targetResourceGroupName := matches[1]
	targetCacheName := matches[2]

	// The geo-replication link between two PREMIUM caches. Azure names
	// the link after the LINKED (secondary) cache -- there is no name
	// argument. Every argument is ForceNew; replacing the link
	// re-establishes replication without touching cached data on the
	// primary.
	//
	// DELETING this resource IS the failover operation: unlinking makes
	// the secondary writable. Applications that point at the
	// geo_replicated_primary_host_name output (instead of either cache's
	// own hostname) keep working across failovers without a config
	// change.
	//
	// Azure's requirements, enforced at link time: both caches PREMIUM,
	// in different regions, and the secondary at least as large as the
	// primary. Establishing the link takes several minutes on top of the
	// caches' own provisioning; the secondary rejects writes while
	// linked. No tags: ARM does not support tags on linked servers.
	createdLinkedServer, err := redis.NewLinkedServer(ctx,
		targetCacheName,
		&redis.LinkedServerArgs{
			TargetRedisCacheName: pulumi.String(targetCacheName),
			ResourceGroupName:    pulumi.String(targetResourceGroupName),

			LinkedRedisCacheId:       pulumi.String(locals.LinkedRedisCacheId),
			LinkedRedisCacheLocation: pulumi.String(spec.LinkedRedisCacheLocation.GetValue()),

			ServerRole: pulumi.String(serverRoleStrings[spec.ServerRole]),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create redis linked server on cache %s", targetCacheName)
	}

	// Export outputs. Azure names the link after the linked
	// (secondary) cache; the geo hostname follows the CURRENT primary
	// across failovers.
	ctx.Export(OpLinkedServerId, createdLinkedServer.ID())
	ctx.Export(OpLinkedServerName, createdLinkedServer.Name)
	ctx.Export(OpGeoReplicatedPrimaryHostName, createdLinkedServer.GeoReplicatedPrimaryHostName)

	return nil
}
