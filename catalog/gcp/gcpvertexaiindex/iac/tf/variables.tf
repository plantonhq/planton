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
  description = "GcpVertexAiIndex specification"
  type = object({
    # GCP project where the index will be created.
    # If omitted, the index is created in the provider's default project
    # (from the credential or ambient configuration).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Region where the index will be created (e.g., "us-central1").
    # Deployed indexes must live in the same region. Immutable after
    # creation.
    location = string

    # Display name of the index (up to 128 UTF-8 characters). The
    # primary human-readable identifier; the numeric resource ID is
    # GCP-assigned.
    display_name = string

    # Description of the index.
    description = optional(string, "")

    # How the index contents are updated after creation. Immutable —
    # migrating between regimes means recreating the index.
    # BATCH_UPDATE: bulk rebuilds from Cloud Storage (GCP's default).
    # STREAM_UPDATE: near-real-time upserts via the API.
    index_update_method = optional(string)

    # Cloud Storage DIRECTORY (not file) holding the initial or delta
    # vector data, e.g. "gs://my-bucket/embeddings/". File format and
    # layout: https://cloud.google.com/vertex-ai/docs/vector-search/setup/format-structure
    #
    # May be omitted at creation (an empty index; the norm for
    # STREAM_UPDATE) and set later to load data. Two provider quirks,
    # both harmless but worth knowing: a change to this field travels in
    # its own single-field update, and the value is write-only — GCP
    # never reports it back on read, so out-of-band loads show up as a
    # one-field diff on the next plan.
    #
    # A plain string (not a reference) because the gs:// directory URI
    # has no matching stack output shape on the GCS kinds; compose by
    # writing the bucket name into the URI.
    contents_delta_uri = optional(string, "")

    # If true, an update that carries contents_delta_uri REPLACES the
    # whole index contents with the files at the URI; if false (the
    # default), the files are treated as a delta (upserts/deletes) on
    # top of the existing contents. Only meaningful together with
    # contents_delta_uri.
    is_complete_overwrite = optional(bool, false)

    # Vector-search geometry: dimensions, algorithm, sharding, distance
    # measure. Required — an index cannot exist without its geometry.
    # The entire block is immutable.
    config = object({
      # Number of dimensions of the input vectors — the embedding model's
      # output size (e.g. 768 for many text encoders). Immutable.
      dimensions = number

      # Number of neighbors found via approximate search before exact
      # reordering (a more expensive distance computation over the
      # candidates). Required by the API when tree-AH is used; not
      # meaningful for brute-force. Immutable.
      approximate_neighbors_count = optional(number, 0)

      # Physical shard size for the index data. Determines how much data
      # each shard holds and which machine types can serve it when
      # deployed. If omitted, GCP picks a size based on the data.
      # SHARD_SIZE_SMALL: 2 GB per shard.
      # SHARD_SIZE_MEDIUM: 20 GB per shard.
      # SHARD_SIZE_LARGE: 50 GB per shard.
      # Immutable.
      shard_size = optional(string, "")

      # Distance measure used in nearest-neighbor search. Match it to how
      # the embedding model was trained — a mismatched measure silently
      # degrades result quality.
      # SQUARED_L2_DISTANCE: Euclidean (L2) distance.
      # L1_DISTANCE: Manhattan (L1) distance.
      # COSINE_DISTANCE: 1 - cosine similarity.
      # DOT_PRODUCT_DISTANCE: negative dot product (GCP's default).
      # Immutable.
      distance_measure_type = optional(string)

      # Normalization applied to each vector before indexing.
      # UNIT_L2_NORM: unit L2 normalization (with DOT_PRODUCT_DISTANCE this
      #   makes ranking equivalent to cosine similarity).
      # NONE: no normalization (GCP's default).
      # Immutable.
      feature_norm_type = optional(string)

      # Tree-AH approximate search configuration. Mutually exclusive with
      # brute_force_config. Immutable.
      tree_ah_config = optional(object({
        # Number of embeddings on each leaf node of the tree. Larger leaves
        # mean fewer tree levels (faster build, coarser pruning); smaller
        # leaves prune more aggressively per query. GCP's default is 1000.
        leaf_node_embedding_count = optional(number)

        # Percentage of leaf nodes any single query may search, 1-100
        # inclusive. Raising it improves recall at the cost of latency.
        # GCP's default is 10.
        leaf_nodes_to_search_percent = optional(number)
      }))

      # Brute-force (exact) search. Mutually exclusive with tree_ah_config.
      # Immutable.
      brute_force_config = optional(object({}))
    })

    # User-defined labels to organize the index (cost attribution, team
    # ownership, environment tagging). Keys and values must follow GCP
    # label rules: lowercase letters, digits, underscores, and dashes,
    # at most 63 characters. Merged with the platform's attribution
    # labels; on key conflicts the platform labels win. Mutable in place.
    labels = optional(map(string), {})

    # Cloud KMS key for customer-managed encryption at rest (CMEK) of the
    # index data, as the full key resource path
    # projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
    # — a GcpKmsKey reference resolves to it. The key must live in the
    # same region as the index, and the Vertex AI service agent needs
    # roles/cloudkms.cryptoKeyEncrypterDecrypter on it. If omitted, data
    # is encrypted with Google-managed keys. Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_name = optional(string, "")

    # Deletion policy for the index — what happens when this resource is
    # destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the index and every vector in it are deleted
    #   "PREVENT" -- destroy FAILS; a guard for a corpus that took hours
    #                of batch builds (or months of streaming upserts) to
    #                load
    #   "ABANDON" -- the index is removed from management but left
    #                standing (and billing for its stored vectors) in GCP
    deletion_policy = optional(string, "")
  })
}
