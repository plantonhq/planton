# DigitalOcean Reserved IP
#
# Provisions a static public IP reservation -- one kind covering the
# provider's four resources: digitalocean_reserved_ip (IPv4, with the
# genuinely mutable inline droplet assignment), digitalocean_reserved_ipv6,
# and digitalocean_reserved_ipv6_assignment. The v4/v6 API asymmetries live
# entirely in this module:
#
#   - v4 assigns through the resource's own droplet_id argument, which the
#     provider updates in place (assign / re-point / unassign without ever
#     replacing the address).
#   - v6 CANNOT assign inline: the provider's v6 create silently ignores a
#     droplet id and the resource has no update function (an inline set
#     would produce a plan that never converges), so assignment goes
#     through the separate digitalocean_reserved_ipv6_assignment resource.
#     Re-pointing replaces just the assignment, never the address.
#
# BILLING: an UNASSIGNED reserved IPv4 accrues charges (~$5/month,
# prorated); an assigned one is free. IPv6 reservations are free either
# way. Destroy the component when the address is no longer needed.

locals {
  is_ipv6 = var.spec.ip_version == "ipv6"

  # Droplet references resolve to the literal numeric droplet id before the
  # module runs; empty means unassigned.
  droplet_id = try(var.spec.droplet, "") != "" ? tonumber(var.spec.droplet) : null
}

# --- IPv4 -------------------------------------------------------------

resource "digitalocean_reserved_ip" "ipv4" {
  count = local.is_ipv6 ? 0 : 1

  # The provider lowercases the region into state; the spec's region enum
  # value names are already the lowercase slugs, so no normalization diff.
  region     = var.spec.region
  droplet_id = local.droplet_id
}

# --- IPv6 -------------------------------------------------------------

resource "digitalocean_reserved_ipv6" "ipv6" {
  count = local.is_ipv6 ? 1 : 0

  # Same region concept; the v6 resource names the argument region_slug.
  region_slug = var.spec.region
}

resource "digitalocean_reserved_ipv6_assignment" "ipv6_assignment" {
  count = local.is_ipv6 && local.droplet_id != null ? 1 : 0

  ip         = digitalocean_reserved_ipv6.ipv6[0].ip
  droplet_id = local.droplet_id
}
