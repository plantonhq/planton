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
  description = "AwsS3ObjectSet specification"
  type = object({
    # The AWS region where the target bucket lives. Must match the bucket's own
    # region — S3 PutObject is a regional API.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # The target S3 bucket for every object in the set.
    # Can be a literal bucket name or a reference to an AwsS3Bucket component
    # (resolved from status.outputs.bucket_id). The literal arm also accepts an
    # access-point ARN or an S3 Express directory-bucket name for buckets
    # managed outside this catalog.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    bucket = string

    # The objects to upload. At least one. Each object's `key` must be unique
    # within the set — both engines key the underlying provider resources by it.
    objects = list(object({
      # The S3 object key — the object's full path within the bucket and its
      # identity in this set. Changing the key replaces the object.
      # Examples: "config/app.json", "assets/logo.png", "data/seed.csv"
      key = string

      # Inline UTF-8 text content.
      # Suitable for configuration files, JSON, YAML, HTML, and other text.
      content = optional(string)

      # Base64-encoded binary content.
      # Suitable for images, zip archives, or any binary payload small enough
      # to carry in a manifest.
      content_base64 = optional(string)

      # Server-side copy of an existing S3 object (any size — the bytes move
      # inside S3, never through the deploy host). Suitable for promoting
      # build artifacts between buckets, seeding environments from a golden
      # bucket, or re-homing objects with new placement/encryption. The copy
      # is taken at deploy time; it does not track later changes to the
      # source object.
      copy_from = optional(object({
        # The bucket holding the source object.
        # Can be a literal bucket name or a reference to an AwsS3Bucket component
        # (resolved from status.outputs.bucket_id). The literal arm also accepts
        # an S3 access-point ARN (`arn:aws:s3:<region>:<account>:accesspoint/
        # <name>`) for sources reached through an access point. The deploying
        # principal needs s3:GetObject on the source.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        source_bucket = string

        # The key of the source object within `source_bucket`.
        # Example: "releases/v1.2.3/app.zip"
        source_key = string

        # Replace the destination object's user metadata and content headers
        # instead of preserving the source's (S3's REPLACE metadata directive).
        # When true, the owning object's `metadata`, `content_type`,
        # `cache_control`, `content_encoding`, `content_disposition`,
        # `content_language`, and `website_redirect` are written to the copy;
        # when false (the default) those fields must stay unset and the copy
        # keeps everything the source object carried.
        replace_metadata = optional(bool, false)

        # Copy only if the source object's current ETag matches this value
        # (a precondition evaluated by S3 at copy time; a failed precondition
        # fails the deploy). Guards promotions against a source that changed
        # since the ETag was recorded.
        copy_if_match = optional(string, "")

        # Copy only if the source object's current ETag does NOT match this
        # value.
        copy_if_none_match = optional(string, "")

        # Copy only if the source object was modified after this RFC 3339
        # timestamp (e.g. "2026-08-01T00:00:00Z").
        copy_if_modified_since = optional(string, "")

        # Copy only if the source object has NOT been modified since this RFC
        # 3339 timestamp.
        copy_if_unmodified_since = optional(string, "")

        # The Expires header stored with the copied object, as an RFC 3339
        # timestamp — the date after which the content is considered stale by
        # caches. Only the copy path offers this header (the provider's
        # content-object resource carries no expires argument).
        expires = optional(string, "")

        # Confirms the copier pays request and data-transfer costs when the
        # SOURCE bucket is a Requester Pays bucket. The only accepted value is
        # "requester" — S3 rejects copies from Requester Pays buckets without
        # this acknowledgment.
        request_payer = optional(string, "")
      }))

      # The MIME content type stored with the object and returned as the
      # Content-Type header on every GET. Set it correctly for anything a
      # browser will fetch (e.g. "text/html", "application/json", "image/png") —
      # the default is the generic binary type, which browsers download instead
      # of rendering.
      # Default: application/octet-stream
      content_type = optional(string)

      # The Cache-Control header stored with the object, governing CDN and
      # browser caching. Examples: "max-age=86400" (cache for 24h),
      # "no-cache" (revalidate every time), "public, max-age=31536000, immutable"
      # (fingerprinted static assets).
      cache_control = optional(string, "")

      # The Content-Encoding header stored with the object. Set this when the
      # content is pre-compressed (e.g. "gzip", "br") so clients decompress it
      # transparently. The content itself must already be encoded — S3 does not
      # compress.
      content_encoding = optional(string, "")

      # The Content-Disposition header stored with the object, controlling
      # whether browsers render inline or download. Examples: "inline",
      # "attachment; filename=\"report.pdf\"".
      content_disposition = optional(string, "")

      # The Content-Language header stored with the object — the natural
      # language(s) of the content, e.g. "en-US".
      content_language = optional(string, "")

      # User-defined metadata stored with the object and returned as
      # x-amz-meta-* headers. Keys must be lowercase (the S3 API lowercases them
      # on storage; requiring lowercase here keeps the manifest identical to
      # what reads back and avoids phantom drift). Note: metadata is immutable
      # in S3 — changing it rewrites the object.
      metadata = optional(map(string), {})

      # The x-amz-website-redirect-location value: when the bucket is configured
      # for static website hosting, a GET of this object's key redirects to the
      # target instead of serving content — either another key in the same site
      # ("/new-page.html") or an absolute URL ("https://example.com/moved").
      # Stored but inert on buckets without website hosting.
      website_redirect = optional(string, "")

      # The storage class for the object, trading retrieval latency/cost against
      # storage cost. Unset uses STANDARD.
      #
      # - "STANDARD": general purpose; the default.
      # - "STANDARD_IA" / "ONEZONE_IA": infrequent access — cheaper storage,
      #   per-GB retrieval fee, 30-day minimum billing (ONEZONE_IA is single-AZ).
      # - "INTELLIGENT_TIERING": automatic tiering by access pattern; the safe
      #   choice when access patterns are unknown.
      # - "GLACIER_IR": archive with millisecond access; 90-day minimum.
      # - "GLACIER" / "DEEP_ARCHIVE": archive requiring asynchronous restore
      #   before reads (minutes-to-hours / hours); 90/180-day minimums. Rarely
      #   right for declaratively managed objects, which are typically read.
      # - "REDUCED_REDUNDANCY": legacy, not recommended.
      # - "EXPRESS_ONEZONE": S3 Express — valid only when the bucket is a
      #   directory bucket (literal bucket arm).
      # - "OUTPOSTS" / "SNOW" / "FSX_OPENZFS" / "FSX_ONTAP": valid only when the
      #   literal bucket arm targets the matching special infrastructure
      #   (Outposts access point, Snow device, FSx-backed access point).
      storage_class = optional(string, "")

      # Per-object server-side encryption OVERRIDE. Leave unset to inherit the
      # bucket's default encryption (the recommended posture — configure it on
      # AwsS3Bucket). Set only for individual objects that must diverge.
      #
      # - "AES256": SSE-S3 (S3-managed keys).
      # - "aws:kms": SSE-KMS with the key in `kms_key` (or the account's
      #   aws/s3 key when no key is given).
      # - "aws:kms:dsse": dual-layer SSE-KMS (two independent encryption layers;
      #   compliance-driven).
      # - "aws:fsx": valid only when the literal bucket arm targets an
      #   FSx-backed access point.
      server_side_encryption = optional(string, "")

      # The KMS key for SSE-KMS encryption of this object, as a key ARN.
      # Can be a literal ARN or a reference to an AwsKmsKey component. Setting a
      # key implies "aws:kms" encryption when `server_side_encryption` is unset;
      # combining it with "AES256" is contradictory and rejected. The deploying
      # principal needs kms:GenerateDataKey on the key.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key = optional(string, "")

      # Whether this object uses an S3 Bucket Key for SSE-KMS, batching KMS
      # requests to cut KMS costs. Only meaningful with "aws:kms" encryption.
      # Unset inherits the bucket's default; an explicit false opts this object
      # out even when the bucket enables bucket keys.
      bucket_key_enabled = optional(bool)

      # The checksum algorithm S3 uses to verify upload integrity and store an
      # additional object checksum (retrievable via GetObjectAttributes).
      # "CRC64NVME" is AWS's current full-object recommendation; the SHA variants
      # suit compliance regimes that mandate them. Unset relies on the default
      # TLS/MD5 integrity protection only.
      checksum_algorithm = optional(string, "")

      # Object Lock retention mode for this object version. Requires the BUCKET
      # to have Object Lock enabled (an immutable create-time bucket setting)
      # and must be paired with `object_lock_retain_until_date`.
      #
      # - "GOVERNANCE": privileged principals (s3:BypassGovernanceRetention) can
      #   override the retention.
      # - "COMPLIANCE": nobody — including the root user — can shorten it.
      #   The object version genuinely cannot be deleted until the date passes;
      #   destroy will fail until then.
      object_lock_mode = optional(string, "")

      # The date until which this object version is retained under
      # `object_lock_mode`, as an RFC 3339 timestamp (e.g.
      # "2027-01-01T00:00:00Z"). Required with, and only valid with,
      # `object_lock_mode`.
      object_lock_retain_until_date = optional(string, "")

      # Object Lock legal hold for this object version — an indefinite deletion
      # block independent of retention mode, toggled on ("ON") and off ("OFF")
      # by principals with s3:PutObjectLegalHold. Requires an Object
      # Lock-enabled bucket.
      object_lock_legal_hold_status = optional(string, "")

      # Canned ACL for this object. Leave unset for every modern bucket:
      # AwsS3Bucket's default ownership posture (BucketOwnerEnforced) DISABLES
      # object ACLs, and setting one fails at apply time with
      # AccessControlListNotSupported. Only buckets whose object_ownership is
      # relaxed to BucketOwnerPreferred or ObjectWriter accept ACLs — a legacy
      # pattern; prefer bucket policies for public access.
      acl = optional(string, "")

      # Allow this object to be deleted even while under GOVERNANCE-mode Object
      # Lock retention or a legal hold, by sending the governance-bypass flag
      # with the delete (the deploying principal needs
      # s3:BypassGovernanceRetention). ONLY valid on Object Lock-enabled
      # buckets — S3 rejects the bypass flag on regular buckets, failing the
      # destroy. Not needed for ordinary versioned-bucket cleanup: destroying an
      # object always removes all of its versions. COMPLIANCE-mode retention
      # cannot be bypassed by anyone.
      force_destroy = optional(bool, false)

      # Tags specific to this object. Merged with set-level tags (object tags
      # take precedence on key collisions).
      tags = optional(map(string), {})
    }))

    # Tags applied to every object in the set.
    # Individual object tags are merged on top, with object-level tags taking
    # precedence on key collisions.
    tags = optional(map(string), {})
  })
}
