# outputs.tf

output "record_id" {
  description = "The unique identifier of the created DNS record"
  value       = cloudflare_dns_record.main.id
}

output "hostname" {
  description = "The name of the DNS record"
  value       = cloudflare_dns_record.main.name
}

output "record_type" {
  description = "The DNS record type that was created"
  value       = cloudflare_dns_record.main.type
}

output "proxied" {
  description = "Whether the record is proxied through Cloudflare"
  value       = cloudflare_dns_record.main.proxied
}
