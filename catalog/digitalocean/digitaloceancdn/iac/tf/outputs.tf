# Outputs — exactly the DigitalOceanCdnOutputs contract,
# identical across both provisioners.

output "cdn_id" {
  description = "UUID of the CDN endpoint (its API identity and import id)"
  value       = digitalocean_cdn.cdn.id
}

output "endpoint" {
  description = "The fully-qualified domain name the CDN serves content from"
  value       = digitalocean_cdn.cdn.endpoint
}
