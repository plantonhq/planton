variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}

variable "spec" {
  description = "AwsGuardDuty specification"
  type = object({
    region = string
    enable = optional(bool)
    finding_publishing_frequency = optional(string, "")
    features = optional(list(object({
      name = optional(string, "")
      enabled = optional(bool)
      additional_configuration = optional(list(object({
        name = optional(string, "")
        enabled = optional(bool)
      })), [])
    })), [])
    filters = optional(list(object({
      name = string
      description = optional(string, "")
      action = optional(string, "")
      rank = optional(number, 0)
      criteria = list(object({
        field = string
        equals = optional(list(string), [])
        not_equals = optional(list(string), [])
        matches = optional(list(string), [])
        not_matches = optional(list(string), [])
        greater_than = optional(string, "")
        greater_than_or_equal = optional(string, "")
        less_than = optional(string, "")
        less_than_or_equal = optional(string, "")
      }))
    })), [])
    ip_sets = optional(list(object({
      name = string
      format = optional(string, "")
      location = string
      activate = optional(bool, false)
    })), [])
    threat_intel_sets = optional(list(object({
      name = string
      format = optional(string, "")
      location = string
      activate = optional(bool, false)
    })), [])
    publishing_destination = optional(object({
      bucket_arn = string
      kms_key_arn = string
    }))
    organization = optional(object({
      admin_account_id = optional(string, "")
      auto_enable_organization_members = optional(string, "")
      features = optional(list(object({
        name = optional(string, "")
        auto_enable = optional(string, "")
        additional_configuration = optional(list(object({
          name = optional(string, "")
          auto_enable = optional(string, "")
        })), [])
      })), [])
    }))
    members = optional(list(object({
      account_id = optional(string, "")
      email = string
      invite = optional(bool)
      invitation_message = optional(string, "")
      disable_email_notification = optional(bool, false)
      features = optional(list(object({
        name = optional(string, "")
        enabled = optional(bool)
        additional_configuration = optional(list(object({
          name = optional(string, "")
          enabled = optional(bool)
        })), [])
      })), [])
    })), [])
    accept_invitation_from_account_id = optional(string, "")
  })
}