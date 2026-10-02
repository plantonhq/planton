output "name" {
  description = "Full resource name of the key handle (projects/{project}/locations/{location}/keyHandles/{name})"
  value       = google_kms_key_handle.this.id
}

output "kms_key" {
  description = "The Cloud KMS key Autokey assigned to the handle"
  value       = google_kms_key_handle.this.kms_key
}
