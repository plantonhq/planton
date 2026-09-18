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
  description = "GcpFirestoreIndex specification"
  type = object({
    # GCP project owning the database. Can be a literal project ID or a
    # reference to a GcpProject resource. If omitted, the provider's
    # default project is used.
    # Immutable: changing the project destroys and recreates the index.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Firestore database the index belongs to — the database name (a
    # GcpFirestoreDatabase reference resolves to it). Empty falls back to
    # the project's "(default)" database. Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    database = optional(string, "")

    # The collection (group) ID the index applies to. Immutable after
    # creation.
    collection = string

    # Scope of the queries the index serves.
    # COLLECTION: queries against a single collection at this path
    #   (the default).
    # COLLECTION_GROUP: queries against every collection with this ID
    #   anywhere in the database.
    # COLLECTION_RECURSIVE: queries against the collection and everything
    #   beneath it (Datastore Mode only).
    query_scope = optional(string)

    # Which API surface the index serves.
    # ANY_API: Firestore Native queries (the default).
    # DATASTORE_MODE_API: Datastore Mode queries.
    # MONGODB_COMPATIBLE_API: Firestore Enterprise (MongoDB-compatible)
    #   queries — required for multikey and search-config indexes, and
    #   only valid on an ENTERPRISE-edition database. Requires
    #   query_scope COLLECTION_GROUP explicitly (the API rejects the
    #   COLLECTION default for this scope).
    api_scope = optional(string)

    # Index density. Leave empty for GCP's default (SPARSE_ALL).
    # DENSE indexes also include documents missing the indexed fields —
    # required for some Datastore Mode query shapes.
    density = optional(string, "")

    # The indexed fields, in query order (equality filters first, then
    # inequality/sort fields; a vector field, if any, last). At least one
    # field. Firestore appends __name__ automatically.
    fields = list(object({
      # Dot-separated field path in the document (e.g. "user.age").
      field_path = string

      # Sort order for a scalar field: ASCENDING or DESCENDING.
      order = optional(string, "")

      # CONTAINS declares an array-membership index on the field (enables
      # array-contains queries).
      array_config = optional(string, "")

      # Vector index configuration for nearest-neighbor queries.
      vector_config = optional(object({
        # Dimension of the vectors indexed on this field (the embedding
        # model's output size, e.g. 768).
        dimension = number
      }))

      # Text or geo search index configuration for the field — the
      # Firestore Enterprise search surface. Requires an ENTERPRISE-edition
      # database (GcpFirestoreDatabase database_edition: ENTERPRISE); text
      # search additionally pairs with api_scope MONGODB_COMPATIBLE_API,
      # while geo search works under the default scope. Search and
      # non-search fields cannot mix in one index (API-enforced): when any
      # field carries search_config, every field must.
      search_config = optional(object({
        # Text search index specification for the field.
        text_spec = optional(object({
          # Index specifications; at least one.
          index_specs = list(object({
            # How the text field value is indexed (e.g. "TOKENIZED").
            index_type = optional(string, "")

            # How the text field value is matched (e.g. "MATCH_GLOBALLY").
            match_type = optional(string, "")
          }))
        }))

        # Geo search index specification for the field.
        geo_spec = optional(object({
          # If true, disables GeoJSON indexing for the field (GeoJSON points are
          # indexed by default). Firestore GeoPoints are indexed regardless.
          geo_json_indexing_disabled = optional(bool, false)
        }))
      }))
    }))

    # Whether the index is multikey: at most one indexed path may reach or
    # traverse an array (MongoDB-style array indexing). Only valid with
    # api_scope MONGODB_COMPATIBLE_API. Immutable after creation.
    multikey = optional(bool, false)

    # Whether the index enforces uniqueness: all values of the indexed
    # field(s) must be unique across documents. Immutable after creation.
    unique = optional(bool, false)

    # If true, the deploy returns as soon as index creation is REQUESTED
    # instead of waiting for the (potentially long) background build to
    # finish. The index serves queries only once the build completes —
    # use when orchestrating many indexes and the caller polls readiness
    # itself.
    skip_wait = optional(bool, false)

    # Deletion policy — what happens when this resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the index is deleted
    #   "PREVENT" -- destroy FAILS; protects indexes whose rebuild would
    #                be expensive on a large collection
    #   "ABANDON" -- the index is removed from management but kept
    #                serving queries in GCP
    deletion_policy = optional(string, "")
  })
}
