locals {
  environment_name = var.spec.environment_name != "" ? var.spec.environment_name : var.metadata.name

  labels = merge(
    {
      "resource"      = "true"
      "resource-name" = local.environment_name
      "resource-kind" = "gcpcloudcomposerenvironment"
    },
    var.metadata.org != "" ? { "organization" = var.metadata.org } : {},
    var.metadata.env != "" ? { "environment" = var.metadata.env } : {},
    var.metadata.id != "" ? { "resource-id" = var.metadata.id } : {},
  )
}
