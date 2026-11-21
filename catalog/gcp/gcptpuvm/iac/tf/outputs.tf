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

output "state" {
  description = "The TPU's state as Google reported it at the end of the deploy"
  value       = google_tpu_v2_vm.this.state
}

output "health" {
  description = "The TPU's health as Google reported it at the end of the deploy"
  value       = google_tpu_v2_vm.this.health
}

# Worker 0 first: the endpoint a runtime client reaches first.
output "worker_internal_ips" {
  description = "The internal IP address of every TPU worker"
  value       = [for endpoint in google_tpu_v2_vm.this.network_endpoints : endpoint.ip_address]
}

output "worker_external_ips" {
  description = "The external IP address of every TPU worker that has one"
  value = flatten([
    for endpoint in google_tpu_v2_vm.this.network_endpoints : [
      for access in endpoint.access_config : access.external_ip if access.external_ip != null && access.external_ip != ""
    ]
  ])
}
