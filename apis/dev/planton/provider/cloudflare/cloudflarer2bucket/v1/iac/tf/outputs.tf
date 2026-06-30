output "bucket_name" {
  description = "The name of the R2 bucket"
  value       = cloudflare_r2_bucket.main.name
}

output "bucket_url" {
  description = "The S3-compatible API URL for the bucket"
  value       = local.bucket_url
}

output "custom_domain_urls" {
  description = "URLs of the configured custom domains (one per enabled custom domain)"
  value       = [for domain, cd in cloudflare_r2_custom_domain.main : "https://${cd.domain}"]
}

output "public_url" {
  description = "The Cloudflare-managed r2.dev public URL, when public access is enabled"
  value       = local.public_access ? "https://${cloudflare_r2_managed_domain.main[0].domain}" : ""
}
