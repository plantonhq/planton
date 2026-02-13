# ── 1. Dedicated Flexible IP ──────────────────────────────────────────────────
#
# A dedicated public IPv4 address for the gateway. Creating it as a separate
# resource (rather than letting the gateway auto-create one) gives explicit
# control over the IP lifecycle -- the IP survives gateway replacement.
resource "scaleway_vpc_public_gateway_ip" "ip" {
  reverse = local.reverse_dns != "" ? local.reverse_dns : null
  tags    = local.standard_tags
  zone    = local.zone
}

# ── 2. Public Gateway ────────────────────────────────────────────────────────
#
# The managed network appliance that provides NAT, SSH bastion, and port
# forwarding for the attached Private Network.
resource "scaleway_vpc_public_gateway" "gateway" {
  name              = local.gateway_name
  type              = local.gateway_type
  ip_id             = scaleway_vpc_public_gateway_ip.ip.id
  bastion_enabled   = local.bastion_enabled
  bastion_port      = local.bastion_port
  allowed_ip_ranges = local.allowed_ip_ranges
  enable_smtp       = local.enable_smtp
  tags              = local.standard_tags
  zone              = local.zone
}

# ── 3. Gateway-to-Network attachment ─────────────────────────────────────────
#
# The glue resource that connects the gateway to the Private Network.
# Without it, the gateway exists but serves no network.
# enable_masquerade controls whether NAT is active.
resource "scaleway_vpc_gateway_network" "attachment" {
  gateway_id         = scaleway_vpc_public_gateway.gateway.id
  private_network_id = local.private_network_id
  enable_masquerade  = local.enable_masquerade
  zone               = local.zone
}

# ── 4. PAT rules (optional) ─────────────────────────────────────────────────
#
# Port forwarding rules that map public ports on the gateway's IP to
# private IP:port pairs inside the attached Private Network.
resource "scaleway_vpc_public_gateway_pat_rule" "rules" {
  for_each = { for i, rule in local.pat_rules : tostring(i) => rule }

  gateway_id   = scaleway_vpc_public_gateway.gateway.id
  private_ip   = each.value.private_ip
  private_port = each.value.private_port
  public_port  = each.value.public_port
  protocol     = each.value.protocol != "" ? each.value.protocol : "both"
  zone         = local.zone
}
