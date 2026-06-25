variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "CloudflareEmailRoutingRuleSpec defines a single Email Routing rule"
  type = object({
    # (Required) The zone ID. StringValueOrRef is flattened to a plain string.
    zone_id = optional(string)

    # (Optional) Descriptive rule name.
    name = optional(string, "")

    # (Optional) Whether the rule is active (default true).
    enabled = optional(bool, true)

    # (Optional) Evaluation priority (lower first; default 0).
    priority = optional(number, 0)

    # (Required) Matchers selecting which messages this rule applies to.
    matchers = list(object({
      type  = string
      field = optional(string, "")
      value = optional(string, "")
    }))

    # (Required) The action taken on matched messages. forward_to and worker are
    # StringValueOrRef lists/values flattened to plain strings.
    action = object({
      type       = string
      forward_to = optional(list(string), [])
      worker     = optional(string, "")
    })
  })
}
