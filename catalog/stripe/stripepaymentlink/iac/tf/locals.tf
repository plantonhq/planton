# Local values for the StripePaymentLink module.
#
# The provider nests everything as nested attributes, so each block below is one object the
# resource takes whole. An optional value the spec leaves at its zero value renders as null, so the
# provider never sends it and Stripe keeps its own default. Enums arrive as their value names,
# which are Stripe's own strings (required, if_required, subscribe). Every price and shipping rate
# arrives as a resolved id: a reference to a StripePrice or StripeShippingRate is resolved before
# this module runs.
#
# The spec models three of Stripe's "type" discriminators as the block that is set: after
# completion (hosted_confirmation or redirect), a custom field (dropdown, numeric or text), and an
# account reference (a connected account, or the link's own account when none is named). The
# module writes each type from that shape.
locals {
  inactive_message           = try(var.spec.inactive_message, "") != "" ? var.spec.inactive_message : null
  billing_address_collection = try(var.spec.billing_address_collection, "") != "" ? var.spec.billing_address_collection : null
  currency                   = try(var.spec.currency, "") != "" ? var.spec.currency : null
  customer_creation          = try(var.spec.customer_creation, "") != "" ? var.spec.customer_creation : null
  payment_method_collection  = try(var.spec.payment_method_collection, "") != "" ? var.spec.payment_method_collection : null
  payment_method_types       = length(try(var.spec.payment_method_types, [])) > 0 ? var.spec.payment_method_types : null
  submit_type                = try(var.spec.submit_type, "") != "" ? var.spec.submit_type : null
  on_behalf_of               = try(var.spec.on_behalf_of, "") != "" ? var.spec.on_behalf_of : null
  metadata                   = length(try(var.spec.metadata, {})) > 0 ? var.spec.metadata : null
  # The spec's default is active; the provider needs the value only to deactivate.
  active = try(var.spec.active, null) == null ? true : var.spec.active

  line_items = [
    for item in var.spec.line_items : {
      price    = item.price
      quantity = item.quantity
      adjustable_quantity = try(item.adjustable_quantity, null) == null ? null : {
        enabled = item.adjustable_quantity.enabled
        minimum = try(item.adjustable_quantity.minimum, null)
        maximum = try(item.adjustable_quantity.maximum, null)
      }
    }
  ]

  # The ordered prices the link sells, with their quantities. Stripe can't change a line item's
  # price on an existing link, and the provider never stores a line item's quantity (it is
  # write-only), so the tracker in main.tf holds both and replaces the link when either changes.
  line_items_tracked = [for item in var.spec.line_items : { price = item.price, quantity = item.quantity }]

  optional_items = length(try(var.spec.optional_items, [])) == 0 ? null : [
    for item in var.spec.optional_items : {
      price    = item.price
      quantity = item.quantity
      adjustable_quantity = try(item.adjustable_quantity, null) == null ? null : {
        enabled = item.adjustable_quantity.enabled
        minimum = try(item.adjustable_quantity.minimum, null)
        maximum = try(item.adjustable_quantity.maximum, null)
      }
    }
  ]

  after_completion_spec = try(var.spec.after_completion, null)
  after_completion = local.after_completion_spec == null ? null : {
    type = try(local.after_completion_spec.redirect, null) != null ? "redirect" : "hosted_confirmation"
    redirect = try(local.after_completion_spec.redirect, null) == null ? null : {
      url = local.after_completion_spec.redirect.url
    }
    hosted_confirmation = try(local.after_completion_spec.hosted_confirmation, null) == null ? null : {
      custom_message = try(local.after_completion_spec.hosted_confirmation.custom_message, "") != "" ? local.after_completion_spec.hosted_confirmation.custom_message : null
    }
  }

  automatic_tax = try(var.spec.automatic_tax, null) == null ? null : {
    enabled = var.spec.automatic_tax.enabled
    liability = try(var.spec.automatic_tax.liability, null) == null ? null : {
      type    = try(var.spec.automatic_tax.liability.account, "") != "" ? "account" : "self"
      account = try(var.spec.automatic_tax.liability.account, "") != "" ? var.spec.automatic_tax.liability.account : null
    }
  }

  consent_collection = try(var.spec.consent_collection, null) == null ? null : {
    payment_method_reuse_agreement = try(var.spec.consent_collection.payment_method_reuse_agreement, null) == null ? null : {
      position = var.spec.consent_collection.payment_method_reuse_agreement.position
    }
    promotions       = try(var.spec.consent_collection.promotions, "") != "" ? var.spec.consent_collection.promotions : null
    terms_of_service = try(var.spec.consent_collection.terms_of_service, "") != "" ? var.spec.consent_collection.terms_of_service : null
  }

  custom_fields = length(try(var.spec.custom_fields, [])) == 0 ? null : [
    for field in var.spec.custom_fields : {
      key      = field.key
      label    = { type = "custom", custom = field.label }
      type     = try(field.dropdown, null) != null ? "dropdown" : try(field.numeric, null) != null ? "numeric" : "text"
      optional = try(field.optional, null)
      dropdown = try(field.dropdown, null) == null ? null : {
        options       = [for option in field.dropdown.options : { label = option.label, value = option.value }]
        default_value = try(field.dropdown.default_value, "") != "" ? field.dropdown.default_value : null
      }
      numeric = try(field.numeric, null) == null ? null : {
        default_value  = try(field.numeric.default_value, "") != "" ? field.numeric.default_value : null
        minimum_length = try(field.numeric.minimum_length, null)
        maximum_length = try(field.numeric.maximum_length, null)
      }
      text = try(field.text, null) == null ? null : {
        default_value  = try(field.text.default_value, "") != "" ? field.text.default_value : null
        minimum_length = try(field.text.minimum_length, null)
        maximum_length = try(field.text.maximum_length, null)
      }
    }
  ]

  custom_text = try(var.spec.custom_text, null) == null ? null : {
    after_submit                = try(var.spec.custom_text.after_submit, null)
    shipping_address            = try(var.spec.custom_text.shipping_address, null)
    submit                      = try(var.spec.custom_text.submit, null)
    terms_of_service_acceptance = try(var.spec.custom_text.terms_of_service_acceptance, null)
  }

  invoice_data_spec = try(var.spec.invoice_creation.invoice_data, null)
  invoice_creation = try(var.spec.invoice_creation, null) == null ? null : {
    enabled = var.spec.invoice_creation.enabled
    invoice_data = local.invoice_data_spec == null ? null : {
      account_tax_ids = length(try(local.invoice_data_spec.account_tax_ids, [])) > 0 ? local.invoice_data_spec.account_tax_ids : null
      custom_fields   = length(try(local.invoice_data_spec.custom_fields, [])) > 0 ? [for f in local.invoice_data_spec.custom_fields : { name = f.name, value = f.value }] : null
      description     = try(local.invoice_data_spec.description, "") != "" ? local.invoice_data_spec.description : null
      footer          = try(local.invoice_data_spec.footer, "") != "" ? local.invoice_data_spec.footer : null
      metadata        = length(try(local.invoice_data_spec.metadata, {})) > 0 ? local.invoice_data_spec.metadata : null
      issuer = try(local.invoice_data_spec.issuer, null) == null ? null : {
        type    = try(local.invoice_data_spec.issuer.account, "") != "" ? "account" : "self"
        account = try(local.invoice_data_spec.issuer.account, "") != "" ? local.invoice_data_spec.issuer.account : null
      }
      rendering_options = try(local.invoice_data_spec.rendering_options, null) == null ? null : {
        amount_tax_display = try(local.invoice_data_spec.rendering_options.amount_tax_display, "") != "" ? local.invoice_data_spec.rendering_options.amount_tax_display : null
        template           = try(local.invoice_data_spec.rendering_options.template, "") != "" ? local.invoice_data_spec.rendering_options.template : null
      }
    }
  }

  managed_payments = try(var.spec.managed_payments, null) == null ? null : {
    enabled = try(var.spec.managed_payments.enabled, null)
  }

  name_collection = try(var.spec.name_collection, null) == null ? null : {
    business = try(var.spec.name_collection.business, null) == null ? null : {
      enabled  = var.spec.name_collection.business.enabled
      optional = try(var.spec.name_collection.business.optional, null)
    }
    individual = try(var.spec.name_collection.individual, null) == null ? null : {
      enabled  = var.spec.name_collection.individual.enabled
      optional = try(var.spec.name_collection.individual.optional, null)
    }
  }

  payment_intent_data = try(var.spec.payment_intent_data, null) == null ? null : {
    capture_method              = try(var.spec.payment_intent_data.capture_method, "") != "" ? var.spec.payment_intent_data.capture_method : null
    description                 = try(var.spec.payment_intent_data.description, "") != "" ? var.spec.payment_intent_data.description : null
    metadata                    = length(try(var.spec.payment_intent_data.metadata, {})) > 0 ? var.spec.payment_intent_data.metadata : null
    setup_future_usage          = try(var.spec.payment_intent_data.setup_future_usage, "") != "" ? var.spec.payment_intent_data.setup_future_usage : null
    statement_descriptor        = try(var.spec.payment_intent_data.statement_descriptor, "") != "" ? var.spec.payment_intent_data.statement_descriptor : null
    statement_descriptor_suffix = try(var.spec.payment_intent_data.statement_descriptor_suffix, "") != "" ? var.spec.payment_intent_data.statement_descriptor_suffix : null
    transfer_group              = try(var.spec.payment_intent_data.transfer_group, "") != "" ? var.spec.payment_intent_data.transfer_group : null
  }

  brands_blocked = try(var.spec.payment_method_options.card.restrictions.brands_blocked, [])
  payment_method_options = try(var.spec.payment_method_options, null) == null ? null : {
    card = try(var.spec.payment_method_options.card, null) == null ? null : {
      restrictions = try(var.spec.payment_method_options.card.restrictions, null) == null ? null : {
        brands_blocked = length(local.brands_blocked) > 0 ? local.brands_blocked : null
      }
    }
  }

  phone_number_collection = try(var.spec.phone_number_collection, null) == null ? null : {
    enabled = var.spec.phone_number_collection.enabled
  }

  restrictions = try(var.spec.restrictions, null) == null ? null : {
    completed_sessions = { limit = var.spec.restrictions.completed_sessions.limit }
  }

  shipping_address_collection = try(var.spec.shipping_address_collection, null) == null ? null : {
    allowed_countries = var.spec.shipping_address_collection.allowed_countries
  }

  shipping_options = length(try(var.spec.shipping_options, [])) == 0 ? null : [
    for option in var.spec.shipping_options : { shipping_rate = option.shipping_rate }
  ]

  subscription_data_spec = try(var.spec.subscription_data, null)
  subscription_data = local.subscription_data_spec == null ? null : {
    description       = try(local.subscription_data_spec.description, "") != "" ? local.subscription_data_spec.description : null
    metadata          = length(try(local.subscription_data_spec.metadata, {})) > 0 ? local.subscription_data_spec.metadata : null
    trial_period_days = try(local.subscription_data_spec.trial_period_days, null)
    invoice_settings = try(local.subscription_data_spec.invoice_settings, null) == null ? null : {
      issuer = try(local.subscription_data_spec.invoice_settings.issuer, null) == null ? null : {
        type    = try(local.subscription_data_spec.invoice_settings.issuer.account, "") != "" ? "account" : "self"
        account = try(local.subscription_data_spec.invoice_settings.issuer.account, "") != "" ? local.subscription_data_spec.invoice_settings.issuer.account : null
      }
    }
    trial_settings = try(local.subscription_data_spec.trial_settings, null) == null ? null : {
      end_behavior = {
        missing_payment_method = local.subscription_data_spec.trial_settings.end_behavior.missing_payment_method
      }
    }
  }

  tax_id_collection = try(var.spec.tax_id_collection, null) == null ? null : {
    enabled  = var.spec.tax_id_collection.enabled
    required = try(var.spec.tax_id_collection.required, "") != "" ? var.spec.tax_id_collection.required : null
  }

  transfer_data = try(var.spec.transfer_data, null) == null ? null : {
    destination = var.spec.transfer_data.destination
    amount      = try(var.spec.transfer_data.amount, null)
  }
}
