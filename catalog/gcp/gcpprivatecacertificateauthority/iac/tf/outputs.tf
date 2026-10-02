output "name" {
  description = "Full resource name of the authority (projects/{project}/locations/{location}/caPools/{pool}/certificateAuthorities/{id}) -- what subordinates and certificates reference"
  value       = google_privateca_certificate_authority.this.name
}

output "certificate_authority_id" {
  description = "The authority's ID"
  value       = google_privateca_certificate_authority.this.certificate_authority_id
}

output "state" {
  description = "The authority's state: ENABLED, DISABLED, STAGED, or AWAITING_USER_ACTIVATION"
  value       = google_privateca_certificate_authority.this.state
}

output "pem_ca_certificate" {
  description = "The authority's own CA certificate, PEM"
  value       = try(google_privateca_certificate_authority.this.pem_ca_certificates[0], "")
}

output "pem_ca_certificates" {
  description = "The authority's certificate chain, PEM, its own certificate first and the root last"
  value       = google_privateca_certificate_authority.this.pem_ca_certificates
}

output "ca_certificate_access_url" {
  description = "Where Google publishes the CA certificate"
  value       = try(google_privateca_certificate_authority.this.access_urls[0].ca_certificate_access_url, "")
}

output "crl_access_urls" {
  description = "Where Google publishes the authority's CRLs"
  value       = try(google_privateca_certificate_authority.this.access_urls[0].crl_access_urls, [])
}
