# DigitalOcean Spaces Bucket Outputs
# Matches stack_outputs.proto fields

output "bucket_id" {
  description = "The unique ID of the Spaces bucket (format: region:bucket-name)"
  value       = digitalocean_spaces_bucket.main.id
}

output "endpoint" {
  description = "The FQDN endpoint for the Spaces bucket"
  value       = digitalocean_spaces_bucket.main.bucket_domain_name
}

output "bucket_name" {
  description = "The name of the bucket"
  value       = digitalocean_spaces_bucket.main.name
}

output "region" {
  description = "The region where the bucket is deployed"
  value       = digitalocean_spaces_bucket.main.region
}

output "urn" {
  description = "The uniform resource name (URN) of the bucket"
  value       = digitalocean_spaces_bucket.main.urn
}

output "bucket_domain_name" {
  description = "The FQDN of the bucket"
  value       = digitalocean_spaces_bucket.main.bucket_domain_name
}

