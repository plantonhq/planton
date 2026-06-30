output "instance_name" {
  description = "Name of the Compute Engine instance"
  value       = google_compute_instance.instance.name
}

output "instance_id" {
  description = "Instance ID (unique numeric identifier)"
  value       = google_compute_instance.instance.instance_id
}

output "self_link" {
  description = "Self link URL of the instance"
  value       = google_compute_instance.instance.self_link
}

output "internal_ip" {
  description = "Internal (private) IP address of the instance"
  value       = length(google_compute_instance.instance.network_interface) > 0 ? google_compute_instance.instance.network_interface[0].network_ip : null
}

output "external_ip" {
  description = "External (public) IP address of the instance"
  value       = (
    length(google_compute_instance.instance.network_interface) > 0 &&
    length(google_compute_instance.instance.network_interface[0].access_config) > 0
    ? google_compute_instance.instance.network_interface[0].access_config[0].nat_ip
    : null
  )
}

output "status" {
  description = "Current status of the instance"
  value       = google_compute_instance.instance.current_status
}

output "zone" {
  description = "Zone where the instance is located"
  value       = google_compute_instance.instance.zone
}

output "machine_type" {
  description = "Machine type of the instance"
  value       = google_compute_instance.instance.machine_type
}

output "cpu_platform" {
  description = "CPU platform of the instance"
  value       = google_compute_instance.instance.cpu_platform
}
