output "list_id" {
  description = "The Cloudflare-assigned identifier of the list"
  value       = cloudflare_list.main.id
}

output "name" {
  description = "The list name (used in rule expressions)"
  value       = cloudflare_list.main.name
}

output "kind" {
  description = "The list kind (ip, redirect, hostname, or asn)"
  value       = cloudflare_list.main.kind
}
