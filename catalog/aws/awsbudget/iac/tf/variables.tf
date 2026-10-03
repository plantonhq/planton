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
  description = "AwsBudget specification"
  type = object({
    region           = string
    budget_name      = string
    budget_type      = optional(string, "")
    time_unit        = optional(string, "")
    account_id       = optional(string, "")
    billing_view_arn = optional(string, "")
    limit = optional(object({
      amount = optional(string, "")
      unit   = string
    }))
    planned_limits = optional(list(object({
      start_time = optional(string, "")
      amount     = optional(string, "")
      unit       = string
    })), [])
    auto_adjust = optional(object({
      auto_adjust_type         = optional(string, "")
      budget_adjustment_period = optional(number, 0)
    }))
    time_period_start = optional(string, "")
    time_period_end   = optional(string, "")
    cost_types = optional(object({
      include_credit             = optional(bool)
      include_discount           = optional(bool)
      include_other_subscription = optional(bool)
      include_recurring          = optional(bool)
      include_refund             = optional(bool)
      include_subscription       = optional(bool)
      include_support            = optional(bool)
      include_tax                = optional(bool)
      include_upfront            = optional(bool)
      use_amortized              = optional(bool)
      use_blended                = optional(bool)
    }))
    cost_filters = optional(list(object({
      name   = string
      values = list(string)
    })), [])
    metric = optional(string, "")
    filter_expression = optional(object({
      dimension = optional(object({
        key           = optional(string, "")
        match_options = optional(list(string), [])
        values        = list(string)
      }))
      tag = optional(object({
        key           = optional(string, "")
        match_options = optional(list(string), [])
        values        = optional(list(string), [])
      }))
      cost_category = optional(object({
        key           = optional(string, "")
        match_options = optional(list(string), [])
        values        = optional(list(string), [])
      }))
      and = optional(list(object({
        dimension = optional(object({
          key           = optional(string, "")
          match_options = optional(list(string), [])
          values        = list(string)
        }))
        tag = optional(object({
          key           = optional(string, "")
          match_options = optional(list(string), [])
          values        = optional(list(string), [])
        }))
        cost_category = optional(object({
          key           = optional(string, "")
          match_options = optional(list(string), [])
          values        = optional(list(string), [])
        }))
        and = optional(list(object({
          dimension = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = list(string)
          }))
          tag = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = optional(list(string), [])
          }))
          cost_category = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = optional(list(string), [])
          }))
        })), [])
        or = optional(list(object({
          dimension = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = list(string)
          }))
          tag = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = optional(list(string), [])
          }))
          cost_category = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = optional(list(string), [])
          }))
        })), [])
        not = optional(object({
          dimension = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = list(string)
          }))
          tag = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = optional(list(string), [])
          }))
          cost_category = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = optional(list(string), [])
          }))
        }))
      })), [])
      or = optional(list(object({
        dimension = optional(object({
          key           = optional(string, "")
          match_options = optional(list(string), [])
          values        = list(string)
        }))
        tag = optional(object({
          key           = optional(string, "")
          match_options = optional(list(string), [])
          values        = optional(list(string), [])
        }))
        cost_category = optional(object({
          key           = optional(string, "")
          match_options = optional(list(string), [])
          values        = optional(list(string), [])
        }))
        and = optional(list(object({
          dimension = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = list(string)
          }))
          tag = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = optional(list(string), [])
          }))
          cost_category = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = optional(list(string), [])
          }))
        })), [])
        or = optional(list(object({
          dimension = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = list(string)
          }))
          tag = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = optional(list(string), [])
          }))
          cost_category = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = optional(list(string), [])
          }))
        })), [])
        not = optional(object({
          dimension = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = list(string)
          }))
          tag = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = optional(list(string), [])
          }))
          cost_category = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = optional(list(string), [])
          }))
        }))
      })), [])
      not = optional(object({
        dimension = optional(object({
          key           = optional(string, "")
          match_options = optional(list(string), [])
          values        = list(string)
        }))
        tag = optional(object({
          key           = optional(string, "")
          match_options = optional(list(string), [])
          values        = optional(list(string), [])
        }))
        cost_category = optional(object({
          key           = optional(string, "")
          match_options = optional(list(string), [])
          values        = optional(list(string), [])
        }))
        and = optional(list(object({
          dimension = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = list(string)
          }))
          tag = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = optional(list(string), [])
          }))
          cost_category = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = optional(list(string), [])
          }))
        })), [])
        or = optional(list(object({
          dimension = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = list(string)
          }))
          tag = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = optional(list(string), [])
          }))
          cost_category = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = optional(list(string), [])
          }))
        })), [])
        not = optional(object({
          dimension = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = list(string)
          }))
          tag = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = optional(list(string), [])
          }))
          cost_category = optional(object({
            key           = optional(string, "")
            match_options = optional(list(string), [])
            values        = optional(list(string), [])
          }))
        }))
      }))
    }))
    notifications = optional(list(object({
      comparison_operator        = optional(string, "")
      notification_type          = optional(string, "")
      threshold                  = optional(number, 0)
      threshold_type             = optional(string, "")
      subscriber_email_addresses = optional(list(string), [])
      subscriber_sns_topic_arns  = optional(list(string), [])
    })), [])
    actions = optional(list(object({
      name               = string
      action_type        = optional(string, "")
      approval_model     = optional(string, "")
      notification_type  = optional(string, "")
      execution_role_arn = string
      action_threshold = object({
        action_threshold_type  = optional(string, "")
        action_threshold_value = optional(number, 0)
      })
      subscribers = list(object({
        address           = string
        subscription_type = optional(string, "")
      }))
      iam_action_definition = optional(object({
        policy_arn = string
        groups     = optional(list(string), [])
        roles      = optional(list(string), [])
        users      = optional(list(string), [])
      }))
      scp_action_definition = optional(object({
        policy_id  = string
        target_ids = list(string)
      }))
      ssm_action_definition = optional(object({
        action_sub_type = optional(string, "")
        region          = string
        instance_ids    = list(string)
      }))
    })), [])
  })
}