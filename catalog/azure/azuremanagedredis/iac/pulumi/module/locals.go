package module

import (
	"strings"

	azuremanagedredisv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremanagedredis/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureManagedRedis *azuremanagedredisv1alpha1.AzureManagedRedis
	ResourceGroupName string
	AzureTags         map[string]string
	// SkuName is ARM's {Family}_{Size} SKU value, mapped from the spec
	// enum (the spec's BALANCED_B0 style becomes Azure's Balanced_B0
	// style).
	SkuName string
}

// skuStrings maps the spec's sku enum to Azure's {Family}_{Size} wire
// values. Spelled out row by row so a vocabulary drift fails loudly at
// preview time instead of deploying a wrong SKU.
var skuStrings = map[azuremanagedredisv1alpha1.AzureManagedRedisSku]string{
	azuremanagedredisv1alpha1.AzureManagedRedisSku_BALANCED_B0:            "Balanced_B0",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_BALANCED_B1:            "Balanced_B1",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_BALANCED_B3:            "Balanced_B3",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_BALANCED_B5:            "Balanced_B5",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_BALANCED_B10:           "Balanced_B10",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_BALANCED_B20:           "Balanced_B20",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_BALANCED_B50:           "Balanced_B50",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_BALANCED_B100:          "Balanced_B100",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_BALANCED_B150:          "Balanced_B150",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_BALANCED_B250:          "Balanced_B250",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_BALANCED_B350:          "Balanced_B350",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_BALANCED_B500:          "Balanced_B500",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_BALANCED_B700:          "Balanced_B700",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_BALANCED_B1000:         "Balanced_B1000",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_COMPUTE_OPTIMIZED_X3:   "ComputeOptimized_X3",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_COMPUTE_OPTIMIZED_X5:   "ComputeOptimized_X5",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_COMPUTE_OPTIMIZED_X10:  "ComputeOptimized_X10",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_COMPUTE_OPTIMIZED_X20:  "ComputeOptimized_X20",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_COMPUTE_OPTIMIZED_X50:  "ComputeOptimized_X50",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_COMPUTE_OPTIMIZED_X100: "ComputeOptimized_X100",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_COMPUTE_OPTIMIZED_X150: "ComputeOptimized_X150",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_COMPUTE_OPTIMIZED_X250: "ComputeOptimized_X250",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_COMPUTE_OPTIMIZED_X350: "ComputeOptimized_X350",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_COMPUTE_OPTIMIZED_X500: "ComputeOptimized_X500",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_COMPUTE_OPTIMIZED_X700: "ComputeOptimized_X700",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_MEMORY_OPTIMIZED_M10:   "MemoryOptimized_M10",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_MEMORY_OPTIMIZED_M20:   "MemoryOptimized_M20",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_MEMORY_OPTIMIZED_M50:   "MemoryOptimized_M50",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_MEMORY_OPTIMIZED_M100:  "MemoryOptimized_M100",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_MEMORY_OPTIMIZED_M150:  "MemoryOptimized_M150",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_MEMORY_OPTIMIZED_M250:  "MemoryOptimized_M250",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_MEMORY_OPTIMIZED_M350:  "MemoryOptimized_M350",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_MEMORY_OPTIMIZED_M500:  "MemoryOptimized_M500",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_MEMORY_OPTIMIZED_M700:  "MemoryOptimized_M700",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_MEMORY_OPTIMIZED_M1000: "MemoryOptimized_M1000",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_MEMORY_OPTIMIZED_M1500: "MemoryOptimized_M1500",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_MEMORY_OPTIMIZED_M2000: "MemoryOptimized_M2000",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_FLASH_OPTIMIZED_A250:   "FlashOptimized_A250",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_FLASH_OPTIMIZED_A500:   "FlashOptimized_A500",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_FLASH_OPTIMIZED_A700:   "FlashOptimized_A700",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_FLASH_OPTIMIZED_A1000:  "FlashOptimized_A1000",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_FLASH_OPTIMIZED_A1500:  "FlashOptimized_A1500",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_FLASH_OPTIMIZED_A2000:  "FlashOptimized_A2000",
	azuremanagedredisv1alpha1.AzureManagedRedisSku_FLASH_OPTIMIZED_A4500:  "FlashOptimized_A4500",
}

