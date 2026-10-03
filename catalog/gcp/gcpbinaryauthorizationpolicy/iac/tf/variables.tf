variable "metadata" {
  description = "Catalog object metadata"
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
  description = "GcpBinaryAuthorizationPolicy specification"
  type = object({
    # The project the policy governs: a literal project ID or a GcpProject
    # reference. Empty means the provider's default project. The module
    # enables binaryauthorization.googleapis.com there.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # What the policy is for.
    description = optional(string, "")

    # Google's global policy for its own system images (GKE add-ons and
    # similar), evaluated before this policy:
    #   "ENABLE"  -- Google-maintained system images are always admitted, so
    #                a strict default rule cannot break the cluster's own
    #                components (recommended with REQUIRE_ATTESTATION)
    #   "DISABLE" -- system images face this policy like any other image
    # Empty sends nothing and keeps Google's current value.
    global_policy_evaluation_mode = optional(string, "")

    # Image name patterns admitted regardless of every rule, in the form
    # registry/path/to/image; a trailing * is a wildcard and may appear only
    # after the registry/ part, e.g. "us-docker.pkg.dev/my-project/base/*".
    # Use them for images no build pipeline signs.
    admission_whitelist_patterns = optional(list(string), [])

    # The rule for every cluster without its own rule below. Required.
    default_admission_rule = object({
      # How images are judged:
      #   "ALWAYS_ALLOW"        -- every image is admitted
      #   "REQUIRE_ATTESTATION" -- an image is admitted only when every attestor
      #                            in require_attestations_by has signed it
      #   "ALWAYS_DENY"         -- every image is denied (lock a cluster down)
      evaluation_mode = string

      # What a denial does:
      #   "ENFORCED_BLOCK_AND_AUDIT_LOG" -- the pod is blocked and the denial
      #                                     logged
      #   "DRYRUN_AUDIT_LOG_ONLY"        -- the pod runs and the would-be denial
      #                                     is logged (the safe rollout mode)
      enforcement_mode = string

      # The attestors that must all have signed an image: GcpBinaryAuthorizationAttestor
      # references or literals. A bare attestor name resolves to the policy's
      # project; an attestor in another project must be the full
      # projects/{project}/attestors/{attestor}. Each attestor must exist
      # before the policy names it.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      require_attestations_by = optional(list(string), [])
    })

    # Per-cluster rules, one per cluster, overriding the default rule there.
    cluster_admission_rules = optional(list(object({
      # The cluster, as Google keys it: {location}.{cluster_name}, where the
      # location is the cluster's zone (e.g. "us-central1-a.prod") or region
      # (e.g. "us-central1.prod").
      cluster = string

      # How images are judged on this cluster; see
      # GcpBinaryAuthorizationPolicyAdmissionRule.evaluation_mode.
      evaluation_mode = string

      # What a denial does on this cluster; see
      # GcpBinaryAuthorizationPolicyAdmissionRule.enforcement_mode.
      enforcement_mode = string

      # The attestors that must all have signed an image on this cluster; see
      # GcpBinaryAuthorizationPolicyAdmissionRule.require_attestations_by.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      require_attestations_by = optional(list(string), [])
    })), [])

    # What destroying this block does:
    #   "" / "DELETE" -- Google's default policy (allow every image) is
    #                    written back
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the block leaves management and this policy stays in
    #                    force
    deletion_policy = optional(string, "")
  })
}
