# Input variables for the TestCloudResourceGeneric Terraform module.
#
# The variable shapes mirror the proto spec exactly (snake_case, one variable
# per envelope section) so the proto->tfvars conversion pipeline is exercised
# against every generic field class: scalars with and without defaults, a
# nested message, value-or-reference strings (already resolved to plain
# strings by the time tfvars reach a module), sensitive strings, maps, and
# repeated fields.

variable "metadata" {
  description = "Metadata for the test resource"
  type = object({
    name = string
    org  = optional(string)
    env  = optional(string)
  })
}

variable "spec" {
  description = "Specification for the test resource (every generic field class)"
  type = object({
    display_name = optional(string)
    int32_field       = optional(number)
    int64_field       = optional(number)
    uint32_field      = optional(number)
    uint64_field      = optional(number)
    float_field       = optional(number)
    double_field      = optional(number)
    bool_field        = optional(bool)

    nested = optional(object({
      nested_string = optional(string)
      nested_int    = optional(number)
    }))

    required_ref        = string
    optional_ref        = optional(string)
    annotated_ref       = optional(string)
    kindless_access_ref = optional(string)

    labels = optional(map(string), {})
    steps = optional(list(object({
      command = string
    })), [])
    replicas = optional(number)

    sensitive_string = optional(string)
    sensitive_ref    = optional(string)
  })
}