// clientProtocolStrings maps the TLS-posture enum to ARM's values.
var clientProtocolStrings = map[azuremanagedredisv1alpha1.AzureManagedRedisClientProtocol]string{
	azuremanagedredisv1alpha1.AzureManagedRedisClientProtocol_ENCRYPTED: "Encrypted",
	azuremanagedredisv1alpha1.AzureManagedRedisClientProtocol_PLAINTEXT: "Plaintext",
}

// clusteringPolicyStrings maps the shard-distribution enum to ARM's
// values.
var clusteringPolicyStrings = map[azuremanagedredisv1alpha1.AzureManagedRedisClusteringPolicy]string{
	azuremanagedredisv1alpha1.AzureManagedRedisClusteringPolicy_ENTERPRISE_CLUSTER: "EnterpriseCluster",
	azuremanagedredisv1alpha1.AzureManagedRedisClusteringPolicy_OSS_CLUSTER:        "OSSCluster",
	azuremanagedredisv1alpha1.AzureManagedRedisClusteringPolicy_NO_CLUSTER:         "NoCluster",
}

// evictionPolicyStrings maps the eviction enum to ARM's values.
var evictionPolicyStrings = map[azuremanagedredisv1alpha1.AzureManagedRedisEvictionPolicy]string{
	azuremanagedredisv1alpha1.AzureManagedRedisEvictionPolicy_ALL_KEYS_LFU:    "AllKeysLFU",
	azuremanagedredisv1alpha1.AzureManagedRedisEvictionPolicy_ALL_KEYS_LRU:    "AllKeysLRU",
	azuremanagedredisv1alpha1.AzureManagedRedisEvictionPolicy_ALL_KEYS_RANDOM: "AllKeysRandom",
	azuremanagedredisv1alpha1.AzureManagedRedisEvictionPolicy_VOLATILE_LRU:    "VolatileLRU",
	azuremanagedredisv1alpha1.AzureManagedRedisEvictionPolicy_VOLATILE_LFU:    "VolatileLFU",
	azuremanagedredisv1alpha1.AzureManagedRedisEvictionPolicy_VOLATILE_TTL:    "VolatileTTL",
	azuremanagedredisv1alpha1.AzureManagedRedisEvictionPolicy_VOLATILE_RANDOM: "VolatileRandom",
	azuremanagedredisv1alpha1.AzureManagedRedisEvictionPolicy_NO_EVICTION:     "NoEviction",
}

// identityTypeStrings maps the identity-type enum to ARM's values.
var identityTypeStrings = map[azuremanagedredisv1alpha1.AzureManagedRedisIdentityType]string{
	azuremanagedredisv1alpha1.AzureManagedRedisIdentityType_SYSTEM_ASSIGNED:          "SystemAssigned",
	azuremanagedredisv1alpha1.AzureManagedRedisIdentityType_USER_ASSIGNED:            "UserAssigned",
	azuremanagedredisv1alpha1.AzureManagedRedisIdentityType_SYSTEM_AND_USER_ASSIGNED: "SystemAssigned, UserAssigned",
}

func initializeLocals(ctx *pulumi.Context, stackInput *azuremanagedredisv1alpha1.AzureManagedRedisStackInput) *Locals {
	locals := &Locals{}

	locals.AzureManagedRedis = stackInput.Target
	target := stackInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// The sku enum is spec-required (never unspecified), so the map hit
	// is guaranteed by validation.
	locals.SkuName = skuStrings[target.Spec.SkuName]

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzureManagedRedis.String()),
	}

	if target.Metadata.Id != "" {
		locals.AzureTags[azuretagkeys.ResourceId] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.AzureTags[azuretagkeys.Organization] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.AzureTags[azuretagkeys.Environment] = target.Metadata.Env
	}

	for key, value := range target.Spec.Tags {
		locals.AzureTags[key] = value
	}

	return locals
}
