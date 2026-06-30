variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "OciContainerEngineNodePool specification"
  type = object({
    compartment_id = object({
      value = string
    })

    cluster_id = object({
      value = string
    })

    name               = optional(string, "")
    kubernetes_version = optional(string, "")
    node_shape         = string

    node_shape_config = optional(object({
      ocpus         = optional(number, 0)
      memory_in_gbs = optional(number, 0)
    }), null)

    node_source_details = optional(object({
      image_id              = string
      boot_volume_size_in_gbs = optional(number, 0)
    }), null)

    node_config_details = object({
      placement_configs = list(object({
        availability_domain = string
        subnet_id = object({
          value = string
        })
        fault_domains = optional(list(string), [])
        capacity_reservation_id = optional(object({
          value = string
        }), null)
        preemptible_node_config = optional(object({
          is_preserve_boot_volume = optional(bool, null)
        }), null)
      }))
      size = number
      nsg_ids = optional(list(object({
        value = string
      })), [])
      kms_key_id = optional(object({
        value = string
      }), null)
      is_pv_encryption_in_transit_enabled = optional(bool, false)
      pod_network_option_details = optional(object({
        cni_type         = string
        max_pods_per_node = optional(number, 0)
        pod_nsg_ids = optional(list(object({
          value = string
        })), [])
        pod_subnet_ids = optional(list(object({
          value = string
        })), [])
      }), null)
    })

    ssh_public_key = optional(string, "")

    initial_node_labels = optional(list(object({
      key   = string
      value = string
    })), [])

    node_metadata = optional(map(string), {})

    node_eviction_settings = optional(object({
      eviction_grace_duration             = optional(string, "")
      is_force_action_after_grace_duration = optional(bool, null)
      is_force_delete_after_grace_duration = optional(bool, null)
    }), null)

    node_pool_cycling_details = optional(object({
      is_node_cycling_enabled = optional(bool, false)
      maximum_surge           = optional(string, "")
      maximum_unavailable     = optional(string, "")
    }), null)
  })
}
