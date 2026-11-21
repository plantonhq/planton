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
  description = "GcpBigQueryCapacityCommitment specification"
  type = object({
    # The administration project the commitment is bought in: a literal
    # project ID or a GcpProject reference. If omitted, the provider's
    # default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Where the slots are: a multi-region (US, EU) or a region. Only
    # reservations in the same admin project and location use them; a
    # commitment in another project or location does nothing for them. Empty
    # means US (Google's default). Write US and EU in upper case and a region
    # in lower case: the provider never reads the location back, so another
    # spelling is another location to it, and replaces (re-buys) the
    # commitment. Immutable.
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
    # shorter one: Google refuses a shorter plan. After a renewal Google
    # reports the renewal plan as the commitment's plan, so keep plan equal
    # to renewal_plan, or update plan after each renewal -- otherwise the
    # next deploy re-commits to a longer plan or fails on a shorter one.
    plan = string

    # What the commitment becomes when its term ends:
    #   NONE       -- it ends: Google removes it at the end of its term (it
    #                 needs edition set). Once Google removes it, remove this
    #                 block, or destroy it with ABANDON before the term ends:
    #                 a deploy of the unchanged manifest after the term buys
    #                 a new commitment.
    #   ANNUAL     -- it renews for another year, and keeps renewing
    #   THREE_YEAR -- it renews for three years
    # (and the legacy flat-rate plans). Changing it extends the committed
    # period. Empty leaves Google's default, and ANNUAL commitments renew by
    # default.
    renewal_plan = optional(string, "")

    # The edition the slots are for: STANDARD, ENTERPRISE, or
    # ENTERPRISE_PLUS -- it must match the reservations that draw on the
    # commitment. Empty lets Google choose. Immutable.
    edition = optional(string, "")

    # Google's guard that fails the purchase if another project in the
    # organization already holds a capacity commitment. The pinned provider
    # records this value but never sends it to Google (Google takes it as a
    # query parameter of the create call, which the provider leaves out), so
    # it guards nothing; changing it still replaces -- re-buys -- the
    # commitment. Leave it unset. Immutable.
    enforce_single_admin_project_per_org = optional(bool, false)

    # What happens to the commitment when this resource is destroyed:
    #   "" / "DELETE" -- deleted, which Google refuses before the term ends
    #                    (the destroy then fails)
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the commitment leaves management and keeps billing
    #                    in GCP: to the end of its term when renewal_plan is
    #                    NONE, and through every renewal otherwise
    # The policy also governs a replacing change (see the message comment).
    deletion_policy = optional(string, "")
  })
}
