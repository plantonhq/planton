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
  description = "GcpLoggingSink specification"
  type = object({
    # Where the sink lives. Omit entirely for a project sink in the
    # provider's default project — the common case.
    scope = optional(object({
      # Project sink: the owning project — a literal project ID or a reference
      # to a GcpProject resource.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      project_id = optional(string, "")

      # Folder sink: the folder ID (numeric, with or without the "folders/"
      # prefix).
      folder_id = optional(string, "")

      # Organization sink: the numeric organization ID.
      organization_id = optional(string, "")

      # Billing-account sink: the billing account ID
      # (e.g. 012345-6789AB-CDEF01).
      billing_account = optional(string, "")
    }))

    # The sink name in GCP. Defaults to metadata.name when left empty.
    # Immutable: changing it destroys and recreates the sink (a new
    # writer_identity is minted — re-grant destination access).
    sink_name = optional(string, "")

    # Where matching log entries are exported. Exactly one destination arm.
    destination = object({
      # Export to a Cloud Storage bucket (hourly batches of JSON files). The
      # bucket NAME — a literal or a reference to a GcpGcsBucket resource.
      # Rendered as storage.googleapis.com/{bucket}. Grant the sink's
      # writer_identity roles/storage.objectCreator on the bucket.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      gcs_bucket = optional(string, "")

      # Export to a BigQuery dataset (near-real-time, queryable). Accepts the
      # dataset SELF LINK (https://bigquery.googleapis.com/bigquery/v2/projects/
      # {p}/datasets/{d} — the GcpBigQueryDataset self_link output) or a bare
      # projects/{p}/datasets/{d} path; the module normalizes either into the
      # bigquery.googleapis.com/... destination URI. Grant the writer_identity
      # roles/bigquery.dataEditor on the dataset. The reference is
      # containment-exempt: the sink EXPORTS INTO the dataset, it does not live
      # inside it (the sink's home is its scope).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      bigquery_dataset = optional(string, "")

      # BigQuery destinations only: write into date-partitioned tables
      # (recommended — enables partition pruning and expiration) instead of
      # date-sharded table names. Requires unique_writer_identity true (the
      # provider's own constraint).
      use_partitioned_tables = optional(bool, false)

      # Export to a Pub/Sub topic (streaming; the front door to third-party
      # log pipelines). The full topic path projects/{p}/topics/{t} — a literal
      # or a reference to a GcpPubSubTopic resource (its topic_id output).
      # Rendered as pubsub.googleapis.com/projects/{p}/topics/{t}. Grant the
      # writer_identity roles/pubsub.publisher on the topic.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      pubsub_topic = optional(string, "")

      # Escape hatch: a complete destination URI passed through verbatim, for
      # destinations without a first-class arm — chiefly Cloud Logging buckets
      # (logging.googleapis.com/projects/{p}/locations/{l}/buckets/{b}) until a
      # dedicated log-bucket kind exists, and cross-project Logging API
      # destinations.
      raw_uri = optional(string, "")
    })

    # The logs filter selecting entries to export
    # (https://cloud.google.com/logging/docs/view/logging-query-language),
    # e.g. severity>=ERROR or resource.type="cloud_run_revision".
    # Empty exports EVERYTHING in scope — deliberate for audit archives,
    # expensive for anything else.
    filter = optional(string, "")

    # Why this sink exists and where its data lands — write it for the
    # operator auditing log routing later. At most 8000 characters.
    description = optional(string, "")

    # If true, the sink keeps its configuration but exports nothing — the
    # safe way to pause an export without losing the writer identity and
    # destination grants.
    disabled = optional(bool, false)

    # Log entries matching ANY of these exclusion filters are NOT exported,
    # even when they match `filter` — carve noisy sub-streams (health checks,
    # debug logs) out of a broad export.
    exclusions = optional(list(object({
      # Identifier for this exclusion (letters, digits, underscores, hyphens,
      # periods; must start alphanumeric; at most 100 characters).
      name = string

      # The logs filter selecting entries to EXCLUDE. Use the sample()
      # function to exclude only a percentage (e.g.
      # sample(insertId, 0.9) drops 90% of matches).
      filter = string

      # What this exclusion carves out and why.
      description = optional(string, "")

      # If true, this exclusion is ignored (its entries export normally) —
      # stage or pause a carve-out without deleting it.
      disabled = optional(bool, false)
    })), [])

    # Folder/organization scopes only: also export logs from all CHILD
    # resources (subfolders and projects). Off by default — the sink then
    # sees only logs at the scope itself.
    include_children = optional(bool, false)

    # Folder/organization scopes only (requires include_children semantics):
    # INTERCEPT matching logs — they are exported by this sink and NOT routed
    # onward to the children's own sinks. For centralized compliance capture;
    # use deliberately, it changes what child projects see.
    intercept_children = optional(bool, false)

    # Project scope only. If true (default), GCP mints a dedicated service
    # account as the sink's writer identity — required for exporting across
    # projects and for BigQuery destinations. Setting false uses the legacy
    # shared cloud-logs@ account (single-project GCS/Pub/Sub only; not
    # recommended). Folder/organization/billing sinks always get a unique
    # writer. Both IaC engines send the value explicitly on project sinks so
    # behavior is identical regardless of engine.
    unique_writer_identity = optional(bool)

    # Project scope only: use a caller-provided service account as the
    # writer identity instead of a GCP-minted one (format:
    # serviceAccount:{email} is NOT included — pass the bare email). Requires
    # unique_writer_identity true semantics; GCP rejects the combination
    # with false.
    custom_writer_identity = optional(string, "")

    # Deletion policy — what happens when this resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the sink is deleted; export stops immediately (already
    #                exported data stays in the destination)
    #   "PREVENT" -- destroy FAILS; protects a compliance-mandated export
    #                pipeline from accidental teardown
    #   "ABANDON" -- the sink is removed from management but keeps exporting
    #                in GCP
    deletion_policy = optional(string, "")
  })
}
