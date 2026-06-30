variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Specification for GCP Compute Instance"
  type = object({

    # GCP project ID where the Compute Instance will be created.
    # Supports StringValueOrRef: either { value = "..." } or { value_from = { kind = "...", name = "...", field_path = "..." } }
    project_id = object({
      value      = optional(string)
      value_from = optional(object({
        kind       = optional(string)
        env        = optional(string)
        name       = string
        field_path = optional(string)
      }))
    })

    # Zone where the instance will be deployed, for example "us-central1-a".
    zone = string

    # Machine type for the instance, for example "e2-medium", "n1-standard-1".
    machine_type = string

    # Boot disk configuration for the instance.
    boot_disk = object({
      # Source image for the boot disk.
      image = string

      # Size of the boot disk in GB.
      size_gb = optional(number, 10)

      # Type of the boot disk (pd-standard, pd-ssd, pd-balanced).
      type = optional(string, "pd-standard")

      # Whether to auto-delete the boot disk when the instance is deleted.
      auto_delete = optional(bool, true)
    })

    # Network interface configurations for the instance.
    network_interfaces = list(object({
      # VPC network for this interface.
      # Supports StringValueOrRef.
      network = optional(object({
        value      = optional(string)
        value_from = optional(object({
          kind       = optional(string)
          env        = optional(string)
          name       = string
          field_path = optional(string)
        }))
      }))

      # Subnetwork for this interface.
      # Supports StringValueOrRef.
      subnetwork = optional(object({
        value      = optional(string)
        value_from = optional(object({
          kind       = optional(string)
          env        = optional(string)
          name       = string
          field_path = optional(string)
        }))
      }))

      # Access configurations for external IPs.
      access_configs = optional(list(object({
        # Static NAT IP address.
        nat_ip = optional(string)

        # Network tier (PREMIUM or STANDARD).
        network_tier = optional(string)
      })), [])

      # Alias IP ranges for this interface.
      alias_ip_ranges = optional(list(object({
        ip_cidr_range         = string
        subnetwork_range_name = optional(string)
      })), [])
    }))

    # Additional attached disks.
    attached_disks = optional(list(object({
      # Source disk self-link or name.
      source = string

      # Device name for the disk.
      device_name = optional(string)

      # Mode of the disk (READ_WRITE or READ_ONLY).
      mode = optional(string, "READ_WRITE")
    })), [])

    # Service account configuration.
    service_account = optional(object({
      # Email of the service account.
      # Supports StringValueOrRef: either { value = "..." } or { value_from = { kind = "...", name = "...", field_path = "..." } }
      email = optional(object({
        value      = optional(string)
        value_from = optional(object({
          kind       = optional(string)
          env        = optional(string)
          name       = string
          field_path = optional(string)
        }))
      }))

      # List of OAuth scopes.
      scopes = optional(list(string), ["https://www.googleapis.com/auth/cloud-platform"])
    }))

    # Whether the instance should be preemptible.
    preemptible = optional(bool, false)

    # Whether the instance should be a Spot VM.
    spot = optional(bool, false)

    # Whether to enable deletion protection.
    deletion_protection = optional(bool, false)

    # Custom metadata key-value pairs.
    metadata = optional(map(string), {})

    # Labels to apply to the instance.
    labels = optional(map(string), {})

    # Network tags for firewall rules.
    tags = optional(list(string), [])

    # SSH keys for accessing the instance.
    ssh_keys = optional(list(string), [])

    # Startup script to run when the instance boots.
    startup_script = optional(string)

    # Whether to allow stopping the instance for update operations.
    allow_stopping_for_update = optional(bool, true)

    # Scheduling configuration.
    scheduling = optional(object({
      # Whether the instance is preemptible.
      preemptible = optional(bool, false)

      # Automatic restart policy on failure.
      automatic_restart = optional(bool, true)

      # Behavior when host maintenance occurs (MIGRATE or TERMINATE).
      on_host_maintenance = optional(string, "MIGRATE")

      # Provisioning model (STANDARD or SPOT).
      provisioning_model = optional(string)

      # Instance termination action for Spot VMs (STOP or DELETE).
      instance_termination_action = optional(string)
    }))
  })
}
