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
  description = "AwsOrganizationAccount specification"
  type = object({
    region                     = string
    account_name               = string
    email                      = string
    parent_id                  = optional(string, "")
    role_name                  = optional(string, "")
    iam_user_access_to_billing = optional(string, "")
    close_on_deletion          = optional(bool, false)
    create_govcloud            = optional(bool, false)
    alternate_contacts = optional(object({
      billing = optional(object({
        name          = string
        title         = string
        email_address = optional(string, "")
        phone_number  = optional(string, "")
      }))
      operations = optional(object({
        name          = string
        title         = string
        email_address = optional(string, "")
        phone_number  = optional(string, "")
      }))
      security = optional(object({
        name          = string
        title         = string
        email_address = optional(string, "")
        phone_number  = optional(string, "")
      }))
    }))
    primary_contact = optional(object({
      full_name          = string
      company_name       = optional(string, "")
      address_line_1     = string
      address_line_2     = optional(string, "")
      address_line_3     = optional(string, "")
      city               = string
      district_or_county = optional(string, "")
      state_or_region    = optional(string, "")
      postal_code        = string
      country_code       = optional(string, "")
      phone_number       = optional(string, "")
      website_url        = optional(string, "")
    }))
    regions = optional(list(object({
      region_name = string
      enabled     = optional(bool, false)
    })), [])
  })
}