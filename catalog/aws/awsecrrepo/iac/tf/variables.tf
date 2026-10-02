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
  description = "AwsEcrRepo specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Name of the ECR repository. Must be unique within the AWS account and
    # region, and cannot be changed after creation (ForceNew).
    #
    # ECR repository names support slash-separated namespaces (e.g.
    # "team-blue/checkout-service"), which is why the name is an explicit spec
    # field rather than being derived from metadata.name: the namespace path is
    # registry structure, a genuinely different concept from the graph node's
    # display name.
    #
    # Allowed characters: lowercase letters, numbers, and separators
    # (hyphen, underscore, period) between alphanumeric runs; path segments
    # separated by "/".
    repository_name = string

    # Controls whether image tags can be overwritten after they are pushed.
    #
    # Valid values:
    # - "MUTABLE" (AWS default): any tag can be overwritten by a later push.
    #   Convenient for development, but a moving "v1.2.3" undermines deploy
    #   reproducibility.
    # - "IMMUTABLE": no tag can ever be overwritten. The production
    #   recommendation — a tag permanently identifies one image digest.
    # - "IMMUTABLE_WITH_EXCLUSION": immutable EXCEPT for tags matching
    #   image_tag_mutability_exclusion_filters. The best of both: release tags
    #   are frozen while floating tags like "latest" or "dev-*" stay movable.
    # - "MUTABLE_WITH_EXCLUSION": mutable EXCEPT for tags matching the
    #   exclusion filters (the filters select the IMMUTABLE tags).
    image_tag_mutability = optional(string)

    # Wildcard tag filters that invert the repository's base mutability for
    # matching tags. Only allowed (and then required) when image_tag_mutability
    # is IMMUTABLE_WITH_EXCLUSION or MUTABLE_WITH_EXCLUSION.
    #
    # Each filter is a tag pattern of letters, numbers, ".", "_", "-" and up to
    # two "*" wildcards. Examples: "latest", "dev-*", "*-snapshot".
    # Maximum 5 filters. AWS currently supports only wildcard-type filters, so
    # the filter type is implied and not modeled.
    image_tag_mutability_exclusion_filters = optional(list(string), [])

    # How ECR encrypts stored images. Changing this after creation REPLACES the
    # repository (images are not migrated).
    #
    # Valid values:
    # - "AES256" (default): server-side encryption with an Amazon S3-managed key.
    #   Zero configuration, no KMS cost.
    # - "KMS": server-side encryption with AWS KMS — either the AWS-managed
    #   "aws/ecr" key (when kms_key_id is omitted) or a customer-managed key.
    #   Required when compliance demands key rotation control or CloudTrail
    #   visibility of decrypt calls.
    # - "KMS_DSSE": dual-layer server-side encryption (two independent KMS
    #   envelope layers). For workloads subject to DoD CC SRG / top-secret
    #   data-at-rest requirements.
    encryption_type = optional(string)

    # Customer-managed KMS key used when encryption_type is "KMS" or "KMS_DSSE".
    # Accepts a key ARN or key ID, or a reference to an AwsKmsKey resource.
    # When omitted with KMS/KMS_DSSE, AWS uses the AWS-managed "aws/ecr" key.
    # Must not be set when encryption_type is "AES256".
    # Create-time only (ForceNew, together with encryption_type).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_id = optional(string, "")

    # Enables automatic vulnerability scanning on every image push (the
    # repository-level "basic scanning" toggle). A production security
    # essential — shift-left detection of known CVEs before deploy.
    # Note: if the account has registry-level enhanced scanning configured
    # (Amazon Inspector), that account-wide setting supersedes this toggle.
    scan_on_push = optional(bool)

    # When true, deleting the repository also deletes all contained images.
    # When false (default), deletion fails if any image is present — protection
    # against losing the only copy of a production image.
    force_delete = optional(bool, false)

    # Automated image-expiration rules for storage cost control. Active CI/CD
    # pipelines generate images far faster than anyone deletes them; without
    # lifecycle rules, storage grows unbounded.
    #
    # Rules are evaluated by ascending rule_priority; an image is expired by the
    # first rule that selects it. AWS constraints (enforced here so they fail at
    # authoring time, not at apply): rule priorities must be unique, and at most
    # one rule may use tag_status "any" PER storage class — the provider's own
    # tested archive example pairs an "any" transition rule (standard images)
    # with an "any" expire rule selecting storage_class "archive", so the cap
    # applies per selection tier, not per policy.
    lifecycle_rules = optional(list(object({
      # Evaluation order: lower priorities are evaluated first, and an image is
      # expired by the first rule that selects it. Must be unique across rules.
      # A tag_status "any" rule must carry the highest priority in the policy.
      rule_priority = number

      # Human-readable description of what the rule does, stored in the policy.
      # Example: "Expire untagged images after 14 days".
      description = optional(string, "")

      # Which images the rule selects by tag state:
      # - "tagged": only images with tags matching tag_prefixes or tag_patterns
      #   (exactly one of the two lists is required).
      # - "untagged": only images with no tags (typically superseded layers and
      #   failed builds). No prefix/pattern lists allowed.
      # - "any": every image in the repository. No prefix/pattern lists allowed;
      #   at most one such rule per policy and it must have the highest priority.
      tag_status = string

      # Tag prefixes selecting tagged images (e.g. "release-" matches
      # "release-1.2.3"). Only with tag_status "tagged", and mutually exclusive
      # with tag_patterns. Maximum 100 prefixes.
      tag_prefixes = optional(list(string), [])

      # Wildcard tag patterns selecting tagged images (e.g. "*-snapshot",
      # "v1.*"). Each pattern may contain up to four "*" wildcards. Only with
      # tag_status "tagged", and mutually exclusive with tag_prefixes.
      # Maximum 100 patterns.
      tag_patterns = optional(list(string), [])

      # How the rule counts images:
      # - "imageCountMoreThan": keep the newest count_number selected images and
      #   act on the rest ("keep last N").
      # - "sinceImagePushed": act on selected images older than count_number days
      #   ("by age").
      # - "sinceImagePulled": act on selected images not pulled for count_number
      #   days ("by disuse") — the natural selector for archive transitions.
      # - "sinceImageTransitioned": act on selected images that have been in the
      #   archive storage class for count_number days (requires storage_class
      #   "archive").
      count_type = string

      # The count for count_type: an image count for "imageCountMoreThan", or a
      # number of days for the three "since..." types (the "days" unit is the
      # only unit ECR supports and is implied).
      count_number = number

      # Selects images by their CURRENT storage class: "standard" (the default
      # when unset — freshly pushed images) or "archive" (images a transition
      # rule already moved). "archive" is required with count_type
      # "sinceImageTransitioned" and is only valid there — with every other
      # count_type ECR supports only "standard".
      storage_class = optional(string, "")

      # What the rule does to selected images:
      # - "expire" (the default when unset): delete them.
      # - "transition": move them to the cheaper archive storage class instead of
      #   deleting (pair with target_storage_class "archive"). The canonical
      #   archive policy is two rules: transition after N days without a pull,
      #   then expire after M days in archive (count_type
      #   "sinceImageTransitioned" + storage_class "archive").
      action_type = optional(string, "")

      # For action_type "transition": the storage class images move to. "archive"
      # is the only value ECR supports.
      target_storage_class = optional(string, "")
    })), [])

    # Resource-based access policy for the repository, as a standard IAM policy
    # document. The primary uses are cross-account pulls (granting another
    # account's principals ecr:BatchGetImage / ecr:GetDownloadUrlForLayer) and
    # service access such as AWS Lambda pulling container images. AWS models
    # this as a separate API resource keyed by the repository; it is folded here
    # because the policy has no identity of its own and follows the repository's
    # lifecycle.
    repository_policy = optional(any)
  })
}
