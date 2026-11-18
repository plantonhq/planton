# The binding's resource name, tagBindings/{encoded parent}/{tagValues/id},
# from whichever of the two resources was created. The provider's `name`
# attribute arrives without the tagBindings/ prefix its documentation shows
# (measured on both resources' API), so the prefix is put back once.
output "name" {
  description = "The tag binding's resource name"
  value       = "tagBindings/${trimprefix(local.is_location_scoped ? google_tags_location_tag_binding.this[0].name : google_tags_tag_binding.this[0].name, "tagBindings/")}"
}

# The full resource name the tag is bound to, as sent -- the resolved project
# number when the manifest named the project by ID or not at all.
output "parent" {
  description = "The full resource name of the tagged resource"
  value       = local.is_location_scoped ? google_tags_location_tag_binding.this[0].parent : google_tags_tag_binding.this[0].parent
}

# The bound value's resource name, tagValues/{id}.
output "tag_value" {
  description = "The bound tag value's resource name"
  value       = local.is_location_scoped ? google_tags_location_tag_binding.this[0].tag_value : google_tags_tag_binding.this[0].tag_value
}
