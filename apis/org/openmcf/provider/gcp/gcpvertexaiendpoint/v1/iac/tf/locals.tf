locals {
  project_id   = var.spec.project_id.value
  location     = var.spec.location
  display_name = var.spec.display_name

  # Endpoint name: use explicit value or auto-generated numeric ID.
  endpoint_name = var.spec.endpoint_name != "" ? var.spec.endpoint_name : random_integer.endpoint_name[0].result

  # Framework GCP labels.
  gcp_labels = merge(
    {
      "openmcf-resource"      = "true"
      "openmcf-resource-name" = lower(var.metadata.name)
      "openmcf-resource-kind" = "gcpvertexaiendpoint"
    },
    var.metadata.org != "" ? { "openmcf-organization" = var.metadata.org } : {},
    var.metadata.env != "" ? { "openmcf-environment" = var.metadata.env } : {},
    var.metadata.id != "" ? { "openmcf-resource-id" = var.metadata.id } : {},
  )
}
