# main.tf

# Create the OpenStack Neutron floating IP.
resource "openstack_networking_floatingip_v2" "main" {
  # "pool" is the TF name for the external network ID.
  pool = local.floating_network_id

  # Port association (optional -- omit for allocation-only mode).
  port_id = local.port_id

  # Fixed IP (only meaningful when port_id is set and port has multiple IPs).
  fixed_ip = var.spec.fixed_ip != "" ? var.spec.fixed_ip : null

  # Allocate from a specific subnet within the external network.
  subnet_id = var.spec.subnet_id != "" ? var.spec.subnet_id : null

  # Request a specific floating IP address.
  address = var.spec.address != "" ? var.spec.address : null

  # Description (empty means unset).
  description = var.spec.description != "" ? var.spec.description : null

  # Tags applied to the OpenStack resource.
  tags = length(var.spec.tags) > 0 ? toset(var.spec.tags) : null

  # Region override (optional).
  region = var.spec.region != "" ? var.spec.region : null
}
