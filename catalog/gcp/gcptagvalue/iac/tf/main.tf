# The Resource Manager tag value under its key.
#
# The key (the provider's parent, tagKeys/{id} -- the GcpTagKey name output)
# and the short name are immutable; only the description updates in place.
# Optional inputs are sent only when set so the provider's defaults stay the
# provider's; deletion_policy likewise.
resource "google_tags_tag_value" "this" {
  parent     = var.spec.tag_key
  short_name = local.short_name

  description     = var.spec.description != "" ? var.spec.description : null
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
