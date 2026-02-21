locals {
  resource_id = coalesce(var.metadata.id, var.metadata.name)
}
