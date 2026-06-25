output "sitekey" {
  description = "The public site key embedded in the page frontend"
  value       = cloudflare_turnstile_widget.main.sitekey
}

output "secret" {
  description = "The secret key used server-side for /siteverify (sensitive)"
  value       = cloudflare_turnstile_widget.main.secret
  sensitive   = true
}

output "created_on" {
  description = "RFC3339 timestamp of when the widget was created"
  value       = cloudflare_turnstile_widget.main.created_on
}

output "modified_on" {
  description = "RFC3339 timestamp of when the widget was last modified"
  value       = cloudflare_turnstile_widget.main.modified_on
}
