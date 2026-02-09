# main.tf

# Create the OpenStack Neutron network.
resource "openstack_networking_network_v2" "main" {
  name        = local.network_name
  description = var.spec.description != "" ? var.spec.description : null

  admin_state_up = var.spec.admin_state_up

  # Shared and external are admin-only operations.
  # Only set when explicitly true to avoid permission errors for tenant users.
  shared   = var.spec.shared ? true : null
  external = var.spec.external ? true : null

  # MTU override (0 means unset, let OpenStack decide)
  mtu = var.spec.mtu > 0 ? var.spec.mtu : null

  # DNS domain (empty means unset)
  dns_domain = var.spec.dns_domain != "" ? var.spec.dns_domain : null

  # Port security (null means let OpenStack deployment decide)
  port_security_enabled = var.spec.port_security_enabled

  # Tags applied to the OpenStack resource
  tags = length(var.spec.tags) > 0 ? toset(var.spec.tags) : null

  # Region override (optional)
  region = var.spec.region != "" ? var.spec.region : null
}
