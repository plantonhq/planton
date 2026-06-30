locals {
  project_id    = var.spec.project_id.value
  instance_name = var.spec.instance_name != "" ? var.spec.instance_name : var.metadata.name
  location      = var.spec.location

  # Framework GCP labels.
  gcp_labels = merge(
    {
      "planton-resource"      = "true"
      "planton-resource-name" = local.instance_name
      "planton-resource-kind" = "gcpvertexainotebook"
    },
    var.metadata.org != "" ? { "planton-organization" = var.metadata.org } : {},
    var.metadata.env != "" ? { "planton-environment" = var.metadata.env } : {},
    var.metadata.id != "" ? { "planton-resource-id" = var.metadata.id } : {},
  )
}
