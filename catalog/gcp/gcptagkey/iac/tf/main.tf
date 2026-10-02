# The Resource Manager tag key.
#
# The owner, short name, purpose, and purpose data are immutable -- a change
# to any recreates the key, which Google refuses while the key has values --
# while the description and allowed-values regex update in place. Optional
# inputs are sent only when set so the provider's defaults stay the
# provider's; deletion_policy likewise.
resource "google_tags_tag_key" "this" {
  parent     = local.parent
  short_name = local.short_name

  description          = var.spec.description != "" ? var.spec.description : null
  purpose              = var.spec.purpose != "" ? var.spec.purpose : null
  purpose_data         = length(var.spec.purpose_data) > 0 ? var.spec.purpose_data : null
  allowed_values_regex = var.spec.allowed_values_regex != "" ? var.spec.allowed_values_regex : null
  deletion_policy      = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
