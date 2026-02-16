locals {
  project_id  = var.spec.project_id.value
  policy_name = var.spec.policy_name != "" ? var.spec.policy_name : var.metadata.name

  # Framework GCP labels.
  gcp_labels = merge(
    {
      "openmcf-resource"      = "true"
      "openmcf-resource-name" = lower(var.metadata.name)
      "openmcf-resource-kind" = "gcpcloudarmorpolicy"
    },
    var.metadata.org != "" ? { "openmcf-organization" = var.metadata.org } : {},
    var.metadata.env != "" ? { "openmcf-environment" = var.metadata.env } : {},
    var.metadata.id != "" ? { "openmcf-resource-id" = var.metadata.id } : {},
  )
}
