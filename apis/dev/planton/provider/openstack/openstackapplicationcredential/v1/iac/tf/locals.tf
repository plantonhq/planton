# locals.tf

locals {
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )
  resource_name = var.metadata.name
}
