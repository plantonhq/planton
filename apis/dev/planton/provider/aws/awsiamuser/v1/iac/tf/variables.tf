variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "AwsIamUser specification"
  type = object({
    region = string
    user_name = string
    path = optional(string, "")
    # managed_policy_arns arrive pre-resolved: the orchestrator replaces each
    # valueFrom reference with the referenced AwsIamPolicy's policy_arn before
    # the module runs, so the module sees a plain list of ARN strings.
    managed_policy_arns = optional(list(string), [])
    # inline_policies is free-form JSON (map<string, google.protobuf.Struct>);
    # typed `any` because its entries have heterogeneous shapes.
    inline_policies = optional(any, {})
    # permissions_boundary arrives pre-resolved (an AwsIamPolicy's policy_arn
    # or a literal ARN).
    permissions_boundary = optional(string, "")
    disable_access_keys = optional(bool, false)
    force_destroy = optional(bool, false)
  })
}
