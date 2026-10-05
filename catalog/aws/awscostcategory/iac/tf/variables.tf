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
  description = "AwsCostCategory specification"
  type = object({
    region          = string
    category_name   = string
    default_value   = optional(string, "")
    effective_start = optional(string, "")
    rules = list(object({
      type  = optional(string, "")
      value = optional(string, "")
      rule = optional(object({
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
          key           = string
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
            key           = string
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
              key           = string
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
              key           = string
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
              key           = string
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
            key           = string
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
              key           = string
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
              key           = string
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
              key           = string
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
            key           = string
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
              key           = string
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
              key           = string
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
              key           = string
              match_options = optional(list(string), [])
              values        = optional(list(string), [])
            }))
          }))
        }))
      }))
      inherited_value = optional(object({
        dimension_name = optional(string, "")
        dimension_key  = optional(string, "")
      }))
    }))
    split_charge_rules = optional(list(object({
      source  = string
      targets = list(string)
      method  = optional(string, "")
      parameters = optional(list(object({
        type   = optional(string, "")
        values = list(string)
      })), [])
    })), [])
  })
}