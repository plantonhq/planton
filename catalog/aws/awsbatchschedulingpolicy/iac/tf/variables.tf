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
  description = "AwsBatchSchedulingPolicy specification"
  type = object({
    # The AWS region where the scheduling policy is created. A policy can
    # only be attached to job queues in the same region.
    # Example: "us-west-2", "eu-west-1".
    region = string

    # The percentage (0-99) of the queue's capacity held back for share
    # identifiers that are NOT currently represented among running jobs --
    # headroom so a quiet team's first job does not wait behind a busy
    # team's backlog. The effective reservation is computed as
    # (compute_reservation/100)^N where N is the number of active shares,
    # so the held-back slice shrinks as more shares become active.
    compute_reservation = optional(number)

    # The sliding window, in seconds (0-604800, up to 7 days), over which
    # past usage counts against a share's fair allocation. Longer windows
    # make fairness account for history ("you had the cluster all morning");
    # 0 considers only currently-running jobs.
    share_decay_seconds = optional(number)

    # The relative weight of each share identifier. Shares absent from this
    # list use weight 1.0. AWS allows up to 500 entries per policy.
    share_distributions = optional(list(object({
      # The share identifier jobs carry at submission time (SubmitJob's
      # shareIdentifier). End with "*" to match a prefix -- e.g. "analytics*"
      # covers "analyticsDaily" and "analyticsAdhoc" as one share. Up to 255
      # characters, alphanumeric plus the trailing wildcard.
      share_identifier = string

      # The share's relative weight, 0.0001-999.9999. LOWER weight means MORE
      # capacity: a share with weight 0.5 receives twice the capacity of a
      # weight-1.0 share. Defaults to 1.0 when unset.
      weight_factor = optional(number, 0)
    })), [])
  })
}
