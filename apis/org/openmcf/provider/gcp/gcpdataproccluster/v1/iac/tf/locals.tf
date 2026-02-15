locals {
  labels = merge(
    {
      "openmcf-resource"      = "true"
      "openmcf-resource-name" = var.spec.cluster_name
      "openmcf-resource-kind" = "gcpdataproccluster"
    },
    var.metadata.org != "" ? { "openmcf-organization" = var.metadata.org } : {},
    var.metadata.env != "" ? { "openmcf-environment" = var.metadata.env } : {},
    var.metadata.id != "" ? { "openmcf-resource-id" = var.metadata.id } : {},
  )
}
