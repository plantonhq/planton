# An account-scoped Cloudflare List. Items are managed separately as
# CloudflareListItem resources, so this resource never declares inline `items`.
resource "cloudflare_list" "main" {
  account_id  = var.spec.account_id
  kind        = var.spec.kind
  name        = var.spec.name
  description = try(var.spec.description, "") != "" ? var.spec.description : null
}
