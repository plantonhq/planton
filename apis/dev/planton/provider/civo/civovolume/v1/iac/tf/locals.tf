locals {
  # Volume name from spec
  volume_name = var.spec.volume_name

  # Size in GiB
  size_gib = var.spec.size_gib

  # Region
  region = var.spec.region

  # Filesystem type (informational - not used by Civo provider)
  filesystem_type = var.spec.filesystem_type

  # Snapshot ID (informational - not supported on public Civo cloud)
  snapshot_id = var.spec.snapshot_id

  # Tags (informational - not supported by Civo Volume provider)
  tags = var.spec.tags

  # Planton labels (for internal tracking)
  # Note: These are stored in metadata but not applied to Civo resource
  # as the Civo Volume provider doesn't support tags/labels
  planton_labels = {
    "planton.org/resource"      = "true"
    "planton.org/resource-kind" = "CivoVolume"
    "planton.org/resource-id"   = var.metadata.id
    "planton.org/resource-name" = var.metadata.name
    "planton.org/organization"  = var.metadata.org
    "planton.org/environment"   = var.metadata.env
  }
}

