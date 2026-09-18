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
  description = "AwsRoute53Zone specification"
  type = object({
    # The AWS region where the resource will be created.
    # Route 53 itself is a global service; this selects the region used for
    # provider API calls and for regional companions (e.g. a private zone's
    # default VPC region).
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Human-readable comment stored on the hosted zone, visible in the AWS
    # console and GetHostedZone responses. Useful for ownership or purpose
    # notes (e.g. "production apex zone — owned by platform team").
    # Maximum 256 characters.
    comment = optional(string, "")

    # Marks the zone as a private hosted zone that resolves only inside the
    # VPCs listed in vpc_associations. Private zones require at least one VPC
    # association at creation (AWS bakes the first VPC into CreateHostedZone).
    # Default: false (public zone, resolves globally).
    is_private = optional(bool, false)

    # VPCs that can resolve this private zone. Required (at least one) when
    # is_private is true; must be empty for public zones.
    #
    # Each associated VPC must have DNS support and DNS hostnames enabled.
    # All associations here must be same-account: associating a VPC from
    # ANOTHER account requires an authorization handshake between the two
    # accounts (a separate cross-account surface, deliberately not modeled).
    vpc_associations = optional(list(object({
      # The VPC to associate. Can reference an AwsVpc resource.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      vpc_id = string

      # The region the VPC lives in. Defaults to the zone's region when omitted
      # — set it only for VPCs in other regions (a private zone can be resolved
      # from VPCs across regions within the same account).
      vpc_region = optional(string, "")
    })), [])

    # ID of a reusable delegation set to assign this zone's name servers from.
    # Reusable delegation sets give many zones the same four name servers —
    # the white-label DNS pattern (vanity name servers) and bulk zone
    # migrations. Public zones only (AWS rejects it for private zones), and
    # create-time immutable (ForceNew).
    # When omitted, Route 53 assigns a fresh set of name servers to the zone.
    delegation_set_id = optional(string, "")

    # When true, deleting the zone first purges every record in it (except the
    # required NS/SOA pair) and disables DNSSEC signing, so deletion cannot
    # fail on "zone not empty". Leave false (default) to protect a zone that
    # still carries live records from accidental teardown.
    force_destroy = optional(bool, false)

    # Enables accelerated recovery for the hosted zone: Route 53 pre-stages
    # the zone's data so control-plane changes propagate faster during
    # regional recovery events. Public zones only.
    #
    # Tri-state on purpose: AWS keeps the feature's CURRENT state when the
    # field is omitted, so switching it off requires an EXPLICIT false — the
    # provider documents that removing the argument does not disable it.
    # Unset = leave as-is (off for new zones), true = enable, false = disable.
    enable_accelerated_recovery = optional(bool)

    # DNS query logging to CloudWatch Logs — who is querying which names, with
    # response codes. Used for security monitoring, debugging resolution
    # issues, and understanding query patterns. Public zones only.
    # Warning: high-traffic domains generate large log volumes (and cost).
    query_logging = optional(object({
      # ARN of the destination CloudWatch Logs log group (must be in us-east-1).
      # Can reference an AwsCloudwatchLogGroup resource.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      cloudwatch_log_group_arn = string
    }))

    # DNSSEC signing for the zone: Route 53 signs the zone's records with a
    # key-signing key (KSK) backed by an asymmetric KMS key, protecting
    # resolvers from spoofed responses. Public zones only.
    #
    # Enabling signing here is half the chain of trust — to complete it, the
    # DS record from the signed zone must also be registered with the parent
    # (the domain registrar), which is outside this resource.
    dnssec = optional(object({
      # The asymmetric KMS key backing the key-signing key. Can reference an
      # AwsKmsKey resource (see the message comment for the us-east-1 /
      # ECC_NIST_P256 / key-policy requirements).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key_arn = string

      # Name of the key-signing key inside Route 53 (3–128 characters; letters,
      # numbers, "._-"). Defaults to a name derived from the zone when omitted.
      key_signing_key_name = optional(string, "")

      # Operational status of the key-signing key: "ACTIVE" (default when
      # omitted) or "INACTIVE". Deactivating the KSK stops Route 53 from using
      # it to sign — the monitoring/troubleshooting lever AWS documents for
      # investigating signing problems without tearing the DNSSEC config down.
      # Updatable in place.
      #
      # Note: while signing is enabled, an INACTIVE KSK means signatures cannot
      # refresh — deactivate only while diagnosing, or after a replacement KSK
      # is active. AWS allows at most two KSKs per hosted zone (the rotation
      # pair); this resource models the one steady-state KSK, so a
      # zero-downtime rotation (create second KSK, retire the first) is
      # performed with AWS tooling and then reconciled here by pointing
      # kms_key_arn/key_signing_key_name at the new key.
      key_signing_key_status = optional(string, "")
    }))
  })
}
