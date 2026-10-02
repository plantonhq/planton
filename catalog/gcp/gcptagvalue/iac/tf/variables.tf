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
  description = "GcpTagValue specification"
  type = object({
    # The key this value belongs to: a reference to a GcpTagKey resource
    # (its `name` output, `tagKeys/{id}`) or that name as a literal.
    # Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    tag_key = string

    # The value as written in tag conditions and shown in the console:
    # `prod`, `pci`. Defaults to metadata.name when empty. Unique among the
    # key's values; 1-256 characters of any UTF-8 except the four Google
    # forbids -- `/`, `\`, `'`, `"`. Must match the key's
    # allowed_values_regex when the key has one. Immutable: a rename is a
    # delete and a create.
    #
    # A deleted value's short name stays reserved for the 30-day soft-delete
    # window; a fresh value cannot reuse it under the same key until then.
    short_name = optional(string, "")

    # What the value means, shown in the console. At most 256 characters.
    # Mutable.
    description = optional(string, "")

    # What destroying this resource does to the value in GCP:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the value is deleted; fails while any binding still
    #                uses it (destroy the bindings first -- a chart's
    #                dependency order does this when they reference the value)
    #   "PREVENT" -- destroy FAILS; the guard for a value every policy tests
    #   "ABANDON" -- the value is removed from management but stays live in
    #                GCP with its bindings
    deletion_policy = optional(string, "")
  })
}
