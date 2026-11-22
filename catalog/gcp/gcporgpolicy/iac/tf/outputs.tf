# The policy's full resource name, {parent}/policies/{constraint}. It is the
# resource id: the provider keeps only the bare constraint in its name
# attribute after the first read, so an output of name would change on every
# re-plan (measured on Google provider 8.7).
output "name" {
  description = "The policy's full resource name"
  value       = google_org_policy_policy.this.id
}

# Opaque version marker Google changes on every update. After an update this
# module applies, it lags the live etag until the next refresh (measured), so
# read it from a fresh plan, never right after an apply.
output "etag" {
  description = "The policy's etag"
  value       = google_org_policy_policy.this.etag
}
