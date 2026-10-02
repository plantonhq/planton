# The value's resource name, tagValues/{numeric_id} -- what a GcpTagBinding's
# tag_value references and what a create-time tags map uses as its value.
output "name" {
  description = "The tag value's resource name (tagValues/{id})"
  value       = google_tags_tag_value.this.name
}

# {parent}/{key_short_name}/{short_name} -- the human-readable handle.
output "namespaced_name" {
  description = "The tag value's namespaced name"
  value       = google_tags_tag_value.this.namespaced_name
}

# The bare numeric id, derived from the name so both engines export the same
# three handles.
output "tag_value_id" {
  description = "The tag value's numeric ID"
  value       = trimprefix(google_tags_tag_value.this.name, "tagValues/")
}

output "create_time" {
  description = "When the value was created (RFC 3339)"
  value       = google_tags_tag_value.this.create_time
}
