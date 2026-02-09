locals {
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  template_name    = var.metadata.name
  image            = var.spec.image.value
  keypair          = try(var.spec.keypair.value, null)
  external_network = try(var.spec.external_network.value, null)
  fixed_network    = try(var.spec.fixed_network.value, null)
  fixed_subnet     = try(var.spec.fixed_subnet.value, null)
}
