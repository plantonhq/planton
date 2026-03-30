# Atlas MongoDB Cluster Resource
# This resource creates a Atlas MongoDB cluster with all configuration parameters
# from the AtlasMongodbSpec proto definition
#
# Documentation: https://registry.terraform.io/providers/mongodb/atlasmongodb/latest/docs/resources/cluster
# or: https://registry.terraform.io/providers/mongodb/atlasmongodb/latest/docs/resources/advanced_cluster

# Main Atlas MongoDB Cluster resource
# Using atlasmongodb_advanced_cluster for production-grade M10+ clusters
# For M0/M2/M5 shared tiers, use atlasmongodb_cluster resource instead
resource "atlasmongodb_advanced_cluster" "main" {
  # Required parameters
  project_id = local.project_id
  name       = local.resource_name

  # Cluster type: REPLICASET, SHARDED, or GEOSHARDED
  cluster_type = local.cluster_type

  # MongoDB version
  mongo_db_major_version = local.mongo_db_major_version

  # Backup configuration
  backup_enabled = local.cloud_backup_enabled

  # Auto-scaling configuration
  # Note: disk_gb_enabled is specified at the replication_specs level in advanced_cluster
  
  # Replication specification
  # This defines the topology, regions, and node configuration
  replication_specs {
    # Number of shards (zones) - typically 1 for REPLICASET, more for SHARDED
    num_shards = local.is_sharded || local.is_geo_sharded ? 2 : 1

    # Region configuration
    # In a production setup, you would have multiple regions for high availability
    # For this basic implementation, we'll use a single region configuration
    region_configs {
      # Provider and region settings
      provider_name = local.provider_name
      
      # Region name depends on provider:
      # AWS: e.g., US_EAST_1, US_WEST_2
      # GCP: e.g., CENTRAL_US, EASTERN_US
      # AZURE: e.g., US_EAST_2, US_WEST
      # For this implementation, use a default region per provider
      region_name = (
        local.is_aws_provider ? "US_EAST_1" :
        local.is_gcp_provider ? "CENTRAL_US" :
        local.is_azure_provider ? "US_EAST_2" :
        "US_EAST_1"
      )

      # Election priority (7 is highest, identifies preferred region)
      priority = local.priority

      # Electable specifications (nodes that can become primary)
      electable_specs {
        instance_size = local.provider_instance_size_name
        node_count    = local.electable_nodes
      }

      # Read-only specifications (optional)
      # Only create if read_only_nodes > 0
      dynamic "read_only_specs" {
        for_each = local.has_read_only ? [1] : []
        content {
          instance_size = local.provider_instance_size_name
          node_count    = local.read_only_nodes
        }
      }

      # Auto-scaling configuration
      auto_scaling {
        disk_gb_enabled            = local.auto_scaling_disk_enabled
        compute_enabled            = false
        compute_scale_down_enabled = false
      }
    }
  }

  # Labels (tags) for resource organization
  # Atlas MongoDB uses labels for resource tracking
  dynamic "labels" {
    for_each = local.tags
    content {
      key   = labels.key
      value = labels.value
    }
  }
}

# Alternative resource for shared tiers (M0, M2, M5)
# Uncomment and use this instead of advanced_cluster for free/shared tiers
# resource "atlasmongodb_cluster" "shared" {
#   project_id   = local.project_id
#   name         = local.resource_name
#   
#   # Provider settings for shared clusters
#   provider_name               = local.provider_name
#   backing_provider_name       = local.provider_name
#   provider_instance_size_name = local.provider_instance_size_name
#   
#   # Region
#   provider_region_name = (
#     local.is_aws_provider ? "US_EAST_1" :
#     local.is_gcp_provider ? "CENTRAL_US" :
#     local.is_azure_provider ? "US_EAST_2" :
#     "US_EAST_1"
#   )
#   
#   # MongoDB version
#   mongo_db_major_version = local.mongo_db_major_version
#   
#   # Auto-scaling
#   auto_scaling_disk_gb_enabled = local.auto_scaling_disk_enabled
#   
#   # Backup (not available for M0)
#   cloud_backup = local.cloud_backup_enabled
# }

