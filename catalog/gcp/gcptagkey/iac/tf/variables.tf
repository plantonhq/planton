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
  description = "GcpTagKey specification"
  type = object({
    # Who owns the key: the organization or one project. Exactly one arm.
    # Immutable.
    parent = object({
      # An organization-owned key: the numeric organization ID, without the
      # `organizations/` prefix. Its values can be bound to any resource in
      # the organization -- the landing-zone default.
      organization_id = optional(string, "")

      # A project-owned key: a literal project ID (Google also accepts the
      # number; the provider treats the two as equal) or a reference to a
      # GcpProject resource. Its values can be bound only to resources in that
      # project.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      project_id = optional(string, "")
    })

    # The key's name as written in tag conditions and shown in the console:
    # `environment`, `cost-center`. Defaults to metadata.name when empty.
    # Unique among the owner's keys; 1-256 characters of any UTF-8 except
    # the four Google forbids -- `/`, `\`, `'`, `"` -- because
    # `{parent}/{key}/{value}` and `resource.matchTag('{org}/{key}',
    # '{value}')` address tags by short name. Immutable: a rename is a
    # delete and a create.
    #
    # A deleted key's short name stays reserved for the 30-day soft-delete
    # window; a fresh key cannot reuse it under the same owner until then.
    short_name = optional(string, "")

    # What the key is for, shown in the console. At most 256 characters.
    # Mutable.
    description = optional(string, "")

    # Reserve the key for one of Google's system uses, which changes how the
    # key behaves:
    #   ""                -- an ordinary tag key (the usual case)
    #   "GCE_FIREWALL"    -- the key's values may be used as targets and
    #                        sources in network firewall policy rules
    #                        (secure tags); requires purpose_data.network
    #   "DATA_GOVERNANCE" -- the key's values classify data for Sensitive
    #                        Data Protection and BigQuery policy tags
    # Immutable: a purpose cannot be added, changed, or removed once the key
    # exists.
    purpose = optional(string, "")

    # Data the purpose needs, as Google defines it per purpose. For
    # GCE_FIREWALL exactly one entry, `network`, naming the VPC the secure
    # tags are scoped to as `{project_id}/{network_name}` (or the network's
    # full resource name). Only meaningful with a purpose; immutable with it.
    purpose_data = optional(map(string), {})

    # An RE2 regular expression every value's short name must match. When
    # set, the key also becomes a DYNAMIC key: bindings may carry values
    # that were never declared, as long as they match the regex (the
    # pattern for high-cardinality tags such as a ticket number or a team
    # code). Leave empty for a fixed vocabulary of declared GcpTagValues.
    # Mutable. Google compiles it on apply; an invalid expression is
    # rejected there.
    allowed_values_regex = optional(string, "")

    # What destroying this resource does to the key in GCP:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the key is deleted; fails while any of its values still
    #                exist (destroy the values first -- a chart's dependency
    #                order does this when the values reference the key)
    #   "PREVENT" -- destroy FAILS; the guard for a key every policy keys on
    #   "ABANDON" -- the key is removed from management but stays live in
    #                GCP with its values and bindings
    deletion_policy = optional(string, "")
  })
}
