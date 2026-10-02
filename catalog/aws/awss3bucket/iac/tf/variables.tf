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
  description = "AwsS3Bucket specification"
  type = object({
    # The AWS region where the bucket will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Delete all objects (including all versions and delete markers) when the
    # bucket is destroyed, so a non-empty bucket does not block teardown.
    # Irreversible — leave false for production buckets holding real data.
    force_destroy = optional(bool, false)

    # Enable S3 Object Lock on the bucket (WORM — write once, read many).
    # Cannot be changed after creation and requires versioning. Enabling the
    # flag alone only makes the bucket lock-capable; pair it with
    # `object_lock_default_retention` to apply a default retention window to
    # every new object.
    object_lock_enabled = optional(bool, false)

    # Bucket namespace. Empty keeps the classic "global" namespace (bucket
    # names unique across all AWS accounts worldwide); "account-regional"
    # scopes the name to this account and region instead — names only need to
    # be unique within the account+region, and the bucket is addressed through
    # account-scoped endpoints. Chosen at creation; changing it replaces the
    # bucket.
    bucket_namespace = optional(string, "")

    # Bucket versioning state. When empty the bucket is left unversioned (the
    # AWS default). "Enabled" keeps every version of every object — the
    # foundation for replication, Object Lock, and protection against
    # accidental overwrite/delete. "Suspended" stops minting new versions but
    # keeps existing ones. AWS does not allow returning to the never-versioned
    # state once versioning has been enabled, so flipping Enabled back to
    # empty is rejected by AWS at apply time — use Suspended instead.
    # Versioned buckets accrue storage for every version; pair with lifecycle
    # `noncurrent_version_expiration` to bound costs.
    versioning_status = optional(string, "")

    # Default server-side encryption for new objects. When unset, AWS applies
    # its own SSE-S3 (AES256) default — set this only to move to KMS-based
    # encryption or to tune the bucket key.
    encryption = optional(object({
      # Encryption algorithm for new objects. "AES256" is SSE-S3 (S3-managed
      # keys, free — also the AWS account-wide default when this block is
      # absent). "aws:kms" is SSE-KMS: CloudTrail-audited key usage and
      # customer-controlled key policy, billed per KMS request unless the bucket
      # key is enabled. "aws:kms:dsse" is dual-layer SSE-KMS for workloads that
      # require two independent layers of encryption.
      sse_algorithm = optional(string, "")

      # Customer-managed KMS key for SSE-KMS/DSSE-KMS. Accepts a key ARN or a
      # reference to an AwsKmsKey resource. When omitted with a KMS algorithm,
      # AWS uses the account's AWS-managed `aws/s3` key — functional, but without
      # customer control over the key policy or rotation.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key_id = optional(string, "")

      # Use an S3 Bucket Key to reduce SSE-KMS costs. A bucket-level key is
      # derived from the KMS key and reused for objects, cutting KMS API calls
      # by up to 99% on KMS-encrypted buckets. Recommended whenever SSE-KMS is
      # in use; no effect under AES256.
      bucket_key_enabled = optional(bool, false)

      # Encryption types PUT requests may no longer use on this bucket.
      # "SSE-C" blocks customer-provided-key encryption (AWS blocks SSE-C on
      # new buckets by default since March 2026 — set this to state the posture
      # explicitly, or to re-permit SSE-C by managing the block without it);
      # "NONE" blocks unencrypted uploads that override the bucket default.
      blocked_encryption_types = optional(list(string), [])
    }))

    # Public access guard rails. When unset, ALL FOUR guards are enabled — the
    # secure default for every new bucket. Set this block (flipping specific
    # guards to false) only when the bucket must serve public content directly,
    # e.g. a public static website. Each guard is independent; see the block's
    # field comments for what each one controls.
    public_access_block = optional(object({
      # Reject new ACLs that grant public access (PUT requests carrying public
      # ACLs fail). Unset keeps the guard on.
      block_public_acls = optional(bool)

      # Reject bucket policies that grant public access (the policy PUT fails).
      # Unset keeps the guard on.
      block_public_policy = optional(bool)

      # Ignore all existing public ACLs when evaluating access. Unset keeps the
      # guard on.
      ignore_public_acls = optional(bool)

      # Restrict access to this bucket to AWS service principals and authorized
      # users within the bucket owner's account, even if a policy grants public
      # access. Unset keeps the guard on.
      restrict_public_buckets = optional(bool)
    }))

    # Object Ownership setting. Empty defaults to "BucketOwnerEnforced" — ACLs
    # are disabled and the bucket owner owns every object; access is managed
    # purely through policies (the AWS-recommended model). "BucketOwnerPreferred"
    # and "ObjectWriter" re-enable ACLs for legacy cross-account upload patterns
    # that depend on them.
    object_ownership = optional(string, "")

    # Canned ACL applied to the bucket. Only meaningful when `object_ownership`
    # re-enables ACLs (BucketOwnerPreferred or ObjectWriter) — enforced by
    # validation. Prefer bucket policies over ACLs; this exists for legacy
    # integrations (e.g. "log-delivery-write" for classic log delivery,
    # "aws-exec-read" for EC2 AMI bundle reads).
    acl = optional(string, "")

    # Bucket resource policy as a standard IAM policy document. This is the
    # primary access-control surface for a bucket: cross-account read/write
    # grants, TLS-only conditions, public-read statements for website buckets,
    # and service permissions (CloudFront OAC, log delivery) all live here.
    # Note: statements granting public access also require the corresponding
    # `public_access_block` guards to be relaxed, otherwise AWS blocks the
    # policy at apply time.
    policy = optional(any)

    # Minimum object size for the default transition behavior across all
    # lifecycle rules. "all_storage_classes_128K" (the AWS default) skips
    # transitioning objects smaller than 128 KB, which would otherwise cost
    # more in transition requests than they save in storage;
    # "varies_by_storage_class" applies the legacy per-class minimums.
    transition_default_minimum_object_size = optional(string, "")

    # Lifecycle rules that transition objects to cheaper storage classes and
    # expire objects/versions on a schedule. The standard levers for cost
    # control: tier logs to Glacier, expire temporary data, prune noncurrent
    # versions on versioned buckets, and abort stale multipart uploads.
    lifecycle_rules = optional(list(object({
      # Unique identifier for the rule within the bucket.
      id = string

      # Rule state. Empty defaults to "Enabled"; set "Disabled" to keep a rule
      # defined but inactive.
      status = optional(string, "")

      # Which objects the rule applies to. When unset the rule covers every
      # object in the bucket.
      filter = optional(object({
        # Key prefix, e.g. "logs/".
        prefix = optional(string, "")

        # Object tags that must all be present.
        tags = optional(map(string), {})

        # Minimum object size in bytes (exclusive). 0 means no lower bound.
        object_size_greater_than = optional(number, 0)

        # Maximum object size in bytes (exclusive). 0 means no upper bound.
        object_size_less_than = optional(number, 0)
      }))

      # Storage-class transitions for current object versions, e.g. to
      # STANDARD_IA after 30 days and DEEP_ARCHIVE after 365. Each entry names a
      # target class and when to move.
      transitions = optional(list(object({
        # Days after object creation to transition. Mutually exclusive with `date`.
        # Presence-typed because AWS allows 0 ("transition on the upload day" —
        # common for INTELLIGENT_TIERING/GLACIER_IR at ingest): unset means "use
        # date instead", an explicit 0 is a real day-zero transition.
        days = optional(number)

        # Absolute transition date in RFC3339 format (e.g. "2027-01-01T00:00:00Z").
        # Mutually exclusive with `days`.
        date = optional(string, "")

        # Target storage class.
        storage_class = string
      })), [])

      # Expiration (deletion) of current object versions.
      expiration = optional(object({
        # Days after object creation to expire.
        days = optional(number, 0)

        # Absolute expiration date in RFC3339 format. Mutually exclusive with `days`.
        date = optional(string, "")

        # Remove delete markers that have no remaining noncurrent versions
        # ("expired object delete markers") — housekeeping for versioned buckets
        # that also prune noncurrent versions.
        expired_object_delete_marker = optional(bool, false)
      }))

      # Storage-class transitions for noncurrent (superseded) versions on
      # versioned buckets.
      noncurrent_version_transitions = optional(list(object({
        # Days after an object version becomes noncurrent to transition it.
        noncurrent_days = optional(number, 0)

        # Keep this many newest noncurrent versions in their current class; the
        # transition applies only to older ones. 0 transitions all noncurrent
        # versions on schedule.
        newer_noncurrent_versions = optional(number, 0)

        # Target storage class.
        storage_class = string
      })), [])

      # Permanent deletion of noncurrent versions — the essential cost control
      # for versioned buckets.
      noncurrent_version_expiration = optional(object({
        # Days after an object version becomes noncurrent to delete it permanently.
        noncurrent_days = optional(number, 0)

        # Keep this many newest noncurrent versions regardless of age; only older
        # ones are deleted. 0 deletes all noncurrent versions on schedule.
        newer_noncurrent_versions = optional(number, 0)
      }))

      # Abort incomplete multipart uploads this many days after initiation,
      # reclaiming storage from failed uploads. 7 days is a common choice.
      abort_incomplete_multipart_upload_days = optional(number, 0)
    })), [])

    # Replication configuration. Copies objects (asynchronously) to one or more
    # destination buckets — cross-region for disaster recovery, same-region for
    # cross-account backup or log aggregation. Requires versioning on both the
    # source (enforced here) and every destination bucket.
    replication = optional(object({
      # IAM role S3 assumes to replicate. The role must trust
      # `s3.amazonaws.com`, be able to read from this bucket, and be able to
      # write (`s3:ReplicateObject`/`s3:ReplicateDelete`) to every destination.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      role_arn = string

      # Replication rules. Rules with overlapping scopes are disambiguated by
      # `priority`.
      rules = list(object({
        # Unique identifier for the rule.
        id = string

        # Priority among rules with overlapping scopes; higher wins. Must be unique
        # across rules.
        priority = optional(number, 0)

        # Rule state. Empty defaults to "Enabled".
        status = optional(string, "")

        # Which objects the rule replicates. When unset the rule covers the whole
        # bucket. Predicates combine with AND (same convention as lifecycle
        # filters; tag-scoped rules require delete-marker replication to stay
        # disabled per AWS rules — AWS validates this at apply time).
        filter = optional(object({
          # Key prefix, e.g. "important/".
          prefix = optional(string, "")

          # Object tags that must all be present.
          tags = optional(map(string), {})
        }))

        # Where and how replicas are stored.
        destination = object({
          # Destination bucket. Accepts a bucket ARN (arn:aws:s3:::name) or a
          # reference to an AwsS3Bucket resource. The destination must have
          # versioning enabled.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          bucket_arn = string

          # Destination AWS account ID for cross-account replication. Empty for
          # same-account.
          account = optional(string, "")

          # Storage class for replicas. Empty keeps each source object's class.
          # The domain mirrors AWS's PUT Bucket replication contract at provider
          # 6.58.0: the full storage-class set including EXPRESS_ONEZONE (directory
          # buckets), OUTPOSTS, SNOW, and FSX_ONTAP — minus FSX_OPENZFS, which AWS's
          # own API contract states "is not an accepted value when replicating
          # objects". OUTPOSTS/SNOW/FSX_ONTAP apply only to their matching
          # destination bucket types; AWS validates that pairing at apply time.
          storage_class = optional(string, "")

          # Transfer replica ownership to the destination account (cross-account
          # replication where the destination owns its copies). Requires `account`.
          change_replica_ownership_to_destination = optional(bool, false)

          # KMS key (in the destination region/account) used to encrypt replicas of
          # SSE-KMS objects. Required when the rule sets
          # `replicate_sse_kms_encrypted_objects`.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          replica_kms_key_id = optional(string, "")

          # Emit replication metrics (bytes pending, latency) to CloudWatch.
          # Required for — and implied by — Replication Time Control.
          metrics_enabled = optional(bool, false)

          # Enable S3 Replication Time Control: an SLA that 99.99% of objects
          # replicate within 15 minutes, with metrics and events. Extra per-GB cost;
          # requires `metrics_enabled`.
          replication_time_control_enabled = optional(bool, false)
        })

        # Replicate delete markers to the destination, keeping deletions in sync.
        # AWS default is NOT to replicate them (replicas outlive source deletions).
        delete_marker_replication = optional(bool, false)

        # Replicate existing objects (those created before the rule). Requires an
        # AWS Support-activated Batch Replication entitlement on some accounts;
        # most new setups replicate only new objects and backfill with S3 Batch
        # Operations.
        existing_object_replication = optional(bool, false)

        # Replicate changes to replica metadata (replica modification sync) — used
        # for bi-directional replication topologies.
        replicate_replica_modifications = optional(bool, false)

        # Replicate objects encrypted with SSE-KMS (skipped by default). Requires
        # `destination.replica_kms_key_id` (enforced by validation), and the
        # replication role must be able to decrypt with the source key and encrypt
        # with the destination key.
        replicate_sse_kms_encrypted_objects = optional(bool, false)
      }))
    }))

    # Static website hosting configuration. Serves bucket content over the
    # region's website endpoint (HTTP only, no TLS). For production sites,
    # front the bucket with CloudFront (which adds TLS, caching, and lets the
    # bucket stay private via Origin Access Control) and leave this unset;
    # direct website hosting suits internal or throwaway sites.
    website = optional(object({
      # Index document object suffix served for directory-style requests,
      # e.g. "index.html".
      index_document_suffix = optional(string, "")

      # Object key served on 4XX errors, e.g. "error.html".
      error_document_key = optional(string, "")

      # Redirect every request to another host — makes this a pure redirect
      # bucket (e.g. apex-domain → www).
      redirect_all_requests_to = optional(object({
        # Host to redirect to, e.g. "www.example.com".
        host_name = string

        # Protocol for the redirect. Empty preserves the request protocol.
        protocol = optional(string, "")
      }))

      # Conditional redirect rules evaluated per request (e.g. redirect a prefix
      # to another host or rewrite key prefixes).
      routing_rules = optional(list(object({
        # When the rule applies. Omit the condition entirely to apply the redirect
        # to every request; a present-but-empty condition is rejected (AWS requires
        # at least one condition field when the Condition element is sent).
        condition = optional(object({
          # Apply when the response would have this HTTP error code, e.g. "404".
          http_error_code_returned_equals = optional(string, "")

          # Apply to requests whose key starts with this prefix, e.g. "docs/".
          key_prefix_equals = optional(string, "")
        }))

        # What the redirect does.
        redirect = object({
          # Redirect target host. Empty keeps the original host.
          host_name = optional(string, "")

          # HTTP redirect code, e.g. "301". Empty uses the AWS default (301).
          http_redirect_code = optional(string, "")

          # Protocol for the redirect. Empty preserves the request protocol.
          protocol = optional(string, "")

          # Replace the matched key prefix with this value (prefix rewrite).
          # Mutually exclusive with replace_key_with.
          replace_key_prefix_with = optional(string, "")

          # Replace the entire key with this value. Mutually exclusive with
          # replace_key_prefix_with.
          replace_key_with = optional(string, "")
        })
      })), [])
    }))

    # Server access logging. Delivers request logs to another bucket (with some
    # delay). The target bucket must be in the same region and must allow log
    # delivery — either via its ACL (legacy) or, under BucketOwnerEnforced
    # ownership, via a bucket policy granting `logging.s3.amazonaws.com`.
    logging = optional(object({
      # Destination bucket for access logs. Accepts a bucket name or a reference
      # to an AwsS3Bucket resource. Must be in the same region; must not be the
      # bucket itself (logging loops amplify storage).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      target_bucket = string

      # Key prefix for log objects, e.g. "logs/my-bucket/".
      target_prefix = optional(string, "")

      # Date source for partitioned log-object keys
      # ("[prefix]/[account]/[region]/[bucket]/[yyyy]/[mm]/[dd]/..."), which makes
      # logs directly queryable by Athena partitions. "EventTime" partitions by
      # when the request happened, "DeliveryTime" by when the log was delivered.
      # Empty uses the flat (non-partitioned) key format.
      partitioned_prefix_date_source = optional(string, "")
    }))

    # CORS rules for browser-based access, required when a web application on
    # another origin reads from or writes to the bucket directly (presigned
    # uploads, font/asset serving).
    cors_rules = optional(list(object({
      # Optional identifier for the rule (shows up in error messages).
      id = optional(string, "")

      # HTTP methods the origin may use, e.g. ["GET", "PUT"].
      allowed_methods = list(string)

      # Origins allowed to make requests, e.g. ["https://example.com"]. "*"
      # allows any origin.
      allowed_origins = list(string)

      # Headers allowed in the actual request (matched against
      # Access-Control-Request-Headers in preflight).
      allowed_headers = optional(list(string), [])

      # Response headers the browser is allowed to read.
      expose_headers = optional(list(string), [])

      # Seconds the browser may cache the preflight response.
      max_age_seconds = optional(number, 0)
    })), [])

    # Event notifications for object-level events (created, removed, restored,
    # replicated...). Targets are Lambda functions, SQS queues, SNS topics, or
    # EventBridge. Note: SQS/SNS/Lambda targets must grant S3 permission to
    # deliver BEFORE the notification is configured (queue/topic policy or
    # Lambda resource permission), or AWS rejects the configuration at apply
    # time; the EventBridge arm needs no such grant.
    notification = optional(object({
      # Deliver all bucket events to Amazon EventBridge (the default event bus),
      # where rules route them anywhere. The most flexible arm and the only one
      # requiring no delivery permission setup.
      eventbridge = optional(bool, false)

      # Lambda function targets.
      lambda_functions = optional(list(object({
        # Lambda function to invoke. Accepts a function ARN or a reference to an
        # AwsLambda resource. The function must carry a resource-based permission
        # allowing `s3.amazonaws.com` to invoke it (AwsLambda's
        # `invoke_permissions` models this).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        lambda_function_arn = string

        # Event types to deliver, e.g. ["s3:ObjectCreated:*", "s3:ObjectRemoved:*"].
        events = list(string)

        # Only deliver events for keys starting with this prefix.
        filter_prefix = optional(string, "")

        # Only deliver events for keys ending with this suffix, e.g. ".jpg".
        filter_suffix = optional(string, "")
      })), [])

      # SQS queue targets. The queue policy must allow `s3.amazonaws.com` to
      # send messages (scoped to this bucket's ARN) before the notification is
      # configured.
      queues = optional(list(object({
        # SQS queue to deliver to. Accepts a queue ARN or a reference to an
        # AwsSqsQueue resource.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        queue_arn = string

        # Event types to deliver, e.g. ["s3:ObjectCreated:*"].
        events = list(string)

        # Only deliver events for keys starting with this prefix.
        filter_prefix = optional(string, "")

        # Only deliver events for keys ending with this suffix.
        filter_suffix = optional(string, "")
      })), [])

      # SNS topic targets. The topic policy must allow `s3.amazonaws.com` to
      # publish (scoped to this bucket's ARN) before the notification is
      # configured.
      topics = optional(list(object({
        # SNS topic to publish to. Accepts a topic ARN or a reference to an
        # AwsSnsTopic resource.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        topic_arn = string

        # Event types to deliver, e.g. ["s3:ObjectCreated:*"].
        events = list(string)

        # Only deliver events for keys starting with this prefix.
        filter_prefix = optional(string, "")

        # Only deliver events for keys ending with this suffix.
        filter_suffix = optional(string, "")
      })), [])
    }))

    # Default Object Lock retention applied to every new object. Requires
    # `object_lock_enabled` (enforced by validation). GOVERNANCE mode can be
    # bypassed by principals with special permission; COMPLIANCE mode cannot be
    # shortened or bypassed by anyone — including the root account — until the
    # retention period expires, so treat COMPLIANCE with care.
    object_lock_default_retention = optional(object({
      # Retention mode. GOVERNANCE allows privileged bypass
      # (s3:BypassGovernanceRetention); COMPLIANCE is immutable for everyone
      # until expiry.
      mode = string

      # Retention period in days. Mutually exclusive with `years`.
      days = optional(number, 0)

      # Retention period in years. Mutually exclusive with `days`.
      years = optional(number, 0)
    }))

    # Transfer Acceleration state. "Enabled" routes uploads/downloads through
    # CloudFront edge locations for faster long-distance transfers (extra cost
    # per GB). Once enabled, use "Suspended" to turn it off — the setting
    # cannot be removed entirely.
    acceleration_status = optional(string, "")

    # Who pays for requests and data transfer. Empty defaults to "BucketOwner".
    # "Requester" shifts request/transfer costs to the caller — common for
    # large public datasets.
    request_payer = optional(string, "")

    # Archive-tier configurations for objects stored in the INTELLIGENT_TIERING
    # storage class. Each named configuration opts a scope of objects into the
    # Archive Access and/or Deep Archive Access tiers after a period without
    # access. Only affects objects already in INTELLIGENT_TIERING (via lifecycle
    # transition or direct upload).
    intelligent_tiering_configurations = optional(list(object({
      # Name of the configuration, unique within the bucket.
      name = string

      # Configuration state. Empty defaults to "Enabled".
      status = optional(string, "")

      # Key prefix scoping the configuration. Combined (AND) with filter_tags.
      filter_prefix = optional(string, "")

      # Object tags scoping the configuration. Combined (AND) with filter_prefix.
      filter_tags = optional(map(string), {})

      # Archive tiers to enable and when. ARCHIVE_ACCESS requires at least 90
      # days without access, DEEP_ARCHIVE_ACCESS at least 180 (enforced by
      # validation, mirroring AWS limits; both max 730).
      tiers = list(object({
        # Archive tier to move objects into.
        access_tier = string

        # Consecutive days without access before objects move to this tier.
        days = optional(number, 0)
      }))
    })), [])

    # ABAC state for the bucket. "Enabled" lets IAM policies authorize requests
    # by comparing principal tags against the bucket's tags (attribute-based
    # access control); "Disabled" turns it off explicitly. Empty leaves the
    # bucket at AWS's default (disabled) without managing the setting.
    abac_status = optional(string, "")

    # Storage-class-analysis configurations. Each named configuration observes
    # access patterns for a scope of objects and (optionally) exports daily
    # findings to a bucket as CSV — the data source for choosing lifecycle
    # transition ages. Analysis observes for at least 30 days before results
    # stabilize.
    analytics_configurations = optional(list(object({
      # Name of the configuration, unique within the bucket (up to 64
      # characters per the AWS analytics-id contract).
      name = string

      # Key prefix scoping the analysis. Combined (AND) with filter_tags; leave
      # both empty to analyze the whole bucket.
      filter_prefix = optional(string, "")

      # Object tags scoping the analysis. Combined (AND) with filter_prefix.
      filter_tags = optional(map(string), {})

      # Daily CSV export of the analysis findings to a bucket. Without it the
      # findings are only visible in the S3 console. (AWS's V_1 schema and CSV
      # format are the only values the API accepts, so only the destination is
      # configurable.)
      export = optional(object({
        # Destination bucket for the CSV findings. Accepts a bucket ARN
        # (arn:aws:s3:::name) or a reference to an AwsS3Bucket resource. The
        # bucket must allow `s3.amazonaws.com` to `s3:PutObject` (same-account
        # destinations get the grant via bucket policy).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        bucket_arn = string

        # Account ID that owns the destination bucket. Empty assumes the
        # bucket owner's account.
        bucket_account_id = optional(string, "")

        # Key prefix for exported findings, e.g. "analytics/".
        prefix = optional(string, "")
      }))
    })), [])

    # Inventory report configurations. Each named configuration delivers a
    # scheduled manifest of objects and selected metadata (encryption status,
    # replication status, object-lock state, ...) to a destination bucket —
    # the audit backbone for large buckets where ListObjects is impractical.
    # The destination bucket needs a policy allowing `s3.amazonaws.com` to
    # `s3:PutObject` (AWS documents the exact statement); delivery starts
    # within 48 hours of configuration.
    inventory_configurations = optional(list(object({
      # Name of the configuration, unique within the bucket (up to 64
      # characters per the AWS inventory-id contract).
      name = string

      # Keep the configuration defined but stop generating reports. Unset means
      # active (the AWS default) — the zero value never fights the provider's
      # enabled-by-default contract.
      disabled = optional(bool, false)

      # Which object versions the report covers: "All" (every version on
      # versioned buckets) or "Current" (only the latest of each object).
      included_object_versions = string

      # Report generation schedule.
      frequency = string

      # Metadata columns to include beyond bucket/key/version. The report is the
      # cheap way to audit encryption, replication, and object-lock posture at
      # scale.
      optional_fields = optional(list(string), [])

      # Key prefix scoping which objects the report covers. Empty covers the
      # whole bucket.
      filter_prefix = optional(string, "")

      # Where and how the report files are delivered.
      destination = object({
        # Destination bucket for report files. Accepts a bucket ARN
        # (arn:aws:s3:::name) or a reference to an AwsS3Bucket resource — a bucket
        # may deliver its inventory to itself (its own ARN is derived from the
        # name: arn:aws:s3:::<name>). The destination must allow `s3.amazonaws.com`
        # to `s3:PutObject` via bucket policy.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        bucket_arn = string

        # Report file format.
        format = string

        # Account ID that owns the destination bucket. Empty assumes the bucket
        # owner's account.
        account_id = optional(string, "")

        # Key prefix for report files, e.g. "inventory/".
        prefix = optional(string, "")

        # Encrypt report files with SSE-KMS using this key. Accepts a key ARN or
        # a reference to an AwsKmsKey resource; the key policy must allow
        # `s3.amazonaws.com` to use it. Mutually exclusive with `sse_s3`.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        sse_kms_key_id = optional(string, "")

        # Encrypt report files with SSE-S3 (S3-managed keys). Mutually exclusive
        # with `sse_kms_key_id`; leave both unset to deliver with the destination
        # bucket's default encryption.
        sse_s3 = optional(bool, false)
      })
    })), [])

    # Request-metrics configurations. Each named configuration publishes
    # CloudWatch request metrics (AllRequests, 4xxErrors, FirstByteLatency,
    # ...) for a scope of objects, at CloudWatch's standard per-metric cost.
    # Whole-bucket storage metrics are free and always on; these add
    # request-level visibility for a prefix, tag set, or access point.
    metrics_configurations = optional(list(object({
      # Name of the configuration, unique within the bucket (up to 64
      # characters per the AWS metrics-id contract).
      name = string

      # Scope metrics to requests through this access point (access point ARN).
      filter_access_point_arn = optional(string, "")

      # Key prefix scoping the metrics. Combined (AND) with the other predicates.
      filter_prefix = optional(string, "")

      # Object tags scoping the metrics. Combined (AND) with the other
      # predicates.
      filter_tags = optional(map(string), {})
    })), [])

    # S3 Metadata configuration. Maintains queryable Apache Iceberg tables of
    # this bucket's object metadata in an AWS-managed table bucket: a journal
    # table records every change (uploads, deletes, metadata updates) and an
    # optional live inventory table mirrors the current object set. Query them
    # with Athena/Spark to find objects without listing the bucket. Table
    # storage and updates bill per S3 Tables pricing.
    metadata_configuration = optional(object({
      # Maintain the live inventory table (current state of every object) in
      # addition to the journal. Off keeps only the change journal.
      inventory_table_enabled = optional(bool, false)

      # Encryption for the inventory table. Only meaningful while
      # `inventory_table_enabled` is true.
      inventory_table_encryption = optional(object({
        # Encryption algorithm for the table: "AES256" (S3-managed keys) or
        # "aws:kms" (customer-managed KMS key).
        sse_algorithm = string

        # KMS key for aws:kms encryption. Accepts a key ARN or a reference to an
        # AwsKmsKey resource; the key policy must allow the S3 Tables maintenance
        # principal to use it.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        kms_key_arn = optional(string, "")
      }))

      # Expiration policy for journal-table records — the required cost control
      # for the change log (records accrue per object change).
      journal_record_expiration = object({
        # Expire journal records after `days`. Disabled keeps records forever
        # (stated explicitly — AWS requires the policy either way).
        enabled = optional(bool, false)

        # Days to retain journal records (at least 7). Only set when `enabled`.
        days = optional(number, 0)
      })

      # Encryption for the journal table.
      journal_table_encryption = optional(object({
        # Encryption algorithm for the table: "AES256" (S3-managed keys) or
        # "aws:kms" (customer-managed KMS key).
        sse_algorithm = string

        # KMS key for aws:kms encryption. Accepts a key ARN or a reference to an
        # AwsKmsKey resource; the key policy must allow the S3 Tables maintenance
        # principal to use it.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        kms_key_arn = optional(string, "")
      }))
    }))
  })
}
