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
  description = "AwsBatchJobQueue specification"
  type = object({
    # The AWS region where the job queue is created. Must match the region of
    # the compute environments it maps onto.
    # Example: "us-west-2", "eu-west-1".
    region = string

    # Whether the queue accepts newly submitted jobs. When DISABLED, new
    # submissions are rejected but jobs already in the queue finish normally
    # -- the drain switch for maintenance or decommissioning.
    state = optional(string)

    # The queue's priority relative to OTHER queues that share a compute
    # environment: when two queues compete for the same capacity, the one
    # with the higher value is scheduled first. 0 (the default when omitted)
    # is the lowest priority. Priority is between queues, not between jobs --
    # ordering within one queue comes from submission order or the
    # fair-share scheduling policy.
    priority = optional(number, 0)

    # The compute environments this queue dispatches jobs to, in preference
    # order (lowest `order` value is tried first). AWS allows at most three,
    # all of one family: EC2-based (EC2/SPOT) or Fargate-based
    # (FARGATE/FARGATE_SPOT), never mixed. Every referenced environment must
    # be in the VALID state before it can be associated.
    compute_environment_order = list(object({
      # The preference position of this environment: the scheduler tries the
      # LOWEST order first and moves to higher values when capacity runs out.
      # Values must be unique within the queue.
      order = optional(number, 0)

      # The compute environment. Reference an AwsBatchComputeEnvironment's
      # compute_environment_arn output or pass a literal ARN.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      compute_environment = string
    }))

    # The fair-share scheduling policy that orders jobs WITHIN this queue by
    # share identifier instead of first-in-first-out. Reference an
    # AwsBatchSchedulingPolicy's scheduling_policy_arn output or pass a
    # literal ARN. AWS QUIRK: once a queue has a scheduling policy it can be
    # replaced with another, but never removed -- removing fair-share
    # scheduling requires recreating the queue.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    scheduling_policy = optional(string, "")

    # Automatic actions for jobs stuck at the head of the queue in a given
    # state past a time threshold -- the guard against a mis-sized job
    # blocking the whole queue in RUNNABLE forever.
    job_state_time_limit_actions = optional(list(object({
      # The action to take. "CANCEL" is the only action AWS supports for
      # queues mapped to compute environments (the TERMINATE action exists
      # only for SageMaker Training service-environment queues -- a separate
      # Batch surface this catalog does not model).
      action = string

      # Seconds a job may remain in the monitored state before the action
      # fires. Range: 600-86400 (10 minutes to 24 hours).
      max_time_seconds = number

      # WHICH stuck jobs the action applies to -- despite the name, this is a
      # MATCHER against AWS's own stuck-in-RUNNABLE causes, not a free-text log
      # message (CreateJobQueue rejects anything else as "Invalid job reason"):
      #   CAPACITY:INSUFFICIENT_INSTANCE_CAPACITY        -- no capacity in any
      #       mapped compute environment could place the job.
      #   MISCONFIGURATION:COMPUTE_ENVIRONMENT_MAX_RESOURCE -- the job asks for
      #       more than the environment's max_vcpus can ever offer.
      #   MISCONFIGURATION:JOB_RESOURCE_REQUIREMENT      -- the job's resource
      #       requirements cannot be satisfied by the mapped environments'
      #       instance types.
      # Add one entry per cause you want the fuse to cover.
      reason = string

      # The job state to monitor. "RUNNABLE" is the only state AWS supports
      # today -- it catches jobs the scheduler can never place (mis-sized
      # resource requirements, exhausted capacity).
      state = string
    })), [])
  })
}
