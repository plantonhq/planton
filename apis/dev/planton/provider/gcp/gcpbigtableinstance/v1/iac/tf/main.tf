resource "google_bigtable_instance" "this" {
  name                = var.spec.instance_name
  project             = var.spec.project_id
  labels              = local.labels
  deletion_protection = var.spec.deletion_protection
  force_destroy       = var.spec.force_destroy

  display_name = var.spec.display_name != "" ? var.spec.display_name : null

  dynamic "cluster" {
    for_each = var.spec.clusters
    content {
      cluster_id         = cluster.value.cluster_id
      zone               = cluster.value.zone
      num_nodes          = cluster.value.autoscaling_config == null ? (cluster.value.num_nodes > 0 ? cluster.value.num_nodes : null) : null
      storage_type       = cluster.value.storage_type
      kms_key_name       = cluster.value.kms_key_name != "" ? cluster.value.kms_key_name : null
      node_scaling_factor = cluster.value.node_scaling_factor != "" ? cluster.value.node_scaling_factor : null

      dynamic "autoscaling_config" {
        for_each = cluster.value.autoscaling_config != null ? [cluster.value.autoscaling_config] : []
        content {
          min_nodes      = autoscaling_config.value.min_nodes
          max_nodes      = autoscaling_config.value.max_nodes
          cpu_target     = autoscaling_config.value.cpu_target
          storage_target = autoscaling_config.value.storage_target > 0 ? autoscaling_config.value.storage_target : null
        }
      }
    }
  }
}
