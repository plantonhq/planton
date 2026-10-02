# Enable the Memorystore for Redis API -- the control plane that owns the
# cluster. disable_on_destroy is false: tearing down one cluster must
# never disable the API for everything else in the project.
resource "google_project_service" "redis_api" {
  project = local.project_id
  service = "redis.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# Enable the Network Connectivity API -- the service connectivity
# automation that places this cluster's PSC endpoints is driven through
# it (the GcpServiceConnectionPolicy prerequisite lives there too).
resource "google_project_service" "networkconnectivity_api" {
  project = local.project_id
  service = "networkconnectivity.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The Memorystore for Redis Cluster. Connectivity is Private Service
# Connect: with psc_configs set, service connectivity automation places
# the endpoints (a GcpServiceConnectionPolicy for the gcp-memorystore-redis
# class must already exist on that network in this region); without it,
# the cluster only publishes service attachments and consumers register
# their own forwarding rules through GcpRedisClusterEndpointSet.
#
# The immutables (ForceNew): name, region, authorization_mode,
# transit_encryption_mode, zone_distribution_config, and the seed sources.
# shard_count, replica_count, node_type, redis_configs, kms_key,
# psc_configs, persistence, backups, maintenance, the replication role,
# labels, and deletion protection all update in place.
resource "google_redis_cluster" "this" {
  name        = local.cluster_name
  project     = local.project_id
  region      = local.region
  shard_count = local.shard_count

  # 0 is an explicit "no replicas" -- sent through rather than nulled so
  # the manifest value is authoritative (identical to the Pulumi module).
  replica_count = local.replica_count

  node_type     = local.node_type
  redis_configs = length(var.spec.redis_configs) > 0 ? var.spec.redis_configs : null

  # Both immutable security modes are always sent (Google's defaults when
  # the spec leaves them unset) so the posture never depends on a provider
  # default.
  authorization_mode      = local.authorization_mode
  transit_encryption_mode = local.transit_encryption_mode

  # Server certificate authority for the TLS-enabled cluster; a
  # customer-managed CA pool is consumed only in
  # SERVER_CA_MODE_CUSTOMER_MANAGED_CAS_CA mode (spec-enforced pairing).
  server_ca_mode = local.server_ca_mode
  server_ca_pool = local.server_ca_pool

  kms_key = local.kms_key

  # Self-service maintenance: a newer available version applies the update
  # now instead of waiting for Google's rollout. Update-only and
  # forward-only at the API.
  maintenance_version = local.maintenance_version

  # Shared Redis ACL policy attached by full resource name. Sent only when
  # set so a cluster without one keeps its built-in default ACL.
  acl_policy = local.acl_policy

  # Always sent explicitly (spec defaults TRUE) so destroy behavior is
  # identical on both engines: a manifest that never mentions deletion
  # protection must behave the same everywhere.
  deletion_protection_enabled = var.spec.deletion_protection_enabled

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
  # Sent only when set so the provider default stays in charge otherwise;
  # evaluated only after deletion_protection_enabled allows the destroy.
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  labels = local.final_labels

  # Google-placed PSC endpoints. network arrives as the VPC's relative
  # resource path (the GcpVpcNetwork network_id output) -- the only form
  # the Service Connectivity API accepts.
  dynamic "psc_configs" {
    for_each = var.spec.psc_configs
    content {
      network = psc_configs.value.network
    }
  }

  # Persistence (RDB snapshots or AOF log); every leaf is Optional+Computed
  # and sent only when set.
  dynamic "persistence_config" {
    for_each = var.spec.persistence_config != null ? [var.spec.persistence_config] : []
    content {
      mode = persistence_config.value.mode != "" ? persistence_config.value.mode : null

      dynamic "rdb_config" {
        for_each = persistence_config.value.rdb_config != null ? [persistence_config.value.rdb_config] : []
        content {
          rdb_snapshot_period     = rdb_config.value.rdb_snapshot_period != "" ? rdb_config.value.rdb_snapshot_period : null
          rdb_snapshot_start_time = rdb_config.value.rdb_snapshot_start_time != "" ? rdb_config.value.rdb_snapshot_start_time : null
        }
      }

      dynamic "aof_config" {
        for_each = persistence_config.value.aof_config != null ? [persistence_config.value.aof_config] : []
        content {
          append_fsync = aof_config.value.append_fsync != "" ? aof_config.value.append_fsync : null
        }
      }
    }
  }

  # Zone distribution (immutable).
  dynamic "zone_distribution_config" {
    for_each = var.spec.zone_distribution_config != null ? [var.spec.zone_distribution_config] : []
    content {
      mode = zone_distribution_config.value.mode != "" ? zone_distribution_config.value.mode : null
      zone = zone_distribution_config.value.zone != "" ? zone_distribution_config.value.zone : null
    }
  }

  # Weekly maintenance window. Hours only: Google exposes the Redis Cluster
  # window start by the hour (gcloud has no minute flag), so the
  # TimeOfDay's finer fields are never sent.
  dynamic "maintenance_policy" {
    for_each = var.spec.maintenance_policy != null ? [var.spec.maintenance_policy] : []
    content {
      weekly_maintenance_window {
        day = maintenance_policy.value.weekly_maintenance_window.day
        start_time {
          hours = maintenance_policy.value.weekly_maintenance_window.hour
        }
      }
    }
  }

  # Daily automated backups into the cluster's managed backup collection.
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

  # Cross-region DR: PRIMARY lists its secondaries; SECONDARY names its
  # primary. Cluster references arrive as full resource paths (the other
  # cluster's name output).
  dynamic "cross_cluster_replication_config" {
    for_each = var.spec.cross_cluster_replication_config != null ? [var.spec.cross_cluster_replication_config] : []
    content {
      cluster_role = cross_cluster_replication_config.value.cluster_role != "" ? cross_cluster_replication_config.value.cluster_role : null

      dynamic "primary_cluster" {
        for_each = cross_cluster_replication_config.value.primary_cluster != null ? [cross_cluster_replication_config.value.primary_cluster] : []
        content {
          cluster = primary_cluster.value.cluster
        }
      }

      dynamic "secondary_clusters" {
        for_each = cross_cluster_replication_config.value.secondary_clusters
        content {
          cluster = secondary_clusters.value.cluster
        }
      }
    }
  }

  # Seed sources (mutually exclusive, ForceNew -- seeding happens once).
  dynamic "gcs_source" {
    for_each = var.spec.gcs_source != null ? [var.spec.gcs_source] : []
    content {
      uris = gcs_source.value.uris
    }
  }

  dynamic "managed_backup_source" {
    for_each = var.spec.managed_backup_source != null ? [var.spec.managed_backup_source] : []
    content {
      backup = managed_backup_source.value.backup
    }
  }

  depends_on = [
    google_project_service.redis_api,
    google_project_service.networkconnectivity_api,
  ]
}
