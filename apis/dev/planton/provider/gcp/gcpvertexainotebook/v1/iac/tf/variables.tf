variable "metadata" {
  description = "Planton resource metadata"
  type = object({
    name    = string
    id      = optional(string, "")
    org     = optional(string, "")
    env     = optional(string, "")
    labels  = optional(map(string), {})
    tags    = optional(list(string), [])
    version = optional(string, "")
  })
}

variable "spec" {
  description = "GcpVertexAiNotebook spec"
  type = object({
    project_id   = object({ value = string })
    location     = string
    machine_type = string

    instance_name        = optional(string, "")
    instance_owners      = optional(list(string), [])
    desired_state        = optional(string, "")
    disable_proxy_access = optional(bool, false)
    metadata             = optional(map(string), {})

    boot_disk = optional(object({
      disk_type    = optional(string, "")
      disk_size_gb = optional(number, 0)
      kms_key      = optional(object({ value = string }), null)
    }), null)

    data_disk = optional(object({
      disk_type    = optional(string, "")
      disk_size_gb = optional(number, 0)
      kms_key      = optional(object({ value = string }), null)
    }), null)

    accelerator_config = optional(object({
      type       = optional(string, "")
      core_count = optional(number, 0)
    }), null)

    network_interface = optional(object({
      network  = optional(object({ value = string }), null)
      subnet   = optional(object({ value = string }), null)
      nic_type = optional(string, "")
    }), null)

    disable_public_ip    = optional(bool, false)
    enable_ip_forwarding = optional(bool, false)

    service_account = optional(object({ value = string }), null)

    tags = optional(list(string), [])

    vm_image = optional(object({
      project = optional(string, "")
      family  = optional(string, "")
      name    = optional(string, "")
    }), null)

    container_image = optional(object({
      repository = string
      tag        = optional(string, "")
    }), null)

    shielded_instance_config = optional(object({
      enable_secure_boot          = optional(bool, false)
      enable_vtpm                 = optional(bool, false)
      enable_integrity_monitoring = optional(bool, false)
    }), null)
  })
}

variable "provider_config" {
  description = "GCP provider configuration"
  type = object({
    service_account_key = optional(string, "")
  })
  default = { service_account_key = "" }
}
