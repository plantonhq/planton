# Local values for the StripeProduct module.
#
# An optional value the spec leaves at its zero value renders as null, so the provider never
# sends it and Stripe keeps its own default. type arrives as its enum value name, which is
# Stripe's own string (good, service). features arrive as resolved feature ids: a reference to a
# StripeEntitlementFeature is resolved to its id before this module runs.
locals {
  description          = try(var.spec.description, "") != "" ? var.spec.description : null
  type                 = try(var.spec.type, "") != "" ? var.spec.type : null
  images               = length(try(var.spec.images, [])) > 0 ? var.spec.images : null
  shippable            = try(var.spec.shippable, null)
  statement_descriptor = try(var.spec.statement_descriptor, "") != "" ? var.spec.statement_descriptor : null
  tax_code             = try(var.spec.tax_code, "") != "" ? var.spec.tax_code : null
  unit_label           = try(var.spec.unit_label, "") != "" ? var.spec.unit_label : null
  url                  = try(var.spec.url, "") != "" ? var.spec.url : null
  metadata             = length(try(var.spec.metadata, {})) > 0 ? var.spec.metadata : null
  marketing_features   = try(var.spec.marketing_features, [])
  package_dimensions   = try(var.spec.package_dimensions, null)
  # The spec's default is active; the provider needs the value only to archive.
  active      = try(var.spec.active, null) == null ? true : var.spec.active
  feature_ids = { for id in try(var.spec.features, []) : id => id }
}
