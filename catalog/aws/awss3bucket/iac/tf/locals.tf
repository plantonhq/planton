locals {
  # metadata.name is the bucket's cloud name. S3 bucket names are globally
  # unique and immutable, so the name IS the identity — there is no separate
  # display name to derive.
  bucket_name = var.metadata.name

  # Resource-identity tags match the Pulumi module key-for-key. S3 identity
  # tagging is the only tagging surface this module manages; user-defined
  # custom tags are a platform-wide concern, not per-kind spec surface.
  aws_tags = {
    "Name"                     = var.metadata.name
    "planton.ai/resource"      = "true"
    "planton.ai/organization"  = var.metadata.org
    "planton.ai/environment"   = var.metadata.env
    "planton.ai/resource-kind" = "AwsS3Bucket"
    "planton.ai/resource-id"   = var.metadata.id
  }

  # Public access block — ABSENCE of the spec block means fully private (all
  # four guards on). The satellite resource is therefore always created; the
  # spec block only exists to relax specific guards. Inside the block each
  # guard is presence-tracked and defaults to ON, so a guard that arrives null
  # stays on and only an explicit false relaxes it.
  public_access_block = {
    block_public_acls       = var.spec.public_access_block == null ? true : coalesce(var.spec.public_access_block.block_public_acls, true)
    block_public_policy     = var.spec.public_access_block == null ? true : coalesce(var.spec.public_access_block.block_public_policy, true)
    ignore_public_acls      = var.spec.public_access_block == null ? true : coalesce(var.spec.public_access_block.ignore_public_acls, true)
    restrict_public_buckets = var.spec.public_access_block == null ? true : coalesce(var.spec.public_access_block.restrict_public_buckets, true)
  }

  # Object Ownership — empty means the modern BucketOwnerEnforced (ACLs
  # disabled). Spelled out here so the satellite is always created and the
  # posture is explicit in state.
  object_ownership = var.spec.object_ownership != "" ? var.spec.object_ownership : "BucketOwnerEnforced"

  # Encryption — when the spec block is absent, S3's own account-level SSE-S3
  # default applies and no satellite is created. When present, empty
  # sse_algorithm still means AES256 (stated rather than implied so a
  # bucket_key-only or future tweak keeps working).
  manage_encryption = var.spec.encryption != null
  # try() guards the attribute access: HCL evaluates both operands of &&,
  # so touching var.spec.encryption.* directly would error when the whole
  # encryption block is absent (null).
  sse_algorithm = try(var.spec.encryption.sse_algorithm, "") != "" ? var.spec.encryption.sse_algorithm : "AES256"
  # kms key only applies to the KMS algorithms; the generated contract
  # flattens the StringValueOrRef to a plain string.
  sse_kms_key_id = try(var.spec.encryption.kms_key_id, "") != "" ? var.spec.encryption.kms_key_id : null

  # Bucket policy — the Struct arrives from the tfvars layer as a nested
  # object; the provider wants the document as a JSON string.
  policy = var.spec.policy != null ? jsonencode(var.spec.policy) : null

  # Lifecycle rules — zero-value normalization: the generated contract fills
  # unset optionals with zero values ("", 0, {}), while the provider treats
  # null and zero differently for several arguments. Normalize once here so
  # main.tf reads cleanly.
  lifecycle_rules = [
    for r in var.spec.lifecycle_rules : {
      id     = r.id
      status = r.status != "" ? r.status : "Enabled"
      # A filter block distinguishing "no filter" (rule covers the whole
      # bucket) from "filter with predicates". AWS expresses one predicate
      # directly and multiple predicates inside an `and` wrapper — that
      # shaping happens in main.tf; here we only count predicates.
      filter = r.filter != null ? {
        prefix                   = r.filter.prefix
        tags                     = r.filter.tags
        object_size_greater_than = r.filter.object_size_greater_than
        object_size_less_than    = r.filter.object_size_less_than
        predicate_count = (
          (r.filter.prefix != "" ? 1 : 0) +
          length(r.filter.tags) +
          (r.filter.object_size_greater_than > 0 ? 1 : 0) +
          (r.filter.object_size_less_than > 0 ? 1 : 0)
        )
      } : null
      transitions                            = r.transitions
      expiration                             = r.expiration
      noncurrent_version_transitions         = r.noncurrent_version_transitions
      noncurrent_version_expiration          = r.noncurrent_version_expiration
      abort_incomplete_multipart_upload_days = r.abort_incomplete_multipart_upload_days
    }
  ]

  # Replication rules — same zero-value normalization + predicate counting as
  # lifecycle filters. delete_marker_replication is a bool in the spec but a
  # required Enabled/Disabled block on filter-carrying rules in the provider.
  replication_rules = var.spec.replication != null ? [
    for r in var.spec.replication.rules : {
      id       = r.id
      priority = r.priority
      status   = r.status != "" ? r.status : "Enabled"
      filter = r.filter != null ? {
        prefix = r.filter.prefix
        tags   = r.filter.tags
        predicate_count = (
          (r.filter.prefix != "" ? 1 : 0) +
          length(r.filter.tags)
        )
      } : null
      destination                         = r.destination
      delete_marker_replication           = r.delete_marker_replication
      existing_object_replication         = r.existing_object_replication
      replicate_replica_modifications     = r.replicate_replica_modifications
      replicate_sse_kms_encrypted_objects = r.replicate_sse_kms_encrypted_objects
    }
  ] : []

  # Website — index/error mode vs redirect-all mode (CEL guarantees
  # exclusivity, so presence of redirect_all_requests_to is the mode switch).
  website_redirect_all = var.spec.website != null ? var.spec.website.redirect_all_requests_to : null

  # Intelligent tiering configurations keyed by name — each is its own
  # provider resource (many-per-bucket satellite), so for_each needs a map.
  intelligent_tiering_configurations = {
    for c in var.spec.intelligent_tiering_configurations : c.name => c
  }

  # The other three many-per-bucket satellites follow the same name-keyed
  # for_each convention (the name is each instance's identity in the
  # provider's import address).
  analytics_configurations = {
    for c in var.spec.analytics_configurations : c.name => c
  }
  inventory_configurations = {
    for c in var.spec.inventory_configurations : c.name => c
  }
  metrics_configurations = {
    for c in var.spec.metrics_configurations : c.name => c
  }
}
