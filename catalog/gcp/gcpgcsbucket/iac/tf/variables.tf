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
  description = "GcpGcsBucket specification"
  type = object({
    # The GCP project that owns the bucket.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the GCS bucket. Globally unique across ALL of GCP (not just
    # your project), 3-63 characters: lowercase letters, numbers, hyphens,
    # dots; must start and end with a letter or number. Deliberately
    # required: bucket names are a global namespace, so the name deserves an
    # explicit, stable choice rather than a derived default.
    # Immutable after creation.
    bucket_name = string

    # The location for the bucket: a region ("us-east1"), a predefined
    # dual-region ("NAM4"), or a multi-region ("US", "EU", "ASIA"). For a
    # custom dual-region, set a multi-region here (e.g. "US") plus
    # custom_placement_config naming the two regions. Immutable after
    # creation.
    location = string

    # Default storage class for objects written without an explicit class.
    # Values: "STANDARD" (default; hot data), "NEARLINE" (~monthly access),
    # "COLDLINE" (~quarterly), "ARCHIVE" (~yearly, cheapest at-rest).
    # Legacy classes ("MULTI_REGIONAL", "REGIONAL") remain valid API values
    # for pre-existing buckets but should not be chosen for new ones.
    # Prefer autoclass over hand-picking a cold class when access patterns
    # are uncertain. Mutable in place (existing objects keep their class).
    storage_class = optional(string, "")

    # If true, deleting the bucket deletes all contained objects first
    # (including noncurrent versions) — required to destroy any non-empty
    # bucket. Defaults to false: destroying a bucket that still holds data
    # fails instead of silently erasing it. Enable for ephemeral/derived
    # data; leave off for anything precious.
    force_destroy = optional(bool, false)

    # Enable Uniform Bucket-Level Access (UBLA): all access is controlled by
    # IAM alone and legacy object ACLs are disabled. Strongly recommended —
    # and required for buckets with hierarchical namespace or managed
    # folders. GCP's default is false (fine-grained ACLs), and UBLA can be
    # permanently locked on by GCP 90 days after enablement.
    uniform_bucket_level_access_enabled = optional(bool, false)

    # Public access prevention policy:
    #   ""          -- same as "inherited" (GCP default)
    #   "inherited" -- inherit the org policy (public access possible unless
    #                  the org forbids it)
    #   "enforced"  -- no public access, ever, regardless of IAM grants
    # Recommended: "enforced" for every bucket not deliberately public.
    # Mutable in place.
    public_access_prevention = optional(string, "")

    # Enable object versioning: overwrites and deletes keep the previous
    # version as a noncurrent object. Pair with a lifecycle rule on
    # num_newer_versions or days_since_noncurrent_time to bound storage
    # growth. Cannot be combined with hierarchical namespace. Mutable.
    versioning_enabled = optional(bool, false)

    # Autoclass: GCS automatically transitions each object between storage
    # classes based on its observed access pattern — the zero-tuning
    # alternative to hand-written SetStorageClass lifecycle rules.
    autoclass = optional(object({
      # Enable autoclass. Objects start in STANDARD and transition to colder
      # classes as they go unread; a read promotes the object back to
      # STANDARD. Toggling autoclass is allowed but restricted by GCP to
      # once per 24 hours. Required inside autoclass: declaring the block takes
      # the feature under management, and the switch says which way. An explicit
      # false is a real statement (it turns a previously enabled bucket's
      # autoclass off), which is why the switch carries presence.
      enabled = bool

      # The coldest class autoclass may transition objects into:
      #   ""         -- GCP default ("NEARLINE")
      #   "NEARLINE" -- stop at NEARLINE
      #   "ARCHIVE"  -- allow transitions all the way to ARCHIVE (also enables
      #                 COLDLINE as an intermediate step)
      terminal_storage_class = optional(string, "")
    }))

    # Lifecycle rules for automatic object management (deletion, storage
    # class transitions, aborting stale multipart uploads). Up to 100 rules;
    # all conditions within a rule must match (logical AND).
    lifecycle_rules = optional(list(object({
      # Action to take when the condition matches.
      action = object({
        # Action type:
        #   "Delete"          -- delete the matching object (or its noncurrent
        #                        version when the condition targets versions)
        #   "SetStorageClass" -- transition the object to storage_class
        #   "AbortIncompleteMultipartUpload" -- abort multipart uploads older
        #                        than the condition's age (reclaims hidden
        #                        storage from abandoned uploads)
        type = string

        # Target storage class for SetStorageClass actions, e.g. "NEARLINE",
        # "COLDLINE", "ARCHIVE".
        storage_class = optional(string, "")
      })

      # Condition selecting the objects the action applies to. All specified
      # criteria must match (logical AND).
      condition = object({
        # Minimum age of the object in days. A set 0 matches all objects.
        age_days = optional(number)

        # Match objects created before this date (RFC 3339 date, "2026-01-01").
        created_before = optional(string, "")

        # Match by version state (requires versioning):
        #   ""        -- any state (default)
        #   "LIVE"     -- only the current version
        #   "ARCHIVED" -- only noncurrent versions
        #   "ANY"      -- both
        with_state = optional(string, "")

        # Match objects currently in any of these storage classes, e.g.
        # ["STANDARD", "NEARLINE"]. Legacy classes ("MULTI_REGIONAL",
        # "REGIONAL", "DURABLE_REDUCED_AVAILABILITY") are valid here for
        # matching long-lived objects.
        matches_storage_class = optional(list(string), [])

        # Match objects whose name starts with any of these prefixes.
        matches_prefix = optional(list(string), [])

        # Match objects whose name ends with any of these suffixes.
        matches_suffix = optional(list(string), [])

        # Match noncurrent versions with at least this many newer versions —
        # the standard "keep the last N versions" cleanup (requires versioning).
        # A set 0 matches every noncurrent version.
        num_newer_versions = optional(number)

        # Match noncurrent versions that became noncurrent at least this many
        # days ago (requires versioning). A set 0 matches immediately.
        days_since_noncurrent_time = optional(number)

        # Match noncurrent versions that became noncurrent before this date
        # (RFC 3339 date; requires versioning).
        noncurrent_time_before = optional(string, "")

        # Match objects whose Custom-Time metadata is at least this many days
        # old. A set 0 matches any object with Custom-Time set.
        days_since_custom_time = optional(number)

        # Match objects whose Custom-Time metadata is before this date
        # (RFC 3339 date).
        custom_time_before = optional(string, "")

        # Match objects LARGER than this many bytes. Combine with
        # size_below_bytes for a size band (e.g. transition only large
        # artifacts to cold storage).
        size_above_bytes = optional(number)

        # Match objects SMALLER than this many bytes.
        size_below_bytes = optional(number)
      })
    })), [])

    # Retention policy for WORM (write once, read many) compliance: objects
    # cannot be deleted or replaced until they reach the retention age.
    retention_policy = optional(object({
      # Minimum object retention period in seconds (max ~100 years,
      # 3155760000s). Objects cannot be deleted or overwritten until they are
      # this old. Mutable while unlocked; can only be increased once locked.
      retention_period_seconds = number

      # Lock the retention policy. IRREVERSIBLE: a locked policy can never be
      # removed or shortened, and the bucket cannot be deleted until every
      # object passes its retention period. Attempting to unlock forces bucket
      # re-creation. Validate the policy against real workloads before locking.
      is_locked = optional(bool, false)
    }))

    # How long deleted objects remain recoverable (and billed) before being
    # permanently removed. GCP's default is 7 days (604800s) even when this
    # block is omitted. Set 0 to disable soft delete entirely — common for
    # high-churn scratch buckets where the 7-day tail is pure cost.
    soft_delete_policy = optional(object({
      # How long deleted objects remain recoverable, in seconds. GCP default
      # is 604800 (7 days); 0 disables soft delete; otherwise the value must
      # be between 7 and 90 days. Soft-deleted storage is billed at the
      # object's storage class rate.
      retention_duration_seconds = optional(number)
    }))

    # Default customer-managed encryption key (CMEK) for objects written
    # without an explicit key. The GCS service agent needs
    # roles/cloudkms.cryptoKeyEncrypterDecrypter on the key. If omitted,
    # Google-managed encryption is used. Accepts the fully qualified crypto
    # key path or a reference to a GcpKmsKey resource. Mutable in place
    # (existing objects keep the key they were written with).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_name = optional(string, "")

    # Requester pays: the caller's project (not the bucket owner) is billed
    # for data access and egress. Useful for widely shared public datasets.
    # Mutable in place.
    requester_pays = optional(bool, false)

    # Automatically place an event-based hold on every new object — the
    # object cannot be deleted until the hold is released AND its retention
    # period (measured from release) expires. For event-driven compliance
    # like "retain 3 years after account closure". Mutable in place.
    default_event_based_hold = optional(bool, false)

    # Enable per-object retention: individual objects can carry their own
    # retention configuration, independent of the bucket-level policy.
    # Can only be set at creation — immutable.
    enable_object_retention = optional(bool, false)

    # Static website serving configuration (main page and 404 page) for
    # requests via the bucket's website endpoint or a load balancer backend
    # bucket. For production HTTPS sites, front the bucket with the L7 load
    # balancer family (GcpBackendBucket + URL map + HTTPS proxy).
    website = optional(object({
      # Object served for directory requests, e.g. "index.html".
      main_page_suffix = optional(string, "")

      # Object served when the requested path does not exist, e.g. "404.html".
      not_found_page = optional(string, "")
    }))

    # CORS rules for direct cross-origin browser access (fonts, direct
    # uploads, XHR downloads). Not needed when all access goes through a
    # load balancer or same-origin paths.
    cors_rules = optional(list(object({
      # Origins allowed to make cross-origin requests, e.g.
      # "https://example.com". "*" allows any origin.
      origins = list(string)

      # HTTP methods allowed, e.g. ["GET", "HEAD"]. "*" allows any method.
      methods = list(string)

      # Response headers browsers are allowed to read.
      response_headers = optional(list(string), [])

      # How long (seconds) browsers may cache the preflight response.
      max_age_seconds = optional(number, 0)
    })), [])

    # Usage/access log delivery to another GCS bucket (classic storage
    # access logs). Most observability needs are better served by Cloud
    # Audit Logs; use this for legacy tooling that parses access-log files.
    logging = optional(object({
      # Destination bucket that receives the log objects. Accepts a literal
      # bucket name or a reference to another GcpGcsBucket resource.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      log_bucket = string

      # Prefix for log object names. Defaults to this bucket's name.
      log_object_prefix = optional(string, "")
    }))

    # Custom dual-region placement: exactly two regions (e.g. ["US-EAST1",
    # "US-WEST1"]) that must belong to the multi-region set in `location`.
    # Gives dual-region durability with region pinning. Immutable after
    # creation.
    custom_placement_config = optional(object({
      # Exactly two regions forming the custom dual-region, e.g.
      # ["US-EAST1", "US-WEST1"]. Both must belong to the multi-region set in
      # `location`. Immutable after creation.
      data_locations = list(string)
    }))

    # Recovery point objective for dual- and multi-region buckets:
    #   ""            -- provider default ("DEFAULT")
    #   "DEFAULT"     -- asynchronous replication with no SLA on lag
    #   "ASYNC_TURBO" -- turbo replication (15-minute RPO SLA; dual-region
    #                    only, additional cost)
    # Mutable in place.
    rpo = optional(string, "")

    # Enable hierarchical namespace (HNS): the bucket gets real folder
    # semantics with atomic folder renames — required for Hadoop/Spark-style
    # workloads that rename directories. Requires uniform bucket-level
    # access and cannot be combined with object versioning. Immutable —
    # only settable at creation.
    hierarchical_namespace_enabled = optional(bool, false)

    # User-defined labels attached to the bucket, for cost attribution and
    # fleet queries. Merged with Planton's platform labels (which win on key
    # conflicts). Mutable in place.
    labels = optional(map(string), {})

    # Additive IAM grants on this bucket. Each entry grants one role to one
    # member and composes safely with grants made by other tools or charts —
    # removal subtracts only that exact (role, member) pair.
    #
    # Common roles:
    #   roles/storage.objectViewer  -- read objects
    #   roles/storage.objectAdmin   -- full object control (no bucket admin)
    #   roles/storage.admin         -- full bucket + object control
    #
    # Public access: grant roles/storage.objectViewer to "allUsers" (also
    # requires public_access_prevention to be "inherited" and the org policy
    # to allow it).
    iam_members = optional(list(object({
      # The role to grant, e.g. "roles/storage.objectViewer",
      # "roles/storage.objectAdmin", "roles/storage.admin", or a custom
      # role's fully-qualified name.
      role = string

      # The identity receiving the grant, in GCP IAM member format:
      #   serviceAccount:<email>  -- a service account (the most common in IaC;
      #                              reference a GcpServiceAccount resource —
      #                              its `member` output is exactly this value)
      #   user:<email> / group:<email> / domain:<domain>
      #   allUsers / allAuthenticatedUsers -- public access (grant with care)
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      member = string

      # Optional IAM Condition restricting when this grant applies (e.g. only
      # objects under a prefix, or before an expiry date). The condition is
      # part of the grant's identity: the same role with and without a
      # condition are two independent grants.
      condition = optional(object({
        # Short human-readable title identifying the condition's intent,
        # e.g. "reports-prefix-only".
        title = string

        # The CEL condition expression, e.g.
        # resource.name.startsWith("projects/_/buckets/b/objects/reports/").
        expression = string

        # Optional longer explanation of what the condition does.
        description = optional(string, "")
      }))
    })), [])

    # Network-layer IP filtering: restrict which public CIDR ranges and
    # which VPC networks may reach the bucket at all, before IAM is even
    # evaluated. Defense-in-depth for data-exfiltration control — IAM
    # decides WHO, the IP filter decides FROM WHERE. Mutable in place.
    ip_filter = optional(object({
      # The filter mode:
      #   "Enabled"  -- only the listed sources may reach the bucket
      #   "Disabled" -- filter retained but inactive (all sources allowed)
      mode = string

      # Public internet sources allowed to access the bucket, as IPv4/IPv6
      # CIDR ranges.
      public_network_source = optional(object({
        # Public IPv4/IPv6 CIDR ranges allowed to access the bucket,
        # e.g. "203.0.113.0/24".
        allowed_ip_cidr_ranges = list(string)
      }))

      # VPC networks allowed to access the bucket, each with its own CIDR
      # allowlist.
      vpc_network_sources = optional(list(object({
        # The VPC network, in the form projects/{project}/global/networks/{name}.
        # Accepts a literal path or a reference to a GcpVpcNetwork resource
        # (its network_id output is exactly this value).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        network = string

        # IPv4/IPv6 CIDR ranges within this network allowed to access the
        # bucket.
        allowed_ip_cidr_ranges = list(string)
      })), [])

      # Allow VPC network sources that belong to a different organization
      # than the bucket.
      allow_cross_org_vpcs = optional(bool, false)

      # Exempt Google service agents (the identities GCP services act as —
      # e.g. the Storage transfer agent) from the IP filter, so managed
      # integrations keep working when the filter is Enabled.
      allow_all_service_agent_access = optional(bool, false)
    }))

    # Encryption-type enforcement for NEW objects: restrict which encryption
    # mechanisms (Google-managed, customer-managed KMS, customer-supplied)
    # may be used when writing objects into this bucket. Applies to new
    # objects only — existing objects keep their encryption. Mutable in
    # place.
    encryption_enforcement = optional(object({
      # Restriction for Google-managed encryption keys (GMEK) — the default
      # encryption objects get when no KMS key applies.
      google_managed_restriction_mode = optional(string, "")

      # Restriction for customer-managed encryption keys (CMEK, Cloud KMS).
      customer_managed_restriction_mode = optional(string, "")

      # Restriction for customer-supplied encryption keys (CSEK — raw keys
      # provided per request).
      customer_supplied_restriction_mode = optional(string, "")
    }))

    # Deletion policy — what happens when this resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the bucket is deleted (subject to force_destroy)
    #   "PREVENT" -- destroy FAILS; a guard rail for buckets that must
    #                never be removed by automation
    #   "ABANDON" -- the bucket is removed from management but left
    #                running in GCP (an orphan by design — reserve for
    #                deliberate hand-offs)
    # Mutable in place.
    deletion_policy = optional(string, "")

    # Folders to create inside the bucket — REAL directories with atomic
    # rename semantics, available only on hierarchical-namespace buckets
    # (hierarchical_namespace_enabled, which itself requires uniform
    # bucket-level access and is create-time only). Parent folders must be
    # listed explicitly: creating "logs/2026/" requires a "logs/" entry too
    # — the Storage API does not auto-create missing parents.
    # The kind-level deletion_policy applies to each folder resource as
    # well as the bucket (PREVENT guards them; ABANDON leaves them behind).
    folders = optional(list(object({
      # The folder path, WITH the trailing slash the Storage API requires:
      # "logs/", "logs/2026/", "a-b/d-f/". Renaming is atomic on HNS buckets,
      # but through this spec a name change is destroy-and-recreate (the
      # path is the resource's identity).
      #
      # The API does NOT auto-create missing parents, so nested paths need
      # every ancestor listed as its own entry ("logs/2026/" needs "logs/"
      # too) — the modules then create parents before children and delete
      # children before parents. Nesting is capped at 5 levels.
      name = string

      # If true, destroying this folder first deletes every object under it
      # (the provider sweeps the prefix client-side, in parallel), then any
      # sub-folders. Defaults to false: destroying a non-empty folder fails
      # instead of silently erasing data — same safe posture as the bucket's
      # own force_destroy, decided per folder.
      force_destroy = optional(bool, false)
    })), [])

    # Managed folders — permission anchor points inside the bucket. Unlike
    # plain folders they exist on ANY bucket with uniform bucket-level
    # access (hierarchical namespace not required) and their purpose is
    # prefix-scoped IAM: grant a role on "reports/" without granting it on
    # the whole bucket. Grants on a managed folder are made through the
    # managed-folder IAM surface (composition), not through this kind's
    # bucket-level iam_members. The kind-level deletion_policy applies to
    # each managed folder as well as the bucket.
    managed_folders = optional(list(object({
      # The managed-folder path, WITH the trailing slash the Storage API
      # requires: "reports/", "reports/2026/". The path is the resource's
      # identity — changing it is destroy-and-recreate. Managed folders are
      # independent prefix anchors, not a hierarchy — "reports/2026/" does
      # not need a "reports/" managed folder to exist.
      name = string

      # If true, the managed folder can be destroyed even while objects live
      # under its prefix. Unlike a plain folder's force_destroy this is
      # SERVER-side and non-destructive to data: the objects survive and
      # simply stop being covered by the managed folder's IAM. Defaults to
      # false (destroy fails while the prefix is non-empty).
      force_destroy = optional(bool, false)
    })), [])

    # Pub/Sub notification configurations: GCS publishes an event to the
    # named topic whenever objects change in this bucket — the standard
    # trigger surface for event-driven pipelines (Cloud Run, Cloud
    # Functions, Eventarc all consume the topic downstream).
    #
    # HARD PREREQUISITE the API enforces at create time: the project's GCS
    # service agent — service-{project_number}@gs-project-accounts.iam.gserviceaccount.com,
    # where {project_number} is this kind's project_number output — must
    # hold roles/pubsub.publisher on the topic (or project-wide). The agent
    # is NOT granted automatically; compose the grant alongside this
    # resource (GcpProjectIamMember for project scope, or the topic's own
    # IAM surface) exactly like the CMEK key grant for kms_key_name.
    notifications = optional(list(object({
      # The Pub/Sub topic that receives the events, as the FULLY-QUALIFIED
      # resource path "projects/{project}/topics/{name}" — the API accepts
      # only this form (a bare topic name is rejected). Reference a
      # GcpPubSubTopic — its topic_id output is exactly this value. The
      # topic may live in a different project than the bucket.
      #
      # Before creation the project's GCS service agent must hold
      # roles/pubsub.publisher on this topic — see the notifications field
      # comment on the spec for the composition pattern.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      topic = string

      # The payload attached to each event message:
      #   "JSON_API_V1" -- the object's full JSON representation (the common
      #                    choice; consumers read metadata without a GET)
      #   "NONE"        -- no payload; event attributes only (cheapest)
      payload_format = string

      # Which object events publish to the topic. Empty means ALL event
      # types. Values:
      #   "OBJECT_FINALIZE"        -- a new object (or new version) is written
      #   "OBJECT_METADATA_UPDATE" -- an object's metadata changed
      #   "OBJECT_DELETE"          -- an object (or version) was deleted or
      #                               overwritten
      #   "OBJECT_ARCHIVE"         -- a live version became noncurrent
      #                               (versioned buckets only)
      event_types = optional(list(string), [])

      # Only objects whose names start with this prefix generate events —
      # scope a busy bucket's feed to one subtree (e.g. "uploads/").
      object_name_prefix = optional(string, "")

      # Custom key:value attributes stamped onto every event message the
      # topic receives — routing hints for subscribers (e.g. env: prod).
      custom_attributes = optional(map(string), {})
    })), [])
  })
}
