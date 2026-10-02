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
