output "network_self_link" {
  description = "The full self-link URL of the created VPC network (useful for connecting subnets or other resources)"
  value       = google_compute_network.vpc.self_link
}

output "private_services_ip_range_name" {
  description = "Name of the allocated IP range for private services (only set if private_services_access is enabled)"
  value       = local.enable_private_services ? google_compute_global_address.private_services_range[0].name : ""
}

output "private_services_ip_range_cidr" {
  description = "CIDR of the allocated IP range for private services (only set if private_services_access is enabled)"
  value       = local.enable_private_services ? "${google_compute_global_address.private_services_range[0].address}/${local.private_services_prefix_length}" : ""
}

