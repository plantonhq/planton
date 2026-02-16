locals {
  project_id    = var.spec.project_id.value
  instance_name = var.spec.instance_name != "" ? var.spec.instance_name : var.metadata.name
  location      = var.spec.location

  # Framework GCP labels.
  gcp_labels = merge(
    {
      "openmcf-resource"      = "true"
      "openmcf-resource-name" = local.instance_name
      "openmcf-resource-kind" = "gcpvertexainotebook"
    },
    var.metadata.org != "" ? { "openmcf-organization" = var.metadata.org } : {},
    var.metadata.env != "" ? { "openmcf-environment" = var.metadata.env } : {},
    var.metadata.id != "" ? { "openmcf-resource-id" = var.metadata.id } : {},
  )
}
