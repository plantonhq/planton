# Stack outputs matching ScalewayBlockVolumeStackOutputs proto.

output "volume_id" {
  description = "The unique identifier of the volume (format: {zone}/{uuid})"
  value       = scaleway_block_volume.volume.id
}

output "volume_name" {
  description = "The name of the volume as it exists in Scaleway Block Storage"
  value       = scaleway_block_volume.volume.name
}

output "zone" {
  description = "The Availability Zone where the volume is deployed"
  value       = local.zone
}

# Complete outputs object matching stack_outputs.proto structure.
# Used by the Planton platform to populate status.outputs.
output "outputs" {
  description = "Complete volume outputs for integration with other resources"
  value = {
    volume_id   = scaleway_block_volume.volume.id
    volume_name = scaleway_block_volume.volume.name
    zone        = local.zone
  }
}
