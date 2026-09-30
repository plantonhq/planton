# Local values for the StripeRadarValueList module.
#
# item_type arrives as its enum value name, which is Stripe's own string (country, email, ...);
# unset renders as null and Stripe uses string. An empty metadata map renders as null, so the
# provider never sends it.
locals {
  item_type = try(var.spec.item_type, "") != "" ? var.spec.item_type : null
  metadata  = length(try(var.spec.metadata, {})) > 0 ? var.spec.metadata : null
  items     = { for value in try(var.spec.items, []) : value => value }
}
