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
  description = "AwsEksNodeGroup specification"
  type = object({
    # The AWS region the node group is created in. Must match the cluster's
    # region. Example: "us-west-2", "eu-west-1".
    region = string

    # The EKS cluster the nodes register with. Reference an AwsEksCluster's
    # name output or pass a literal cluster name for a cluster managed
    # outside Planton. Create-only in AWS.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    cluster_name = string

    # The IAM role every node assumes. It must trust ec2.amazonaws.com and
    # carry the worker policies (AmazonEKSWorkerNodePolicy,
    # AmazonEC2ContainerRegistryReadOnly, AmazonEKS_CNI_Policy) -- attach
    # them on the AwsIamRole itself; this component never modifies a role it
    # merely references. AmazonEC2ContainerRegistryReadOnly is what lets the
    # kubelet pull private images from ECR in this account with no pull
    # secret: ECR issues only twelve-hour tokens, so the node's own identity
    # (or IRSA on the pod) is the only way a cluster ever pulls from ECR.
    # Reference an AwsIamRole's role_arn output or pass a literal ARN.
    # Create-only in AWS.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    node_role_arn = string

    # The subnets nodes launch into -- typically the cluster VPC's private
    # subnets. One subnet is a legitimate zonal topology (e.g. a stateful
    # pool pinned to its EBS volumes' zone); use two-plus zones for fleets
    # that should survive a zone impairment. Reference AwsSubnet subnet_id
    # outputs or pass literal subnet IDs. Create-only in AWS. A node group
    # is a member of its cluster and lives there on a diagram; the subnets
    # are where its nodes attach, so the reference is access, not
    # placement -- otherwise a group on subnets its cluster does not name
    # would be drawn outside the cluster it belongs to.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_ids = list(string)

    # Launch the nodes from an AwsLaunchTemplate instead of the inline
    # knobs: custom AMI + bootstrap user data, IMDSv2 enforcement, encrypted
    # or provisioned-IOPS volumes, extra ENI/tag configuration. When set,
    # AWS forbids disk_size_gb, instance_types, and remote_access on the
    # node group (enforced below); ami_type stays valid unless the template
    # pins a custom AMI.
    launch_template = optional(object({
      # The launch template. Reference an AwsLaunchTemplate's
      # launch_template_id output or pass a literal template ID ("lt-...").
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      launch_template_id = string

      # Which template version the nodes launch from: a numeric version for a
      # hard pin, "$Default", or "$Latest". Empty keeps the template's default
      # version. Changing it rolls the group onto the new version -- the
      # template-driven fleet-rollout mechanism.
      version = optional(string, "")
    }))

    # The EC2 instance types AWS may launch (e.g. ["m6i.large"]). Several
    # types is a Spot best practice (pool diversity); On-Demand groups use
    # the first type. Empty keeps the AWS default (t3.medium). Mutually
    # exclusive with launch_template. Create-only in AWS.
    instance_types = optional(list(string), [])

    # The EKS-optimized AMI family: "AL2023_x86_64_STANDARD" /
    # "AL2023_ARM_64_STANDARD" (current default generation),
    # "AL2023_x86_64_NVIDIA" / "AL2023_ARM_64_NVIDIA" /
    # "AL2023_x86_64_NEURON" (accelerated), the Bottlerocket families
    # (container-optimized, incl. FIPS/NVIDIA variants), the Windows Core/
    # Full families (2019/2022/2025), legacy AL2 ("AL2_x86_64",
    # "AL2_x86_64_GPU", "AL2_ARM_64"), or "CUSTOM" (launch template with
    # your own AMI). Empty lets AWS pick from the instance types. With a
    # launch template that pins a custom AMI, leave this empty. Create-only
    # in AWS.
    ami_type = optional(string, "")

    # Purchase model for the fleet: on_demand (default), spot (interruptible
    # at steep discount -- pair with several instance_types), or
    # capacity_block (pre-purchased ML capacity reservations). Create-only
    # in AWS.
    capacity_type = optional(string, "")

    # Root EBS volume size per node, in GiB. 0 keeps the AWS default (20
    # GiB Linux / 50 GiB Windows); 100 is a comfortable production default
    # for image-heavy workloads. Mutually exclusive with launch_template
    # (size the template's block device instead). Create-only in AWS.
    disk_size_gb = optional(number, 0)

    # Node counts the group scales between. desired_size is where the group
    # starts (and what AWS holds until an autoscaler moves it); min 0 with
    # desired 0 expresses an intentionally dormant pool.
    scaling = object({
      # The floor the group never shrinks below. 0 is valid -- a pool that
      # scales to zero when idle.
      min_size = optional(number, 0)

      # The ceiling the group never grows above. At least 1.
      max_size = optional(number, 0)

      # The node count AWS creates and maintains (until a cluster autoscaler
      # adjusts it). Must sit within [min_size, max_size].
      desired_size = optional(number, 0)
    })

    # SSH/security-group access to the nodes. Immutable in AWS -- changing
    # it replaces the group. Mutually exclusive with launch_template (put
    # key and security groups in the template instead).
    remote_access = optional(object({
      # The name of an existing EC2 key pair enabling SSH to the nodes.
      ec2_ssh_key = optional(string, "")

      # Security groups allowed to reach the nodes over SSH. Empty with an
      # ec2_ssh_key set means AWS opens port 22 to 0.0.0.0/0 -- always scope
      # this when enabling SSH. Reference AwsSecurityGroup security_group_id
      # outputs or pass literal IDs.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      source_security_group_ids = optional(list(string), [])
    }))

    # Kubernetes labels applied to every node (visible to schedulers and
    # nodeSelectors). Updates in place.
    labels = optional(map(string), {})

    # Kubernetes taints applied to every node -- the reservation mechanism
    # that keeps ordinary pods off dedicated capacity (GPU pools, ingress
    # tiers) until a pod tolerates the taint. Updates in place.
    taints = optional(list(object({
      # Taint key (e.g. "dedicated"). Required.
      key = string

      # Taint value (e.g. "gpu"). Optional -- key-only taints are valid.
      value = optional(string, "")

      # The scheduling effect: "NO_SCHEDULE" (new pods need a toleration),
      # "PREFER_NO_SCHEDULE" (soft), or "NO_EXECUTE" (also evicts running
      # pods without a toleration). Required.
      effect = string
    })), [])

    # How aggressively version updates roll nodes (surge/unavailability
    # budget). Unset keeps AWS defaults (1 node unavailable, DEFAULT
    # strategy).
    update_config = optional(object({
      # Maximum number of nodes updated (unavailable) at once, 1-100.
      # Exactly one of max_unavailable or max_unavailable_percentage must be
      # set.
      max_unavailable = optional(number, 0)

      # Maximum percentage of nodes updated at once, 1-100. Exactly one of
      # max_unavailable or max_unavailable_percentage must be set.
      max_unavailable_percentage = optional(number, 0)

      # The rollout strategy: "DEFAULT" (respect the unavailability budget
      # only) or "MINIMAL" (additionally launch replacements before
      # terminating -- the surge, capacity-safe rollout). Empty keeps the AWS
      # default (DEFAULT).
      update_strategy = optional(string, "")
    }))

    # Automatic replacement/reboot of nodes the cluster reports unhealthy --
    # AWS's managed node auto-repair.
    node_repair_config = optional(object({
      # Turn node auto-repair on.
      enabled = optional(bool, false)

      # Maximum number of nodes repaired in parallel. Mutually exclusive with
      # max_parallel_nodes_repaired_percentage. 0 keeps the AWS default.
      max_parallel_nodes_repaired_count = optional(number, 0)

      # Maximum percentage of nodes repaired in parallel, 1-100. Mutually
      # exclusive with max_parallel_nodes_repaired_count.
      max_parallel_nodes_repaired_percentage = optional(number, 0)

      # Repair pauses when more than this many nodes are unhealthy (a signal
      # the problem is systemic, not per-node). Mutually exclusive with
      # max_unhealthy_node_threshold_percentage. 0 keeps the AWS default.
      max_unhealthy_node_threshold_count = optional(number, 0)

      # Percentage form of the unhealthy-node pause threshold, 1-100.
      # Mutually exclusive with max_unhealthy_node_threshold_count.
      max_unhealthy_node_threshold_percentage = optional(number, 0)

      # Per-condition overrides of the repair action and wait time.
      overrides = optional(list(object({
        # Minutes to wait after the condition is observed before repairing.
        # Required, at least 1.
        min_repair_wait_time_mins = optional(number, 0)

        # The node monitoring condition this override applies to (e.g.
        # "AcceleratedHardwareReady"). Required.
        node_monitoring_condition = string

        # The unhealthy reason this override applies to. Required.
        node_unhealthy_reason = string

        # What to do: "Replace", "Reboot", or "NoAction". Required.
        repair_action = string
      })), [])
    }))

    # The Kubernetes version of the nodes, e.g. "1.31". Empty follows the
    # cluster's version at creation. Set it to pin nodes during a control-
    # plane upgrade, then bump to roll them; nodes may trail the control
    # plane by up to two minors during an upgrade window.
    version = optional(string, "")

    # The exact EKS-optimized AMI release to run (e.g.
    # "1.31.3-20241109"), for byte-identical fleets and controlled AMI
    # rollouts. Empty keeps the latest release for `version`. Changing it
    # rolls the group.
    release_version = optional(string, "")

    # Force a version update even if pods cannot be drained within their
    # disruption budgets (otherwise the update fails and rolls back). Only
    # consulted while version/release_version/launch_template change.
    force_update_version = optional(bool, false)

    # A pool of pre-initialized instances that cuts scale-out latency from
    # minutes to seconds -- worth its cost exactly when boot time (AMI +
    # bootstrap + image pulls) dominates how fast new capacity serves
    # pods. Updates in place.
    warm_pool_config = optional(object({
      # The state pooled instances wait in: "STOPPED" (AWS default --
      # near-zero compute cost, seconds to start), "RUNNING" (instant but
      # full price), or "HIBERNATED" (RAM restored from disk -- fast
      # JVM/cache warmup without running cost). Empty keeps the AWS default.
      pool_state = optional(string, "")

      # Minimum number of instances always kept in the pool.
      min_size = optional(number, 0)

      # Ceiling on pool size. Unset keeps the AWS default: the gap between
      # the group's max_size and its desired capacity (AWS represents that
      # default internally as -1 -- it never appears here). Explicit 0 is
      # meaningful (no prepared capacity beyond min_size), which is why this
      # field is optional.
      max_group_prepared_capacity = optional(number)

      # Return scaled-in instances to the pool instead of terminating them
      # -- reuse the warm boot instead of paying for it again.
      reuse_on_scale_in = optional(bool, false)
    }))
  })
}
