# Local values for the StripeTaxRegistration module.
#
# The provider nests a registration's options under the country they belong to
# (country_options.de, country_options.us, ...), one block per country with its own shape. The
# spec is flat, and this module writes the single block the declared country takes, filling only
# the parts the declared type uses: the place-of-supply scheme under standard, the province under
# province_standard, the jurisdiction under the local tax the type names, and the elections under
# state_sales_tax. The spec's validation already refuses a part the country or type does not
# take, so every part written here is one Stripe accepts.
#
# Enum values arrive as their value names, which are Stripe's own strings (oss_union,
# state_sales_tax, inbound_goods). A value the spec leaves at its zero value renders as null, so
# the provider never sends it.
locals {
  country_key = lower(var.spec.country)

  place_of_supply_scheme = try(var.spec.place_of_supply_scheme, "") != "" ? var.spec.place_of_supply_scheme : null
  province               = try(var.spec.province, "") != "" ? var.spec.province : null
  state                  = try(var.spec.state, "") != "" ? var.spec.state : null
  jurisdiction           = try(var.spec.jurisdiction, "") != "" ? var.spec.jurisdiction : null
  elections = [
    for election in try(var.spec.state_sales_tax_elections, []) : {
      type         = election.type
      jurisdiction = try(election.jurisdiction, "") != "" ? election.jurisdiction : null
    }
  ]

  # The one country block, built from the parts the type uses. merge() drops the parts a
  # registration does not take, so the block carries exactly the attributes its country defines.
  country_option = merge(
    { type = var.spec.type },
    local.place_of_supply_scheme == null ? {} : { standard = { place_of_supply_scheme = local.place_of_supply_scheme } },
    local.province == null ? {} : { province_standard = { province = local.province } },
    local.state == null ? {} : { state = local.state },
    var.spec.type == "local_amusement_tax" ? { local_amusement_tax = { jurisdiction = local.jurisdiction } } : {},
    var.spec.type == "local_lease_tax" ? { local_lease_tax = { jurisdiction = local.jurisdiction } } : {},
    length(local.elections) == 0 ? {} : { state_sales_tax = { elections = local.elections } },
  )

  # expires_at is optional; unset, the registration never expires.
  expires_at = try(var.spec.expires_at, null)
}
