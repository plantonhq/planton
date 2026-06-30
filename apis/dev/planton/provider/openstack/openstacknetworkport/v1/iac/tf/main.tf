# main.tf

# Create the OpenStack Neutron port.
resource "openstack_networking_port_v2" "main" {
  name       = var.metadata.name
  network_id = local.network_id

  # Fixed IP allocations (optional).
  # Each entry assigns an IP from a subnet, with optional specific IP.
  dynamic "fixed_ip" {
    for_each = var.spec.fixed_ips
    content {
      subnet_id  = fixed_ip.value.subnet_id != null ? fixed_ip.value.subnet_id.value : null
      ip_address = fixed_ip.value.ip_address != "" ? fixed_ip.value.ip_address : null
    }
  }

  # Security group IDs (optional, mutually exclusive with no_security_groups).
  security_group_ids = length(local.security_group_ids) > 0 ? toset(local.security_group_ids) : null

  # Explicitly remove all security groups including the default.
  no_security_groups = var.spec.no_security_groups ? true : null

  # Administrative state (default true from variables.tf).
  admin_state_up = var.spec.admin_state_up

  # Specific MAC address (optional, ForceNew).
  mac_address = var.spec.mac_address != "" ? var.spec.mac_address : null

  # Port security enforcement (optional, inherits from network if null).
  port_security_enabled = var.spec.port_security_enabled

  # Description (empty means unset).
  description = var.spec.description != "" ? var.spec.description : null

  # Tags applied to the OpenStack resource.
  tags = length(var.spec.tags) > 0 ? toset(var.spec.tags) : null

  # Region override (optional).
  region = var.spec.region != "" ? var.spec.region : null
}
