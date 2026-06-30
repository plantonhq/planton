output "zone_id" {
  description = "The managed zone ID (numeric identifier)"
  value       = google_dns_managed_zone.managed_zone.id
}

output "zone_name" {
  description = "The name of the created Managed Zone"
  value       = google_dns_managed_zone.managed_zone.name
}

output "nameservers" {
  description = "The list of nameservers assigned to this Managed Zone. Configure these at your domain registrar."
  value       = google_dns_managed_zone.managed_zone.name_servers
}

output "gcp_project_id" {
  description = "The GCP project ID where the Managed Zone is created"
  value       = google_dns_managed_zone.managed_zone.project
}
