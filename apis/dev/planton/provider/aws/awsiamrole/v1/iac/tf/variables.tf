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
  description = "AwsIamRole specification"
  type = object({
    region = string
    description = optional(string, "")
    path = optional(string, "")
    # trust_policy is free-form JSON (google.protobuf.Struct); typed `any`
    # because trust documents have heterogeneous shapes.
    trust_policy = any
    # managed_policy_arns arrive pre-resolved: the orchestrator replaces each
    # valueFrom reference with the referenced AwsIamPolicy's policy_arn before
    # the module runs, so the module sees a plain list of ARN strings.
    managed_policy_arns = optional(list(string), [])
    # inline_policies is free-form JSON (map<string, google.protobuf.Struct>);
    # typed `any` because its entries have heterogeneous shapes.
    inline_policies = optional(any, {})
    max_session_duration = optional(number, 0)
    # permissions_boundary arrives pre-resolved (an AwsIamPolicy's policy_arn
    # or a literal ARN).
    permissions_boundary = optional(string, "")
    force_detach_policies = optional(bool, false)
  })
}
