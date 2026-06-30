# OpenFgaStore Local Values
# This file computes local values from the input variables.

locals {
  # Store name from spec (used directly in the openfga_store resource)
  store_name = var.spec.name
}
