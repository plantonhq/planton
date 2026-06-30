# OpenFgaStore Main Resources
# This file creates the OpenFGA store resource.
#
# A store is a logical container for authorization data in OpenFGA.
# Each store contains authorization models and relationship tuples.
#
# Reference: https://registry.terraform.io/providers/openfga/openfga/latest/docs/resources/store

resource "openfga_store" "this" {
  name = local.store_name
}
