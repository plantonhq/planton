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
  description = "GcpVertexAiRagEngineConfig specification"
  type = object({
    # The GCP project whose RAG Engine is configured: a literal project ID or
    # a GcpProject reference. If omitted, the provider's default project is
    # used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Vertex AI location (region) whose RAG Engine is configured, e.g.
    # "us-central1". One configuration exists per project per location.
    # Immutable.
    location = string

    # The managed database tier for this location: BASIC, SCALED, or
    # UNPROVISIONED (see the message comment; UNPROVISIONED deletes the
    # managed database's data). Mutable in place -- raising BASIC to SCALED
    # is a live upgrade.
    tier = string

    # What happens to the configuration when this resource is destroyed:
    #   "" / "DELETE" -- the location is set to UNPROVISIONED, which deletes
    #                    the managed database's data (Google's delete
    #                    semantics for this singleton)
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the configuration leaves management and the
    #                    location keeps its tier and data
    deletion_policy = optional(string, "")
  })
}
