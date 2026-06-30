# Create Civo Network (VPC)
resource "civo_network" "main" {
  label  = local.network_name
  region = local.region

  # Optional: Explicit CIDR block (if empty, Civo auto-allocates)
  cidr_v4 = local.cidr_block != "" ? local.cidr_block : null

  # Optional: Set as default network for the region
  # Note: Only one default network is allowed per region
  default = local.is_default
}

# Informational outputs for limitations
# Description specified: ${local.description}
# Note: The Civo Network provider doesn't currently support description field.
# This is recorded in Planton metadata only.

