output "record_id" {
  description = "The unique identifier of the created DNS record"
  value       = digitalocean_record.dns_record.id
}

output "hostname" {
  description = "The fully qualified hostname of the DNS record"
  value       = local.hostname
}

output "record_type" {
  description = "The DNS record type that was created"
  value       = local.type
}

output "domain" {
  description = "The domain name where the record was created"
  value       = local.domain
}

output "ttl_seconds" {
  description = "The TTL applied to the record in seconds"
  value       = local.ttl_seconds
}
