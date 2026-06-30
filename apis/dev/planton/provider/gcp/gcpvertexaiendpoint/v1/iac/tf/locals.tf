locals {
  project_id   = var.spec.project_id.value
  location     = var.spec.location
  display_name = var.spec.display_name

  # Endpoint name: use explicit value or auto-generated numeric ID.
  endpoint_name = var.spec.endpoint_name != "" ? var.spec.endpoint_name : random_integer.endpoint_name[0].result

  # Framework GCP labels.
  gcp_labels = merge(
    {
      "planton-resource"      = "true"
      "planton-resource-name" = lower(var.metadata.name)
      "planton-resource-kind" = "gcpvertexaiendpoint"
    },
    var.metadata.org != "" ? { "planton-organization" = var.metadata.org } : {},
    var.metadata.env != "" ? { "planton-environment" = var.metadata.env } : {},
    var.metadata.id != "" ? { "planton-resource-id" = var.metadata.id } : {},
  )
}
