locals {
  cluster_name = coalesce(var.spec.cluster_name, var.metadata.name)

  labels = merge(
    {
      "planton-resource"      = "true"
      "planton-resource-kind" = "gcpdataprocvirtualcluster"
      "planton-resource-name" = local.cluster_name
    },
    var.metadata.org != "" ? { "planton-organization" = var.metadata.org } : {},
    var.metadata.env != null ? { "planton-environment" = var.metadata.env.id } : {},
    var.metadata.id != "" ? { "planton-resource-id" = var.metadata.id } : {},
  )
}
