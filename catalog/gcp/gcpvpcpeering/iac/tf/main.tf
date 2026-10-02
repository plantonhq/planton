# The CREATE form: this side's peering entry on `network`, pointing at
# `peer_network`, with the route exchange it allows.
#
# The name, both networks, and the two public-IP subnet-route flags are
# immutable (a change recreates the peering); the custom-route flags,
# stack_type, and update_strategy change in place. All four route-exchange
# flags are always sent (false is a real choice; the public-IP pair carries
# the spec's defaults when unset -- see locals.tf). Optional enums are sent
# only when set so the provider's defaults stay the provider's.
resource "google_compute_network_peering" "this" {
  count = local.is_create_form ? 1 : 0

  name         = local.peering_name
  network      = var.spec.network
  peer_network = var.spec.peer_network

  export_custom_routes                = var.spec.export_custom_routes
  import_custom_routes                = var.spec.import_custom_routes
  export_subnet_routes_with_public_ip = local.export_subnet_routes_with_public_ip
  import_subnet_routes_with_public_ip = local.import_subnet_routes_with_public_ip

  stack_type      = var.spec.stack_type != "" ? var.spec.stack_type : null
  update_strategy = var.spec.update_strategy != "" ? var.spec.update_strategy : null
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}

# The ROUTES-CONFIG form: the route exchange of a peering that already exists
# on `network` under peering_name -- typically the
# servicenetworking-googleapis-com peering Google creates for Cloud SQL
# private IP and other private-services-access products.
#
# The provider REQUIRES both custom-route flags here, so they are always
# sent; the two public-IP flags are always sent too, with the spec's defaults
# when unset (see locals.tf). Destroying this resource is a no-op in GCP --
# the peering keeps whatever flags it last had -- so there is no
# deletion_policy argument to send.
resource "google_compute_network_peering_routes_config" "this" {
  count = local.is_create_form ? 0 : 1

  peering = local.peering_name
  network = local.network_name
  project = local.network_project

  export_custom_routes                = var.spec.export_custom_routes
  import_custom_routes                = var.spec.import_custom_routes
  export_subnet_routes_with_public_ip = local.export_subnet_routes_with_public_ip
  import_subnet_routes_with_public_ip = local.import_subnet_routes_with_public_ip
}
