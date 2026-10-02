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
  description = "GcpBigQueryReservation specification"
  type = object({
    # The administration project the reservation lives in and bills to: a
    # literal project ID or a GcpProject reference. If omitted, the
    # provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Where the slots are: a multi-region (US, EU) or a region
    # (asia-northeast1). Assigned projects' jobs use the reservation only
    # for data in this location. Empty means US (Google's default).
    # Immutable.
    location = optional(string, "")

    # The reservation's name -- letters, digits, and dashes. Defaults to
    # metadata.name. Immutable.
    reservation_name = optional(string, "")

    # Baseline slots, always allocated and billed -- 0 is valid with
    # autoscaling, for pay-while-used capacity. Queries may exceed it by
    # borrowing idle slots unless ignore_idle_slots is set. Mutable in
    # place.
    slot_capacity = optional(number, 0)

    # The BigQuery edition, which sets the features and the per-slot price:
    #   STANDARD        -- autoscaling only, no commitments, fewer features
    #   ENTERPRISE      -- commitments, BigQuery ML, BI Engine, and more
    #   ENTERPRISE_PLUS -- adds managed disaster recovery and compliance
    #                      controls
    # Empty lets Google choose. Immutable.
    edition = optional(string, "")

    # The most slots autoscaling may add on top of slot_capacity (a
    # multiple of 50). 0 or empty turns autoscaling off.
    autoscale_max_slots = optional(number, 0)

    # true caps queries at this reservation's own slots; false (Google's
    # default) lets them borrow idle slots from other reservations in the
    # same admin project.
    ignore_idle_slots = optional(bool, false)

    # The soft limit on queries running at once. 0 (Google's default) sizes
    # it automatically from the reservation's slots.
    concurrency = optional(number, 0)

    # The group the reservation belongs to, which pools idle slots among its
    # members before other reservations: a GcpBigQueryReservationGroup
    # reference (its name output) or a literal group name or full path.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    reservation_group = optional(string, "")

    # Managed disaster recovery: the location of a secondary replica
    # (Enterprise Plus). Setting it on create makes a failover reservation;
    # setting or clearing it later converts one way or the other. Failover
    # itself is an operational action outside the spec.
    secondary_location = optional(string, "")

    # Labels on the reservation. The platform attribution labels are added
    # on top and win on a key conflict.
    labels = optional(map(string), {})

    # The projects, folders, or organization whose jobs use the
    # reservation, each for one job type. Each is unique by assignee,
    # job_type, and principal. An assignment is PENDING until the
    # reservation has capacity to give.
    assignments = optional(list(object({
      # Whose jobs use the reservation.
      assignee = object({
        # A project: a literal project ID or a GcpProject reference. The modules
        # send projects/{id}.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        project_id = optional(string, "")

        # A folder: its numeric ID, literal or a GcpFolder reference. The
        # modules send folders/{id}.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        folder_id = optional(string, "")

        # An organization: the numeric organization ID, without the
        # organizations/ prefix. The modules send organizations/{id}.
        organization_id = optional(string, "")
      })

      # Which jobs:
      #   QUERY      -- interactive and batch queries, and scripts
      #   PIPELINE   -- load, export, and copy jobs
      #   CONTINUOUS -- continuous queries
      job_type = string

      # Narrow the assignment to one principal's jobs: jobs that principal
      # runs use this reservation, everyone else's fall back to the
      # project/folder/organization assignment (then on-demand). Formats:
      #   principal://goog/subject/USER_EMAIL
      #   principal://iam.googleapis.com/projects/-/serviceAccounts/SA_EMAIL
      #   principal://iam.googleapis.com/projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/POOL/subject/SUBJECT
      # Empty covers every principal.
      principal = optional(string, "")
    })), [])

    # What happens to the reservation and its assignments when this
    # resource is destroyed:
    #   "" / "DELETE" -- deleted (assigned jobs fall back to on-demand)
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- they leave management and keep running (and
    #                    billing) in GCP
    deletion_policy = optional(string, "")
  })
}
