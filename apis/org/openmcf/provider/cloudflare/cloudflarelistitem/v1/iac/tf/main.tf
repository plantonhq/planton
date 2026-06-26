# A single entry in a Cloudflare List. Exactly one of ip / asn / hostname /
# redirect is set, matching the parent list's kind.
resource "cloudflare_list_item" "main" {
  account_id = var.spec.account_id
  list_id    = var.spec.list_id
  comment    = try(var.spec.comment, "") != "" ? var.spec.comment : null

  ip  = try(var.spec.ip, null)
  asn = try(var.spec.asn, null)

  hostname = try(var.spec.hostname, null) != null ? {
    url_hostname           = var.spec.hostname.url_hostname
    exclude_exact_hostname = try(var.spec.hostname.exclude_exact_hostname, null)
  } : null

  redirect = try(var.spec.redirect, null) != null ? {
    source_url            = var.spec.redirect.source_url
    target_url            = var.spec.redirect.target_url
    status_code           = try(var.spec.redirect.status_code, 0) != 0 ? var.spec.redirect.status_code : null
    include_subdomains    = try(var.spec.redirect.include_subdomains, null)
    preserve_path_suffix  = try(var.spec.redirect.preserve_path_suffix, null)
    preserve_query_string = try(var.spec.redirect.preserve_query_string, null)
    subpath_matching      = try(var.spec.redirect.subpath_matching, null)
  } : null
}
