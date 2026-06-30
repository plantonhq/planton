resource "oci_containerengine_node_pool" "this" {
  compartment_id     = var.spec.compartment_id.value
  cluster_id         = var.spec.cluster_id.value
  name               = local.display_name
  node_shape         = var.spec.node_shape
  freeform_tags      = local.freeform_tags
  kubernetes_version = var.spec.kubernetes_version != "" ? var.spec.kubernetes_version : null
  ssh_public_key     = var.spec.ssh_public_key != "" ? var.spec.ssh_public_key : null
  node_metadata      = length(var.spec.node_metadata) > 0 ? var.spec.node_metadata : null

  dynamic "node_shape_config" {
    for_each = var.spec.node_shape_config != null ? [var.spec.node_shape_config] : []
    content {
      ocpus         = node_shape_config.value.ocpus > 0 ? node_shape_config.value.ocpus : null
      memory_in_gbs = node_shape_config.value.memory_in_gbs > 0 ? node_shape_config.value.memory_in_gbs : null
    }
  }

  dynamic "node_source_details" {
    for_each = var.spec.node_source_details != null ? [var.spec.node_source_details] : []
    content {
      image_id              = node_source_details.value.image_id
      source_type           = "IMAGE"
      boot_volume_size_in_gbs = node_source_details.value.boot_volume_size_in_gbs > 0 ? node_source_details.value.boot_volume_size_in_gbs : null
    }
  }

  node_config_details {
    size          = var.spec.node_config_details.size
    freeform_tags = local.freeform_tags
    nsg_ids       = length(local.node_nsg_ids) > 0 ? local.node_nsg_ids : null
    kms_key_id    = var.spec.node_config_details.kms_key_id != null ? var.spec.node_config_details.kms_key_id.value : null

    is_pv_encryption_in_transit_enabled = var.spec.node_config_details.is_pv_encryption_in_transit_enabled ? "true" : null

    dynamic "placement_configs" {
      for_each = var.spec.node_config_details.placement_configs
      content {
        availability_domain   = placement_configs.value.availability_domain
        subnet_id             = placement_configs.value.subnet_id.value
        fault_domains         = length(placement_configs.value.fault_domains) > 0 ? placement_configs.value.fault_domains : null
        capacity_reservation_id = placement_configs.value.capacity_reservation_id != null ? placement_configs.value.capacity_reservation_id.value : null

        dynamic "preemptible_node_config" {
          for_each = placement_configs.value.preemptible_node_config != null ? [placement_configs.value.preemptible_node_config] : []
          content {
            preemption_action {
              type                    = "TERMINATE"
              is_preserve_boot_volume = preemptible_node_config.value.is_preserve_boot_volume
            }
          }
        }
      }
    }

    dynamic "node_pool_pod_network_option_details" {
      for_each = var.spec.node_config_details.pod_network_option_details != null ? [var.spec.node_config_details.pod_network_option_details] : []
      content {
        cni_type        = lookup(local.cni_type_map, node_pool_pod_network_option_details.value.cni_type, node_pool_pod_network_option_details.value.cni_type)
        max_pods_per_node = node_pool_pod_network_option_details.value.max_pods_per_node > 0 ? node_pool_pod_network_option_details.value.max_pods_per_node : null
        pod_nsg_ids     = length(local.pod_nsg_ids) > 0 ? local.pod_nsg_ids : null
        pod_subnet_ids  = length(local.pod_subnet_ids) > 0 ? local.pod_subnet_ids : null
      }
    }
  }

  dynamic "initial_node_labels" {
    for_each = var.spec.initial_node_labels
    content {
      key   = initial_node_labels.value.key
      value = initial_node_labels.value.value
    }
  }

  dynamic "node_eviction_node_pool_settings" {
    for_each = var.spec.node_eviction_settings != null ? [var.spec.node_eviction_settings] : []
    content {
      eviction_grace_duration              = node_eviction_node_pool_settings.value.eviction_grace_duration != "" ? node_eviction_node_pool_settings.value.eviction_grace_duration : null
      is_force_action_after_grace_duration = node_eviction_node_pool_settings.value.is_force_action_after_grace_duration
      is_force_delete_after_grace_duration = node_eviction_node_pool_settings.value.is_force_delete_after_grace_duration
    }
  }

  dynamic "node_pool_cycling_details" {
    for_each = var.spec.node_pool_cycling_details != null ? [var.spec.node_pool_cycling_details] : []
    content {
      is_node_cycling_enabled = node_pool_cycling_details.value.is_node_cycling_enabled
      maximum_surge           = node_pool_cycling_details.value.maximum_surge != "" ? node_pool_cycling_details.value.maximum_surge : null
      maximum_unavailable     = node_pool_cycling_details.value.maximum_unavailable != "" ? node_pool_cycling_details.value.maximum_unavailable : null
    }
  }
}
