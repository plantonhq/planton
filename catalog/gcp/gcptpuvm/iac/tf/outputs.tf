output "name" {
  description = "Full resource name of the TPU (projects/{project}/locations/{zone}/nodes/{node_id})"
  value       = google_tpu_v2_vm.this.id
}

output "node_id" {
  description = "The TPU's id"
  value       = google_tpu_v2_vm.this.name
}

output "zone" {
  description = "The zone the TPU runs in"
  value       = google_tpu_v2_vm.this.zone
}
