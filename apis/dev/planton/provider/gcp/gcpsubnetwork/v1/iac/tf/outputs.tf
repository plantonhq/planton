# Output the self-link of the created subnetwork
# This is useful for GKE clusters or other resources that need to reference this subnet
output "subnetwork_self_link" {
  description = "Self-link URL of the created subnetwork"
  value       = google_compute_subnetwork.main.self_link
}

# Output the region where this subnetwork resides
output "region" {
  description = "The region where this subnetwork resides"
  value       = google_compute_subnetwork.main.region
}

# Output the primary IPv4 CIDR of the subnet
output "ip_cidr_range" {
  description = "The primary IPv4 CIDR of the subnet"
  value       = google_compute_subnetwork.main.ip_cidr_range
}

# Output the list of secondary ranges created in this subnet
output "secondary_ranges" {
  description = "List of secondary ranges with their names and CIDRs"
  value = [
    for range in google_compute_subnetwork.main.secondary_ip_range : {
      range_name    = range.range_name
      ip_cidr_range = range.ip_cidr_range
    }
  ]
}

