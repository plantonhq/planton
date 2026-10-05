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
  description = "AwsBedrockPrompt specification"
  type = object({
    region = string
    description = optional(string, "")
    customer_encryption_key_arn = optional(string, "")
    default_variant = optional(string, "")
    # Deliberately `any`: variant entries are heterogeneous (text vs chat
    # arms, JSON-document members like input_schema whose concrete types
    # differ per entry), and HCL cannot unify `any`-typed members across
    # list elements. The module reads every optional field with try().
    variants = any
  })
}
