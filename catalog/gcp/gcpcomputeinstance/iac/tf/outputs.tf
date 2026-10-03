# Semantic outputs mirroring GcpComputeInstanceOutputs — names and
# shapes byte-identical to the Pulumi module's exports.

output "instance_name" {
  description = "Name of the Compute Engine instance"
  value       = google_compute_instance.this.name
}

output "instance_id" {
  description = "Server-assigned unique numeric identifier of the instance"
  value       = google_compute_instance.this.instance_id
}

output "self_link" {
  description = "Self-link URL of the instance"
  value       = google_compute_instance.this.self_link
}

output "internal_ip" {
  description = "Primary internal IP address (first network interface)"
  value       = google_compute_instance.this.network_interface[0].network_ip
}

# Empty when the VM has no external access config — private VMs export ""
# rather than failing, so downstream consumers can branch on presence.
output "external_ip" {
  description = "External IP address of the first interface, if configured"
  value = try(
    google_compute_instance.this.network_interface[0].access_config[0].nat_ip,
    ""
  )
}

output "status" {
  description = "Current status of the instance (RUNNING, TERMINATED, ...)"
  value       = google_compute_instance.this.current_status
}

output "zone" {
  description = "Zone of the instance"
  value       = google_compute_instance.this.zone
}

output "machine_type" {
  description = "Machine type of the instance"
  value       = google_compute_instance.this.machine_type
}

output "cpu_platform" {
  description = "CPU platform the instance landed on"
  value       = google_compute_instance.this.cpu_platform
}
