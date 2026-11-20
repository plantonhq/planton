output "name" {
  description = "Full resource name of the certificate (projects/{project}/locations/{location}/caPools/{pool}/certificates/{certificate_id})"
  value       = google_privateca_certificate.this.id
}

output "certificate_id" {
  description = "The certificate's ID"
  value       = google_privateca_certificate.this.name
}

output "pem_certificate" {
  description = "The signed certificate, PEM"
  value       = google_privateca_certificate.this.pem_certificate
}

output "pem_certificate_chain" {
  description = "The chain that verifies the certificate, PEM, issuer first and root last"
  value       = google_privateca_certificate.this.pem_certificate_chain
}

output "issuer_certificate_authority" {
  description = "The full name of the authority that signed the certificate"
  value       = google_privateca_certificate.this.issuer_certificate_authority
}

# The validity window, read from the provider's description of the signed
# certificate (a one-element list nested twice).
output "not_before_time" {
  description = "When the certificate becomes valid (RFC 3339)"
  value       = try(google_privateca_certificate.this.certificate_description[0].subject_description[0].not_before_time, "")
}

output "not_after_time" {
  description = "When the certificate expires (RFC 3339) -- the date a successor must be installed by"
  value       = try(google_privateca_certificate.this.certificate_description[0].subject_description[0].not_after_time, "")
}
