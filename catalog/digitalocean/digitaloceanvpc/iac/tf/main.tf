# DigitalOcean VPC.
#
# The IP range is immutable: DigitalOcean assigns one when ip_range is unset
# (Optional+Computed, so the assigned range lands in state without a diff),
# and an explicit change to a set range REPLACES the VPC -- surfaced honestly
# at plan time rather than silently ignored.
#
# Destroy caveat the module cannot fix: in a region with no VPC yet, the
# first VPC created becomes the region's DEFAULT, and the API refuses to
# delete default VPCs (403). The `default` attribute is computed and cannot
# be steered from here; the kind's GUIDE carries the account-level recovery.
resource "digitalocean_vpc" "vpc" {
  name   = local.vpc_name
  region = var.spec.region

  description = var.spec.description != "" ? var.spec.description : null

  ip_range = var.spec.ip_range_cidr != "" ? var.spec.ip_range_cidr : null
}
