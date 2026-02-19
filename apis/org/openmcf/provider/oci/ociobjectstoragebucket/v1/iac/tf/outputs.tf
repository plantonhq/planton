output "bucket_id" {
  description = "OCID of the Object Storage bucket"
  value       = oci_objectstorage_bucket.this.bucket_id
}
