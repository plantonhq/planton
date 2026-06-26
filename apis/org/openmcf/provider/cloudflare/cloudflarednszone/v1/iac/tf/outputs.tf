# outputs.tf

output "zone_id" {
  description = "The unique identifier of the created Cloudflare zone"
  value       = cloudflare_zone.main.id
}

output "nameservers" {
  description = "The Cloudflare nameservers assigned to this zone"
  value       = cloudflare_zone.main.name_servers
}

output "status" {
  description = "The zone status on Cloudflare"
  value       = cloudflare_zone.main.status
}

output "dnssec_status" {
  description = "DNSSEC status (empty when DNSSEC is not enabled)"
  value       = local.has_dnssec ? cloudflare_zone_dnssec.main[0].status : ""
}

output "dnssec_ds" {
  description = "The full DS record to enter at the registrar"
  value       = local.has_dnssec ? cloudflare_zone_dnssec.main[0].ds : ""
}

output "dnssec_digest" {
  description = "The DS record digest"
  value       = local.has_dnssec ? cloudflare_zone_dnssec.main[0].digest : ""
}

output "dnssec_digest_type" {
  description = "The DS digest type code"
  value       = local.has_dnssec ? cloudflare_zone_dnssec.main[0].digest_type : ""
}

output "dnssec_digest_algorithm" {
  description = "The DS digest algorithm"
  value       = local.has_dnssec ? cloudflare_zone_dnssec.main[0].digest_algorithm : ""
}

output "dnssec_algorithm" {
  description = "The DNSKEY algorithm code"
  value       = local.has_dnssec ? cloudflare_zone_dnssec.main[0].algorithm : ""
}

output "dnssec_key_tag" {
  description = "The DNSKEY key tag"
  value       = local.has_dnssec ? tostring(cloudflare_zone_dnssec.main[0].key_tag) : ""
}

output "dnssec_public_key" {
  description = "The DNSKEY public key"
  value       = local.has_dnssec ? cloudflare_zone_dnssec.main[0].public_key : ""
}

output "dnssec_flags" {
  description = "The DNSKEY flags"
  value       = local.has_dnssec ? tostring(cloudflare_zone_dnssec.main[0].flags) : ""
}
