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
    # StringValueOrRef fields arrive from the proto→tfvars converter as
    # plain strings (already resolved), never as object({value}). If
    # project_id is empty, the provider's default project is used
    # (see locals.tf).
    project_id = optional(string, "")

    # Zone (not region). Immutable (ForceNew).
    location = string

    machine_type = string

    # Instance name; falls back to metadata.name. Immutable (ForceNew).
    instance_name = optional(string, "")

    instance_owners      = optional(list(string), [])
    desired_state        = optional(string, "")
    disable_proxy_access = optional(bool, false)
    metadata             = optional(map(string), {})

    # User labels; merged beneath the platform attribution labels.
    labels = optional(map(string), {})

    boot_disk = optional(object({
      disk_type    = optional(string, "")
      disk_size_gb = optional(number, 0)
      # Resolved from a GcpKmsKey reference to the key path.
      kms_key = optional(string, "")
    }), null)

    data_disk = optional(object({
      disk_type    = optional(string, "")
      disk_size_gb = optional(number, 0)
      kms_key      = optional(string, "")
    }), null)

    accelerator_config = optional(object({
      type       = optional(string, "")
      core_count = optional(number, 0)
    }), null)

    network_interface = optional(object({
      # Resolved from GcpVpcNetwork / GcpSubnetwork references to self links.
      network  = optional(string, "")
      subnet   = optional(string, "")
      nic_type = optional(string, "")
      # Resolved from a GcpAddress reference to the literal IP address.
      external_ip = optional(string, "")
    }), null)

    disable_public_ip    = optional(bool, false)
    enable_ip_forwarding = optional(bool, false)

    # Resolved from a GcpServiceAccount reference to the SA email.
    service_account = optional(string, "")

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
      enable_secure_boot          = optional(bool)
      enable_vtpm                 = optional(bool)
      enable_integrity_monitoring = optional(bool)
    }), null)

    confidential_instance_config = optional(object({
      enabled                    = optional(bool)
      confidential_instance_type = optional(string, "")
    }), null)

    reservation_affinity = optional(object({
      consume_reservation_type = optional(string, "")
      key                      = optional(string, "")
      values                   = optional(list(string), [])
    }), null)

    enable_managed_euc          = optional(bool, false)
    enable_third_party_identity = optional(bool, false)

    # Client-side destroy behavior: DELETE (default), PREVENT, ABANDON.
    deletion_policy = optional(string, "")
  })

  validation {
    condition     = var.spec.location != ""
    error_message = "location is required."
  }

  validation {
    condition     = var.spec.machine_type != ""
    error_message = "machine_type is required."
  }
}
