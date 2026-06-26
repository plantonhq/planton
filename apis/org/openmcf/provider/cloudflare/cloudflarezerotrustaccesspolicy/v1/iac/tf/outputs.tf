output "policy_id" {
  description = "The Cloudflare-assigned ID of the Access policy (reference it from an application's policies list)"
  value       = cloudflare_zero_trust_access_policy.main.id
}
