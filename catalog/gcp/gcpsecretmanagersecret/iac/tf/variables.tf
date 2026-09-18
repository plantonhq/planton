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
  description = "GcpSecretManagerSecret specification"
  type = object({
    # The GCP project that owns the secret. Can be a literal project ID or a
    # reference to a GcpProject resource. If omitted, the provider's default
    # project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The secret ID — the last segment of the secret's resource name
    # (projects/{p}/secrets/{id}). Defaults to metadata.name when left
    # empty. Letters, numbers, underscores, and hyphens; at most 255
    # characters. Immutable: changing it destroys and recreates the secret
    # (and every version with it).
    secret_id = optional(string, "")

    # Region for a REGIONAL secret (e.g. "us-central1") whose payloads never
    # leave that region — the data-residency posture. Leave empty for a
    # GLOBAL secret with replication control. Immutable: a secret cannot
    # move between scopes.
    region = optional(string, "")

    # GLOBAL secrets only: where payload replicas live. Omit for automatic
    # replication (Google chooses placement — the right default when no
    # residency regime applies); the module then configures the API's `auto`
    # mode. Set user_managed to pin replicas to specific regions.
    # Immutable: replication cannot change after create.
    replication = optional(object({
      # Google chooses replica placement. Set this arm explicitly (over
      # omitting replication) only to attach a CMEK key to automatic
      # replication.
      auto = optional(object({
        # Encrypt payloads with a customer-managed KMS key. Must be a GLOBAL
        # KMS key for automatic replication; the Secret Manager service agent
        # needs roles/cloudkms.cryptoKeyEncrypterDecrypter on it.
        customer_managed_encryption = optional(object({
          # Full KMS crypto key resource path
          # (projects/{p}/locations/{l}/keyRings/{r}/cryptoKeys/{k}) — a literal or
          # a reference to a GcpKmsKey resource.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          kms_key = string
        }))
      }))

      # Pin payload replicas to specific regions — the residency-controlled
      # form.
      user_managed = optional(object({
        # The regions holding payload replicas (at least one). Reads are served
        # from the nearest replica; writes go to all.
        replicas = list(object({
          # The replica's region (e.g. "us-east1").
          location = string

          # Encrypt this replica with a customer-managed KMS key IN THE SAME
          # region. With multiple replicas, either every replica sets a key or
          # none does (the API's own rule).
          customer_managed_encryption = optional(object({
            # Full KMS crypto key resource path
            # (projects/{p}/locations/{l}/keyRings/{r}/cryptoKeys/{k}) — a literal or
            # a reference to a GcpKmsKey resource.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            kms_key = string
          }))
        }))
      }))
    }))

    # REGIONAL secrets only: encrypt payloads with a customer-managed KMS
    # key instead of Google-managed encryption. The key must be in the same
    # region as the secret, and the Secret Manager service agent needs
    # roles/cloudkms.cryptoKeyEncrypterDecrypter on it.
    customer_managed_encryption = optional(object({
      # Full KMS crypto key resource path
      # (projects/{p}/locations/{l}/keyRings/{r}/cryptoKeys/{k}) — a literal or
      # a reference to a GcpKmsKey resource.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key = string
    }))

    # User labels attached to the secret, merged with Planton's platform
    # labels (which win on key conflicts).
    labels = optional(map(string), {})

    # Annotations attached to the secret — freeform non-identifying
    # metadata (up to 64 entries; keys at most 63 characters). Unlike
    # labels, annotation values are not queryable in filters.
    annotations = optional(map(string), {})

    # Resource manager tags bound at secret creation, as
    # tagKeys/{id} -> tagValues/{id} pairs — the org-policy surface (IAM
    # conditions, CMEK enforcement). Immutable: changing tags destroys and
    # recreates the secret.
    tags = optional(map(string), {})

    # Timestamp when the secret auto-DELETES (RFC3339 UTC, e.g.
    # "2027-01-01T00:00:00Z"). Deletion removes every version — data
    # unrecoverable. At most one of expire_time or ttl.
    expire_time = optional(string, "")

    # Lifetime after which the secret auto-DELETES, as a seconds duration
    # (e.g. "7776000s" for 90 days). At most one of expire_time or ttl.
    ttl = optional(string, "")

    # Version aliases: friendly name -> version NUMBER (e.g. "prod" -> "3").
    # Consumers can then address projects/{p}/secrets/{id}/versions/prod and
    # re-pointing the alias re-targets every consumer without touching them.
    # TEMPORAL CONSTRAINT (live API truth): GCP validates aliases against
    # EXISTING versions at secret create/update ("Aliases cannot be assigned
    # to versions that don't exist"). A first apply that both seeds
    # initial_version and aliases it is therefore rejected — the version is
    # created after the secret. Deploy first, then add the alias on a
    # subsequent apply once the version exists.
    version_aliases = optional(map(string), {})

    # Delayed version destruction: when set (a seconds duration, minimum
    # "86400s" — 24h), destroying a version first DISABLES it for this
    # window, during which it can be restored — the undo buffer for fat
    # fingers. Empty destroys immediately.
    version_destroy_ttl = optional(string, "")

    # Rotation REMINDERS: GCP publishes a message to `topics` on the
    # schedule — it does not rotate anything itself; the subscriber (a
    # Cloud Function, a pipeline) performs the actual rotation. Requires
    # topics.
    rotation = optional(object({
      # Time between rotation reminders, as a seconds duration — at least
      # "3600s" (1 hour), at most "3153600000s" (100 years). Setting it
      # requires next_rotation_time.
      rotation_period = optional(string, "")

      # When the FIRST (or next) rotation reminder fires (RFC3339 UTC, e.g.
      # "2026-09-01T00:00:00Z"). GCP advances it by rotation_period after each
      # reminder.
      next_rotation_time = optional(string, "")
    }))

    # Pub/Sub topics notified on secret lifecycle events (version added,
    # rotation due, expiry approaching) — at most 10. Each entry is the full
    # topic path projects/{p}/topics/{t}: a literal or a reference to a
    # GcpPubSubTopic resource. The Secret Manager service agent needs
    # roles/pubsub.publisher on each topic.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    topics = optional(list(string), [])

    # The first secret payload, stored as version 1 at create time — so one
    # manifest yields a READABLE secret. Omit to create the container only
    # (versions added later via GCP tooling or rotation pipelines).
    initial_version = optional(object({
      # The secret payload (at most 64KiB). A secret value: the platform
      # stores it as a managed-secret reference and resolves it just-in-time
      # at deploy — it never sits in plaintext in the control plane. In
      # charts, wire it via valueFrom from a producing resource's sensitive
      # output (e.g. a generated credential) instead of a literal.
      # Immutable: changing the payload creates a NEW version through GCP
      # tooling or rotation — this field only seeds version 1.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      data = string

      # Whether version 1 starts ENABLED (default true). A disabled version
      # exists but cannot be accessed — staging a payload ahead of cut-over.
      # Both IaC engines send the value explicitly so behavior is identical
      # regardless of engine.
      enabled = optional(bool)

      # Set true when `data` is base64-encoded binary (the API stores the
      # DECODED bytes). Leave false for text payloads. Immutable.
      is_base64 = optional(bool, false)

      # What destroying this resource does to version 1 specifically:
      #   ""        -- same as "DELETE" (provider default)
      #   "DELETE"  -- the version is destroyed (or disabled-then-destroyed
      #                when the secret sets version_destroy_ttl)
      #   "DISABLE" -- the version is disabled but kept — recoverable
      #   "ABANDON" -- the version is left untouched in GCP
      # Distinct from the secret-level deletion_policy: deleting the whole
      # secret always removes every version regardless of this setting.
      deletion_policy = optional(string, "")
    }))

    # Secret-SCOPED IAM grants — typically roles/secretmanager.secretAccessor
    # to the service account of each workload that reads this secret. Grants
    # are additive (iam_member semantics): they compose safely with grants
    # made elsewhere and never clobber them.
    iam_members = optional(list(object({
      # The role to grant — most commonly roles/secretmanager.secretAccessor
      # (read payloads); also roles/secretmanager.viewer (metadata only) or
      # roles/secretmanager.secretVersionManager (add/destroy versions).
      role = string

      # The identity receiving the grant, in GCP IAM member format:
      #   serviceAccount:<email>  -- a workload identity (the most common in
      #                              IaC; reference a GcpServiceAccount — its
      #                              `member` output is exactly this value)
      #   user:<email> / group:<email> / domain:<domain>
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      member = string

      # Optional IAM Condition restricting when this grant applies. The
      # condition is part of the grant's identity: the same role with and
      # without a condition are two independent grants.
      condition = optional(object({
        # Short title identifying the condition's purpose (shown in the console).
        title = string

        # The CEL condition expression, e.g.
        # request.time < timestamp("2027-01-01T00:00:00Z").
        expression = string

        # What the condition enforces and why — for the operator auditing
        # access later.
        description = optional(string, "")
      }))
    })), [])

    # Engine-side destroy guard (default false). When true in state, any
    # plan that would delete the secret FAILS before reaching the API —
    # flip it to false explicitly, then destroy. Independent of (and
    # evaluated before) deletion_policy.
    deletion_protection = optional(bool, false)

    # Deletion policy — what happens when this resource is destroyed
    # (evaluated only after deletion_protection allows the plan):
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the secret and ALL its versions are deleted; payloads
    #                unrecoverable (version_destroy_ttl does not apply to
    #                whole-secret deletion)
    #   "PREVENT" -- destroy FAILS; belt-and-suspenders with
    #                deletion_protection for production credentials
    #   "ABANDON" -- the secret is removed from management but stays
    #                readable in GCP (IAM grants and versions intact)
    deletion_policy = optional(string, "")
  })
}
