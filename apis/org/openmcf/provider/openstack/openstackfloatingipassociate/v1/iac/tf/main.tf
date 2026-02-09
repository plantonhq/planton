# main.tf

# Create the OpenStack Neutron floating IP association.
# This binds a floating IP to a port, providing external connectivity
# to whatever is attached to the port (typically an instance).
resource "openstack_networking_floatingip_associate_v2" "main" {
  floating_ip = local.floating_ip
  port_id     = local.port_id

  # Fixed IP on the port to map to (only needed for multi-IP ports).
  fixed_ip = var.spec.fixed_ip != "" ? var.spec.fixed_ip : null

  # Region override (optional, ForceNew).
  region = var.spec.region != "" ? var.spec.region : null
}
