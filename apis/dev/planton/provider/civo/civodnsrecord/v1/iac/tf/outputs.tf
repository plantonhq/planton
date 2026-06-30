# outputs.tf

output "record_id" {
  description = "The unique identifier of the created DNS record"
  value       = civo_dns_domain_record.main.id
}

output "hostname" {
  description = "The fully qualified hostname of the DNS record"
  value       = civo_dns_domain_record.main.name
}

output "record_type" {
  description = "The DNS record type that was created"
  value       = civo_dns_domain_record.main.type
}

output "account_id" {
  description = "The Civo account ID"
  value       = civo_dns_domain_record.main.account_id
}
