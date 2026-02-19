output "placement_group_id" {
  description = "The Hetzner Cloud numeric ID of the placement group"
  value       = hcloud_placement_group.this.id
}
