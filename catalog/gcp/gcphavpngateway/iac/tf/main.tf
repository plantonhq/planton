# Enable the Compute Engine API so a fresh project can host the gateway.
# disable_on_destroy is false: tearing down one gateway must never disable
# the API for everything else in the project.
resource "google_project_service" "compute_api" {
  project = local.project_id
  service = "compute.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The HA VPN gateway: the Google Cloud end of every IPsec VPN from this VPC,
# with two interfaces that each get a public IP (or, pinned to Interconnect
# attachments, none).
#
# Nearly everything is immutable: the network, region, IP version, stack
# type, and interface pinning recreate the gateway -- and a recreated gateway
# has NEW public IPs. Labels change in place. Optional inputs are sent only
# when set so the provider's defaults stay the provider's.
resource "google_compute_ha_vpn_gateway" "this" {
  project = local.project_id
  name    = local.gateway_name
  region  = var.spec.region
  network = var.spec.network

  description        = var.spec.description != "" ? var.spec.description : null
  gateway_ip_version = var.spec.gateway_ip_version != "" ? var.spec.gateway_ip_version : null
  stack_type         = var.spec.stack_type != "" ? var.spec.stack_type : null
  labels             = local.final_labels

  # HA VPN over Interconnect: pin each interface to its VLAN attachment.
  # Empty means the ordinary internet-facing gateway, whose interfaces the
  # API fills in with public IPs (the block is Optional+Computed, so it stays
  # out of the payload when unset).
  dynamic "vpn_interfaces" {
    for_each = var.spec.vpn_interfaces
    content {
      id                      = vpn_interfaces.value.id
      interconnect_attachment = vpn_interfaces.value.interconnect_attachment
    }
  }

  # Create-time resource-manager tags (org policy / IAM conditions).
  dynamic "params" {
    for_each = length(var.spec.resource_manager_tags) > 0 ? [1] : []
    content {
      resource_manager_tags = var.spec.resource_manager_tags
    }
  }

  deletion_policy = local.deletion_policy

  depends_on = [google_project_service.compute_api]
}

# The Cloud Router the gateway's tunnels terminate BGP on. One router per
# gateway: Google allows many routers per network and region, so a VPN
# router beside a NAT router is the normal topology.
#
# The ASN and encrypted-interconnect flag recreate the router; the
# description and BGP advertisement change in place. The BGP block is
# required by the spec: an HA VPN router exists to run BGP and every tunnel
# session needs the ASN.
resource "google_compute_router" "this" {
  project = local.project_id
  name    = local.router_name
  region  = var.spec.region
  network = var.spec.network

  description = var.spec.router.description != "" ? var.spec.router.description : null

  # Dedicates the router to encrypted VLAN attachments (HA VPN over
  # Interconnect). Immutable; an encrypted router cannot be converted.
  encrypted_interconnect_router = var.spec.router.encrypted_interconnect_router

  bgp {
    asn = var.spec.router.bgp.asn

    # advertise_mode empty means DEFAULT; custom groups/ranges are legal only
    # in CUSTOM mode (spec CELs mirror the provider).
    advertise_mode     = var.spec.router.bgp.advertise_mode != "" ? var.spec.router.bgp.advertise_mode : null
    advertised_groups  = length(var.spec.router.bgp.advertised_groups) > 0 ? var.spec.router.bgp.advertised_groups : null
    keepalive_interval = var.spec.router.bgp.keepalive_interval > 0 ? var.spec.router.bgp.keepalive_interval : null

    # identifier_range is Optional+Computed: an empty value must stay out of
    # the payload or it would fight the API's computed value.
    identifier_range = var.spec.router.bgp.identifier_range != "" ? var.spec.router.bgp.identifier_range : null

    dynamic "advertised_ip_ranges" {
      for_each = var.spec.router.bgp.advertised_ip_ranges
      content {
        range       = advertised_ip_ranges.value.range
        description = advertised_ip_ranges.value.description != "" ? advertised_ip_ranges.value.description : null
      }
    }
  }

  dynamic "params" {
    for_each = length(var.spec.router.resource_manager_tags) > 0 ? [1] : []
    content {
      resource_manager_tags = var.spec.router.resource_manager_tags
    }
  }

  deletion_policy = local.deletion_policy

  depends_on = [google_project_service.compute_api]
}
