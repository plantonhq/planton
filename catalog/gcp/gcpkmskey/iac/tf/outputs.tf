# The resource ID is the fully qualified crypto key path
# (projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{name})
# — the exact string every CMEK consumer (BigQuery, Spanner, GKE, Cloud SQL,
# Pub/Sub, ...) passes to its kms_key_name-style attribute.
output "key_id" {
  description = "Fully qualified crypto key resource path"
  value       = google_kms_crypto_key.this.id
}

output "key_name" {
  description = "The short name of the crypto key"
  value       = google_kms_crypto_key.this.name
}

# GCP populates primary only for ENCRYPT_DECRYPT keys; the outputs are
# empty for asymmetric/raw/MAC keys and for keys created without an
# initial version.
output "primary_version_name" {
  description = "Fully qualified resource name of the current primary CryptoKeyVersion"
  value       = try(google_kms_crypto_key.this.primary[0].name, "")
}

output "primary_state" {
  description = "Lifecycle state of the current primary CryptoKeyVersion"
  value       = try(google_kms_crypto_key.this.primary[0].state, "")
}

# Version 1 is the one Google creates with the key, whatever its purpose;
# asymmetric-sign keys have no primary, so this is the version their
# consumers name. Empty when no initial version was created.
output "initial_version_name" {
  description = "Fully qualified resource name of the version created with the key (version 1)"
  value       = var.spec.skip_initial_version_creation ? "" : "${google_kms_crypto_key.this.id}/cryptoKeyVersions/1"
}
