locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CloudResourceKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azuremanagedredis"
    "resource_name" = var.metadata.name
  }

  org_tag = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  env_tag = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "environment" = var.metadata.env } : {}

  id_tag = (
    var.metadata.id != null && var.metadata.id != ""
  ) ? { "resource_id" = var.metadata.id } : {}

  # Metadata-derived tags first, then the user's spec tags merged over
  # them: user tags deliberately win so an org's governance conventions
  # (cost center, owner) can override the derived values where they
  # collide.
  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, local.id_tag, var.spec.tags)

  # The spec's sku enum arrives as the FULL proto value name
  # (BALANCED_B0 style); Azure spells the same SKU {Family}_{Size}
  # (Balanced_B0 style). The mapping is mechanical -- family word
  # capitalization -- but spelled out row by row so the plan renders the
  # exact wire value and a vocabulary drift fails loudly at plan time.
  sku_map = {
    "BALANCED_B0"            = "Balanced_B0"
    "BALANCED_B1"            = "Balanced_B1"
    "BALANCED_B3"            = "Balanced_B3"
    "BALANCED_B5"            = "Balanced_B5"
    "BALANCED_B10"           = "Balanced_B10"
    "BALANCED_B20"           = "Balanced_B20"
    "BALANCED_B50"           = "Balanced_B50"
    "BALANCED_B100"          = "Balanced_B100"
    "BALANCED_B150"          = "Balanced_B150"
    "BALANCED_B250"          = "Balanced_B250"
    "BALANCED_B350"          = "Balanced_B350"
    "BALANCED_B500"          = "Balanced_B500"
    "BALANCED_B700"          = "Balanced_B700"
    "BALANCED_B1000"         = "Balanced_B1000"
    "COMPUTE_OPTIMIZED_X3"   = "ComputeOptimized_X3"
    "COMPUTE_OPTIMIZED_X5"   = "ComputeOptimized_X5"
    "COMPUTE_OPTIMIZED_X10"  = "ComputeOptimized_X10"
    "COMPUTE_OPTIMIZED_X20"  = "ComputeOptimized_X20"
    "COMPUTE_OPTIMIZED_X50"  = "ComputeOptimized_X50"
    "COMPUTE_OPTIMIZED_X100" = "ComputeOptimized_X100"
    "COMPUTE_OPTIMIZED_X150" = "ComputeOptimized_X150"
    "COMPUTE_OPTIMIZED_X250" = "ComputeOptimized_X250"
    "COMPUTE_OPTIMIZED_X350" = "ComputeOptimized_X350"
    "COMPUTE_OPTIMIZED_X500" = "ComputeOptimized_X500"
    "COMPUTE_OPTIMIZED_X700" = "ComputeOptimized_X700"
    "MEMORY_OPTIMIZED_M10"   = "MemoryOptimized_M10"
    "MEMORY_OPTIMIZED_M20"   = "MemoryOptimized_M20"
    "MEMORY_OPTIMIZED_M50"   = "MemoryOptimized_M50"
    "MEMORY_OPTIMIZED_M100"  = "MemoryOptimized_M100"
    "MEMORY_OPTIMIZED_M150"  = "MemoryOptimized_M150"
    "MEMORY_OPTIMIZED_M250"  = "MemoryOptimized_M250"
    "MEMORY_OPTIMIZED_M350"  = "MemoryOptimized_M350"
    "MEMORY_OPTIMIZED_M500"  = "MemoryOptimized_M500"
    "MEMORY_OPTIMIZED_M700"  = "MemoryOptimized_M700"
    "MEMORY_OPTIMIZED_M1000" = "MemoryOptimized_M1000"
    "MEMORY_OPTIMIZED_M1500" = "MemoryOptimized_M1500"
    "MEMORY_OPTIMIZED_M2000" = "MemoryOptimized_M2000"
    "FLASH_OPTIMIZED_A250"   = "FlashOptimized_A250"
    "FLASH_OPTIMIZED_A500"   = "FlashOptimized_A500"
    "FLASH_OPTIMIZED_A700"   = "FlashOptimized_A700"
    "FLASH_OPTIMIZED_A1000"  = "FlashOptimized_A1000"
    "FLASH_OPTIMIZED_A1500"  = "FlashOptimized_A1500"
    "FLASH_OPTIMIZED_A2000"  = "FlashOptimized_A2000"
    "FLASH_OPTIMIZED_A4500"  = "FlashOptimized_A4500"
  }
  sku_name = local.sku_map[var.spec.sku_name]

  # Database enums: spec enum names -> ARM's wire values. Absent fields
  # deploy Azure's own defaults (Encrypted / OSSCluster / VolatileLRU)
  # explicitly, so both engines send identical bodies.
  client_protocol_map = {
    "ENCRYPTED" = "Encrypted"
    "PLAINTEXT" = "Plaintext"
  }
  client_protocol = local.client_protocol_map[coalesce(var.spec.default_database.client_protocol, "ENCRYPTED")]

  clustering_policy_map = {
    "ENTERPRISE_CLUSTER" = "EnterpriseCluster"
    "OSS_CLUSTER"        = "OSSCluster"
    "NO_CLUSTER"         = "NoCluster"
  }
  clustering_policy = local.clustering_policy_map[coalesce(var.spec.default_database.clustering_policy, "OSS_CLUSTER")]

  eviction_policy_map = {
    "ALL_KEYS_LFU"    = "AllKeysLFU"
    "ALL_KEYS_LRU"    = "AllKeysLRU"
    "ALL_KEYS_RANDOM" = "AllKeysRandom"
    "VOLATILE_LRU"    = "VolatileLRU"
    "VOLATILE_LFU"    = "VolatileLFU"
    "VOLATILE_TTL"    = "VolatileTTL"
    "VOLATILE_RANDOM" = "VolatileRandom"
    "NO_EVICTION"     = "NoEviction"
  }
  eviction_policy = local.eviction_policy_map[coalesce(var.spec.default_database.eviction_policy, "VOLATILE_LRU")]

  # Identity type: spec enum name -> ARM's value.
  identity_type_map = {
    "SYSTEM_ASSIGNED"          = "SystemAssigned"
    "USER_ASSIGNED"            = "UserAssigned"
    "SYSTEM_AND_USER_ASSIGNED" = "SystemAssigned, UserAssigned"
  }

  # The provider models public network access as an Enabled/Disabled
  # string; the spec's bool maps onto it.
  public_network_access = (
    coalesce(var.spec.public_network_access_enabled, true)
    ? "Enabled"
    : "Disabled"
  )
}
