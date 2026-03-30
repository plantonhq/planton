# Local variables for Atlas MongoDB cluster configuration
# These locals help centralize computed values and improve maintainability

locals {
  # Resource identification
  resource_id   = var.metadata.id != null && var.metadata.id != "" ? var.metadata.id : var.metadata.name
  resource_name = var.metadata.name

  # Tags and labels for resource organization
  # Convert metadata labels to tags if needed
  tags = merge(
    var.metadata.labels != null ? var.metadata.labels : {},
    {
      "managed-by"  = "openmcf"
      "component"   = "atlas-mongodb"
      "environment" = var.metadata.env != null && var.metadata.env != "" ? var.metadata.env : "default"
      "org"         = var.metadata.org != null && var.metadata.org != "" ? var.metadata.org : "default"
    }
  )

  # Cluster configuration from spec
  cluster_config = var.spec.cluster_config

  # Atlas MongoDB cluster settings
  project_id                  = local.cluster_config.project_id
  cluster_type                = local.cluster_config.cluster_type
  electable_nodes             = local.cluster_config.electable_nodes
  priority                    = local.cluster_config.priority
  read_only_nodes             = local.cluster_config.read_only_nodes
  cloud_backup_enabled        = local.cluster_config.cloud_backup
  auto_scaling_disk_enabled   = local.cluster_config.auto_scaling_disk_gb_enabled
  mongo_db_major_version      = local.cluster_config.mongo_db_major_version
  provider_name               = local.cluster_config.provider_name
  provider_instance_size_name = local.cluster_config.provider_instance_size_name

  # Configuration validation flags
  is_replica_set   = local.cluster_type == "REPLICASET"
  is_sharded       = local.cluster_type == "SHARDED"
  is_geo_sharded   = local.cluster_type == "GEOSHARDED"
  has_read_only    = local.read_only_nodes > 0
  is_multi_node    = local.electable_nodes > 1
  is_production    = local.provider_instance_size_name != "M0" && local.provider_instance_size_name != "M2" && local.provider_instance_size_name != "M5"

  # Provider-specific flags
  is_aws_provider   = local.provider_name == "AWS"
  is_gcp_provider   = local.provider_name == "GCP"
  is_azure_provider = local.provider_name == "AZURE"
  is_tenant_provider = local.provider_name == "TENANT"
}

