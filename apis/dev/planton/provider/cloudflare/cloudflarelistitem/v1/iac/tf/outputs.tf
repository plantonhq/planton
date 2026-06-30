output "item_id" {
  description = "The Cloudflare-assigned identifier of the list item"
  value       = cloudflare_list_item.main.id
}

output "list_id" {
  description = "The list ID the entry was written to"
  value       = cloudflare_list_item.main.list_id
}
