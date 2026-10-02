# Enable the Compute Engine API so a fresh project can host the group.
# disable_on_destroy is false: tearing down one group must never disable
# Compute Engine for everything else in the project.
resource "google_project_service" "compute_api" {
  project = local.project_id
  service = "compute.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# ----------------------------------------------------------------------------
# Instance template — what every VM in the group looks like.
#
# Templates are IMMUTABLE in GCP (labels excepted): name_prefix +
# create_before_destroy is the native rotation pattern — every template
# change creates a fresh "<mig-name>-<timestamp>" template FIRST, the
# group manager repoints its version, then the old template is deleted.
# The group is never left referencing a deleted template.
#
# Zonal groups use the global instance template resource, whose
# self_link_unique carries the ?uniqueId= that keys the group manager's
# rotation-aware diff suppression. The regional template has no unique
# variant (rotation is carried by the name_prefix change producing a new
# self_link) and — unlike the zonal resource — carries a deletion_policy.
# ----------------------------------------------------------------------------

resource "google_compute_instance_template" "this" {
  count = local.is_regional ? 0 : 1

  name_prefix = local.template_name_prefix
  project     = local.project_id

  machine_type         = var.spec.template.machine_type
  description          = var.spec.template.description != "" ? var.spec.template.description : null
  instance_description = var.spec.template.instance_description != "" ? var.spec.template.instance_description : null

  labels = local.final_labels

  dynamic "disk" {
    for_each = var.spec.template.disks
    content {
      boot            = disk.value.boot
      source_image    = disk.value.source_image != "" ? disk.value.source_image : null
      source_snapshot = disk.value.source_snapshot != "" ? disk.value.source_snapshot : null
      source          = disk.value.source != "" ? disk.value.source : null
      disk_size_gb    = disk.value.size_gb != 0 ? disk.value.size_gb : null
      disk_type       = disk.value.disk_type != "" ? disk.value.disk_type : null
      type            = disk.value.type != "" ? disk.value.type : null
      # Defaults true (matching GCP); explicit-send so a stated false
      # always reaches the engine.
      auto_delete = coalesce(disk.value.auto_delete, true)
      device_name = disk.value.device_name != "" ? disk.value.device_name : null
      disk_name   = disk.value.disk_name != "" ? disk.value.disk_name : null
      mode        = disk.value.mode != "" ? disk.value.mode : null
      # Google-advice-only lever — sent only when explicitly set.
      interface              = disk.value.interface != "" ? disk.value.interface : null
      labels                 = length(disk.value.disk_labels) > 0 ? disk.value.disk_labels : null
      provisioned_iops       = disk.value.provisioned_iops
      provisioned_throughput = disk.value.provisioned_throughput
      architecture           = disk.value.architecture != "" ? disk.value.architecture : null
      guest_os_features      = length(disk.value.guest_os_features) > 0 ? disk.value.guest_os_features : null
      # The provider flattens this max-1 list; the spec's repeated shape
      # (capped at 1 by validation) mirrors the provider API.
      resource_policies     = length(disk.value.resource_policies) > 0 ? disk.value.resource_policies : null
      resource_manager_tags = length(disk.value.resource_manager_tags) > 0 ? disk.value.resource_manager_tags : null
      storage_pool          = disk.value.storage_pool != "" ? disk.value.storage_pool : null

      # CMEK only — CSEK raw-key arms are deliberately not modeled
      # (secure-by-default).
      dynamic "disk_encryption_key" {
        for_each = disk.value.disk_encryption != null ? [disk.value.disk_encryption] : []
        content {
          kms_key_self_link       = disk_encryption_key.value.kms_key
          kms_key_service_account = disk_encryption_key.value.kms_key_service_account != "" ? disk_encryption_key.value.kms_key_service_account : null
        }
      }
      dynamic "source_image_encryption_key" {
        for_each = disk.value.source_image_encryption != null ? [disk.value.source_image_encryption] : []
        content {
          kms_key_self_link       = source_image_encryption_key.value.kms_key
          kms_key_service_account = source_image_encryption_key.value.kms_key_service_account != "" ? source_image_encryption_key.value.kms_key_service_account : null
        }
      }
      dynamic "source_snapshot_encryption_key" {
        for_each = disk.value.source_snapshot_encryption != null ? [disk.value.source_snapshot_encryption] : []
        content {
          kms_key_self_link       = source_snapshot_encryption_key.value.kms_key
          kms_key_service_account = source_snapshot_encryption_key.value.kms_key_service_account != "" ? source_snapshot_encryption_key.value.kms_key_service_account : null
        }
      }
    }
  }

  dynamic "network_interface" {
    for_each = var.spec.template.network_interfaces
    content {
      network            = network_interface.value.network != "" ? network_interface.value.network : null
      subnetwork         = network_interface.value.subnetwork != "" ? network_interface.value.subnetwork : null
      subnetwork_project = network_interface.value.subnetwork_project != "" ? network_interface.value.subnetwork_project : null
      # PSC consumer interface — legal on its own with no
      # network/subnetwork (the spec's CEL mirrors that rule).
      network_attachment          = network_interface.value.network_attachment != "" ? network_interface.value.network_attachment : null
      network_ip                  = network_interface.value.network_ip != "" ? network_interface.value.network_ip : null
      stack_type                  = network_interface.value.stack_type != "" ? network_interface.value.stack_type : null
      nic_type                    = network_interface.value.nic_type != "" ? network_interface.value.nic_type : null
      queue_count                 = network_interface.value.queue_count
      vlan                        = network_interface.value.vlan
      igmp_query                  = network_interface.value.igmp_query != "" ? network_interface.value.igmp_query : null
      ipv6_address                = network_interface.value.ipv6_address != "" ? network_interface.value.ipv6_address : null
      internal_ipv6_prefix_length = network_interface.value.internal_ipv6_prefix_length

      # Presence grants an external IPv4; absence keeps the fleet
      # private (pair with Cloud NAT for egress — the secure default).
      dynamic "access_config" {
        for_each = network_interface.value.access_configs
        content {
          nat_ip       = access_config.value.nat_ip != "" ? access_config.value.nat_ip : null
          network_tier = access_config.value.network_tier != "" ? access_config.value.network_tier : null
        }
      }
      dynamic "ipv6_access_config" {
        for_each = network_interface.value.ipv6_access_configs
        content {
          network_tier = ipv6_access_config.value.network_tier
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
    for_each = var.spec.template.service_account != null ? [var.spec.template.service_account] : []
    content {
      email  = service_account.value.email != "" ? service_account.value.email : null
      scopes = service_account.value.scopes
    }
  }

  # Spot semantics: SPOT requires the API's legacy preemptible flag and
  # forbids automatic restart — both derived here, identically to the
  # Pulumi module, so the spec's single switch stays honest.
  dynamic "scheduling" {
    for_each = var.spec.template.scheduling != null ? [var.spec.template.scheduling] : []
    content {
      preemptible                 = scheduling.value.provisioning_model == "SPOT"
      automatic_restart           = local.scheduling_is_reclaimable ? false : coalesce(scheduling.value.automatic_restart, true)
      on_host_maintenance         = scheduling.value.on_host_maintenance != "" ? scheduling.value.on_host_maintenance : null
      provisioning_model          = scheduling.value.provisioning_model != "" ? scheduling.value.provisioning_model : null
      instance_termination_action = scheduling.value.instance_termination_action != "" ? scheduling.value.instance_termination_action : null
      availability_domain         = scheduling.value.availability_domain
      min_node_cpus               = scheduling.value.min_node_cpus
      termination_time            = scheduling.value.termination_time != "" ? scheduling.value.termination_time : null

      dynamic "max_run_duration" {
        for_each = scheduling.value.max_run_duration_seconds != null ? [scheduling.value.max_run_duration_seconds] : []
        content {
          seconds = max_run_duration.value
        }
      }
      dynamic "on_instance_stop_action" {
        for_each = scheduling.value.discard_local_ssds_on_stop != null ? [scheduling.value.discard_local_ssds_on_stop] : []
        content {
          discard_local_ssd = on_instance_stop_action.value
        }
      }
      dynamic "node_affinities" {
        for_each = scheduling.value.node_affinities
        content {
          key      = node_affinities.value.key
          operator = node_affinities.value.operator
          values   = node_affinities.value.values
        }
      }
      dynamic "local_ssd_recovery_timeout" {
        for_each = scheduling.value.local_ssd_recovery_timeout_seconds != null ? [scheduling.value.local_ssd_recovery_timeout_seconds] : []
        content {
          seconds = local_ssd_recovery_timeout.value
        }
      }
      # Host-error detection timeout (90..330 s, steps of 30). Sent only
      # when set so an unset spec keeps Compute Engine's default recovery.
      host_error_timeout_seconds = scheduling.value.host_error_timeout_seconds
    }
  }

  # Shielded VM: unset booleans follow GCP defaults (secure boot off,
  # vTPM on, integrity monitoring on) — sent explicitly.
  dynamic "shielded_instance_config" {
    for_each = var.spec.template.shielded_instance_config != null ? [var.spec.template.shielded_instance_config] : []
    content {
      enable_secure_boot          = coalesce(shielded_instance_config.value.enable_secure_boot, false)
      enable_vtpm                 = coalesce(shielded_instance_config.value.enable_vtpm, true)
      enable_integrity_monitoring = coalesce(shielded_instance_config.value.enable_integrity_monitoring, true)
    }
  }

  # The typed field is the modern surface; the legacy enable flag stays
  # unset (SEV-only, headed for deprecation).
  dynamic "confidential_instance_config" {
    for_each = var.spec.template.confidential_instance_config != null ? [var.spec.template.confidential_instance_config] : []
    content {
      confidential_instance_type = confidential_instance_config.value.confidential_instance_type
    }
  }

  dynamic "advanced_machine_features" {
    for_each = var.spec.template.advanced_machine_features != null ? [var.spec.template.advanced_machine_features] : []
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
    for_each = var.spec.template.guest_accelerators
    content {
      type  = guest_accelerator.value.type
      count = guest_accelerator.value.count
    }
  }

  dynamic "reservation_affinity" {
    for_each = var.spec.template.reservation_affinity != null ? [var.spec.template.reservation_affinity] : []
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
    for_each = var.spec.template.total_egress_bandwidth_tier != "" ? [var.spec.template.total_egress_bandwidth_tier] : []
    content {
      total_egress_bandwidth_tier = network_performance_config.value
    }
  }

  metadata = length(var.spec.template.metadata) > 0 ? var.spec.template.metadata : null
  # The dedicated attribute (never plain metadata) so the script re-runs
  # on every boot exactly as GCP documents.
  metadata_startup_script = var.spec.template.startup_script != "" ? var.spec.template.startup_script : null

  tags                  = length(var.spec.template.tags) > 0 ? var.spec.template.tags : null
  resource_manager_tags = length(var.spec.template.resource_manager_tags) > 0 ? var.spec.template.resource_manager_tags : null

  min_cpu_platform           = var.spec.template.min_cpu_platform != "" ? var.spec.template.min_cpu_platform : null
  can_ip_forward             = var.spec.template.can_ip_forward
  key_revocation_action_type = var.spec.template.key_revocation_action_type != "" ? var.spec.template.key_revocation_action_type : null
  resource_policies          = length(var.spec.template.resource_policies) > 0 ? var.spec.template.resource_policies : null

  # Managed workload identity: a SPIFFE ID issued to each VM, optionally
  # with X.509 identity certificates for mutual TLS. A template field, so
  # changing it rotates the template.
  dynamic "workload_identity_config" {
    for_each = var.spec.template.workload_identity_config != null ? [var.spec.template.workload_identity_config] : []
    content {
      identity                     = workload_identity_config.value.identity
      identity_certificate_enabled = workload_identity_config.value.identity_certificate_enabled
    }
  }

  lifecycle {
    create_before_destroy = true
  }

  depends_on = [google_project_service.compute_api]
}

resource "google_compute_region_instance_template" "this" {
  count = local.is_regional ? 1 : 0

  name_prefix = local.template_name_prefix
  project     = local.project_id
  region      = var.spec.region

  machine_type         = var.spec.template.machine_type
  description          = var.spec.template.description != "" ? var.spec.template.description : null
  instance_description = var.spec.template.instance_description != "" ? var.spec.template.instance_description : null

  labels = local.final_labels

  dynamic "disk" {
    for_each = var.spec.template.disks
    content {
      boot                   = disk.value.boot
      source_image           = disk.value.source_image != "" ? disk.value.source_image : null
      source_snapshot        = disk.value.source_snapshot != "" ? disk.value.source_snapshot : null
      source                 = disk.value.source != "" ? disk.value.source : null
      disk_size_gb           = disk.value.size_gb != 0 ? disk.value.size_gb : null
      disk_type              = disk.value.disk_type != "" ? disk.value.disk_type : null
      type                   = disk.value.type != "" ? disk.value.type : null
      auto_delete            = coalesce(disk.value.auto_delete, true)
      device_name            = disk.value.device_name != "" ? disk.value.device_name : null
      disk_name              = disk.value.disk_name != "" ? disk.value.disk_name : null
      mode                   = disk.value.mode != "" ? disk.value.mode : null
      interface              = disk.value.interface != "" ? disk.value.interface : null
      labels                 = length(disk.value.disk_labels) > 0 ? disk.value.disk_labels : null
      provisioned_iops       = disk.value.provisioned_iops
      provisioned_throughput = disk.value.provisioned_throughput
      architecture           = disk.value.architecture != "" ? disk.value.architecture : null
      guest_os_features      = length(disk.value.guest_os_features) > 0 ? disk.value.guest_os_features : null
      resource_policies      = length(disk.value.resource_policies) > 0 ? disk.value.resource_policies : null
      resource_manager_tags  = length(disk.value.resource_manager_tags) > 0 ? disk.value.resource_manager_tags : null
      storage_pool           = disk.value.storage_pool != "" ? disk.value.storage_pool : null

      dynamic "disk_encryption_key" {
        for_each = disk.value.disk_encryption != null ? [disk.value.disk_encryption] : []
        content {
          kms_key_self_link       = disk_encryption_key.value.kms_key
          kms_key_service_account = disk_encryption_key.value.kms_key_service_account != "" ? disk_encryption_key.value.kms_key_service_account : null
        }
      }
      dynamic "source_image_encryption_key" {
        for_each = disk.value.source_image_encryption != null ? [disk.value.source_image_encryption] : []
        content {
          kms_key_self_link       = source_image_encryption_key.value.kms_key
          kms_key_service_account = source_image_encryption_key.value.kms_key_service_account != "" ? source_image_encryption_key.value.kms_key_service_account : null
        }
      }
      dynamic "source_snapshot_encryption_key" {
        for_each = disk.value.source_snapshot_encryption != null ? [disk.value.source_snapshot_encryption] : []
        content {
          kms_key_self_link       = source_snapshot_encryption_key.value.kms_key
          kms_key_service_account = source_snapshot_encryption_key.value.kms_key_service_account != "" ? source_snapshot_encryption_key.value.kms_key_service_account : null
        }
      }
    }
  }

  dynamic "network_interface" {
    for_each = var.spec.template.network_interfaces
    content {
      network                     = network_interface.value.network != "" ? network_interface.value.network : null
      subnetwork                  = network_interface.value.subnetwork != "" ? network_interface.value.subnetwork : null
      subnetwork_project          = network_interface.value.subnetwork_project != "" ? network_interface.value.subnetwork_project : null
      network_attachment          = network_interface.value.network_attachment != "" ? network_interface.value.network_attachment : null
      network_ip                  = network_interface.value.network_ip != "" ? network_interface.value.network_ip : null
      stack_type                  = network_interface.value.stack_type != "" ? network_interface.value.stack_type : null
      nic_type                    = network_interface.value.nic_type != "" ? network_interface.value.nic_type : null
      queue_count                 = network_interface.value.queue_count
      vlan                        = network_interface.value.vlan
      igmp_query                  = network_interface.value.igmp_query != "" ? network_interface.value.igmp_query : null
      ipv6_address                = network_interface.value.ipv6_address != "" ? network_interface.value.ipv6_address : null
      internal_ipv6_prefix_length = network_interface.value.internal_ipv6_prefix_length

      dynamic "access_config" {
        for_each = network_interface.value.access_configs
        content {
          nat_ip       = access_config.value.nat_ip != "" ? access_config.value.nat_ip : null
          network_tier = access_config.value.network_tier != "" ? access_config.value.network_tier : null
        }
      }
      dynamic "ipv6_access_config" {
        for_each = network_interface.value.ipv6_access_configs
        content {
          network_tier = ipv6_access_config.value.network_tier
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

  dynamic "service_account" {
    for_each = var.spec.template.service_account != null ? [var.spec.template.service_account] : []
    content {
      email  = service_account.value.email != "" ? service_account.value.email : null
      scopes = service_account.value.scopes
    }
  }

  dynamic "scheduling" {
    for_each = var.spec.template.scheduling != null ? [var.spec.template.scheduling] : []
    content {
      preemptible                 = scheduling.value.provisioning_model == "SPOT"
      automatic_restart           = local.scheduling_is_reclaimable ? false : coalesce(scheduling.value.automatic_restart, true)
      on_host_maintenance         = scheduling.value.on_host_maintenance != "" ? scheduling.value.on_host_maintenance : null
      provisioning_model          = scheduling.value.provisioning_model != "" ? scheduling.value.provisioning_model : null
      instance_termination_action = scheduling.value.instance_termination_action != "" ? scheduling.value.instance_termination_action : null
      availability_domain         = scheduling.value.availability_domain
      min_node_cpus               = scheduling.value.min_node_cpus
      termination_time            = scheduling.value.termination_time != "" ? scheduling.value.termination_time : null

      dynamic "max_run_duration" {
        for_each = scheduling.value.max_run_duration_seconds != null ? [scheduling.value.max_run_duration_seconds] : []
        content {
          seconds = max_run_duration.value
        }
      }
      dynamic "on_instance_stop_action" {
        for_each = scheduling.value.discard_local_ssds_on_stop != null ? [scheduling.value.discard_local_ssds_on_stop] : []
        content {
          discard_local_ssd = on_instance_stop_action.value
        }
      }
      dynamic "node_affinities" {
        for_each = scheduling.value.node_affinities
        content {
          key      = node_affinities.value.key
          operator = node_affinities.value.operator
          values   = node_affinities.value.values
        }
      }
      dynamic "local_ssd_recovery_timeout" {
        for_each = scheduling.value.local_ssd_recovery_timeout_seconds != null ? [scheduling.value.local_ssd_recovery_timeout_seconds] : []
        content {
          seconds = local_ssd_recovery_timeout.value
        }
      }
      # Host-error detection timeout (90..330 s, steps of 30). Sent only
      # when set so an unset spec keeps Compute Engine's default recovery.
      host_error_timeout_seconds = scheduling.value.host_error_timeout_seconds
    }
  }

  dynamic "shielded_instance_config" {
    for_each = var.spec.template.shielded_instance_config != null ? [var.spec.template.shielded_instance_config] : []
    content {
      enable_secure_boot          = coalesce(shielded_instance_config.value.enable_secure_boot, false)
      enable_vtpm                 = coalesce(shielded_instance_config.value.enable_vtpm, true)
      enable_integrity_monitoring = coalesce(shielded_instance_config.value.enable_integrity_monitoring, true)
    }
  }

  dynamic "confidential_instance_config" {
    for_each = var.spec.template.confidential_instance_config != null ? [var.spec.template.confidential_instance_config] : []
    content {
      confidential_instance_type = confidential_instance_config.value.confidential_instance_type
    }
  }

  dynamic "advanced_machine_features" {
    for_each = var.spec.template.advanced_machine_features != null ? [var.spec.template.advanced_machine_features] : []
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
    for_each = var.spec.template.guest_accelerators
    content {
      type  = guest_accelerator.value.type
      count = guest_accelerator.value.count
    }
  }

  dynamic "reservation_affinity" {
    for_each = var.spec.template.reservation_affinity != null ? [var.spec.template.reservation_affinity] : []
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
    for_each = var.spec.template.total_egress_bandwidth_tier != "" ? [var.spec.template.total_egress_bandwidth_tier] : []
    content {
      total_egress_bandwidth_tier = network_performance_config.value
    }
  }

  metadata                = length(var.spec.template.metadata) > 0 ? var.spec.template.metadata : null
  metadata_startup_script = var.spec.template.startup_script != "" ? var.spec.template.startup_script : null

  tags                  = length(var.spec.template.tags) > 0 ? var.spec.template.tags : null
  resource_manager_tags = length(var.spec.template.resource_manager_tags) > 0 ? var.spec.template.resource_manager_tags : null

  min_cpu_platform           = var.spec.template.min_cpu_platform != "" ? var.spec.template.min_cpu_platform : null
  can_ip_forward             = var.spec.template.can_ip_forward
  key_revocation_action_type = var.spec.template.key_revocation_action_type != "" ? var.spec.template.key_revocation_action_type : null
  resource_policies          = length(var.spec.template.resource_policies) > 0 ? var.spec.template.resource_policies : null

  # Managed workload identity: a SPIFFE ID issued to each VM, optionally
  # with X.509 identity certificates for mutual TLS. A template field, so
  # changing it rotates the template.
  dynamic "workload_identity_config" {
    for_each = var.spec.template.workload_identity_config != null ? [var.spec.template.workload_identity_config] : []
    content {
      identity                     = workload_identity_config.value.identity
      identity_certificate_enabled = workload_identity_config.value.identity_certificate_enabled
    }
  }

  # Only the REGIONAL template carries a deletion_policy in the
  # provider — the zonal one has none (always deleted on destroy).
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  lifecycle {
    create_before_destroy = true
  }

  depends_on = [google_project_service.compute_api]
}

# ----------------------------------------------------------------------------
# Instance group manager — how many VMs run, how they roll out changes,
# how failed VMs are repaired. The version block references the
# template's self_link_unique (zonal) so a template ROTATION genuinely
# rolls the version; the regional group references the regional
# template's self_link (rotation rides the name_prefix change).
# ----------------------------------------------------------------------------

resource "google_compute_instance_group_manager" "this" {
  count = local.is_regional ? 0 : 1

  name               = local.mig_name
  project            = local.project_id
  zone               = var.spec.zone
  base_instance_name = local.base_instance_name
  description        = var.spec.description != "" ? var.spec.description : null

  dynamic "version" {
    for_each = local.versions
    content {
      name              = version.value.version_name != "" ? version.value.version_name : null
      instance_template = version.value.template_self_link != "" ? version.value.template_self_link : google_compute_instance_template.this[0].self_link_unique
      dynamic "target_size" {
        for_each = version.value.target_size_fixed != null || version.value.target_size_percent != null ? [version.value] : []
        content {
          fixed   = target_size.value.target_size_fixed
          percent = target_size.value.target_size_percent
        }
      }
    }
  }

  # Never set alongside an autoscaler (the spec's CEL walls that off) —
  # the autoscaler would fight a fixed size on every apply.
  target_size = var.spec.target_size

  dynamic "named_port" {
    for_each = var.spec.named_ports
    content {
      name = named_port.value.name
      port = named_port.value.port
    }
  }

  dynamic "update_policy" {
    for_each = var.spec.update_policy != null ? [var.spec.update_policy] : []
    content {
      minimal_action                 = update_policy.value.minimal_action
      type                           = update_policy.value.type
      most_disruptive_allowed_action = update_policy.value.most_disruptive_allowed_action != "" ? update_policy.value.most_disruptive_allowed_action : null
      replacement_method             = update_policy.value.replacement_method != "" ? update_policy.value.replacement_method : null
      max_surge_fixed                = update_policy.value.max_surge_fixed
      max_surge_percent              = update_policy.value.max_surge_percent
      max_unavailable_fixed          = update_policy.value.max_unavailable_fixed
      max_unavailable_percent        = update_policy.value.max_unavailable_percent
    }
  }

  dynamic "auto_healing_policies" {
    for_each = var.spec.auto_healing != null ? [var.spec.auto_healing] : []
    content {
      health_check      = auto_healing_policies.value.health_check
      initial_delay_sec = auto_healing_policies.value.initial_delay_sec
    }
  }

  dynamic "standby_policy" {
    for_each = var.spec.standby_policy != null ? [var.spec.standby_policy] : []
    content {
      initial_delay_sec = standby_policy.value.initial_delay_sec
      mode              = standby_policy.value.mode != "" ? standby_policy.value.mode : null
    }
  }
  target_suspended_size = var.spec.target_suspended_size
  target_stopped_size   = var.spec.target_stopped_size

  dynamic "stateful_disk" {
    for_each = var.spec.stateful_disks
    content {
      device_name = stateful_disk.value.device_name
      delete_rule = stateful_disk.value.delete_rule != "" ? stateful_disk.value.delete_rule : null
    }
  }
  dynamic "stateful_external_ip" {
    for_each = var.spec.stateful_external_ips
    content {
      interface_name = stateful_external_ip.value.interface_name != "" ? stateful_external_ip.value.interface_name : null
      delete_rule    = stateful_external_ip.value.delete_rule != "" ? stateful_external_ip.value.delete_rule : null
    }
  }
  dynamic "stateful_internal_ip" {
    for_each = var.spec.stateful_internal_ips
    content {
      interface_name = stateful_internal_ip.value.interface_name != "" ? stateful_internal_ip.value.interface_name : null
      delete_rule    = stateful_internal_ip.value.delete_rule != "" ? stateful_internal_ip.value.delete_rule : null
    }
  }

  dynamic "instance_lifecycle_policy" {
    for_each = var.spec.instance_lifecycle_policy != null ? [var.spec.instance_lifecycle_policy] : []
    content {
      default_action_on_failure = instance_lifecycle_policy.value.default_action_on_failure != "" ? instance_lifecycle_policy.value.default_action_on_failure : null
      force_update_on_repair    = instance_lifecycle_policy.value.force_update_on_repair != "" ? instance_lifecycle_policy.value.force_update_on_repair : null
      on_failed_health_check    = instance_lifecycle_policy.value.on_failed_health_check != "" ? instance_lifecycle_policy.value.on_failed_health_check : null
      dynamic "on_repair" {
        for_each = instance_lifecycle_policy.value.on_repair_allow_changing_zone != "" ? [instance_lifecycle_policy.value.on_repair_allow_changing_zone] : []
        content {
          allow_changing_zone = on_repair.value
        }
      }
    }
  }

  dynamic "all_instances_config" {
    for_each = var.spec.all_instances_config != null ? [var.spec.all_instances_config] : []
    content {
      labels   = length(all_instances_config.value.labels) > 0 ? all_instances_config.value.labels : null
      metadata = length(all_instances_config.value.metadata) > 0 ? all_instances_config.value.metadata : null
    }
  }

  list_managed_instances_results = var.spec.list_managed_instances_results != "" ? var.spec.list_managed_instances_results : null

  dynamic "resource_policies" {
    for_each = var.spec.workload_policy != "" ? [var.spec.workload_policy] : []
    content {
      workload_policy = resource_policies.value
    }
  }

  target_pools = length(var.spec.target_pools) > 0 ? var.spec.target_pools : null

  wait_for_instances        = var.spec.wait_for_instances
  wait_for_instances_status = var.spec.wait_for_instances_status != "" ? var.spec.wait_for_instances_status : null

  dynamic "target_size_policy" {
    for_each = var.spec.target_size_policy_mode != "" ? [var.spec.target_size_policy_mode] : []
    content {
      mode = target_size_policy.value
    }
  }

  # Empty defers to the provider default (DELETE).
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}

resource "google_compute_region_instance_group_manager" "this" {
  count = local.is_regional ? 1 : 0

  name               = local.mig_name
  project            = local.project_id
  region             = var.spec.region
  base_instance_name = local.base_instance_name
  description        = var.spec.description != "" ? var.spec.description : null

  dynamic "version" {
    for_each = local.versions
    content {
      name              = version.value.version_name != "" ? version.value.version_name : null
      instance_template = version.value.template_self_link != "" ? version.value.template_self_link : google_compute_region_instance_template.this[0].self_link
      dynamic "target_size" {
        for_each = version.value.target_size_fixed != null || version.value.target_size_percent != null ? [version.value] : []
        content {
          fixed   = target_size.value.target_size_fixed
          percent = target_size.value.target_size_percent
        }
      }
    }
  }

  target_size = var.spec.target_size

  dynamic "named_port" {
    for_each = var.spec.named_ports
    content {
      name = named_port.value.name
      port = named_port.value.port
    }
  }

  dynamic "update_policy" {
    for_each = var.spec.update_policy != null ? [var.spec.update_policy] : []
    content {
      minimal_action                 = update_policy.value.minimal_action
      type                           = update_policy.value.type
      most_disruptive_allowed_action = update_policy.value.most_disruptive_allowed_action != "" ? update_policy.value.most_disruptive_allowed_action : null
      replacement_method             = update_policy.value.replacement_method != "" ? update_policy.value.replacement_method : null
      max_surge_fixed                = update_policy.value.max_surge_fixed
      max_surge_percent              = update_policy.value.max_surge_percent
      max_unavailable_fixed          = update_policy.value.max_unavailable_fixed
      max_unavailable_percent        = update_policy.value.max_unavailable_percent
      # Regional-only: PROACTIVE (default) rebalances zones; NONE is
      # required for stateful regional groups.
      instance_redistribution_type = update_policy.value.instance_redistribution_type != "" ? update_policy.value.instance_redistribution_type : null
    }
  }

  dynamic "auto_healing_policies" {
    for_each = var.spec.auto_healing != null ? [var.spec.auto_healing] : []
    content {
      health_check      = auto_healing_policies.value.health_check
      initial_delay_sec = auto_healing_policies.value.initial_delay_sec
    }
  }

  dynamic "standby_policy" {
    for_each = var.spec.standby_policy != null ? [var.spec.standby_policy] : []
    content {
      initial_delay_sec = standby_policy.value.initial_delay_sec
      mode              = standby_policy.value.mode != "" ? standby_policy.value.mode : null
    }
  }
  target_suspended_size = var.spec.target_suspended_size
  target_stopped_size   = var.spec.target_stopped_size

  dynamic "stateful_disk" {
    for_each = var.spec.stateful_disks
    content {
      device_name = stateful_disk.value.device_name
      delete_rule = stateful_disk.value.delete_rule != "" ? stateful_disk.value.delete_rule : null
    }
  }
  dynamic "stateful_external_ip" {
    for_each = var.spec.stateful_external_ips
    content {
      interface_name = stateful_external_ip.value.interface_name != "" ? stateful_external_ip.value.interface_name : null
      delete_rule    = stateful_external_ip.value.delete_rule != "" ? stateful_external_ip.value.delete_rule : null
    }
  }
  dynamic "stateful_internal_ip" {
    for_each = var.spec.stateful_internal_ips
    content {
      interface_name = stateful_internal_ip.value.interface_name != "" ? stateful_internal_ip.value.interface_name : null
      delete_rule    = stateful_internal_ip.value.delete_rule != "" ? stateful_internal_ip.value.delete_rule : null
    }
  }

  dynamic "instance_lifecycle_policy" {
    for_each = var.spec.instance_lifecycle_policy != null ? [var.spec.instance_lifecycle_policy] : []
    content {
      default_action_on_failure = instance_lifecycle_policy.value.default_action_on_failure != "" ? instance_lifecycle_policy.value.default_action_on_failure : null
      force_update_on_repair    = instance_lifecycle_policy.value.force_update_on_repair != "" ? instance_lifecycle_policy.value.force_update_on_repair : null
      on_failed_health_check    = instance_lifecycle_policy.value.on_failed_health_check != "" ? instance_lifecycle_policy.value.on_failed_health_check : null
      dynamic "on_repair" {
        for_each = instance_lifecycle_policy.value.on_repair_allow_changing_zone != "" ? [instance_lifecycle_policy.value.on_repair_allow_changing_zone] : []
        content {
          allow_changing_zone = on_repair.value
        }
      }
    }
  }

  dynamic "all_instances_config" {
    for_each = var.spec.all_instances_config != null ? [var.spec.all_instances_config] : []
    content {
      labels   = length(all_instances_config.value.labels) > 0 ? all_instances_config.value.labels : null
      metadata = length(all_instances_config.value.metadata) > 0 ? all_instances_config.value.metadata : null
    }
  }

  list_managed_instances_results = var.spec.list_managed_instances_results != "" ? var.spec.list_managed_instances_results : null

  dynamic "resource_policies" {
    for_each = var.spec.workload_policy != "" ? [var.spec.workload_policy] : []
    content {
      workload_policy = resource_policies.value
    }
  }

  target_pools = length(var.spec.target_pools) > 0 ? var.spec.target_pools : null

  wait_for_instances        = var.spec.wait_for_instances
  wait_for_instances_status = var.spec.wait_for_instances_status != "" ? var.spec.wait_for_instances_status : null

  # Regional-only spread controls. When omitted, GCP spreads evenly
  # across the region's zones.
  distribution_policy_zones        = var.spec.distribution_policy != null && length(var.spec.distribution_policy.zones) > 0 ? var.spec.distribution_policy.zones : null
  distribution_policy_target_shape = var.spec.distribution_policy != null && var.spec.distribution_policy.target_shape != "" ? var.spec.distribution_policy.target_shape : null

  dynamic "instance_flexibility_policy" {
    for_each = var.spec.instance_flexibility_policy != null ? [var.spec.instance_flexibility_policy] : []
    content {
      dynamic "instance_selections" {
        for_each = instance_flexibility_policy.value.instance_selections
        content {
          name          = instance_selections.value.name
          machine_types = instance_selections.value.machine_types
          rank          = instance_selections.value.rank
        }
      }
    }
  }

  dynamic "target_size_policy" {
    for_each = var.spec.target_size_policy_mode != "" ? [var.spec.target_size_policy_mode] : []
    content {
      mode = target_size_policy.value
    }
  }

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}

# ----------------------------------------------------------------------------
# Autoscaler — grows and shrinks the fleet between min/max replicas.
# Targeted at the group manager by resource reference (dependency order
# and the ambient-project case stay correct).
# ----------------------------------------------------------------------------

resource "google_compute_autoscaler" "this" {
  count = !local.is_regional && var.spec.autoscaler != null ? 1 : 0

  name        = local.autoscaler_name
  project     = local.project_id
  zone        = var.spec.zone
  target      = google_compute_instance_group_manager.this[0].self_link
  description = var.spec.autoscaler.description != "" ? var.spec.autoscaler.description : null

  autoscaling_policy {
    min_replicas    = var.spec.autoscaler.min_replicas
    max_replicas    = var.spec.autoscaler.max_replicas
    cooldown_period = var.spec.autoscaler.cooldown_period
    mode            = var.spec.autoscaler.mode != "" ? var.spec.autoscaler.mode : null

    dynamic "cpu_utilization" {
      for_each = var.spec.autoscaler.cpu_target != null || var.spec.autoscaler.cpu_predictive_method != "" ? [var.spec.autoscaler] : []
      content {
        target            = cpu_utilization.value.cpu_target
        predictive_method = cpu_utilization.value.cpu_predictive_method != "" ? cpu_utilization.value.cpu_predictive_method : null
      }
    }

    dynamic "load_balancing_utilization" {
      for_each = var.spec.autoscaler.load_balancing_target != null ? [var.spec.autoscaler.load_balancing_target] : []
      content {
        target = load_balancing_utilization.value
      }
    }

    dynamic "metric" {
      for_each = var.spec.autoscaler.metrics
      content {
        name                       = metric.value.name
        target                     = metric.value.target
        type                       = metric.value.type != "" ? metric.value.type : null
        filter                     = metric.value.filter != "" ? metric.value.filter : null
        single_instance_assignment = metric.value.single_instance_assignment
      }
    }

    dynamic "scale_in_control" {
      for_each = var.spec.autoscaler.scale_in_control != null ? [var.spec.autoscaler.scale_in_control] : []
      content {
        dynamic "max_scaled_in_replicas" {
          for_each = scale_in_control.value.max_scaled_in_replicas_fixed != null || scale_in_control.value.max_scaled_in_replicas_percent != null ? [scale_in_control.value] : []
          content {
            fixed   = max_scaled_in_replicas.value.max_scaled_in_replicas_fixed
            percent = max_scaled_in_replicas.value.max_scaled_in_replicas_percent
          }
        }
        time_window_sec = scale_in_control.value.time_window_sec
      }
    }

    dynamic "scaling_schedules" {
      for_each = var.spec.autoscaler.schedules
      content {
        name                  = scaling_schedules.value.schedule_name
        schedule              = scaling_schedules.value.schedule
        duration_sec          = scaling_schedules.value.duration_sec
        min_required_replicas = scaling_schedules.value.min_required_replicas
        disabled              = scaling_schedules.value.disabled
        time_zone             = scaling_schedules.value.time_zone != "" ? scaling_schedules.value.time_zone : null
        description           = scaling_schedules.value.description != "" ? scaling_schedules.value.description : null
      }
    }

    stabilization_period = var.spec.autoscaler.stabilization_period
  }

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}

resource "google_compute_region_autoscaler" "this" {
  count = local.is_regional && var.spec.autoscaler != null ? 1 : 0

  name        = local.autoscaler_name
  project     = local.project_id
  region      = var.spec.region
  target      = google_compute_region_instance_group_manager.this[0].self_link
  description = var.spec.autoscaler.description != "" ? var.spec.autoscaler.description : null

  autoscaling_policy {
    min_replicas    = var.spec.autoscaler.min_replicas
    max_replicas    = var.spec.autoscaler.max_replicas
    cooldown_period = var.spec.autoscaler.cooldown_period
    mode            = var.spec.autoscaler.mode != "" ? var.spec.autoscaler.mode : null

    dynamic "cpu_utilization" {
      for_each = var.spec.autoscaler.cpu_target != null || var.spec.autoscaler.cpu_predictive_method != "" ? [var.spec.autoscaler] : []
      content {
        target            = cpu_utilization.value.cpu_target
        predictive_method = cpu_utilization.value.cpu_predictive_method != "" ? cpu_utilization.value.cpu_predictive_method : null
      }
    }

    dynamic "load_balancing_utilization" {
      for_each = var.spec.autoscaler.load_balancing_target != null ? [var.spec.autoscaler.load_balancing_target] : []
      content {
        target = load_balancing_utilization.value
      }
    }

    dynamic "metric" {
      for_each = var.spec.autoscaler.metrics
      content {
        name                       = metric.value.name
        target                     = metric.value.target
        type                       = metric.value.type != "" ? metric.value.type : null
        filter                     = metric.value.filter != "" ? metric.value.filter : null
        single_instance_assignment = metric.value.single_instance_assignment
      }
    }

    dynamic "scale_in_control" {
      for_each = var.spec.autoscaler.scale_in_control != null ? [var.spec.autoscaler.scale_in_control] : []
      content {
        dynamic "max_scaled_in_replicas" {
          for_each = scale_in_control.value.max_scaled_in_replicas_fixed != null || scale_in_control.value.max_scaled_in_replicas_percent != null ? [scale_in_control.value] : []
          content {
            fixed   = max_scaled_in_replicas.value.max_scaled_in_replicas_fixed
            percent = max_scaled_in_replicas.value.max_scaled_in_replicas_percent
          }
        }
        time_window_sec = scale_in_control.value.time_window_sec
      }
    }

    dynamic "scaling_schedules" {
      for_each = var.spec.autoscaler.schedules
      content {
        name                  = scaling_schedules.value.schedule_name
        schedule              = scaling_schedules.value.schedule
        duration_sec          = scaling_schedules.value.duration_sec
        min_required_replicas = scaling_schedules.value.min_required_replicas
        disabled              = scaling_schedules.value.disabled
        time_zone             = scaling_schedules.value.time_zone != "" ? scaling_schedules.value.time_zone : null
        description           = scaling_schedules.value.description != "" ? scaling_schedules.value.description : null
      }
    }

    stabilization_period = var.spec.autoscaler.stabilization_period
  }

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}

# ----------------------------------------------------------------------------
# Per-instance configs — stateful name/disk/IP overrides for individual
# instances. The config name IS the instance name it applies to; the
# manager is referenced by NAME (the provider's expected form).
# ----------------------------------------------------------------------------

resource "google_compute_per_instance_config" "this" {
  for_each = local.is_regional ? {} : local.per_instance_configs

  name                   = each.value.config_name
  project                = local.project_id
  zone                   = var.spec.zone
  instance_group_manager = google_compute_instance_group_manager.this[0].name

  dynamic "preserved_state" {
    for_each = each.value.preserved_state != null ? [each.value.preserved_state] : []
    content {
      metadata = length(preserved_state.value.metadata) > 0 ? preserved_state.value.metadata : null

      dynamic "disk" {
        for_each = preserved_state.value.disks
        content {
          device_name = disk.value.device_name
          source      = disk.value.source
          mode        = disk.value.mode != "" ? disk.value.mode : null
          delete_rule = disk.value.delete_rule != "" ? disk.value.delete_rule : null
        }
      }
      dynamic "external_ip" {
        for_each = preserved_state.value.external_ips
        content {
          interface_name = external_ip.value.interface_name
          auto_delete    = external_ip.value.auto_delete != "" ? external_ip.value.auto_delete : null
          dynamic "ip_address" {
            for_each = external_ip.value.address != "" ? [external_ip.value.address] : []
            content {
              address = ip_address.value
            }
          }
        }
      }
      dynamic "internal_ip" {
        for_each = preserved_state.value.internal_ips
        content {
          interface_name = internal_ip.value.interface_name
          auto_delete    = internal_ip.value.auto_delete != "" ? internal_ip.value.auto_delete : null
          dynamic "ip_address" {
            for_each = internal_ip.value.address != "" ? [internal_ip.value.address] : []
            content {
              address = ip_address.value
            }
          }
        }
      }
    }
  }

  minimal_action                   = each.value.minimal_action != "" ? each.value.minimal_action : null
  most_disruptive_allowed_action   = each.value.most_disruptive_allowed_action != "" ? each.value.most_disruptive_allowed_action : null
  remove_instance_on_destroy       = each.value.remove_instance_on_destroy
  remove_instance_state_on_destroy = each.value.remove_instance_state_on_destroy

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}

resource "google_compute_region_per_instance_config" "this" {
  for_each = local.is_regional ? local.per_instance_configs : {}

  name                          = each.value.config_name
  project                       = local.project_id
  region                        = var.spec.region
  region_instance_group_manager = google_compute_region_instance_group_manager.this[0].name

  dynamic "preserved_state" {
    for_each = each.value.preserved_state != null ? [each.value.preserved_state] : []
    content {
      metadata = length(preserved_state.value.metadata) > 0 ? preserved_state.value.metadata : null

      dynamic "disk" {
        for_each = preserved_state.value.disks
        content {
          device_name = disk.value.device_name
          source      = disk.value.source
          mode        = disk.value.mode != "" ? disk.value.mode : null
          delete_rule = disk.value.delete_rule != "" ? disk.value.delete_rule : null
        }
      }
      dynamic "external_ip" {
        for_each = preserved_state.value.external_ips
        content {
          interface_name = external_ip.value.interface_name
          auto_delete    = external_ip.value.auto_delete != "" ? external_ip.value.auto_delete : null
          dynamic "ip_address" {
            for_each = external_ip.value.address != "" ? [external_ip.value.address] : []
            content {
              address = ip_address.value
            }
          }
        }
      }
      dynamic "internal_ip" {
        for_each = preserved_state.value.internal_ips
        content {
          interface_name = internal_ip.value.interface_name
          auto_delete    = internal_ip.value.auto_delete != "" ? internal_ip.value.auto_delete : null
          dynamic "ip_address" {
            for_each = internal_ip.value.address != "" ? [internal_ip.value.address] : []
            content {
              address = ip_address.value
            }
          }
        }
      }
    }
  }

  minimal_action                   = each.value.minimal_action != "" ? each.value.minimal_action : null
  most_disruptive_allowed_action   = each.value.most_disruptive_allowed_action != "" ? each.value.most_disruptive_allowed_action : null
  remove_instance_on_destroy       = each.value.remove_instance_on_destroy
  remove_instance_state_on_destroy = each.value.remove_instance_state_on_destroy

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}

# ----------------------------------------------------------------------------
# Resize requests — queued one-shot capacity asks (Dynamic Workload
# Scheduler). Immutable: any change queues a new ask; destroying an
# ACTIVE request cancels it.
# ----------------------------------------------------------------------------

resource "google_compute_resize_request" "this" {
  for_each = local.is_regional ? {} : local.resize_requests

  name                   = each.value.request_name
  project                = local.project_id
  zone                   = var.spec.zone
  instance_group_manager = google_compute_instance_group_manager.this[0].name
  resize_by              = each.value.resize_by
  description            = each.value.description != "" ? each.value.description : null

  # The provider models the duration's seconds as a STRING — rendered
  # from the spec's int64 identically on both engines.
  dynamic "requested_run_duration" {
    for_each = each.value.requested_run_duration_seconds != null ? [each.value.requested_run_duration_seconds] : []
    content {
      seconds = tostring(requested_run_duration.value)
    }
  }

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}

resource "google_compute_region_resize_request" "this" {
  for_each = local.is_regional ? local.resize_requests : {}

  name                   = each.value.request_name
  project                = local.project_id
  region                 = var.spec.region
  instance_group_manager = google_compute_region_instance_group_manager.this[0].name
  resize_by              = each.value.resize_by
  description            = each.value.description != "" ? each.value.description : null

  dynamic "requested_run_duration" {
    for_each = each.value.requested_run_duration_seconds != null ? [each.value.requested_run_duration_seconds] : []
    content {
      seconds = tostring(requested_run_duration.value)
    }
  }

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
