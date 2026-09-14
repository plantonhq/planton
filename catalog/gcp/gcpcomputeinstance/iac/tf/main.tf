# Enable the Compute Engine API — the control plane that owns instances.
# disable_on_destroy is false: tearing down one VM must never disable the
# API for everything else in the project.
resource "google_project_service" "compute_api" {
  project = local.project_id
  service = "compute.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The Compute Engine instance. Sharp edges, all taught by the API rather
# than invented here:
#
#   - zone, boot source, NIC count/networks, scratch disks, hostname,
#     confidential mode, and reservation affinity are ForceNew — changing
#     them replaces the VM. machine_type, service_account, and several
#     others update via stop/start, which the provider performs only when
#     allow_stopping_for_update is true.
#
#   - Spot: provisioning_model = "SPOT" requires the API's legacy
#     preemptible flag and no automatic restart — both derived in locals,
#     identically to the Pulumi module, so the spec's single switch stays
#     honest.
#
#   - desired_status starts/suspends/stops in place ("RUNNING",
#     "SUSPENDED", "TERMINATED"); null follows GCP (running).
#
#   - deletion_protection guards the VM object only; data protection
#     lives on the disks (boot auto_delete, GcpComputeDisk lifecycles).
resource "google_compute_instance" "this" {
  name         = local.instance_name
  project      = local.project_id
  zone         = var.spec.zone
  machine_type = var.spec.machine_type
  description  = var.spec.description != "" ? var.spec.description : null
  hostname     = local.hostname
  labels       = local.final_labels

  tags             = length(var.spec.tags) > 0 ? var.spec.tags : null
  min_cpu_platform = local.min_cpu_platform
  can_ip_forward   = var.spec.can_ip_forward
  enable_display   = var.spec.enable_display ? true : null

  deletion_protection        = var.spec.deletion_protection
  desired_status             = local.desired_status
  allow_stopping_for_update  = var.spec.allow_stopping_for_update
  key_revocation_action_type = local.key_revocation_action_type

  # Destroy behavior: null follows the provider default (DELETE);
  # PREVENT fails the destroy; ABANDON forgets the VM but leaves it
  # running in GCP.
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  resource_policies = length(var.spec.resource_policies) > 0 ? var.spec.resource_policies : null

  metadata                = length(local.final_metadata) > 0 ? local.final_metadata : null
  metadata_startup_script = var.spec.startup_script != "" ? var.spec.startup_script : null

  # Boot disk: exactly one source, enforced pre-deploy by the spec's CEL.
  # An existing bootable disk attaches via source; image/snapshot create a
  # fresh disk through initialize_params. size/type default server-side
  # when null (image size, pd-standard family default).
  boot_disk {
    auto_delete                     = var.spec.boot_disk.auto_delete != null ? var.spec.boot_disk.auto_delete : true
    device_name                     = var.spec.boot_disk.device_name != "" ? var.spec.boot_disk.device_name : null
    kms_key_self_link               = var.spec.boot_disk.kms_key != "" ? var.spec.boot_disk.kms_key : null
    disk_encryption_service_account = var.spec.boot_disk.kms_key_service_account != "" ? var.spec.boot_disk.kms_key_service_account : null
    source                          = var.spec.boot_disk.source_disk != "" ? var.spec.boot_disk.source_disk : null
    mode                            = var.spec.boot_disk.mode != "" ? var.spec.boot_disk.mode : null
    # Google-advice-only lever — sent only when explicitly set, so the
    # API's auto-selected interface never produces a diff.
    interface = var.spec.boot_disk.interface != "" ? var.spec.boot_disk.interface : null
    # Regional-disk takeover; forcing a zonal disk is an API error.
    force_attach      = var.spec.boot_disk.force_attach ? true : null
    guest_os_features = length(var.spec.boot_disk.guest_os_features) > 0 ? var.spec.boot_disk.guest_os_features : null

    dynamic "initialize_params" {
      for_each = var.spec.boot_disk.source_disk == "" ? [1] : []
      content {
        image                       = var.spec.boot_disk.image != "" ? var.spec.boot_disk.image : null
        snapshot                    = var.spec.boot_disk.source_snapshot != "" ? var.spec.boot_disk.source_snapshot : null
        size                        = var.spec.boot_disk.size_gb
        type                        = var.spec.boot_disk.type != "" ? var.spec.boot_disk.type : null
        labels                      = length(var.spec.boot_disk.disk_labels) > 0 ? var.spec.boot_disk.disk_labels : null
        provisioned_iops            = var.spec.boot_disk.provisioned_iops
        provisioned_throughput      = var.spec.boot_disk.provisioned_throughput
        architecture                = var.spec.boot_disk.architecture != "" ? var.spec.boot_disk.architecture : null
        enable_confidential_compute = var.spec.boot_disk.enable_confidential_compute ? true : null
        resource_policies           = length(var.spec.boot_disk.resource_policies) > 0 ? var.spec.boot_disk.resource_policies : null
        storage_pool                = var.spec.boot_disk.storage_pool != "" ? var.spec.boot_disk.storage_pool : null
        # Exactly two zones (one the instance's own) converts the boot
        # disk to a regional disk — enforced pre-deploy by the spec.
        replica_zones = length(var.spec.boot_disk.replica_zones) > 0 ? var.spec.boot_disk.replica_zones : null
        # Create-time only; not returned by the API.
        resource_manager_tags = length(var.spec.boot_disk.resource_manager_tags) > 0 ? var.spec.boot_disk.resource_manager_tags : null

        # CMEK decryption of an encrypted source image/snapshot. CSEK
        # raw-key arms are deliberately not modeled (secure-by-default).
        dynamic "source_image_encryption_key" {
          for_each = var.spec.boot_disk.source_image_encryption != null ? [var.spec.boot_disk.source_image_encryption] : []
          content {
            kms_key_self_link       = source_image_encryption_key.value.kms_key
            kms_key_service_account = source_image_encryption_key.value.kms_key_service_account != "" ? source_image_encryption_key.value.kms_key_service_account : null
          }
        }

        dynamic "source_snapshot_encryption_key" {
          for_each = var.spec.boot_disk.source_snapshot_encryption != null ? [var.spec.boot_disk.source_snapshot_encryption] : []
          content {
            kms_key_self_link       = source_snapshot_encryption_key.value.kms_key
            kms_key_service_account = source_snapshot_encryption_key.value.kms_key_service_account != "" ? source_snapshot_encryption_key.value.kms_key_service_account : null
          }
        }
      }
    }
  }

  # Data disks are pre-existing GcpComputeDisk resources attached by
  # reference — the disk's own lifecycle protects the data.
  dynamic "attached_disk" {
    for_each = var.spec.attached_disks
    content {
      source                          = attached_disk.value.source
      device_name                     = attached_disk.value.device_name != "" ? attached_disk.value.device_name : null
      mode                            = attached_disk.value.mode != "" ? attached_disk.value.mode : null
      kms_key_self_link               = attached_disk.value.kms_key != "" ? attached_disk.value.kms_key : null
      disk_encryption_service_account = attached_disk.value.kms_key_service_account != "" ? attached_disk.value.kms_key_service_account : null
      # Regional-disk takeover; forcing a zonal disk is an API error.
      force_attach = attached_disk.value.force_attach ? true : null
    }
  }

  # Ephemeral local SSDs — contents vanish on stop/preemption.
  dynamic "scratch_disk" {
    for_each = var.spec.scratch_disks
    content {
      interface   = scratch_disk.value.interface
      size        = scratch_disk.value.size_gb
      device_name = scratch_disk.value.device_name != "" ? scratch_disk.value.device_name : null
    }
  }

  dynamic "network_interface" {
    for_each = var.spec.network_interfaces
    content {
      network            = network_interface.value.network != "" ? network_interface.value.network : null
      subnetwork         = network_interface.value.subnetwork != "" ? network_interface.value.subnetwork : null
      subnetwork_project = network_interface.value.subnetwork_project != "" ? network_interface.value.subnetwork_project : null
      # PSC consumer interface — legal on its own with no
      # network/subnetwork (the spec's CEL mirrors that rule).
      network_attachment = network_interface.value.network_attachment != "" ? network_interface.value.network_attachment : null
      network_ip         = network_interface.value.network_ip != "" ? network_interface.value.network_ip : null
      stack_type         = network_interface.value.stack_type != "" ? network_interface.value.stack_type : null
      nic_type           = network_interface.value.nic_type != "" ? network_interface.value.nic_type : null
      queue_count        = network_interface.value.queue_count
      # VLAN tag marks a dynamic sub-interface (2-255).
      vlan       = network_interface.value.vlan
      igmp_query = network_interface.value.igmp_query != "" ? network_interface.value.igmp_query : null
      # Static internal IPv6 — requires an IPv6-enabled stack_type and
      # subnetwork; unset lets GCP assign from the subnetwork range.
      ipv6_address                = network_interface.value.ipv6_address != "" ? network_interface.value.ipv6_address : null
      internal_ipv6_prefix_length = network_interface.value.internal_ipv6_prefix_length

      # An access_config grants a static (nat_ip) or ephemeral external
      # IPv4 -- the manifest's `ephemeral: true` is the request for the
      # latter and renders as a config with no nat_ip; absence keeps the VM
      # private (pair with Cloud NAT).
      dynamic "access_config" {
        for_each = network_interface.value.access_configs
        content {
          nat_ip                 = access_config.value.nat_ip != "" ? access_config.value.nat_ip : null
          network_tier           = access_config.value.network_tier != "" ? access_config.value.network_tier : null
          public_ptr_domain_name = access_config.value.public_ptr_domain_name != "" ? access_config.value.public_ptr_domain_name : null
        }
      }

      dynamic "ipv6_access_config" {
        for_each = network_interface.value.ipv6_access_configs
        content {
          network_tier           = ipv6_access_config.value.network_tier
          public_ptr_domain_name = ipv6_access_config.value.public_ptr_domain_name != "" ? ipv6_access_config.value.public_ptr_domain_name : null
          # Unset lets GCP assign the external range; these three are
          # ForceNew — pinning or renaming replaces the VM.
          external_ipv6               = ipv6_access_config.value.external_ipv6 != "" ? ipv6_access_config.value.external_ipv6 : null
          external_ipv6_prefix_length = ipv6_access_config.value.external_ipv6_prefix_length != "" ? ipv6_access_config.value.external_ipv6_prefix_length : null
          name                        = ipv6_access_config.value.name != "" ? ipv6_access_config.value.name : null
        }
      }

      dynamic "alias_ip_range" {
        for_each = network_interface.value.alias_ip_ranges
        content {
          ip_cidr_range         = alias_ip_range.value.ip_cidr_range
          subnetwork_range_name = alias_ip_range.value.subnetwork_range_name != "" ? alias_ip_range.value.subnetwork_range_name : null
        }
      }
    }
  }

  # Omitted block = Compute Engine default service account with its
  # default scopes; an explicit block with no email pins scopes on the
  # default account.
  dynamic "service_account" {
    for_each = var.spec.service_account != null ? [var.spec.service_account] : []
    content {
      email  = service_account.value.email != "" ? service_account.value.email : null
      scopes = service_account.value.scopes
    }
  }

  # preemptible + automatic_restart are DERIVED for Spot (see locals) —
  # the API requires both to agree with provisioning_model.
  dynamic "scheduling" {
    for_each = var.spec.scheduling != null || local.is_spot ? [1] : []
    content {
      provisioning_model          = var.spec.scheduling != null && var.spec.scheduling.provisioning_model != "" ? var.spec.scheduling.provisioning_model : null
      preemptible                 = local.is_spot
      automatic_restart           = local.automatic_restart
      on_host_maintenance         = var.spec.scheduling != null && var.spec.scheduling.on_host_maintenance != "" ? var.spec.scheduling.on_host_maintenance : null
      instance_termination_action = var.spec.scheduling != null && var.spec.scheduling.instance_termination_action != "" ? var.spec.scheduling.instance_termination_action : null
      termination_time            = var.spec.scheduling != null && var.spec.scheduling.termination_time != "" ? var.spec.scheduling.termination_time : null
      availability_domain         = var.spec.scheduling != null ? var.spec.scheduling.availability_domain : null
      min_node_cpus               = var.spec.scheduling != null ? var.spec.scheduling.min_node_cpus : null

      dynamic "max_run_duration" {
        for_each = var.spec.scheduling != null && var.spec.scheduling.max_run_duration_seconds != null ? [var.spec.scheduling.max_run_duration_seconds] : []
        content {
          seconds = max_run_duration.value
        }
      }

      dynamic "on_instance_stop_action" {
        for_each = var.spec.scheduling != null && var.spec.scheduling.discard_local_ssds_on_stop != null ? [var.spec.scheduling.discard_local_ssds_on_stop] : []
        content {
          discard_local_ssd = on_instance_stop_action.value
        }
      }

      dynamic "node_affinities" {
        for_each = var.spec.scheduling != null ? var.spec.scheduling.node_affinities : []
        content {
          key      = node_affinities.value.key
          operator = node_affinities.value.operator
          values   = node_affinities.value.values
        }
      }

      dynamic "local_ssd_recovery_timeout" {
        for_each = var.spec.scheduling != null && var.spec.scheduling.local_ssd_recovery_timeout_seconds != null ? [var.spec.scheduling.local_ssd_recovery_timeout_seconds] : []
        content {
          seconds = local_ssd_recovery_timeout.value
        }
      }
    }
  }

  # Shielded VM: unset booleans follow GCP defaults (secure boot off,
  # vTPM on, integrity monitoring on).
  dynamic "shielded_instance_config" {
    for_each = var.spec.shielded_instance_config != null ? [var.spec.shielded_instance_config] : []
    content {
      enable_secure_boot          = shielded_instance_config.value.enable_secure_boot != null ? shielded_instance_config.value.enable_secure_boot : false
      enable_vtpm                 = shielded_instance_config.value.enable_vtpm != null ? shielded_instance_config.value.enable_vtpm : true
      enable_integrity_monitoring = shielded_instance_config.value.enable_integrity_monitoring != null ? shielded_instance_config.value.enable_integrity_monitoring : true
    }
  }

  # Confidential VM: the typed field is the modern surface; the legacy
  # enable flag stays unset (it only supports SEV and will be deprecated).
  dynamic "confidential_instance_config" {
    for_each = var.spec.confidential_instance_config != null ? [var.spec.confidential_instance_config] : []
    content {
      confidential_instance_type = confidential_instance_config.value.confidential_instance_type
    }
  }

  dynamic "advanced_machine_features" {
    for_each = var.spec.advanced_machine_features != null ? [var.spec.advanced_machine_features] : []
    content {
      enable_nested_virtualization = advanced_machine_features.value.enable_nested_virtualization
      threads_per_core             = advanced_machine_features.value.threads_per_core
      visible_core_count           = advanced_machine_features.value.visible_core_count
      enable_uefi_networking       = advanced_machine_features.value.enable_uefi_networking
      performance_monitoring_unit  = advanced_machine_features.value.performance_monitoring_unit != "" ? advanced_machine_features.value.performance_monitoring_unit : null
      turbo_mode                   = advanced_machine_features.value.turbo_mode != "" ? advanced_machine_features.value.turbo_mode : null
    }
  }

  dynamic "guest_accelerator" {
    for_each = var.spec.guest_accelerators
    content {
      type  = guest_accelerator.value.type
      count = guest_accelerator.value.count
    }
  }

  dynamic "reservation_affinity" {
    for_each = var.spec.reservation_affinity != null ? [var.spec.reservation_affinity] : []
    content {
      type = reservation_affinity.value.type

      dynamic "specific_reservation" {
        for_each = reservation_affinity.value.specific_reservation != null ? [reservation_affinity.value.specific_reservation] : []
        content {
          key    = specific_reservation.value.key
          values = specific_reservation.value.values
        }
      }
    }
  }

  dynamic "network_performance_config" {
    for_each = local.total_egress_bandwidth_tier != null ? [local.total_egress_bandwidth_tier] : []
    content {
      total_egress_bandwidth_tier = network_performance_config.value
    }
  }

  # Resource Manager tags bind at create time only.
  dynamic "params" {
    for_each = length(var.spec.resource_manager_tags) > 0 ? [var.spec.resource_manager_tags] : []
    content {
      resource_manager_tags = params.value
    }
  }

  # Instance-level CMEK (memory and other instance state) — distinct
  # from the per-disk keys. CSEK raw keys are deliberately not modeled.
  dynamic "instance_encryption_key" {
    for_each = var.spec.instance_encryption_key != null ? [var.spec.instance_encryption_key] : []
    content {
      kms_key_self_link       = instance_encryption_key.value.kms_key
      kms_key_service_account = instance_encryption_key.value.kms_key_service_account != "" ? instance_encryption_key.value.kms_key_service_account : null
    }
  }

  depends_on = [
    google_project_service.compute_api,
  ]
}
