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
  description = "AwsSsmPatchBaseline specification"
  type = object({
    region = string
    operating_system = optional(string, "")
    description = optional(string, "")
    approval_rules = optional(list(object({
      patch_filters = list(object({
        key = string
        values = list(string)
      }))
      approve_after_days = optional(number)
      approve_until_date = optional(string, "")
      compliance_level = optional(string, "")
      enable_non_security = optional(bool, false)
    })), [])
    global_filters = optional(list(object({
      key = string
      values = list(string)
    })), [])
    approved_patches = optional(list(string), [])
    approved_patches_compliance_level = optional(string, "")
    approved_patches_enable_non_security = optional(bool, false)
    rejected_patches = optional(list(string), [])
    rejected_patches_action = optional(string, "")
    available_security_updates_compliance_status = optional(string, "")
    sources = optional(list(object({
      name = string
      configuration = string
      products = list(string)
    })), [])
    patch_groups = optional(list(string), [])
    set_as_default_baseline = optional(bool, false)
  })
}
