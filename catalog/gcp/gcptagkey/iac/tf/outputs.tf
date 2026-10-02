# The key's resource name, tagKeys/{numeric_id} -- what a GcpTagValue's
# tag_key references and what a create-time tags map uses as its key.
output "name" {
  description = "The tag key's resource name (tagKeys/{id})"
  value       = google_tags_tag_key.this.name
}

# {org_or_project}/{short_name} -- the form resource.matchTag tests.
output "namespaced_name" {
  description = "The tag key's namespaced name"
  value       = google_tags_tag_key.this.namespaced_name
}

# The bare numeric id, derived from the name so both engines export the same
# three handles.
output "tag_key_id" {
  description = "The tag key's numeric ID"
  value       = trimprefix(google_tags_tag_key.this.name, "tagKeys/")
}

output "create_time" {
  description = "When the key was created (RFC 3339)"
  value       = google_tags_tag_key.this.create_time
}
