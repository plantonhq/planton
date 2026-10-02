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
  description = "GcpBigQueryCapacityCommitment specification"
  type = object({
    # The administration project the commitment is bought in: a literal
    # project ID or a GcpProject reference. If omitted, the provider's
    # default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Where the slots are: a multi-region (US, EU) or a region. Only
    # reservations in the same location use them. Empty means US (Google's
    # default). Immutable.
    location = optional(string, "")

    # The commitment's ID -- lowercase letters, digits, and dashes, not
    # starting or ending with a dash, at most 64 characters. Defaults to
    # metadata.name. Google does not keep the ID if the commitment is later
    # split or merged. Immutable.
    capacity_commitment_id = optional(string, "")

    # Slots committed. Immutable.
    slot_count = optional(number, 0)

    # The commitment plan (Google's CommitmentPlan values). With BigQuery
    # editions the plans on offer are ANNUAL (one year) and THREE_YEAR; the
    # FLEX, MONTHLY, TRIAL, and *_FLAT_RATE plans belong to the legacy
    # flat-rate model. A plan can move to a longer term in place, never a
    # shorter one.
    plan = string

    # The plan the commitment converts to when its term ends (for plans
    # that renew), e.g. ANNUAL to renew for another year. Changing it
    # extends the committed period. Empty leaves Google's default.
    renewal_plan = optional(string, "")

    # The edition the slots are for: STANDARD, ENTERPRISE, or
    # ENTERPRISE_PLUS -- it must match the reservations that draw on the
    # commitment. Empty lets Google choose. Immutable.
    edition = optional(string, "")

    # Fail the purchase if another project in the organization already
    # holds a capacity commitment -- a guard for organizations that keep
    # every commitment in one admin project. Immutable.
    enforce_single_admin_project_per_org = optional(bool, false)

    # What happens to the commitment when this resource is destroyed:
    #   "" / "DELETE" -- deleted, which Google refuses before the term ends
    #                    (the destroy then fails)
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the commitment leaves management and runs out its
    #                    term (and billing) in GCP
    deletion_policy = optional(string, "")
  })
}
