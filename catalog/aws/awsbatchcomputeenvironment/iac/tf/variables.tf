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
  description = "AwsBatchComputeEnvironment specification"
  type = object({
    # The AWS region where the Batch compute environment is created.
    # Example: "us-west-2", "eu-west-1".
    region = string

    # Whether the compute environment accepts jobs from associated queues.
    # When DISABLED, no new jobs are dispatched to it, but running jobs finish
    # and the environment can still scale in. Disabling is also how a compute
    # environment is drained before deletion or replacement behind a queue.
    state = optional(string)

    # The IAM role AWS Batch assumes to manage compute on your behalf.
    # LEAVE UNSET for the recommended path: AWS Batch then uses (and
    # auto-creates) the AWSServiceRoleForBatch service-linked role -- which is
    # also a precondition for in-place infrastructure updates (see the message
    # comment). Set a custom role only when your org mandates one, and expect
    # most compute_resources changes to replace the environment in that mode.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    service_role = optional(string, "")

    # The compute infrastructure backing this environment: resource type,
    # vCPU scaling bounds, VPC placement, and (for EC2/SPOT) instance
    # selection. Required because the modules always create a MANAGED
    # environment, and AWS requires compute resources for MANAGED.
    compute_resources = object({
      # The compute resource type.
      #   EC2:          On-Demand EC2 instances.
      #   SPOT:         EC2 Spot instances (interruptible, up to ~90% cheaper).
      #   FARGATE:      Serverless containers (AWS manages all instances).
      #   FARGATE_SPOT: Serverless containers at Spot pricing.
      # A job queue can only mix environments of one family (EC2/SPOT together,
      # or FARGATE/FARGATE_SPOT together) -- never both families.
      type = string

      # The maximum vCPUs the environment can scale out to. For Fargate this
      # caps total concurrent vCPU capacity across all running jobs. This is
      # the one sizing knob AWS allows updating on EVERY environment,
      # regardless of service role or allocation strategy.
      max_vcpus = number

      # The vCPU floor maintained even when no jobs are runnable. EC2/SPOT
      # only. Keep the default 0 so the environment scales to zero when idle --
      # a non-zero floor keeps instances (and their cost) warm for
      # latency-sensitive queues.
      min_vcpus = optional(number)

      # The initial vCPU target at environment creation. EC2/SPOT only. AWS
      # Batch continuously adjusts the actual desired capacity between
      # min_vcpus and max_vcpus based on queue demand, so treat this as a
      # starting point, not a setpoint. An explicit 0 is indistinguishable
      # from unset all the way down (the provider cannot send a zero here at
      # create or update) -- to keep an idle environment at zero instances,
      # set min_vcpus to 0 and let Batch scale in.
      desired_vcpus = optional(number, 0)

      # The VPC subnets where compute is launched. Spread across multiple
      # Availability Zones for capacity diversity -- especially for SPOT, where
      # more pools mean fewer interruptions.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      subnet_ids = list(string)

      # The security groups attached to compute resources (and to Fargate task
      # ENIs). REQUIRED for FARGATE/FARGATE_SPOT. For EC2/SPOT they may be
      # omitted only when the launch template supplies its own.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      security_group_ids = optional(list(string), [])

      # The EC2 instance types (or families) Batch may launch. EC2/SPOT only.
      # Use "optimal" to let Batch pick from the C, M, and R families to match
      # each job's resource requirements.
      # Examples: ["optimal"], ["m5", "c5"], ["c5.xlarge", "c5.2xlarge"].
      instance_types = optional(list(string), [])

      # How Batch picks instance types (and Spot pools) when scaling out.
      # EC2/SPOT only; AWS defaults to BEST_FIT when omitted.
      #   BEST_FIT:                            cheapest fitting type only; may
      #                                        stall on capacity; no in-place
      #                                        infrastructure updates.
      #   BEST_FIT_PROGRESSIVE:                cheapest fitting types, falling
      #                                        forward when capacity runs out
      #                                        (recommended for EC2).
      #   BEST_FIT_PROGRESSIVE_ORDERED:        like BEST_FIT_PROGRESSIVE, but
      #                                        honors the instance_types list
      #                                        order as preference order.
      #   SPOT_CAPACITY_OPTIMIZED:             deepest Spot pools first --
      #                                        fewest interruptions (SPOT only).
      #   SPOT_PRICE_CAPACITY_OPTIMIZED:       balances Spot price and pool
      #                                        depth (recommended for SPOT).
      #   SPOT_CAPACITY_OPTIMIZED_PRIORITIZED: capacity-optimized, honoring the
      #                                        instance_types order (SPOT only).
      # Only BEST_FIT_PROGRESSIVE, SPOT_CAPACITY_OPTIMIZED, and
      # SPOT_PRICE_CAPACITY_OPTIMIZED support in-place infrastructure updates;
      # the others force replacement on most compute changes.
      allocation_strategy = optional(string, "")

      # The IAM instance profile applied to EC2/SPOT instances -- it wraps the
      # role that lets the ECS agent on each instance register with Batch's
      # underlying ECS cluster. Required for EC2 and SPOT. Reference an
      # AwsIamInstanceProfile's instance_profile_arn output or pass a literal
      # profile ARN.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      instance_role = optional(string, "")

      # The EC2 key pair name for SSH access to instances. EC2/SPOT only.
      # Prefer SSM Session Manager (via the instance profile) over SSH; omit
      # this unless direct SSH is genuinely needed.
      ec2_key_pair = optional(string, "")

      # The maximum Spot price as a percentage of the On-Demand price
      # (e.g. 60 = pay at most 60% of On-Demand). SPOT only. Omit to default
      # to 100% -- with capacity-optimized strategies the actual price is
      # usually far below the cap anyway.
      bid_percentage = optional(number)

      # The IAM role for the Amazon EC2 Spot Fleet that AWS Batch uses under
      # the BEST_FIT allocation strategy. AWS requires it ONLY for SPOT
      # environments using BEST_FIT (or no strategy, which defaults to
      # BEST_FIT); the modern capacity-optimized strategies do not use Spot
      # Fleet and need no role. CREATE-TIME ONLY: changing it replaces the
      # environment.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      spot_iam_fleet_role = optional(string, "")

      # A custom EC2 launch template for instances -- custom AMIs, user data,
      # extra volumes, IMDSv2 posture. EC2/SPOT only. Adding or removing the
      # block replaces the environment; version changes update in place only
      # within the in-place-update envelope (see the spec comment).
      launch_template = optional(object({
        # The launch template. Reference an AwsLaunchTemplate's launch_template_id
        # output or pass a literal template ID ("lt-...").
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        launch_template_id = string

        # The template version to launch: a version number, "$Latest", or
        # "$Default". Omit to use the template's default version. Batch caches
        # the resolved version at scale-out time; within the in-place-update
        # envelope, changing this triggers an infrastructure update that rolls
        # instances per update_policy.
        version = optional(string, "")
      }))

      # AMI selection for EC2/SPOT instances, keyed by image type. Maximum 2
      # entries (AWS allows one Linux and one Windows-family entry). CREATE-TIME
      # in practice: outside the in-place-update envelope any change here
      # replaces the environment.
      ec2_configurations = optional(list(object({
        # The image family Batch should launch. Common values: "ECS_AL2023"
        # (current Amazon Linux 2023 ECS AMI), "ECS_AL2" (Amazon Linux 2 --
        # required for GPU instance types together with ECS_AL2_NVIDIA),
        # "ECS_AL2_NVIDIA" (GPU), "EKS_AL2023"/"EKS_AL2" (Batch on EKS). AWS
        # defaults to ECS_AL2 when the whole block is omitted.
        image_type = optional(string, "")

        # A specific AMI ID that overrides the image_type default -- the way to
        # pin a custom or hardened AMI while keeping Batch's image-family
        # semantics.
        image_id_override = optional(string, "")

        # The EKS-optimized AMI's Kubernetes version, for Batch-on-EKS
        # environments (image types EKS_AL2023/EKS_AL2). Ignored for ECS image
        # types.
        image_kubernetes_version = optional(string, "")
      })), [])

      # The EC2 placement group for tightly-coupled multi-node parallel jobs
      # that need low-latency networking between instances. EC2/SPOT only.
      # CREATE-TIME ONLY: changing it replaces the environment.
      placement_group = optional(string, "")

      # Tags applied to the launched compute resources themselves (EC2
      # instances and Spot Fleet requests) -- these propagate to the EC2
      # console, cost reports, and IAM tag conditions. EC2/SPOT only; Fargate
      # task ENIs cannot be tagged this way. Distinct from the identity tags
      # Planton applies to the compute environment resource.
      resource_tags = optional(map(string), {})
    })

    # Attach this compute environment to an EKS cluster so Batch schedules
    # jobs as Kubernetes pods instead of ECS tasks. CREATE-TIME ONLY: both
    # fields replace the environment when changed. The referenced cluster must
    # exist before the environment is created. The workload half of this
    # pairing is an AwsBatchJobDefinition with its eks arm set.
    eks_configuration = optional(object({
      # The EKS cluster that receives Batch-managed nodes. Reference an
      # AwsEksCluster's cluster_arn output or pass a literal cluster ARN.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      eks_cluster_arn = string

      # The Kubernetes namespace Batch launches job pods into. The namespace
      # must exist in the cluster and be RBAC-configured for Batch before the
      # environment is created.
      kubernetes_namespace = string
    }))

    # How infrastructure updates treat RUNNING jobs when Batch replaces
    # instances during an in-place update (EC2/SPOT environments). When unset,
    # Batch waits for jobs to finish on the old instances (up to 30 minutes)
    # before terminating them. ONCE SET, removing this block later does NOT
    # reset the environment's policy -- AWS keeps the last-applied values (the
    # provider only sends the policy when the block is present); to change
    # course, keep the block and set its fields explicitly.
    update_policy = optional(object({
      # Terminate running jobs when their instance is replaced, instead of
      # waiting for them to finish. Jobs are restarted per their retry
      # strategy. Leave false for long jobs that checkpoint poorly.
      terminate_jobs_on_update = optional(bool, false)

      # How long (in minutes, 1-360) Batch waits for running jobs to finish on
      # old instances before terminating them anyway. Only meaningful when
      # terminate_jobs_on_update is false. AWS's default when the whole policy
      # is absent: 30.
      job_execution_timeout_minutes = optional(number)
    }))
  })
}
