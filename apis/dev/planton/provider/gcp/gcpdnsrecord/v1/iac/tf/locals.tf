locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Extract project ID from StringValueOrRef
  project_id = var.spec.project_id.value

  # Extract managed zone from StringValueOrRef
  managed_zone = var.spec.managed_zone.value
  record_type  = var.spec.type
  name         = var.spec.name
  values       = var.spec.values
  ttl_seconds  = var.spec.ttl_seconds
}
