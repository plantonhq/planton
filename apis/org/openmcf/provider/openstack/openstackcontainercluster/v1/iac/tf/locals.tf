locals {
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  cluster_name     = var.metadata.name
  cluster_template = var.spec.cluster_template.value
  keypair          = try(var.spec.keypair.value, null)
}
