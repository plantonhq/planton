# main.tf

# Create the OpenStack Neutron router interface (attach router to subnet).
# This creates a port on the subnet and attaches it to the router.
resource "openstack_networking_router_interface_v2" "main" {
  router_id = local.router_id
  subnet_id = local.subnet_id

  # Region override (optional)
  region = var.spec.region != "" ? var.spec.region : null
}
