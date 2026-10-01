# DigitalOcean Managed Database Cluster
#
# Provisions a fully-managed database cluster on DigitalOcean, modeling the
# complete digitalocean_database_cluster resource surface: engine/version/
# size/region/node topology, private networking, custom storage with
# autoscale, maintenance windows, backup-restore provisioning,
# engine-conditional tuning (sql_mode, eviction_policy), project placement,
# and tags.
#
# Related DigitalOcean resources (database users, logical databases,
# connection pools, replicas, firewall rules, per-engine config settings)
# are separate resources with independent lifecycles and are not part of
# this module.

resource "digitalocean_database_cluster" "cluster" {
  name       = var.spec.cluster_name
  engine     = local.engine_slug
  version    = var.spec.engine_version
  size       = var.spec.size_slug
  region     = local.region_slug
  node_count = var.spec.node_count

  # Optional VPC integration for private networking (create-only).
  private_network_uuid = local.vpc_uuid

  # Optional DigitalOcean project placement (create-only).
  project_id = local.project_id

  # storage_size_mib is a string in the provider schema; the spec carries
  # GiB for ergonomics, converted here. Storage can only be increased.
  storage_size_mib = local.storage_size_mib

  # Engine-conditional tuning: the provider rejects sql_mode on non-MySQL
  # and eviction_policy on non-Redis/Valkey clusters at plan time, so both
  # are passed only when set (spec CEL rules enforce the engine pairing).
  eviction_policy = local.eviction_policy
  sql_mode        = local.sql_mode

  # Weekly maintenance window. The provider models this as a list; a
  # cluster has exactly one window, so the spec carries a single message.
  dynamic "maintenance_window" {
    for_each = var.spec.maintenance_window != null ? [var.spec.maintenance_window] : []
    content {
      day  = maintenance_window.value.day
      hour = maintenance_window.value.hour
    }
  }

  # Provision-from-backup. Consumed only at creation; DigitalOcean never
  # reports it back, so it produces no drift after create.
  dynamic "backup_restore" {
    for_each = var.spec.backup_restore != null ? [var.spec.backup_restore] : []
    content {
      database_name     = backup_restore.value.database_name
      backup_created_at = backup_restore.value.backup_created_at
    }
  }

  # Automatic storage growth. threshold_percent and increment_gib are
  # optional; DigitalOcean applies its own defaults when they are unset.
  dynamic "storage_autoscale" {
    for_each = var.spec.storage_autoscale != null ? [var.spec.storage_autoscale] : []
    content {
      enabled           = storage_autoscale.value.enabled
      threshold_percent = storage_autoscale.value.threshold_percent != null && storage_autoscale.value.threshold_percent > 0 ? storage_autoscale.value.threshold_percent : null
      increment_gib     = storage_autoscale.value.increment_gib != null && storage_autoscale.value.increment_gib > 0 ? storage_autoscale.value.increment_gib : null
    }
  }

  # User tags plus the standard Planton labels (identical set in both
  # provisioners).
  tags = local.tags

  # Fail loud on DigitalOcean's combined-tags budget before anything
  # renders (see local.tags_combined_budget; twin of the Pulumi module's
  # error).
  lifecycle {
    precondition {
      condition     = length(local.tags_combined) <= local.tags_combined_budget
      error_message = "DigitalOcean caps a database cluster's combined tags (joined by commas) at ${local.tags_combined_budget} characters; this cluster's ${length(local.tags)} tags join to ${length(local.tags_combined)} characters. Shorten metadata.name or metadata.id, or remove entries from spec.tags -- the Planton label tags alone use ${local.planton_tags_length} characters here."
    }
  }
}
