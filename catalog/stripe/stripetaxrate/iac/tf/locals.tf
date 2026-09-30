# Local values for the StripeTaxRate module.
#
# An optional value the spec leaves at its zero value renders as null, so the provider never
# sends it and Stripe keeps its own default. tax_type arrives as its value name, which is Stripe's
# own string (vat, sales_tax). percentage and inclusive are required by the provider, and their
# zero values (a 0% rate, an exclusive rate) are real choices the manifest may leave out.
locals {
  percentage   = try(var.spec.percentage, 0)
  inclusive    = try(var.spec.inclusive, false)
  country      = try(var.spec.country, "") != "" ? var.spec.country : null
  state        = try(var.spec.state, "") != "" ? var.spec.state : null
  jurisdiction = try(var.spec.jurisdiction, "") != "" ? var.spec.jurisdiction : null
  description  = try(var.spec.description, "") != "" ? var.spec.description : null
  tax_type     = try(var.spec.tax_type, "") != "" ? var.spec.tax_type : null
  metadata     = length(try(var.spec.metadata, {})) > 0 ? var.spec.metadata : null
  # The spec's default is active; the provider needs the value only to deactivate.
  active = try(var.spec.active, null) == null ? true : var.spec.active
}
