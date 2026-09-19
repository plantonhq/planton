# The project whose number the binding needs, read only when the manifest
# names the project by ID or names no parent at all (see
# locals.needs_project_lookup). A project ID is passed to the lookup; an
# empty parent reads the provider's default project.
data "google_project" "this" {
  count      = local.needs_project_lookup ? 1 : 0
  project_id = local.parent_project_id != "" ? local.parent_project_id : null
}

# The global tag binding: organizations, folders, projects, and any other
# global resource. Every input is immutable, so any change replaces the
# binding. deletion_policy is sent only when set so the provider's default
# (DELETE) stays the provider's.
resource "google_tags_tag_binding" "this" {
  count = local.is_location_scoped ? 0 : 1

  parent    = local.parent
  tag_value = var.spec.tag_value

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}

# The location-scoped tag binding: a regional or zonal resource, served from
# the region's or zone's endpoint -- selected by spec.location.
resource "google_tags_location_tag_binding" "this" {
  count = local.is_location_scoped ? 1 : 0

  parent    = local.parent
  tag_value = var.spec.tag_value
  location  = var.spec.location

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
