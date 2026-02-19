output "volume_id" {
  description = "The Hetzner Cloud numeric ID of the created volume"
  value       = hcloud_volume.this.id
}

output "linux_device" {
  description = "The Linux device path for the volume on the attached server"
  value       = hcloud_volume.this.linux_device
}
