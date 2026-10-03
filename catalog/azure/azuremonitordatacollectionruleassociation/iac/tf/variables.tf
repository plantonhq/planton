variable "metadata" {
  description = "Catalog object metadata"
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
  description = "AzureMonitorDataCollectionRuleAssociation specification"
  type = object({
    # Resolved ARM id of the machine being attached (VM, VM scale set,
    # or Arc-enabled server).
    target_resource_id = string

    # Required when binding a rule; left empty for endpoint bindings
    # (the provider applies Azure's mandated fixed name).
    name = optional(string, "")

    # Exactly one of the two bindings (spec-enforced by CEL, provider-
    # enforced by ExactlyOneOf). Resolved ARM ids.
    data_collection_rule_id     = optional(string)
    data_collection_endpoint_id = optional(string)

    description = optional(string, "")
  })
}
