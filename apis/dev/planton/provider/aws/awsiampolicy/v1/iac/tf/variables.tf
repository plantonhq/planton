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
  description = "AwsIamPolicy specification"
  type = object({
    region = string
    # policy_document is free-form JSON (google.protobuf.Struct); typed `any`
    # because policy documents have heterogeneous shapes.
    policy_document = any
    description = optional(string, "")
    path = optional(string, "")
  })
}
