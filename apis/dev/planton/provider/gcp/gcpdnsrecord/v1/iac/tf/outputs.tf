output "fqdn" {
  description = "The fully qualified domain name of the created DNS record"
  value       = google_dns_record_set.record.name
}

output "record_type" {
  description = "The DNS record type that was created"
  value       = google_dns_record_set.record.type
}

output "managed_zone" {
  description = "The name of the managed zone containing this record"
  value       = google_dns_record_set.record.managed_zone
}

output "project_id" {
  description = "The GCP project ID where the record was created"
  value       = google_dns_record_set.record.project
}

output "ttl_seconds" {
  description = "The TTL (time to live) in seconds for the DNS record"
  value       = google_dns_record_set.record.ttl
}
