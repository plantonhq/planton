# Scaleway VPC Resource
resource "scaleway_vpc" "vpc" {
  name   = local.vpc_name
  region = local.region
  tags   = local.standard_tags

  # Routing between Private Networks in this VPC.
  # IMPORTANT: Once enabled, routing cannot be disabled.
  enable_routing = local.enable_routing

  # Custom routes propagation between Private Networks.
  # IMPORTANT: Once enabled, this cannot be disabled.
  enable_custom_routes_propagation = local.enable_custom_routes_propagation

  # Lifecycle management
  lifecycle {
    # Routing flags are one-way toggles (can be enabled but never disabled).
    # Ignoring changes prevents Terraform from attempting to disable them
    # if the spec is later modified, which would fail at the API level.
    ignore_changes = [
      enable_routing,
      enable_custom_routes_propagation,
    ]
  }
}
