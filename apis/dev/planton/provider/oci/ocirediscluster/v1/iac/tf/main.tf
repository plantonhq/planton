resource "oci_redis_redis_cluster" "this" {
  compartment_id  = var.spec.compartment_id.value
  display_name    = local.display_name
  subnet_id       = var.spec.subnet_id.value
  node_count      = var.spec.node_count
  node_memory_in_gbs = var.spec.node_memory_in_gbs
  software_version   = var.spec.software_version
  freeform_tags      = local.freeform_tags

  cluster_mode = var.spec.cluster_mode != "" ? lookup(local.cluster_mode_map, var.spec.cluster_mode, null) : null
  shard_count  = var.spec.shard_count > 0 ? var.spec.shard_count : null

  nsg_ids = length(local.nsg_ids) > 0 ? local.nsg_ids : null

  oci_cache_config_set_id = var.spec.config_set_id != null ? var.spec.config_set_id.value : null
}
