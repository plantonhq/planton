output "group_id" {
  description = "The Cloudflare-assigned ID of the Access group (reference it from a policy or another group)"
  value       = cloudflare_zero_trust_access_group.main.id
}
