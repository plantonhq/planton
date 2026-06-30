# Civo Database Terraform Module
#
# This module creates a Civo managed database instance with optional
# high availability, network isolation, and firewall configuration.
#
# The module provisions:
# - Civo managed database (MySQL or PostgreSQL)
# - Network attachment for private access
# - Optional firewall rules
# - Optional custom storage configuration
# - Tags for resource organization

terraform {
  required_version = ">= 1.0"

  required_providers {
    civo = {
      source  = "civo/civo"
      version = "~> 1.1"
    }
  }
}

# Civo Database Resource
resource "civo_database" "this" {
  # Basic configuration
  name    = var.spec.db_instance_name
  engine  = var.spec.engine
  version = var.spec.engine_version
  region  = var.spec.region
  size    = var.spec.size_slug

  # High availability configuration
  # nodes = primary + replicas (e.g., replicas=2 means 3 total nodes)
  nodes = local.total_nodes

  # Network configuration (required for security)
  network_id = local.network_id

  # Firewall configuration (optional, first firewall only)
  firewall_id = local.firewall_id != "" ? local.firewall_id : null

  # Custom storage size (optional)
  size_gb = var.spec.storage_gib != null ? var.spec.storage_gib : null

  # Tags for resource organization
  tags = local.all_tags
}

