output "site_tag" {
  description = "The Cloudflare-assigned site tag (the site's identity in every RUM API path)"
  # The resource's id IS the site tag (the id attribute is a copy of the
  # tag) and is the one value the create response does return.
  value = cloudflare_web_analytics_site.main.id
}

output "site_token" {
  description = "The site's measurement token, embedded by the JavaScript beacon"
  # From the read-after-create data source, never the resource: the create
  # response omits everything but the tag (see main.tf).
  value = data.cloudflare_web_analytics_site.main.site_token
}

output "snippet" {
  description = "The ready-to-embed JavaScript snippet (carries the site token)"
  value       = data.cloudflare_web_analytics_site.main.snippet
}

output "ruleset_id" {
  description = "The site's ruleset ID (the parent object the include/exclude rules live under; empty for host-identified sites, which have no ruleset)"
  value       = try(data.cloudflare_web_analytics_site.main.ruleset.id, "")
}
