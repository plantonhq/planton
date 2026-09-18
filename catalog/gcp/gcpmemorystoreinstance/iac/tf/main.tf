# Enable the Memorystore API — the control plane that owns the instance.
# disable_on_destroy is false: tearing down one instance must never
# disable the API for everything else in the project.
resource "google_project_service" "memorystore_api" {
  project = local.project_id
  service = "memorystore.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# Enable the Network Connectivity API — the service connectivity
# automation that places this instance's PSC endpoints is driven through
# it (the GcpServiceConnectionPolicy prerequisite lives there too).
resource "google_project_service" "networkconnectivity_api" {
  project = local.project_id
  service = "networkconnectivity.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# Resolves the provider's effective project for PSC endpoint entries that
# omit their consumer project — the API requires one per entry, and the
# common case is "same project as the instance". Only instantiated when
# actually needed (see locals.needs_ambient_endpoint_project), mirroring
# the Pulumi module's only-when-needed GetClientConfig lookup.
data "google_project" "this" {
  count = local.needs_ambient_endpoint_project ? 1 : 0
}

# The Memorystore (Valkey) instance. Connectivity is PSC-only and driven
# by service connectivity automation: a service connection policy for the
# gcp-memorystore class must already exist on each endpoint's network in
# this region, or creation fails with a connectivity error — the policy
# is a separate first-class resource, deployed before this one.
#
# The immutables (ForceNew): instance_id, location, mode,
# authorization_mode, transit_encryption_mode, kms_key,
# zone_distribution_config, the PSC endpoints, and the seed sources.
# shard_count and replica_count resize in place; engine_configs,
# persistence, maintenance, backups, labels, and the DR role update in
# place too.
resource "google_memorystore_instance" "this" {
  instance_id = local.instance_name
  project     = local.project_id
  location    = local.location
  shard_count = local.shard_count

  mode           = local.mode
  node_type      = local.node_type
  engine_version = local.engine_version
  engine_configs = length(var.spec.engine_configs) > 0 ? var.spec.engine_configs : null

  # 0 is an explicit "no replicas" — sent through rather than nulled so
  # the manifest value is authoritative (identical to the Pulumi module).
  replica_count = local.replica_count

  authorization_mode      = local.authorization_mode
  transit_encryption_mode = local.transit_encryption_mode
  kms_key                 = local.kms_key

  # Server certificate authority for the TLS-enabled instance: which CA
  # signs the certificate clients verify. A customer-managed CA pool is
  # consumed only in CUSTOMER_MANAGED_CAS_CA mode (spec-enforced pairing).
  # Both are fixed at creation (ForceNew).
  server_ca_mode = local.server_ca_mode
  server_ca_pool = local.server_ca_pool

  # Self-service maintenance: setting a newer available version applies
  # the update now instead of waiting for GCP's rollout. Update-only —
  # the API rejects it at create and rejects downgrades.
  maintenance_version = local.maintenance_version

  # Shared Valkey ACL policy (users, key patterns, allowed commands)
  # attached by full resource name. Sent only when set so an instance
  # without one keeps its built-in default ACL; swapping is in place.
  acl_policy = local.acl_policy

  # Always sent explicitly (spec defaults TRUE) so destroy behavior is
  # identical on both engines: omitting it would let the provider default
  # decide, and a manifest that never mentions deletion protection must
  # behave the same everywhere.
  deletion_protection_enabled = var.spec.deletion_protection_enabled

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
  # Sent only when set so the provider default stays in charge otherwise.
  # Evaluated only after deletion_protection_enabled allows the destroy.
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  labels = local.final_labels

  # PSC auto-created endpoints for VPC connectivity. network arrives as
  # the VPC's relative resource path — the only format the Service
  # Connectivity API accepts (full https:// self-links are rejected).
  # An entry that omits its consumer project rides the provider's
  # effective project (the common same-project case).
  dynamic "desired_auto_created_endpoints" {
    for_each = var.spec.psc_auto_connections
    content {
      network = desired_auto_created_endpoints.value.network
      project_id = (
        desired_auto_created_endpoints.value.project_id != ""
        ? desired_auto_created_endpoints.value.project_id
        : (var.spec.project_id != "" ? var.spec.project_id : data.google_project.this[0].project_id)
      )
    }
  }

  # Persistence configuration (RDB or AOF).
  dynamic "persistence_config" {
    for_each = var.spec.persistence_config != null ? [var.spec.persistence_config] : []
    content {
      mode = persistence_config.value.mode

      dynamic "rdb_config" {
        for_each = persistence_config.value.rdb_config != null ? [persistence_config.value.rdb_config] : []
        content {
          rdb_snapshot_period     = rdb_config.value.rdb_snapshot_period
          rdb_snapshot_start_time = rdb_config.value.rdb_snapshot_start_time != "" ? rdb_config.value.rdb_snapshot_start_time : null
        }
      }

      dynamic "aof_config" {
        for_each = persistence_config.value.aof_config != null ? [persistence_config.value.aof_config] : []
        content {
          append_fsync = aof_config.value.append_fsync
        }
      }
    }
  }

  # Zone distribution configuration.
  dynamic "zone_distribution_config" {
    for_each = var.spec.zone_distribution_config != null ? [var.spec.zone_distribution_config] : []
    content {
      mode = zone_distribution_config.value.mode
      zone = zone_distribution_config.value.zone != "" ? zone_distribution_config.value.zone : null
    }
  }

  # Maintenance policy with weekly maintenance window.
  dynamic "maintenance_policy" {
    for_each = var.spec.maintenance_policy != null ? [var.spec.maintenance_policy] : []
    content {
      weekly_maintenance_window {
        day = maintenance_policy.value.weekly_maintenance_window.day
        # Hours only: the Memorystore API rejects a start_time carrying
        # minutes (400 "Invalid start time, only hours are supported" —
        # live-verified; narrower than Redis).
        start_time {
          hours = maintenance_policy.value.weekly_maintenance_window.hour
        }
      }
    }
  }

  # Automated backup configuration.
  dynamic "automated_backup_config" {
    for_each = var.spec.automated_backup_config != null ? [var.spec.automated_backup_config] : []
    content {
      retention = automated_backup_config.value.retention

      fixed_frequency_schedule {
        start_time {
          hours = automated_backup_config.value.start_hour
        }
      }
    }
  }

  # Cross-region DR: PRIMARY lists its secondaries; SECONDARY points at
  # its primary. Roles are exchanged in place during a planned
  # switchover. Instance references arrive as full resource paths (the
  # other instance's name output).
  dynamic "cross_instance_replication_config" {
    for_each = var.spec.cross_instance_replication_config != null ? [var.spec.cross_instance_replication_config] : []
    content {
      instance_role = cross_instance_replication_config.value.instance_role

      dynamic "primary_instance" {
        for_each = cross_instance_replication_config.value.primary_instance != null ? [cross_instance_replication_config.value.primary_instance] : []
        content {
          instance = primary_instance.value.instance
        }
      }

      dynamic "secondary_instances" {
        for_each = cross_instance_replication_config.value.secondary_instances
        content {
          instance = secondary_instances.value.instance
        }
      }
    }
  }

  # Seed-from-GCS at creation (mutually exclusive with
  # managed_backup_source; both ForceNew — seeding only happens once).
  dynamic "gcs_source" {
    for_each = var.spec.gcs_source != null ? [var.spec.gcs_source] : []
    content {
      uris = gcs_source.value.uris
    }
  }

  # Seed-from-managed-backup at creation.
  dynamic "managed_backup_source" {
    for_each = var.spec.managed_backup_source != null ? [var.spec.managed_backup_source] : []
    content {
      backup = managed_backup_source.value.backup
    }
  }

  depends_on = [
    google_project_service.memorystore_api,
    google_project_service.networkconnectivity_api,
  ]
}
