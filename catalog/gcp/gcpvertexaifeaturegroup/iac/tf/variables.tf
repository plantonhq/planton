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
  description = "GcpVertexAiFeatureGroup specification"
  type = object({
    # The GCP project the feature group lives in: a literal project ID or a
    # GcpProject reference. If omitted, the provider's default project is
    # used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Vertex AI location (region) the feature group lives in, e.g.
    # "us-central1". Online stores that serve it must be in the same
    # location. Immutable.
    location = string

    # The feature group's id -- the last segment of its resource name and
    # what an online store's feature view references. Up to 128 characters
    # of lowercase letters, digits, and underscores; the first character
    # cannot be a digit (hyphens are not allowed, so metadata.name cannot
    # stand in). Unique within the project and location. Treat it as
    # immutable: the provider does not mark it ForceNew, so change it by
    # replacing the block.
    feature_group_id = string

    # Free-text description of the feature group.
    description = optional(string, "")

    # Labels on the feature group. The platform attribution labels are
    # merged in and win on key conflicts.
    labels = optional(map(string), {})

    # The BigQuery table or view the features come from. Google's API has no
    # other source type today.
    big_query = optional(object({
      # The BigQuery table or view holding the feature data: a GcpBigQueryTable
      # reference (its {project}.{dataset}.{table} name), a literal
      # "project.dataset.table", or the bq:// form Google stores
      # ("bq://project.dataset.table"). The modules add the bq:// prefix when
      # it is missing. The source must have at least one entity ID column and
      # a TIMESTAMP column named `feature_timestamp` -- Google's contract for a
      # feature group source. Jobs read it as the Vertex AI Service Agent,
      # which needs roles/bigquery.dataViewer on the table. Immutable.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      input_uri = string

      # The source columns whose values together form an entity's ID (the row
      # key an online store serves by). Empty means the single column
      # `entity_id`. Mutable in place.
      entity_id_columns = optional(list(string), [])
    }))

    # The source columns registered as features, each keyed by feature_id.
    # Add or remove a feature by editing this list.
    features = optional(list(object({
      # The feature's id within the group -- what a feature view's
      # feature_ids names. Up to 128 characters of lowercase letters, digits,
      # and underscores; the first character cannot be a digit. By default it
      # is also the source column the feature reads (see version_column_name).
      # Immutable.
      feature_id = string

      # Free-text description of the feature.
      description = optional(string, "")

      # Labels on the feature. The platform attribution labels are merged in
      # and win on key conflicts.
      labels = optional(map(string), {})

      # The source column that holds this feature's values, when it differs
      # from feature_id. Empty means the column named like the feature. Sent
      # only when set (Google computes it otherwise).
      version_column_name = optional(string, "")
    })), [])

    # What happens to the feature group and its features when this resource
    # is destroyed:
    #   "" / "DELETE" -- the features and the group are deleted (the BigQuery
    #                    source is never touched)
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- everything leaves management and stays in GCP
    deletion_policy = optional(string, "")
  })
}
