# Outputs for the DigitalOcean Volume

output "volume_id" {
  description = "The unique identifier (UUID) of the volume"
  value       = digitalocean_volume.this.id
}

output "volume_urn" {
  description = "The uniform resource name (URN) of the volume"
  value       = digitalocean_volume.this.urn
}

output "volume_name" {
  description = "The name of the volume"
  value       = digitalocean_volume.this.name
}

output "filesystem_type" {
  description = "The filesystem type of the volume"
  value       = digitalocean_volume.this.initial_filesystem_type
}

output "filesystem_label" {
  description = "The filesystem label of the volume"
  value       = digitalocean_volume.this.filesystem_label
}

output "droplet_ids" {
  description = "List of Droplet IDs currently attached to this volume"
  value       = digitalocean_volume.this.droplet_ids
}

# Complete outputs object matching stack_outputs.proto structure
output "outputs" {
  description = "Complete volume outputs for integration with other resources"
  value = {
    volume_id = digitalocean_volume.this.id
  }
}

