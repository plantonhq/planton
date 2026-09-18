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
  description = "AwsEcsCluster specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # CloudWatch Container Insights for the cluster:
    # "enabled" -- metrics and logs at the cluster/service/task level.
    # "enhanced" -- adds container-level observability with automatic
    #   dashboards (recommended for production; higher CloudWatch cost).
    # "disabled" -- no Insights telemetry.
    # Unset keeps the account's default setting. Updatable in place.
    container_insights = optional(string, "")

    # The AWS-managed serverless capacity providers to associate:
    # "FARGATE" and/or "FARGATE_SPOT". These are built into every account
    # -- associating them here is what lets services in this cluster name
    # them in a capacity_provider_strategy. EC2 capacity is defined
    # separately in ec2_capacity_providers; both sets associate onto the
    # cluster together.
    capacity_providers = optional(list(string), [])

    # EC2 capacity providers, each wrapping a referenced auto-scaling
    # group. ECS's managed scaling drives the group's desired count from
    # task demand -- you size the ASG's bounds, ECS turns instances on and
    # off. Each entry materializes as its own capacity provider resource
    # (keyed by name, so adding or removing one never disturbs the others)
    # and is automatically associated with the cluster alongside the
    # Fargate built-ins. Services reference entries by name in their
    # capacity_provider_strategy.
    ec2_capacity_providers = optional(list(object({
      # The capacity provider name -- what services put in their
      # capacity_provider_strategy. 1-255 characters: letters, digits,
      # hyphens, underscores; must not start with "aws", "ecs", or "fargate"
      # (AWS reserves those prefixes).
      name = string

      # The auto-scaling group that provides the instances. Reference an
      # AwsAutoScalingGroup's autoscaling_group_arn output or pass a literal
      # ARN. The group's launch template decides the instance shape (use an
      # ECS-optimized AMI whose agent joins this cluster via user data);
      # ECS's managed scaling then drives the group's desired capacity
      # between the group's own min/max bounds. ForceNew: changing the group
      # replaces the provider.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      auto_scaling_group_arn = string

      # ECS-managed scaling of the auto-scaling group. Leave unset to keep
      # AWS's defaults (managed scaling enabled, target capacity 100); set
      # it to tune headroom and scaling step bounds.
      managed_scaling = optional(object({
        # Managed scaling on or off: "ENABLED" (AWS default -- ECS sizes the
        # group) or "DISABLED" (you size the group yourself; ECS only places
        # tasks on what exists).
        status = optional(string, "")

        # Utilization target for the group, 1-100 percent. 100 (AWS default)
        # runs instances fully packed; a lower value (e.g. 80) keeps headroom
        # so new tasks place without waiting for an instance launch.
        target_capacity = optional(number, 0)

        # Smallest scale-out step, 1-10000 instances. AWS default: 1.
        minimum_scaling_step_size = optional(number, 0)

        # Largest scale-out step, 1-10000 instances. AWS default: 10000.
        maximum_scaling_step_size = optional(number, 0)

        # Seconds a newly launched instance warms up before counting toward
        # capacity metrics, 0-10000. AWS default: 300.
        instance_warmup_period_seconds = optional(number, 0)
      }))

      # Protect instances running non-daemon tasks from scale-in
      # termination: "ENABLED" or "DISABLED". Enabling it requires the
      # auto-scaling group itself to enable new-instance scale-in protection
      # (protect_from_scale_in on the group) -- AWS rejects the provider
      # otherwise. The safe default for task-dense clusters.
      managed_termination_protection = optional(string, "")

      # Gracefully drain tasks off instances the group is terminating:
      # "ENABLED" (AWS default) or "DISABLED". Draining is what makes
      # scale-in and instance refresh invisible to services.
      managed_draining = optional(string, "")
    })), [])

    # The cluster's default capacity provider strategy -- what ECS uses
    # when a service or run-task does not declare its own strategy. Name
    # any associated provider: the Fargate built-ins, an
    # ec2_capacity_providers entry, or a managed_instances_capacity_providers
    # entry. Example: FARGATE base 1 / weight 1 + FARGATE_SPOT weight 4
    # keeps one guaranteed On-Demand task and runs ~80% of scaled capacity
    # on Spot. Known first-apply caveat when naming a managed-instances
    # entry created in the SAME apply: the strategy PUT can race the
    # provider's seconds-long provisioning (AWS rejects it with "not in an
    # ACTIVE state" until it finishes); a re-apply succeeds. Naming
    # built-ins or EC2 providers has no such window.
    default_capacity_provider_strategy = optional(list(object({
      # The capacity provider: "FARGATE", "FARGATE_SPOT", or the name of an
      # ec2_capacity_providers entry.
      capacity_provider = string

      # Minimum number of tasks guaranteed on this provider before weights
      # apply. Only one entry of the strategy may set a non-zero base.
      base = optional(number, 0)

      # Relative share of tasks beyond the bases. Example: FARGATE weight 1 +
      # FARGATE_SPOT weight 4 runs ~80% of scaled tasks on Spot.
      weight = optional(number, 0)
    })), [])

    # ECS Exec auditing for the cluster: where interactive exec sessions
    # (`aws ecs execute-command`) are logged and how session traffic is
    # encrypted. Without this block, exec sessions still work when a
    # service enables them -- they are simply not centrally audited.
    execute_command_configuration = optional(object({
      # Log destination behavior for exec sessions:
      # "DEFAULT" (AWS default) -- sessions log to the task's own awslogs
      #   configuration.
      # "OVERRIDE" -- sessions log to the destinations in log_configuration.
      # "NONE" -- exec works but sessions are not logged (avoid outside
      #   sandboxes; unaudited interactive access defeats compliance).
      logging = optional(string, "")

      # Custom destinations for exec session logs. Only used (and required)
      # when logging is "OVERRIDE".
      log_configuration = optional(object({
        # The CloudWatch log group session logs are sent to. The group must
        # already exist -- ECS does not create it.
        cloud_watch_log_group_name = optional(string, "")

        # Require the CloudWatch log group to be KMS-encrypted; the send fails
        # if it is not. Pair with an encrypted log group for compliance
        # postures.
        cloud_watch_encryption_enabled = optional(bool, false)

        # The S3 bucket session logs are written to.
        s3_bucket_name = optional(string, "")

        # Key prefix for session log objects within the bucket.
        s3_key_prefix = optional(string, "")

        # Require the S3 bucket to be encrypted; the write fails if it is not.
        s3_bucket_encryption_enabled = optional(bool, false)
      }))

      # A KMS key to encrypt the exec session traffic between client and
      # container. Reference an AwsKmsKey's key_arn output or pass a literal
      # key ARN/ID. Unset uses TLS without customer-managed encryption.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key_id = optional(string, "")
    }))

    # Customer-managed KMS encryption for Fargate ephemeral task storage
    # and managed storage -- the compliance posture for regulated
    # workloads. Unset uses AWS-owned keys (data is still encrypted).
    managed_storage_configuration = optional(object({
      # The KMS key encrypting Fargate ephemeral task storage. Reference an
      # AwsKmsKey's key_arn output or pass a literal key ARN. The key policy
      # must grant the Fargate service principal decrypt/generate rights --
      # AWS rejects the cluster configuration otherwise.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      fargate_ephemeral_storage_kms_key_id = optional(string, "")

      # The KMS key for other ECS-managed storage. Reference an AwsKmsKey's
      # key_arn output or pass a literal key ARN.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key_id = optional(string, "")
    }))

    # The AWS Cloud Map namespace (by ARN) that Service Connect uses by
    # default for services in this cluster. Services can override it;
    # setting it here is what lets a whole environment share one service
    # mesh namespace without per-service wiring.
    service_connect_namespace_arn = optional(string, "")

    # ECS Managed Instances capacity providers: ECS launches, patches, and
    # retires the EC2 instances itself -- you describe the compute by
    # attributes (vCPUs, memory, accelerators) and the network to launch
    # into, and ECS owns the fleet end to end (no auto-scaling group, no
    # AMI, no user data). Each entry materializes as its own capacity
    # provider resource that AWS binds to this cluster at creation --
    # unlike EC2 providers there is no association step
    # (PutClusterCapacityProviders neither attaches nor detaches
    # managed-instances providers); services reference entries by name in
    # their capacity_provider_strategy. Requires an infrastructure role the
    # ECS service principal can assume and an instance profile for the
    # launched instances.
    managed_instances_capacity_providers = optional(list(object({
      # The capacity provider name -- what services put in their
      # capacity_provider_strategy. 1-255 characters: letters, digits,
      # hyphens, underscores; must not start with "aws", "ecs", or "fargate"
      # (AWS reserves those prefixes).
      name = string

      # The infrastructure role ECS assumes to launch, patch, and retire the
      # managed instances. Reference an AwsIamRole's role_arn output or pass
      # a literal ARN. The role must trust the ecs.amazonaws.com service
      # principal and carry AmazonECSInfrastructureRolePolicyForManagedInstances
      # (or equivalent); the caller applying the manifest needs iam:PassRole
      # on it.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      infrastructure_role_arn = string

      # What the launched instances look like: instance profile, network
      # placement, and the attribute-based requirements ECS resolves into
      # concrete instance types.
      instance_launch_template = object({
        # The instance profile attached to every launched instance -- the
        # instance-side identity (the ECS agent's permissions come from here).
        # Reference an AwsIamInstanceProfile's instance_profile_arn output or
        # pass a literal ARN.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        ec2_instance_profile_arn = string

        # Where the managed instances launch: the subnets (required) and
        # security groups applied to each instance.
        network_configuration = object({
          # Subnets the instances launch into -- span at least two AZs for
          # availability. Reference AwsSubnet subnet_id outputs or pass literal
          # subnet IDs.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          subnets = list(string)

          # Security groups applied to each instance -- at least one is REQUIRED.
          # Reference AwsSecurityGroup security_group_id outputs or pass literal
          # group IDs. Unlike EC2 launch paths there is NO fall-back to the VPC
          # default group: AWS's CreateCapacityProvider rejects a managed-instances
          # network configuration without security groups (ClientException
          # "must specify a Network Configuration that contain security groups"),
          # even though the Terraform provider's schema marks the argument
          # optional -- the contract lives only server-side.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          security_groups = list(string)
        })

        # Purchase model for the launched capacity: "ON_DEMAND" (AWS default),
        # "SPOT", or "RESERVED" (draw from capacity reservations --
        # capacity_reservations must then be set). Changing this replaces the
        # capacity provider; everything else in the launch template updates in
        # place.
        capacity_option_type = optional(string, "")

        # Which capacity reservations RESERVED capacity draws from. Only legal
        # (and required) when capacity_option_type is "RESERVED".
        capacity_reservations = optional(object({
          # How reservations are used:
          # "RESERVATIONS_ONLY" -- launch only into reservations (pair with
          #   reservation_group_arn to scope which ones).
          # "RESERVATIONS_FIRST" -- prefer reservations, overflow to on-demand.
          # "RESERVATIONS_EXCLUDED" -- never consume reservations.
          # RESERVATIONS_ONLY and RESERVATIONS_FIRST require instance_requirements
          # on the launch template.
          reservation_preference = optional(string, "")

          # A capacity-reservation group ARN scoping which reservations to use.
          # Only legal when reservation_preference is "RESERVATIONS_ONLY".
          reservation_group_arn = optional(string, "")
        }))

        # Attribute-based instance requirements -- describe the compute
        # (memory, vCPUs, accelerators, price protection) and ECS resolves
        # matching instance types at launch. Required when
        # capacity_reservations uses a RESERVATIONS_ONLY or RESERVATIONS_FIRST
        # preference.
        instance_requirements = optional(object({
          # Required. Memory per instance, in MiB. min is required; leave max
          # unset (0) for no upper bound.
          memory_mib = object({
            # Lower bound, inclusive.
            min = optional(number, 0)

            # Upper bound, inclusive. 0 means no upper bound.
            max = optional(number, 0)
          })

          # Required. vCPUs per instance. min is required; leave max unset (0)
          # for no upper bound.
          vcpu_count = object({
            # Lower bound, inclusive.
            min = optional(number, 0)

            # Upper bound, inclusive. 0 means no upper bound.
            max = optional(number, 0)
          })

          # Allow-list of instance types or families, with wildcards
          # ("m5.large", "m5.*", "c*"). At most 400 entries. Mutually exclusive
          # with excluded_instance_types.
          allowed_instance_types = optional(list(string), [])

          # Deny-list of instance types or families, with wildcards. At most 400
          # entries. Mutually exclusive with allowed_instance_types.
          excluded_instance_types = optional(list(string), [])

          # Instance generations to include: "current" and/or "previous". AWS
          # default: any generation matching the other requirements.
          instance_generations = optional(list(string), [])

          # CPU manufacturers to include: "intel", "amd", "amazon-web-services"
          # (Graviton), "apple". AWS default: any.
          cpu_manufacturers = optional(list(string), [])

          # Bare-metal eligibility: "included", "excluded" (AWS default), or
          # "required".
          bare_metal = optional(string, "")

          # Burstable (T-family) eligibility: "included", "excluded" (AWS
          # default), or "required".
          burstable_performance = optional(string, "")

          # Only instance types that support hibernation.
          require_hibernate_support = optional(bool, false)

          # Spot price protection: exclude types whose Spot price exceeds the
          # identified lowest-priced type's Spot price by more than this
          # percentage. Mutually exclusive with
          # max_spot_price_as_percentage_of_optimal_on_demand_price.
          spot_max_price_percentage_over_lowest_price = optional(number, 0)

          # Spot price protection anchored to On-Demand: exclude types whose
          # Spot price exceeds this percentage of the optimal type's On-Demand
          # price. Mutually exclusive with
          # spot_max_price_percentage_over_lowest_price.
          max_spot_price_as_percentage_of_optimal_on_demand_price = optional(number, 0)

          # On-Demand price protection: exclude types whose On-Demand price
          # exceeds the identified lowest-priced type's by more than this
          # percentage. AWS default: 20.
          on_demand_max_price_percentage_over_lowest_price = optional(number, 0)

          # Instance-store (local disk) eligibility: "included" (AWS default),
          # "excluded", or "required".
          local_storage = optional(string, "")

          # Local storage technologies when instance-store is in play: "hdd"
          # and/or "ssd".
          local_storage_types = optional(list(string), [])

          # Total local (instance-store) storage, in GB.
          total_local_storage_gb = optional(object({
            # Lower bound, inclusive.
            min = optional(number, 0)

            # Upper bound, inclusive. 0 means no upper bound.
            max = optional(number, 0)
          }))

          # Memory-to-vCPU ratio, in GiB per vCPU -- a compact way to say
          # "memory optimized" (min 8) or "compute optimized" (max 2) without
          # naming families.
          memory_gib_per_vcpu = optional(object({
            # Lower bound, inclusive.
            min = optional(number, 0)

            # Upper bound, inclusive. 0 means no upper bound.
            max = optional(number, 0)
          }))

          # Number of network interfaces the type must support.
          network_interface_count = optional(object({
            # Lower bound, inclusive.
            min = optional(number, 0)

            # Upper bound, inclusive. 0 means no upper bound.
            max = optional(number, 0)
          }))

          # Network bandwidth, in Gbps.
          network_bandwidth_gbps = optional(object({
            # Lower bound, inclusive.
            min = optional(number, 0)

            # Upper bound, inclusive. 0 means no upper bound.
            max = optional(number, 0)
          }))

          # Baseline (non-burst) EBS bandwidth, in Mbps.
          baseline_ebs_bandwidth_mbps = optional(object({
            # Lower bound, inclusive.
            min = optional(number, 0)

            # Upper bound, inclusive. 0 means no upper bound.
            max = optional(number, 0)
          }))

          # Number of accelerators (GPUs, FPGAs, inference chips). Set min 1 to
          # require accelerated types; to EXCLUDE accelerators, leave this unset
          # and rely on accelerator_types being empty.
          accelerator_count = optional(object({
            # Lower bound, inclusive.
            min = optional(number, 0)

            # Upper bound, inclusive. 0 means no upper bound.
            max = optional(number, 0)
          }))

          # Accelerator manufacturers: "nvidia", "amd", "amazon-web-services",
          # "xilinx", "habana".
          accelerator_manufacturers = optional(list(string), [])

          # Specific accelerator models (e.g. "a100", "v100", "t4",
          # "inferentia", "radeon-pro-v520").
          accelerator_names = optional(list(string), [])

          # Accelerator categories: "gpu", "fpga", "inference".
          accelerator_types = optional(list(string), [])

          # Total accelerator memory, in MiB.
          accelerator_total_memory_mib = optional(object({
            # Lower bound, inclusive.
            min = optional(number, 0)

            # Upper bound, inclusive. 0 means no upper bound.
            max = optional(number, 0)
          }))
        }))

        # Use instance-store (local NVMe) volumes for container storage on
        # instance types that have them. Unset keeps AWS's default placement.
        use_local_storage = optional(bool)

        # CloudWatch monitoring detail for the launched instances: "BASIC"
        # (AWS default) or "DETAILED" (1-minute metrics, billed).
        monitoring = optional(string, "")

        # Root EBS volume size for each launched instance, in GiB (>= 1).
        # Unset keeps AWS's default size.
        storage_size_gib = optional(number, 0)
      })

      # Seconds an empty managed instance idles before ECS scales it in,
      # 0-3600; -1 disables scale-in entirely (instances stay until
      # terminated another way). Unset keeps AWS's default optimization.
      scale_in_after_seconds = optional(number)

      # Propagate the capacity provider's tags to the EC2 instances ECS
      # launches: "CAPACITY_PROVIDER" or "NONE". Unset keeps AWS's default.
      propagate_tags = optional(string, "")
    })), [])
  })
}
