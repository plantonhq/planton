resource "aws_glue_catalog_database" "this" {
  name         = local.database_name
  description  = var.spec.description != "" ? var.spec.description : null
  location_uri = var.spec.location_uri != "" ? var.spec.location_uri : null
  tags         = local.tags
}
