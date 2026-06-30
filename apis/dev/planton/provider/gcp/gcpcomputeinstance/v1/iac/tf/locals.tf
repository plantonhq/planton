locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Extract project_id from StringValueOrRef
  # Note: value_from resolution is not yet implemented - only literal values are supported
  project_id = (
    var.spec.project_id != null
    ? coalesce(var.spec.project_id.value, "")
    : ""
  )

  # Base GCP labels
  base_gcp_labels = {
    "resource"      = "true"
    "resource_kind" = "gcp-compute-instance"
    "resource_name" = var.metadata.name
  }

  # Organization label only if var.metadata.org is non-empty
  org_label = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? {
    "organization" = var.metadata.org
  } : {}

  # Environment label only if var.metadata.env is non-empty
  env_label = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? {
    "environment" = var.metadata.env
  } : {}

  # Merge base, org, environment labels, user labels, and add resource_id
  final_gcp_labels = merge(
    local.base_gcp_labels,
    { "resource_id" = local.resource_id },
    local.org_label,
    local.env_label,
    var.spec.labels
  )

  # Determine if instance is preemptible/spot
  is_preemptible = (
    var.spec.preemptible ||
    var.spec.spot ||
    (var.spec.scheduling != null && var.spec.scheduling.preemptible)
  )

  # Determine provisioning model
  provisioning_model = (
    var.spec.spot ||
    (var.spec.scheduling != null && var.spec.scheduling.provisioning_model == "SPOT")
    ? "SPOT"
    : "STANDARD"
  )

  # Determine on_host_maintenance behavior
  on_host_maintenance = (
    local.is_preemptible
    ? "TERMINATE"
    : (var.spec.scheduling != null ? coalesce(var.spec.scheduling.on_host_maintenance, "MIGRATE") : "MIGRATE")
  )

  # Determine automatic_restart
  automatic_restart = (
    local.is_preemptible
    ? false
    : (var.spec.scheduling != null ? coalesce(var.spec.scheduling.automatic_restart, true) : true)
  )

  # Build SSH keys metadata if specified
  ssh_keys_metadata = (
    length(var.spec.ssh_keys) > 0
    ? { "ssh-keys" = join("\n", var.spec.ssh_keys) }
    : {}
  )

  # Merge user metadata with SSH keys
  final_metadata = merge(var.spec.metadata, local.ssh_keys_metadata)
}
