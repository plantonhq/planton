output "bucket_name" {
  description = "The name of the R2 bucket"
  value       = cloudflare_r2_bucket.main.name
}

output "bucket_url" {
  description = "The path-style S3 API URL for the bucket"
  value       = local.bucket_url
}

output "custom_domain_url" {
  description = "The custom domain URL if configured (e.g., https://media.example.com)"
  value       = local.custom_domain_enabled ? "https://${local.custom_domain_name}" : null
}
