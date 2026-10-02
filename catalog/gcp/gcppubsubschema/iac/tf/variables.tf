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
  description = "GcpPubSubSchema specification"
  type = object({
    # GCP project where the schema will be created.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the Pub/Sub schema.
    # Must be 3-255 characters, start with a letter, and contain only letters,
    # numbers, hyphens, underscores, periods, tildes, plus signs, and percent
    # signs. Names beginning with "goog" are reserved by Google and rejected
    # at create time. Immutable after creation.
    schema_name = string

    # The schema definition language.
    # Valid values:
    #   - "AVRO": the definition is an Avro schema (JSON). Best default —
    #     human-readable, self-describing, and the format Pub/Sub's BigQuery
    #     and Cloud Storage subscription integrations understand natively.
    #   - "PROTOCOL_BUFFER": the definition is a protobuf message definition
    #     (a single message in proto2 or proto3 syntax). Choose when
    #     publishers already serialize protobuf and binary encoding matters.
    # Keep the type stable for the life of the schema: revisions must stay
    # compatible with the messages already validated against it, and topics
    # choose their encoding (JSON/BINARY) against this type.
    type = string

    # The schema definition text.
    # For AVRO: a JSON Avro schema (e.g. {"type":"record","name":"Event",...}).
    # For PROTOCOL_BUFFER: a protobuf message definition (proto2/proto3 syntax).
    # Changing the definition commits a new schema REVISION in place (no
    # replacement); a schema holds at most 20 revisions, and revisions must
    # be backward-compatible with the encoding topics use, or publishers
    # pinned to older revisions will keep validating against those.
    definition = string

    # Deletion policy for the schema — what happens when this resource is
    # destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the schema is deleted. Topics still referencing it fall
    #                back to the "_deleted-schema_" sentinel and their
    #                publishes start FAILING — detach topics first
    #   "PREVENT" -- destroy FAILS; protects a schema that many topics'
    #                validation depends on
    #   "ABANDON" -- the schema is removed from management but left serving
    #                in GCP (attached topics keep validating)
    deletion_policy = optional(string, "")
  })
}
