locals {
  labels = merge(
    {
      "planton-resource"      = "true"
      "planton-resource-name" = var.spec.cluster_name
      "planton-resource-kind" = "gcpdataproccluster"
    },
    var.metadata.org != "" ? { "planton-organization" = var.metadata.org } : {},
    var.metadata.env != "" ? { "planton-environment" = var.metadata.env } : {},
    var.metadata.id != "" ? { "planton-resource-id" = var.metadata.id } : {},
  )
}
