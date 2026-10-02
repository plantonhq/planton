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
  description = "GcpDataprocAutoscalingPolicy specification"
  type = object({
    # GCP project where the policy will be created.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The policy ID (the GCP resource name). 3-50 characters: letters,
    # numbers, underscores, and hyphens; must start and end with a letter
    # or number. Immutable after creation.
    policy_id = string

    # GCP region the policy lives in (e.g., "us-central1"). A cluster can
    # only attach policies in its own region. Immutable after creation.
    location = string

    # Autoscaling bounds and weight for the primary worker group.
    worker_config = object({
      # Maximum number of primary workers the autoscaler may create.
      max_instances = number

      # Minimum number of primary workers. The Dataproc API requires at
      # least 2 primary workers on an autoscaled cluster; 0 leaves the
      # API's default (2).
      min_instances = optional(number, 0)

      # Relative weight of this group when the autoscaler distributes new
      # capacity between primary and secondary workers. Default: 1.
      # Example: primary weight 1 + secondary weight 3 sends ~75% of new
      # nodes to the secondary (spot) group.
      weight = optional(number, 0)
    })

    # Autoscaling bounds and weight for the secondary (preemptible/spot)
    # worker group. Omit to keep the secondary group unscaled.
    secondary_worker_config = optional(object({
      # Maximum number of secondary workers the autoscaler may create.
      # Default: 0 (secondary group not scaled).
      max_instances = optional(number, 0)

      # Minimum number of secondary workers. Default: 0 — the group can
      # scale to zero when idle.
      min_instances = optional(number, 0)

      # Relative weight of this group when the autoscaler distributes new
      # capacity between primary and secondary workers. Default: 1.
      weight = optional(number, 0)
    }))

    # The autoscaling algorithm: evaluation cadence + YARN scaling
    # behavior.
    basic_algorithm = object({
      # How long the autoscaler waits between evaluations, letting the
      # cluster settle after a scaling event. Bounds: 2 minutes (120s) to
      # 1 day (86400s) — the Dataproc API's own limits, enforced below so
      # an out-of-range value fails at spec time instead of at the
      # provider. Format: duration in seconds with "s" suffix.
      # Default: "120s".
      cooldown_period = optional(string, "")

      # YARN memory-based scaling behavior.
      yarn_config = object({
        # How long the autoscaler waits for a graceful YARN decommission
        # before forcefully removing a worker during scale-down. Running
        # tasks get this window to finish. Bounds: 0s to 1 day (86400s) —
        # the Dataproc API's own limit, enforced below so an out-of-range
        # value fails at spec time instead of at the provider.
        # Format: duration in seconds with "s" suffix (e.g., "3600s").
        graceful_decommission_timeout = string

        # Fraction of pending YARN memory the autoscaler adds capacity for
        # per evaluation (0.0-1.0). 1.0 scales up as fast as possible; 0.05
        # adds ~5% of the suggested capacity per cooldown period.
        # Declared optional (explicit presence) so 0.0 — a legitimate API
        # value — is expressible while the field itself stays required.
        scale_up_factor = number

        # Fraction of available (idle) YARN memory the autoscaler removes
        # capacity for per evaluation (0.0-1.0). 1.0 scales down as fast as
        # possible; 0.0 disables scale-down entirely (ever-growing cluster —
        # pair with idle_delete_ttl on the cluster for cost control).
        # Declared optional (explicit presence) so 0.0 — a legitimate API
        # value — is expressible while the field itself stays required.
        scale_down_factor = number

        # Minimum fractional change to the cluster size the autoscaler acts
        # on when scaling up (0.0-1.0). Example: 0.05 means a recommendation
        # must grow the cluster by at least 5% to trigger; 0.0 (default)
        # means any recommendation of at least one worker triggers.
        scale_up_min_worker_fraction = optional(number, 0)

        # Minimum fractional change to the cluster size the autoscaler acts
        # on when scaling down (0.0-1.0). Default: 0.0.
        scale_down_min_worker_fraction = optional(number, 0)
      })
    })

    # Engine-side teardown behavior. "DELETE" (default) deletes the
    # policy; "PREVENT" fails any plan that would delete it; "ABANDON"
    # removes it from IaC management while leaving it in GCP. Note the API
    # refuses to delete a policy while any cluster still references it,
    # regardless of this setting.
    deletion_policy = optional(string, "")
  })
}
