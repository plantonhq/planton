# The policy's full resource name, {parent}/policies/{constraint}. Google
# reports a project scope as the project NUMBER even when the manifest gave
# the ID.
output "name" {
  description = "The policy's full resource name"
  value       = google_org_policy_policy.this.name
}

# Opaque version marker Google changes on every update.
output "etag" {
  description = "The policy's etag"
  value       = google_org_policy_policy.this.etag
}
