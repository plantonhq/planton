# ── 1. Dedicated Flexible IP (optional) ────────────────────────────────────────
#
# A dedicated public IPv4 address for the instance. Created as a separate
# resource to give explicit lifecycle control -- the IP survives instance
# replacement, preserving DNS records and firewall rules.
#
# Only created when spec.public_ip is set.
resource "scaleway_instance_ip" "ip" {
  count = local.create_public_ip ? 1 : 0

  tags = local.standard_tags
  zone = local.zone
}

# ── 2. Additional Local Volumes (optional) ─────────────────────────────────────
#
# Local volumes (l_ssd, scratch) created alongside the instance and attached
# via additional_volume_ids. These volumes share the instance's lifecycle --
# they are destroyed when the instance is terminated.
resource "scaleway_instance_volume" "volumes" {
  for_each = local.additional_volumes_map

  name       = each.value.name
  type       = each.value.volume_type
  size_in_gb = each.value.size_in_gb
  tags       = local.standard_tags
  zone       = local.zone
}

# ── 3. Instance Server ─────────────────────────────────────────────────────────
#
# The compute instance itself. References the Flexible IP (if created),
# additional volumes, optional security group, and optional Private Network
# attachment.
resource "scaleway_instance_server" "server" {
  name  = local.instance_name
  type  = local.instance_type
  image = local.image
  tags  = local.standard_tags
  zone  = local.zone

  # Attach the dedicated Flexible IP if created.
  ip_id = local.create_public_ip ? scaleway_instance_ip.ip[0].id : null

  # Attach additional volumes (both bundled local volumes and any external).
  additional_volume_ids = length(scaleway_instance_volume.volumes) > 0 ? [
    for key, vol in scaleway_instance_volume.volumes : vol.id
  ] : null

  # Root volume configuration.
  dynamic "root_volume" {
    for_each = var.spec.root_volume != null ? [var.spec.root_volume] : []
    content {
      size_in_gb            = local.root_volume_size_in_gb
      volume_type           = local.root_volume_type
      delete_on_termination = local.root_volume_delete_on_termination
      sbs_iops              = local.root_volume_sbs_iops
    }
  }

  # Security group.
  security_group_id = local.security_group_id != "" ? local.security_group_id : null

  # Private Network attachment (inline block on server).
  dynamic "private_network" {
    for_each = local.private_network_id != "" ? [1] : []
    content {
      pn_id = local.private_network_id
    }
  }

  # Cloud-init script.
  cloud_init = local.cloud_init != "" ? local.cloud_init : null

  # Instance state and protection.
  state     = local.state
  protected = local.protected
}
