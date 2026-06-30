# DigitalOcean Managed Database Cluster Resource
#
# This module provisions a fully-managed database cluster on DigitalOcean.
#
# Supported engines:
# - PostgreSQL (pg): ACID-compliant relational database
# - MySQL: Popular open-source RDBMS
# - Redis: In-memory data store for caching
# - MongoDB: Document-oriented NoSQL database
#
# Key features:
# - Automated daily backups (7-day retention)
# - Automatic security patching
# - Multi-node HA with automatic failover
# - VPC-private networking
# - PgBouncer connection pooling (PostgreSQL)

resource "digitalocean_database_cluster" "cluster" {
  name       = var.spec.cluster_name
  engine     = local.engine_slug
  version    = var.spec.engine_version
  size       = var.spec.size_slug
  region     = local.region_slug
  node_count = var.spec.node_count

  # Optional: VPC integration for private networking
  private_network_uuid = local.vpc_uuid

  # Optional: Custom storage size (if not specified, uses default for size_slug)
  # Note: Storage can only be increased, never decreased
  dynamic "storage" {
    for_each = local.has_custom_storage ? [1] : []
    content {
      size_gib = var.spec.storage_gib
    }
  }

  # Lifecycle management for safe cluster updates
  # Prevents accidental deletion during apply operations
  lifecycle {
    prevent_destroy = false  # Set to true for production clusters
  }
}

# Note: The following resources are NOT included in this module to keep it focused on the 80/20 principle:
# - digitalocean_database_firewall (separate resource for network access rules)
# - digitalocean_database_user (separate resource for additional users)
# - digitalocean_database_db (separate resource for additional databases)
# - digitalocean_database_connection_pool (separate resource for PgBouncer pools)
# - digitalocean_database_replica (separate resource for read replicas)
#
# These should be managed separately to allow independent lifecycle management.

