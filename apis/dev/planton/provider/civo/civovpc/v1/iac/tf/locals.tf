locals {
  # Network name from spec
  network_name = var.spec.network_name

  # Region
  region = var.spec.region

  # CIDR block (empty string means auto-allocate)
  cidr_block = var.spec.ip_range_cidr

  # Whether this is the default network for the region
  is_default = var.spec.is_default_for_region

  # Description (informational - not supported by Civo provider)
  description = var.spec.description

  # Planton labels (for internal tracking)
  # Note: These are stored in metadata but not applied to Civo resource
  # as the Civo Network provider doesn't support labels/tags
  planton_labels = {
    "planton.org/resource"      = "true"
    "planton.org/resource-kind" = "CivoVpc"
    "planton.org/resource-id"   = var.metadata.id
    "planton.org/resource-name" = var.metadata.name
    "planton.org/organization"  = var.metadata.org
    "planton.org/environment"   = var.metadata.env
  }
}

