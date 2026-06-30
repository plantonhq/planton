output "endpoint" {
  description = "The registry hostname (e.g., registry.digitalocean.com/my-registry)"
  value       = digitalocean_container_registry.registry.endpoint
}

output "server_url" {
  description = "The full registry URL for Docker login"
  value       = digitalocean_container_registry.registry.server_url
}

output "name" {
  description = "The name of the container registry"
  value       = digitalocean_container_registry.registry.name
}

output "subscription_tier_slug" {
  description = "The subscription tier slug"
  value       = digitalocean_container_registry.registry.subscription_tier_slug
}

output "region" {
  description = "The region where the registry is deployed"
  value       = digitalocean_container_registry.registry.region
}

output "created_at" {
  description = "The timestamp when the registry was created"
  value       = digitalocean_container_registry.registry.created_at
}

output "storage_usage_bytes" {
  description = "Current storage usage in bytes"
  value       = digitalocean_container_registry.registry.storage_usage_bytes
}

output "docker_credentials" {
  description = "Temporary Docker credentials for registry access (sensitive)"
  value       = digitalocean_container_registry_docker_credentials.credentials.docker_credentials
  sensitive   = true
}

