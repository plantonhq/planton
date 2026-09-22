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
  description = "GcpVectorSearchCollection specification"
  type = object({
    # The GCP project the collection lives in: a literal project ID or a
    # GcpProject reference. If omitted, the provider's default project is
    # used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Vector Search location (region), e.g. "us-central1". Immutable.
    location = string

    # The collection's ID -- 1-63 characters, RFC 1035 (lowercase letters,
    # digits, hyphens; starts with a letter). Defaults to metadata.name.
    # Immutable.
    collection_id = optional(string, "")

    # Human-readable name shown in the console.
    display_name = optional(string, "")

    # Free-text description of the collection.
    description = optional(string, "")

    # Labels on the collection.
    labels = optional(map(string), {})

    # JSON Schema for the non-vector fields of each data object, as a JSON
    # string (write it compact; Google normalizes it). Field names must be
    # alphanumeric characters, underscores, and hyphens. Fields named here
    # can be pushed into an index as filter_fields or store_fields and
    # referenced from a text_template.
    data_schema = optional(string, "")

    # The searchable vector fields. Only fields declared here can be
    # indexed and searched.
    vector_schemas = optional(list(object({
      # The field's name -- alphanumeric characters, underscores, and hyphens.
      # An index names this field as its index_field.
      field_name = string

      # A dense vector field, optionally with Vertex AI computing the
      # embeddings.
      dense_vector = optional(object({
        # Number of dimensions in the vector (768 for textembedding-gecko and
        # text-embedding-005 at their default; 3072 for gemini-embedding-001).
        # Must match the embedding model when vertex_embedding_config is set.
        dimensions = optional(number)

        # Have Vector Search compute this field's embeddings from the object's
        # text through a Vertex AI model. Omit to write precomputed vectors.
        vertex_embedding_config = optional(object({
          # The Vertex AI embedding model, e.g. "text-embedding-005" or
          # "textembedding-gecko@003". The model fixes the field's dimensionality;
          # dimensions on the parent must match what the model produces.
          model_id = string

          # The embedding task the model is told to optimize for. Documents stored
          # in the collection are embedded with RETRIEVAL_DOCUMENT; queries use
          # RETRIEVAL_QUERY at search time.
          task_type = string

          # The text handed to the model for each data object, with one or more
          # references to the object's fields in braces, e.g.
          # "Movie Title: {title} ---- Movie Plot: {plot}". Field names come from
          # the collection's data_schema.
          text_template = string
        }))
      }))

      # True declares a sparse vector field (index/value pairs over a large
      # vocabulary, the shape of lexical or hybrid-search embeddings such as
      # SPLADE or BM25). Google's sparse field carries no settings, so the
      # declaration is a flag.
      sparse_vector = optional(bool, false)
    })), [])

    # Customer-managed encryption key protecting the collection and its
    # indexes: a GcpKmsKey reference or a literal
    # projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
    # in the same region. Omit to use Google-managed encryption. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_name = optional(string, "")

    # The approximate-nearest-neighbor indexes over the collection's vector
    # fields, each keyed by its index_id. Add or remove an index by editing
    # this list; any other change to an index replaces it.
    indexes = optional(list(object({
      # The index's ID within the collection -- 1-63 characters, RFC 1035
      # (lowercase letters, digits, hyphens; starts with a letter). Immutable.
      index_id = string

      # The vector field (a vector_schemas[].field_name) this index serves.
      # Immutable.
      index_field = string

      # Human-readable name shown in the console.
      display_name = optional(string, "")

      # Free-text description of the index.
      description = optional(string, "")

      # Labels on the index -- the one mutable setting.
      labels = optional(map(string), {})

      # How similarity is measured: DOT_PRODUCT (Google's default; the right
      # choice for embeddings the model already normalizes) or COSINE_DISTANCE.
      # Sent only when set. Immutable.
      distance_metric = optional(string, "")

      # Feature normalization the ScaNN index applies before search: NONE, or
      # UNIT_L2_NORM to unit-normalize every vector (makes DOT_PRODUCT behave
      # as cosine similarity). Sent only when set. Immutable.
      feature_norm_type = optional(string, "")

      # Data-schema fields pushed into the index so searches can filter on
      # them inline, without a second lookup. Immutable.
      filter_fields = optional(list(string), [])

      # Data-schema fields pushed into the index so search results return
      # them inline, without fetching the data object. Immutable.
      store_fields = optional(list(string), [])

      # Dedicated serving nodes for this index. Omit to serve from the shared
      # pool.
      dedicated_infrastructure = optional(object({
        # STORAGE_OPTIMIZED for large indexes where cost per vector matters more
        # than latency; PERFORMANCE_OPTIMIZED (Google's default) for
        # latency-sensitive serving. Immutable: a change replaces the index.
        mode = optional(string, "")

        # Replica bounds for the dedicated nodes.
        autoscaling_spec = optional(object({
          # Fewest replicas kept serving (1-1000; Google defaults to 2 when unset
          # or 0). Sent only when set.
          min_replica_count = optional(number)

          # Most replicas the index may scale to (at least min_replica_count, at
          # most 1000; Google defaults to the greater of min_replica_count and 2).
          # Sent only when set.
          max_replica_count = optional(number)
        }))
      }))
    })), [])

    # What happens to the collection and its indexes when this resource is
    # destroyed:
    #   "" / "DELETE" -- the indexes and the collection are deleted, data
    #                    included
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- everything leaves management and stays in GCP
    deletion_policy = optional(string, "")
  })
}
