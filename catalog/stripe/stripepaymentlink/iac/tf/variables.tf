variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "StripePaymentLink specification"
  type = object({
    line_items = list(object({
      price    = string
      quantity = optional(number, 0)
      adjustable_quantity = optional(object({
        enabled = optional(bool, false)
        minimum = optional(number)
        maximum = optional(number)
      }))
    }))
    optional_items = optional(list(object({
      price    = string
      quantity = optional(number, 0)
      adjustable_quantity = optional(object({
        enabled = optional(bool, false)
        minimum = optional(number)
        maximum = optional(number)
      }))
    })), [])
    active           = optional(bool)
    inactive_message = optional(string, "")
    after_completion = optional(object({
      hosted_confirmation = optional(object({
        custom_message = optional(string, "")
      }))
      redirect = optional(object({
        url = string
      }))
    }))
    allow_promotion_codes = optional(bool)
    automatic_tax = optional(object({
      enabled = optional(bool, false)
      liability = optional(object({
        account = optional(string, "")
      }))
    }))
    billing_address_collection = optional(string, "")
    consent_collection = optional(object({
      payment_method_reuse_agreement = optional(object({
        position = optional(string, "")
      }))
      promotions       = optional(string, "")
      terms_of_service = optional(string, "")
    }))
    currency = optional(string, "")
    custom_fields = optional(list(object({
      key      = string
      label    = string
      optional = optional(bool)
      dropdown = optional(object({
        options = list(object({
          label = string
          value = string
        }))
        default_value = optional(string, "")
      }))
      numeric = optional(object({
        default_value  = optional(string, "")
        minimum_length = optional(number)
        maximum_length = optional(number)
      }))
      text = optional(object({
        default_value  = optional(string, "")
        minimum_length = optional(number)
        maximum_length = optional(number)
      }))
    })), [])
    custom_text = optional(object({
      after_submit = optional(object({
        message = string
      }))
      shipping_address = optional(object({
        message = string
      }))
      submit = optional(object({
        message = string
      }))
      terms_of_service_acceptance = optional(object({
        message = string
      }))
    }))
    customer_creation = optional(string, "")
    invoice_creation = optional(object({
      enabled = optional(bool, false)
      invoice_data = optional(object({
        account_tax_ids = optional(list(string), [])
        custom_fields = optional(list(object({
          name  = string
          value = string
        })), [])
        description = optional(string, "")
        footer      = optional(string, "")
        issuer = optional(object({
          account = optional(string, "")
        }))
        metadata = optional(map(string), {})
        rendering_options = optional(object({
          amount_tax_display = optional(string, "")
          template           = optional(string, "")
        }))
      }))
    }))
    managed_payments = optional(object({
      enabled = optional(bool)
    }))
    name_collection = optional(object({
      business = optional(object({
        enabled  = optional(bool, false)
        optional = optional(bool)
      }))
      individual = optional(object({
        enabled  = optional(bool, false)
        optional = optional(bool)
      }))
    }))
    payment_intent_data = optional(object({
      capture_method              = optional(string, "")
      description                 = optional(string, "")
      metadata                    = optional(map(string), {})
      setup_future_usage          = optional(string, "")
      statement_descriptor        = optional(string, "")
      statement_descriptor_suffix = optional(string, "")
      transfer_group              = optional(string, "")
    }))
    payment_method_collection = optional(string, "")
    payment_method_options = optional(object({
      card = optional(object({
        restrictions = optional(object({
          brands_blocked = optional(list(string), [])
        }))
      }))
    }))
    payment_method_types = optional(list(string), [])
    phone_number_collection = optional(object({
      enabled = optional(bool, false)
    }))
    restrictions = optional(object({
      completed_sessions = object({
        limit = optional(number, 0)
      })
    }))
    shipping_address_collection = optional(object({
      allowed_countries = list(string)
    }))
    shipping_options = optional(list(object({
      shipping_rate = string
    })), [])
    submit_type = optional(string, "")
    subscription_data = optional(object({
      description = optional(string, "")
      invoice_settings = optional(object({
        issuer = optional(object({
          account = optional(string, "")
        }))
      }))
      metadata          = optional(map(string), {})
      trial_period_days = optional(number)
      trial_settings = optional(object({
        end_behavior = object({
          missing_payment_method = optional(string, "")
        })
      }))
    }))
    tax_id_collection = optional(object({
      enabled  = optional(bool, false)
      required = optional(string, "")
    }))
    application_fee_amount  = optional(number)
    application_fee_percent = optional(number)
    on_behalf_of            = optional(string, "")
    transfer_data = optional(object({
      destination = string
      amount      = optional(number)
    }))
    metadata = optional(map(string), {})
  })
}
