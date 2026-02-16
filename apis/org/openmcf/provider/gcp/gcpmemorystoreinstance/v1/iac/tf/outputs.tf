output "discovery_address" {
  description = "IP address of the instance's discovery endpoint"
  value = try(
    google_memorystore_instance.this.endpoints[0].connections[0].psc_auto_connection[0].ip_address,
    ""
  )
}

output "discovery_port" {
  description = "Port of the instance's discovery endpoint"
  value = try(
    google_memorystore_instance.this.endpoints[0].connections[0].psc_auto_connection[0].port,
    0
  )
}

output "instance_uid" {
  description = "Server-generated unique identifier for the instance"
  value       = google_memorystore_instance.this.uid
}

output "node_size_gb" {
  description = "Memory size per node in GB"
  value = try(
    google_memorystore_instance.this.node_config[0].size_gb,
    0
  )
}
