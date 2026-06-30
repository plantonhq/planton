# DigitalOcean Container Registry Resource
#
# This module provisions a private Docker container registry on DigitalOcean.
# 
# Key features:
# - Subscription tier selection (STARTER/BASIC/PROFESSIONAL)
# - Region selection for data locality
# - OCI-compliant registry for Docker images and Helm charts
#
# Note: Garbage collection is handled by Planton's custom controller
# because Terraform doesn't support GC scheduling via the DigitalOcean API.

resource "digitalocean_container_registry" "registry" {
  name                   = var.spec.name
  subscription_tier_slug = local.subscription_tier_slug
  region                 = local.region_slug
}

# Docker credentials resource
# This generates temporary Docker credentials for accessing the registry
# Note: These credentials expire and should be rotated regularly
resource "digitalocean_container_registry_docker_credentials" "credentials" {
  registry_name = digitalocean_container_registry.registry.name
  write         = true
}

