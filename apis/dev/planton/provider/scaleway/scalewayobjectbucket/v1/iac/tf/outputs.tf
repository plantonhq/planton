# ── Core Outputs ─────────────────────────────────────────────────────────────

output "bucket_id" {
  description = "The unique identifier of the bucket (format: region/bucket-name)"
  value       = scaleway_object_bucket.bucket.id
}

output "endpoint" {
  description = "The FQDN endpoint URL of the bucket (e.g., bucket-name.s3.fr-par.scw.cloud)"
  value       = scaleway_object_bucket.bucket.endpoint
}

output "api_endpoint" {
  description = "The S3 API endpoint URL for the bucket's region (e.g., https://s3.fr-par.scw.cloud)"
  value       = scaleway_object_bucket.bucket.api_endpoint
}

output "bucket_name" {
  description = "The name of the bucket as it exists in Scaleway Object Storage"
  value       = scaleway_object_bucket.bucket.name
}

output "region" {
  description = "The region where the bucket is deployed"
  value       = local.region
}
