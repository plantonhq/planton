locals {
  labels = merge(
    {
      "resource"      = "true"
      "resource-name" = var.spec.instance_name
      "resource-kind" = "gcpbigtableinstance"
    },
    var.metadata.org != "" ? { "organization" = var.metadata.org } : {},
    var.metadata.env != "" ? { "environment" = var.metadata.env } : {},
    var.metadata.id != "" ? { "resource-id" = var.metadata.id } : {},
  )
}
