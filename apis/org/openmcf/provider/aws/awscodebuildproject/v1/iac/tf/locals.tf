locals {
  tags = {
    "planton.org/resource"      = "true"
    "planton.org/organization"  = var.metadata.org
    "planton.org/environment"   = var.metadata.env
    "planton.org/resource-kind" = "AwsCodeBuildProject"
    "planton.org/resource-id"   = var.metadata.id
  }

  has_webhook     = var.spec.webhook != null
  has_vpc_config  = var.spec.vpc_config != null
  has_cache       = var.spec.cache != null && var.spec.cache.type != "NO_CACHE"
  has_logs_config = var.spec.logs_config != null
}
