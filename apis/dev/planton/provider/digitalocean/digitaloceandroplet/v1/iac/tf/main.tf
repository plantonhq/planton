# Create the DigitalOcean Droplet
resource "digitalocean_droplet" "droplet" {
  name   = var.spec.droplet_name
  region = local.region_slug
  size   = var.spec.size
  image  = var.spec.image

  # SSH keys for secure access
  ssh_keys = var.spec.ssh_keys

  # VPC networking
  vpc_uuid = local.vpc_uuid

  # Optional features
  ipv6       = local.ipv6
  backups    = local.backups
  monitoring = local.monitoring

  # Volume attachments
  volume_ids = length(local.volume_ids) > 0 ? local.volume_ids : null

  # Tags for organization and firewall integration
  tags = local.tags

  # Cloud-init user data for first-boot provisioning
  user_data = var.spec.user_data

  # Lifecycle management
  lifecycle {
    create_before_destroy = false
    ignore_changes        = []
  }
}

