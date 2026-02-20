resource "oci_network_firewall_network_firewall_policy" "this" {
  compartment_id = var.spec.compartment_id.value
  display_name   = local.policy_display_name
  freeform_tags  = local.freeform_tags

  description = var.spec.policy.description != "" ? var.spec.policy.description : null
}
