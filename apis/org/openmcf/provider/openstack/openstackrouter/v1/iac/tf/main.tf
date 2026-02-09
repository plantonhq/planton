# main.tf

# Create the OpenStack Neutron router.
resource "openstack_networking_router_v2" "main" {
  name = local.router_name

  # External gateway (optional)
  external_network_id = local.external_network_id

  # Administrative state (default: true via variable default)
  admin_state_up = var.spec.admin_state_up

  # SNAT control (only meaningful when external_network_id is set)
  enable_snat = var.spec.enable_snat

  # Distributed Virtual Router mode (create-time only)
  distributed = var.spec.distributed

  # External fixed IPs (only meaningful when external_network_id is set)
  dynamic "external_fixed_ip" {
    for_each = var.spec.external_fixed_ips
    content {
      subnet_id  = external_fixed_ip.value.subnet_id != "" ? external_fixed_ip.value.subnet_id : null
      ip_address = external_fixed_ip.value.ip_address != "" ? external_fixed_ip.value.ip_address : null
    }
  }

  # Description (empty means unset)
  description = var.spec.description != "" ? var.spec.description : null

  # Tags applied to the OpenStack resource
  tags = length(var.spec.tags) > 0 ? toset(var.spec.tags) : null

  # Region override (optional)
  region = var.spec.region != "" ? var.spec.region : null
}
