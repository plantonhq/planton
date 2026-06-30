output "certificate_id" {
  description = "The unique identifier (UUID) of the created certificate"
  value       = digitalocean_certificate.certificate.id
}

output "expiry_rfc3339" {
  description = "The expiration timestamp of the certificate in RFC 3339 format"
  value       = digitalocean_certificate.certificate.not_after
}

output "certificate_name" {
  description = "The name of the certificate"
  value       = digitalocean_certificate.certificate.name
}

output "certificate_type" {
  description = "The type of certificate (lets_encrypt or custom)"
  value       = digitalocean_certificate.certificate.type
}

output "certificate_state" {
  description = "The state of the certificate (verified, pending, or error)"
  value       = digitalocean_certificate.certificate.state
}

