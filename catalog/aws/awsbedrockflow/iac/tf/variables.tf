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
  description = "AwsBedrockFlow specification"
  type = object({
    region = string
    description = optional(string, "")
    execution_role_arn = string
    customer_encryption_key_arn = optional(string, "")
    definition = optional(object({
      # Deliberately `any`: node entries are heterogeneous (one
      # configuration arm per class, JSON-document members like
      # input_schema whose concrete types differ per entry), and HCL
      # cannot unify `any`-typed members across list elements. The module
      # reads every optional field with try().
      nodes = any
      connections = optional(list(object({
        name = optional(string, "")
        source = optional(string, "")
        target = optional(string, "")
        data = optional(object({
          source_output = optional(string, "")
          target_input = optional(string, "")
        }))
        conditional = optional(object({
          condition = optional(string, "")
        }))
      })), [])
    }))
  })
}
