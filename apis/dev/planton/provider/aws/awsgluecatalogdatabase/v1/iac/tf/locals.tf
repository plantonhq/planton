locals {
  database_name = var.metadata.name

  tags = {
    "planton.dev/resource"      = "true"
    "planton.dev/organization"  = var.metadata.org
    "planton.dev/environment"   = var.metadata.env
    "planton.dev/resource-kind" = "AwsGlueCatalogDatabase"
    "planton.dev/resource-id"   = var.metadata.id
  }
}
