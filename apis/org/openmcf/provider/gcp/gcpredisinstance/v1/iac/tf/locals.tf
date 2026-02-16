locals {
  project_id    = var.spec.project_id.value
  instance_name = var.spec.instance_name
  region        = var.spec.region
  tier          = var.spec.tier
  memory_size_gb = var.spec.memory_size_gb

  redis_version      = var.spec.redis_version != "" ? var.spec.redis_version : null
  display_name       = var.spec.display_name != "" ? var.spec.display_name : null
  location_id        = var.spec.location_id != "" ? var.spec.location_id : null
  authorized_network = var.spec.authorized_network != null ? var.spec.authorized_network.value : null
  connect_mode       = var.spec.connect_mode != "" ? var.spec.connect_mode : null
  reserved_ip_range  = var.spec.reserved_ip_range != "" ? var.spec.reserved_ip_range : null
  transit_encryption_mode = var.spec.transit_encryption_mode != "" ? var.spec.transit_encryption_mode : null
  read_replicas_mode = var.spec.read_replicas_mode != "" ? var.spec.read_replicas_mode : null
  replica_count      = var.spec.replica_count > 0 ? var.spec.replica_count : null
  customer_managed_key = var.spec.customer_managed_key != null ? var.spec.customer_managed_key.value : null

  labels = merge(
    {
      "openmcf-resource"      = "true"
      "openmcf-resource-name" = local.instance_name
      "openmcf-resource-kind" = "gcpredisinstance"
    },
    var.metadata.org != "" ? { "openmcf-organization" = var.metadata.org } : {},
    var.metadata.env != "" ? { "openmcf-environment" = var.metadata.env } : {},
    var.metadata.id != "" ? { "openmcf-resource-id" = var.metadata.id } : {},
  )
}
