# Local values for the StripeRadarValueList module.
#
# item_type arrives as its enum value name, which is Stripe's own string (country, email, ...);
# unset renders as null and Stripe uses string. An empty metadata map renders as null, so the
# provider never sends it.
locals {
  item_type = try(var.spec.item_type, "") != "" ? var.spec.item_type : null
  metadata  = length(try(var.spec.metadata, {})) > 0 ? var.spec.metadata : null

  # Stripe stores the values of string, email and country lists in lower case ("KP" is stored as
  # "kp", "Fraud@Example.COM" as "fraud@example.com"); case_sensitive_string lists, ids,
  # fingerprints, card BINs and IP addresses keep the case they were sent in (verified live). An
  # unset item_type is string. The module sends each value the way Stripe will store it, so the
  # provider's read after create matches what it wrote; the map stays keyed by the value as
  # declared, which is how the item_ids output and the import recipe address each item.
  case_folded = contains(["string", "email", "country"], coalesce(local.item_type, "string"))
  items       = { for value in try(var.spec.items, []) : value => local.case_folded ? lower(value) : value }
}
