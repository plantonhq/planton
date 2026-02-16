locals {
  database_name = var.metadata.name

  tags = {
    "openmcf.org/resource"      = "true"
    "openmcf.org/organization"  = var.metadata.org
    "openmcf.org/environment"   = var.metadata.env
    "openmcf.org/resource-kind" = "AwsGlueCatalogDatabase"
    "openmcf.org/resource-id"   = var.metadata.id
  }
}
