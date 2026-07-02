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
  description = "AwsIamInstanceProfile specification"
  type = object({
    region = string
    # role arrives pre-resolved: the orchestrator replaces a valueFrom reference
    # with the referenced AwsIamRole's role_name before the module runs.
    role = string
    path = optional(string, "")
  })
}
