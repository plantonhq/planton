# Local values for the StripeBillingPortalConfiguration module.
#
# The provider models features, business_profile and login_page as nested objects whose every
# member is optional and computed. A sub-feature the spec leaves out renders as null, and so does
# every optional scalar at its zero value ("" for strings and enums, [] for lists): the provider
# sends only what is set and Stripe keeps its own default for the rest. That same rule means a
# setting removed from the manifest is NOT reset in Stripe -- the provider never sends the
# removal -- which the spec's comments tell the manifest author.
#
# Enum fields arrive as their value names, which are Stripe's own strings (at_period_end,
# create_prorations, ...), so they pass through unchanged.
locals {
  features = var.spec.features

  customer_update = try(local.features.customer_update, null) == null ? null : {
    enabled         = try(local.features.customer_update.enabled, false) == true
    allowed_updates = length(try(local.features.customer_update.allowed_updates, [])) > 0 ? local.features.customer_update.allowed_updates : null
  }

  invoice_history = try(local.features.invoice_history, null) == null ? null : {
    enabled = try(local.features.invoice_history.enabled, false) == true
  }

  payment_method_update = try(local.features.payment_method_update, null) == null ? null : {
    enabled                      = try(local.features.payment_method_update.enabled, false) == true
    payment_method_configuration = try(local.features.payment_method_update.payment_method_configuration, "") != "" ? local.features.payment_method_update.payment_method_configuration : null
  }

  cancel = try(local.features.subscription_cancel, null)
  subscription_cancel = local.cancel == null ? null : {
    enabled            = try(local.cancel.enabled, false) == true
    mode               = try(local.cancel.mode, "") != "" ? local.cancel.mode : null
    proration_behavior = try(local.cancel.proration_behavior, "") != "" ? local.cancel.proration_behavior : null
    cancellation_reason = try(local.cancel.cancellation_reason, null) == null ? null : {
      enabled = try(local.cancel.cancellation_reason.enabled, false) == true
      options = local.cancel.cancellation_reason.options
    }
  }

  update = try(local.features.subscription_update, null)
  subscription_update = local.update == null ? null : {
    enabled                 = try(local.update.enabled, false) == true
    default_allowed_updates = length(try(local.update.default_allowed_updates, [])) > 0 ? local.update.default_allowed_updates : null
    proration_behavior      = try(local.update.proration_behavior, "") != "" ? local.update.proration_behavior : null
    billing_cycle_anchor    = try(local.update.billing_cycle_anchor, "") != "" ? local.update.billing_cycle_anchor : null
    trial_update_behavior   = try(local.update.trial_update_behavior, "") != "" ? local.update.trial_update_behavior : null
    products = length(try(local.update.products, [])) == 0 ? null : [
      for p in local.update.products : {
        product = p.product
        prices  = p.prices
        adjustable_quantity = try(p.adjustable_quantity, null) == null ? null : {
          enabled = try(p.adjustable_quantity.enabled, false) == true
          minimum = try(p.adjustable_quantity.minimum, null)
          maximum = try(p.adjustable_quantity.maximum, null)
        }
      }
    ]
    schedule_at_period_end = try(local.update.schedule_at_period_end, null) == null ? null : {
      conditions = [for c in try(local.update.schedule_at_period_end.conditions, []) : { type = c.type }]
    }
  }

  business_profile = try(var.spec.business_profile, null) == null ? null : {
    headline             = try(var.spec.business_profile.headline, "") != "" ? var.spec.business_profile.headline : null
    privacy_policy_url   = try(var.spec.business_profile.privacy_policy_url, "") != "" ? var.spec.business_profile.privacy_policy_url : null
    terms_of_service_url = try(var.spec.business_profile.terms_of_service_url, "") != "" ? var.spec.business_profile.terms_of_service_url : null
  }

  login_page = try(var.spec.login_page, null) == null ? null : {
    enabled = try(var.spec.login_page.enabled, false) == true
  }

  name               = try(var.spec.name, "") != "" ? var.spec.name : null
  default_return_url = try(var.spec.default_return_url, "") != "" ? var.spec.default_return_url : null
  metadata           = length(try(var.spec.metadata, {})) > 0 ? var.spec.metadata : null
  # The spec's default is active; the provider needs the value only to deactivate.
  active = try(var.spec.active, null) == null ? true : var.spec.active
}
