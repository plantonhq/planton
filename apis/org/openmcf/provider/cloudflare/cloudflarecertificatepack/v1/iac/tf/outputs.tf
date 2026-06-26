output "certificate_pack_id" {
  description = "The certificate pack identifier"
  value       = cloudflare_certificate_pack.main.id
}

output "status" {
  description = "The order/issuance status"
  value       = cloudflare_certificate_pack.main.status
}

output "primary_certificate" {
  description = "The identifier of the primary certificate in the pack"
  value       = cloudflare_certificate_pack.main.primary_certificate
}
