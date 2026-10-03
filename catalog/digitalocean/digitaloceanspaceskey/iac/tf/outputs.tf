# Outputs — exactly the DigitalOceanSpacesKeyOutputs contract,
# identical across both provisioners.

output "access_key" {
  description = "The access key ID (the resource's API identity); pairs with secret_key as S3-style credentials"
  value       = digitalocean_spaces_key.key.access_key
}

output "secret_key" {
  description = "The secret access key -- returned ONLY at creation, never retrievable again"
  value       = digitalocean_spaces_key.key.secret_key
  sensitive   = true
}
