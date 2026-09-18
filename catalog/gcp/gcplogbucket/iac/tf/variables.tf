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
  description = "GcpLogBucket specification"
  type = object({
    # Where the bucket lives. Omit entirely for a project bucket in the
    # provider's default project — the common case. (Folder, organization,
    # and billing-account scopes take folder_id / organization_id /
    # billing_account instead.)
    scope = optional(object({
      # Project bucket: the owning project — a literal project ID or a
      # reference to a GcpProject resource.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      project_id = optional(string, "")

      # Folder bucket: the folder ID (numeric, with or without the "folders/"
      # prefix). ADOPT-only: the Logging API creates new custom buckets only
      # under projects.
      folder_id = optional(string, "")

      # Organization bucket: the numeric organization ID. ADOPT-only.
      organization_id = optional(string, "")

      # Billing-account bucket: the billing account ID
      # (e.g. 012345-6789AB-CDEF01). ADOPT-only.
      billing_account = optional(string, "")
    }))

    # The bucket ID (the last segment of the bucket resource name), e.g.
    # "audit-logs" — or "_Default" to ADOPT the project's built-in default
    # bucket and manage its retention. Changing it REPLACES the bucket.
    bucket_id = string

    # The bucket location (e.g. "global", "us-central1", "eu"). Immutable —
    # changing it REPLACES the bucket (and its stored logs). "global" is the
    # default and the right choice unless data residency demands a region.
    # Both IaC engines send the value explicitly.
    location = optional(string, "")

    # Why this bucket exists and what lands in it.
    description = optional(string, "")

    # How many days log entries are kept before automatic deletion. GCP's
    # default is 30; up to 3650 (10 years). On a LOCKED bucket the retention
    # period can no longer be changed. Both IaC engines send the value
    # explicitly so the spec default (30) is what GCP applies rather than a
    # silently different server-side state.
    retention_days = optional(number, 0)

    # Lock the bucket (project scope only). LOCKING IS ONE-WAY: a locked
    # bucket's retention policy can never be changed or unlocked again, and
    # the bucket can only be deleted once every entry in it has aged out of
    # retention. The compliance posture — enable deliberately.
    locked = optional(bool, false)

    # Enable Log Analytics on the bucket (project scope only): entries
    # become queryable with SQL from the Log Analytics UI and BigQuery
    # (via linked_bigquery_dataset). ONE-WAY per the provider: analytics
    # cannot be disabled once enabled. Left unset, nothing is sent — GCP
    # keeps its own default (disabled) — because analytics enablement is an
    # atomic, separate API operation the provider performs only on explicit
    # configuration.
    enable_analytics = optional(bool)

    # Encrypt the bucket with a customer-managed KMS key (CMEK) instead of
    # Google-managed encryption. The full crypto key resource name
    # (projects/{p}/locations/{l}/keyRings/{r}/cryptoKeys/{k}) — a literal
    # or a reference to a GcpKmsKey resource. ONE-WAY per the provider:
    # CMEK cannot be disabled once set (rotating to a DIFFERENT key is
    # allowed). PREREQUISITE: grant the Logging service account
    # roles/cloudkms.cryptoKeyEncrypterDecrypter on the key first (find the
    # account with `gcloud logging settings describe`), or bucket creation
    # fails.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    cmek_kms_key = optional(string, "")

    # Indexed LogEntry fields for faster targeted queries — at most 20 (the
    # provider's own cap).
    index_configs = optional(list(object({
      # The LogEntry field path to index, e.g. "jsonPayload.request.status".
      field_path = string

      # The indexed data type: INDEX_TYPE_STRING or INDEX_TYPE_INTEGER (the
      # Logging API's documented values — server-validated).
      type = string
    })), [])

    # Log VIEWS on the bucket: named, filtered slices that can be granted
    # independently (roles/logging.viewAccessor on a view shows a reader
    # only the entries matching its filter). Each becomes a
    # google_logging_log_view resource on this bucket.
    log_views = optional(list(object({
      # The view ID (the last segment of the view resource name). Changing it
      # REPLACES the view (views are immutable in name and home; only filter
      # and description update in place).
      view_id = string

      # The filter selecting which of the bucket's entries this view exposes.
      # Empty exposes every entry in the bucket.
      #
      # View filters speak a RESTRICTED grammar — NOT the general log filter
      # language: the API accepts only restrictions on log source
      # (source()), resource type (resource.type=...), apphub fields,
      # user-defined labels, and log ID (LOG_ID(...)). Severity is NOT a
      # legal view dimension — GCP rejects it at create with "Invalid view
      # filter" (live-verified), even though the same expression is legal in
      # sinks and log-based metrics.
      filter = optional(string, "")

      # What this view is for and who should be granted it.
      description = optional(string, "")
    })), [])

    # Link a read-only BigQuery dataset to the bucket so its entries are
    # queryable from BigQuery directly (requires enable_analytics: true).
    linked_bigquery_dataset = optional(object({
      # The link ID — becomes the BigQuery DATASET ID, so it must be a valid
      # dataset name (letters, numbers, underscores). IMMUTABLE: every field
      # of a linked dataset (including the description) is create-time-only;
      # any change REPLACES the link.
      link_id = string

      # What the linked dataset is for. Immutable like the rest of the link.
      description = optional(string, "")
    }))

    # Folder/organization scopes only: the scope's LOGGING SETTINGS —
    # default-sink disable, default CMEK, and default storage location for
    # buckets created under the scope. A singleton that always exists in
    # GCP: creating this surface ADOPTS it, and destroying it is a
    # state-only no-op (the settings object cannot be deleted).
    scope_settings = optional(object({
      # Disable the scope's automatic _Default sink so logs of NEW child
      # projects are not copied into their _Default buckets — the
      # centralized-logging posture (pair it with an aggregating
      # GcpLoggingSink at the same scope).
      disable_default_sink = optional(bool, false)

      # The default CMEK key applied to buckets created under the scope. The
      # full crypto key resource name — a literal or a reference to a
      # GcpKmsKey resource. The same one-way and grant-first caveats as the
      # bucket-level cmek_kms_key apply.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key = optional(string, "")

      # The default storage location for buckets created under the scope
      # (e.g. "us-central1", "eu").
      storage_location = optional(string, "")
    }))

    # Deletion policy — what happens when this resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the bucket (and its stored entries) is deleted; the
    #                built-in _Default/_Required buckets are undeletable and
    #                are simply removed from management
    #   "PREVENT" -- destroy FAILS; protects compliance-mandated log
    #                storage from accidental teardown
    #   "ABANDON" -- the bucket is removed from management but keeps
    #                storing logs in GCP
    # Also applied to the bucket's log views and linked dataset.
    deletion_policy = optional(string, "")
  })
}
