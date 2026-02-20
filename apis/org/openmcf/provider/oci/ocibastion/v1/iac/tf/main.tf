resource "oci_bastion_bastion" "this" {
  compartment_id = var.spec.compartment_id.value
  target_subnet_id = var.spec.target_subnet_id.value
  bastion_type   = "STANDARD"
  name           = local.display_name
  freeform_tags  = local.freeform_tags

  client_cidr_block_allow_list = var.spec.client_cidr_block_allow_list

  max_session_ttl_in_seconds = var.spec.max_session_ttl_in_seconds

  dns_proxy_status = local.dns_proxy_status
}
