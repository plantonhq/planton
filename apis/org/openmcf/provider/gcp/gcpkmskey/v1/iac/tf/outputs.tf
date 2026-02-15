output "key_id" {
  description = "Fully qualified crypto key resource path (projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys/{name})"
  value       = google_kms_crypto_key.this.id
}

output "key_name" {
  description = "The short name of the crypto key"
  value       = google_kms_crypto_key.this.name
}
