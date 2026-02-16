locals {
  cluster_name = coalesce(var.spec.cluster_name, var.metadata.name)

  labels = merge(
    {
      "openmcf-resource"      = "true"
      "openmcf-resource-kind" = "gcpdataprocvirtualcluster"
      "openmcf-resource-name" = local.cluster_name
    },
    var.metadata.org != "" ? { "openmcf-organization" = var.metadata.org } : {},
    var.metadata.env != null ? { "openmcf-environment" = var.metadata.env.id } : {},
    var.metadata.id != "" ? { "openmcf-resource-id" = var.metadata.id } : {},
  )
}
