# StripeProduct Main Resources
#
# stripe_product is one thing the account sells. type is sent only at create, so changing it
# replaces the product; every other field updates in place, and a value removed from the spec is
# left as it is in Stripe (the provider sends only values that are set). default_price_data is
# never set: the price it would create is invisible to this module and never archived -- prices
# are StripePrice resources that name this product.
#
# Destroy archives the product (active = false); Stripe keeps it. The provider has no handling for
# a product it cannot read: the next refresh fails, and the recovery is `tofu state rm` followed
# by an apply, which creates a new product.
resource "stripe_product" "this" {
  name                 = var.spec.name
  description          = local.description
  active               = local.active
  type                 = local.type
  images               = local.images
  shippable            = local.shippable
  statement_descriptor = local.statement_descriptor
  tax_code             = local.tax_code
  unit_label           = local.unit_label
  url                  = local.url
  metadata             = local.metadata

  dynamic "marketing_features" {
    for_each = local.marketing_features
    content {
      name = marketing_features.value.name
    }
  }

  dynamic "package_dimensions" {
    for_each = local.package_dimensions == null ? [] : [local.package_dimensions]
    content {
      height = package_dimensions.value.height
      length = package_dimensions.value.length
      weight = package_dimensions.value.weight
      width  = package_dimensions.value.width
    }
  }
}

# stripe_product_feature attaches one entitlement feature to the product. A link cannot be
# updated: both fields force replacement, and destroy deletes the link, which removes the
# feature from what the product grants. The for_each key is the feature id itself, so the
# product_feature_ids output (and an import recipe) can address each link by the feature it
# grants.
resource "stripe_product_feature" "this" {
  for_each = local.feature_ids

  product             = stripe_product.this.id
  entitlement_feature = each.value
}
