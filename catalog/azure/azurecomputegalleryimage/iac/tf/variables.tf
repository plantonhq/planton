variable "metadata" {
  description = "Catalog object metadata"
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
  description = "AzureComputeGalleryImage specification"
  type = object({
    resource_group = string
    gallery_name   = string
    name           = string
    region         = string
    identifier = object({
      publisher = string
      offer     = string
      sku       = string
    })
    os_type                             = string
    specialized                         = optional(bool, false)
    architecture                        = optional(string, "")
    hyper_v_generation                  = optional(string, "")
    trusted_launch_supported            = optional(bool, false)
    trusted_launch_enabled              = optional(bool, false)
    confidential_vm_supported           = optional(bool, false)
    confidential_vm_enabled             = optional(bool, false)
    accelerated_network_support_enabled = optional(bool, false)
    hibernation_enabled                 = optional(bool, false)
    disk_controller_type_nvme_enabled   = optional(bool, false)
    disk_types_not_allowed              = optional(list(string), [])
    end_of_life_date                    = optional(string, "")
    eula                                = optional(string, "")
    privacy_statement_uri               = optional(string, "")
    release_note_uri                    = optional(string, "")
    description                         = optional(string, "")
    purchase_plan = optional(object({
      name      = string
      publisher = optional(string, "")
      product   = optional(string, "")
    }))
    min_recommended_vcpu_count   = optional(number)
    max_recommended_vcpu_count   = optional(number)
    min_recommended_memory_in_gb = optional(number)
    max_recommended_memory_in_gb = optional(number)
    versions = optional(list(object({
      name = string
      target_regions = list(object({
        name                        = string
        regional_replica_count      = number
        disk_encryption_set_id      = optional(string, "")
        exclude_from_latest_enabled = optional(bool, false)
        storage_account_type        = optional(string, "")
      }))
      blob_uri                                 = optional(string, "")
      storage_account_id                       = optional(string, "")
      os_disk_snapshot_id                      = optional(string, "")
      managed_image_id                         = optional(string, "")
      replication_mode                         = optional(string, "")
      exclude_from_latest                      = optional(bool, false)
      deletion_of_replicated_locations_enabled = optional(bool, false)
      end_of_life_date                         = optional(string, "")
      tags                                     = optional(map(string), {})
    })), [])
    tags = optional(map(string), {})
  })
}
