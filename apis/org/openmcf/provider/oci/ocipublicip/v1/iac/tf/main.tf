resource "oci_core_public_ip" "this" {
  compartment_id = var.spec.compartment_id.value
  lifetime       = var.spec.lifetime
  display_name   = local.display_name
  freeform_tags  = local.freeform_tags

  private_ip_id    = var.spec.private_ip_id != null ? var.spec.private_ip_id.value : null
  public_ip_pool_id = var.spec.public_ip_pool_id != null ? var.spec.public_ip_pool_id.value : null
}
