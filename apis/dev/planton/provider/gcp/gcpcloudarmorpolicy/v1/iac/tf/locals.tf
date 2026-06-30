locals {
  project_id  = var.spec.project_id.value
  policy_name = var.spec.policy_name != "" ? var.spec.policy_name : var.metadata.name

  # Framework GCP labels.
  gcp_labels = merge(
    {
      "planton-resource"      = "true"
      "planton-resource-name" = lower(var.metadata.name)
      "planton-resource-kind" = "gcpcloudarmorpolicy"
    },
    var.metadata.org != "" ? { "planton-organization" = var.metadata.org } : {},
    var.metadata.env != "" ? { "planton-environment" = var.metadata.env } : {},
    var.metadata.id != "" ? { "planton-resource-id" = var.metadata.id } : {},
  )
}
