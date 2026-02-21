output "snapshot_id" {
  description = "The Hetzner Cloud image ID of the created snapshot"
  value       = hcloud_snapshot.this.id
}
