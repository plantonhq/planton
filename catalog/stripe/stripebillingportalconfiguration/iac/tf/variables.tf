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
  description = "StripeBillingPortalConfiguration specification"
  type = object({
    name = optional(string, "")
    features = object({
      customer_update = optional(object({
        enabled         = optional(bool, false)
        allowed_updates = optional(list(string), [])
      }))
      invoice_history = optional(object({
        enabled = optional(bool, false)
      }))
      payment_method_update = optional(object({
        enabled                      = optional(bool, false)
        payment_method_configuration = optional(string, "")
      }))
      subscription_cancel = optional(object({
        enabled            = optional(bool, false)
        mode               = optional(string, "")
        proration_behavior = optional(string, "")
        cancellation_reason = optional(object({
          enabled = optional(bool, false)
          options = list(string)
        }))
      }))
      subscription_update = optional(object({
        enabled                 = optional(bool, false)
        default_allowed_updates = optional(list(string), [])
        proration_behavior      = optional(string, "")
        billing_cycle_anchor    = optional(string, "")
        trial_update_behavior   = optional(string, "")
        products = optional(list(object({
          product = string
          prices  = list(string)
          adjustable_quantity = optional(object({
            enabled = optional(bool, false)
            minimum = optional(number)
            maximum = optional(number)
          }))
        })), [])
        schedule_at_period_end = optional(object({
          conditions = optional(list(object({
            type = optional(string, "")
          })), [])
        }))
      }))
    })
    business_profile = optional(object({
      headline             = optional(string, "")
      privacy_policy_url   = optional(string, "")
      terms_of_service_url = optional(string, "")
    }))
    default_return_url = optional(string, "")
    login_page = optional(object({
      enabled = optional(bool, false)
    }))
    active   = optional(bool)
    metadata = optional(map(string), {})
  })
}
