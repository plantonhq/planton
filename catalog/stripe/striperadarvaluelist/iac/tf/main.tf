# StripeRadarValueList Main Resources
#
# stripe_radar_value_list is a Radar list rules reference by alias. item_type is sent only at
# create, so changing it replaces the list (and, through value_list, every item); alias, name and
# metadata update in place.
#
# Destroy deletes the list. Stripe refuses while a Radar rule still uses the list, and rules are
# outside this module. The provider has no handling for a list it cannot read: the next refresh
# fails, and the recovery is `tofu state rm` followed by an apply.
resource "stripe_radar_value_list" "this" {
  alias     = var.spec.alias
  name      = var.spec.name
  item_type = local.item_type
  metadata  = local.metadata
}

# stripe_radar_value_list_item is one value in the list. An item cannot be updated: both fields
# force replacement, and destroy deletes it. The for_each key is the value itself, so adding or
# removing one value touches only that item, and the item_ids output (and an import recipe) can
# address each item by its value.
resource "stripe_radar_value_list_item" "this" {
  for_each = local.items

  value_list = stripe_radar_value_list.this.id
  value      = each.value
}
