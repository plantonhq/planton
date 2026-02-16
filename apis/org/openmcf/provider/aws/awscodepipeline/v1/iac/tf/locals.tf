locals {
  tags = {
    "planton.org/resource"      = "true"
    "planton.org/organization"  = var.metadata.org
    "planton.org/environment"   = var.metadata.env
    "planton.org/resource-kind" = "AwsCodePipeline"
    "planton.org/resource-id"   = var.metadata.id
  }

  has_triggers  = var.spec.triggers != null && length(var.spec.triggers) > 0
  has_variables = var.spec.variables != null && length(var.spec.variables) > 0

  is_single_region = length(var.spec.artifact_stores) == 1 && var.spec.artifact_stores[0].region == ""
}
