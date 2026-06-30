# locals.tf - Local value transformations for GCP Project

locals {
  # Use project_id from spec, optionally with random suffix if add_suffix is true
  # If add_suffix is true, append a random 3-character suffix for uniqueness
  # If add_suffix is false (default), use project_id as-is
  project_id = var.spec.add_suffix ? "${var.spec.project_id}-${random_string.project_suffix[0].result}" : var.spec.project_id

  # GCP labels: merge user-provided labels with standard Planton labels
  # User labels come first, then standard labels (which override if there's a conflict)
  gcp_labels = merge(
    var.spec.labels != null ? var.spec.labels : {},
    {
      "planton-resource"      = "true"
      "planton-resource-name" = var.metadata.name
      "planton-resource-kind" = "gcpproject"
      "planton-resource-id"   = var.metadata.id != null ? var.metadata.id : var.metadata.name
      "planton-org"           = var.metadata.org != null ? var.metadata.org : "default"
      "planton-env"           = var.metadata.env != null ? var.metadata.env : "default"
    }
  )

  # Determine parent resource configuration
  parent_type      = var.spec.parent_type
  parent_org_id    = local.parent_type == "organization" ? var.spec.parent_id : null
  parent_folder_id = local.parent_type == "folder" ? var.spec.parent_id : null

  # Billing account configuration
  billing_account_id = var.spec.billing_account_id

  # Network configuration
  # disable_default_network defaults to true if not specified
  auto_create_network = var.spec.disable_default_network != null ? !var.spec.disable_default_network : false

  # APIs to enable
  enabled_apis = var.spec.enabled_apis != null ? var.spec.enabled_apis : []

  # IAM owner member (optional)
  owner_member = var.spec.owner_member != null && var.spec.owner_member != "" ? var.spec.owner_member : null

  # Deletion policy configuration
  # When delete_protection is true, set to "PREVENT" to block project deletion
  # When delete_protection is false or not set, use "DELETE" for normal deletion behavior
  deletion_policy = var.spec.delete_protection == true ? "PREVENT" : "DELETE"
}

