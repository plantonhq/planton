variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "AwsInternetGateway specification"
  # The vpc_id StringValueOrRef arrives as a plain string: the tofu generator
  # flattens StringValueOrRef to its string value (see pkg/iac/tofu/generators
  # typerules), and the orchestrator resolves any value_from reference before the
  # module runs.
  type = object({
    region = string

    vpc_id = string
  })
}
