output "certificate_id" {
  description = "The Origin CA certificate identifier"
  value       = cloudflare_origin_ca_certificate.main.id
}

output "certificate" {
  description = "The issued Origin CA certificate (PEM); public material, not a secret"
  value       = cloudflare_origin_ca_certificate.main.certificate
}

output "private_key" {
  description = "The generated private key (PEM); empty when a user-supplied CSR was used (sensitive)"
  value       = local.generate_key ? tls_private_key.origin[0].private_key_pem : ""
  sensitive   = true
}

output "expires_on" {
  description = "RFC3339 timestamp of when the certificate expires"
  value       = cloudflare_origin_ca_certificate.main.expires_on
}
