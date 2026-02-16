resource "aws_eip" "this" {
  domain = "vpc"

  # BYOIP: allocate from a specific IPv4 address pool.
  public_ipv4_pool = local.public_ipv4_pool

  # BYOIP: request a specific IP address from the pool.
  address = local.address

  # Location scope for Local Zones and Wavelength zones.
  network_border_group = local.network_border_group

  tags = local.tags
}
