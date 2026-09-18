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
  description = "AwsAutoScalingGroup specification"
  type = object({
    # The AWS region the group is created in. Must match the region of the
    # subnets, launch template, and any target groups it references.
    # Example: "us-west-2", "eu-west-1".
    region = string

    # The subnets capacity is placed in (AWS calls this the VPC zone
    # identifier). Spread across at least two availability zones for real
    # fault tolerance -- the group automatically rebalances instances across
    # the zones these subnets cover. Reference AwsSubnet subnet_id outputs
    # or pass literal subnet IDs.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnets = list(string)

    # The launch template every instance launches from. Exactly one of
    # launch_template or mixed_instances_policy must be set -- use this one
    # for a single-type, single-purchase-option fleet.
    launch_template = optional(object({
      # The launch template. Reference an AwsLaunchTemplate's
      # launch_template_id output or pass a literal template ID ("lt-...").
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      launch_template_id = string

      # Which template version to launch: "$Default" (follow the template's
      # default version -- the AWS default and the setup that lets a template
      # update roll the fleet), "$Latest" (always the newest version, even
      # one not yet promoted), or a numeric version for a hard pin.
      version = optional(string, "")
    }))

    # Blend instance types and purchase options in one group: a base of
    # On-Demand capacity plus a Spot overflow, drawn from several instance
    # types (or attribute-based requirements) for pool diversity. Exactly
    # one of launch_template or mixed_instances_policy must be set.
    mixed_instances_policy = optional(object({
      # The base launch template the overrides specialize. Required.
      launch_template = object({
        # The launch template. Reference an AwsLaunchTemplate's
        # launch_template_id output or pass a literal template ID ("lt-...").
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        launch_template_id = string

        # Which template version to launch: "$Default" (follow the template's
        # default version -- the AWS default and the setup that lets a template
        # update roll the fleet), "$Latest" (always the newest version, even
        # one not yet promoted), or a numeric version for a hard pin.
        version = optional(string, "")
      })

      # Instance-type overrides. Each entry widens the pool set: an explicit
      # type, a different template, a capacity weight, or attribute-based
      # requirements. With no overrides the group uses only the base
      # template's type.
      overrides = optional(list(object({
        # An explicit instance type for this override (e.g. "m5.large").
        # Mutually exclusive with instance_requirements.
        instance_type = optional(string, "")

        # How many capacity units an instance of this type fulfills, 1-999.
        # Weights let heterogeneous sizes count fairly: an m5.2xlarge at
        # weight 4 next to an m5.large at weight 1 keeps "desired = 8"
        # meaningful. 0 leaves the weight unset (every instance counts as 1).
        weighted_capacity = optional(number, 0)

        # Launch this override from a different template (e.g. an arm64-AMI
        # template for Graviton types next to the x86 base).
        launch_template = optional(object({
          # The launch template. Reference an AwsLaunchTemplate's
          # launch_template_id output or pass a literal template ID ("lt-...").
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          launch_template_id = string

          # Which template version to launch: "$Default" (follow the template's
          # default version -- the AWS default and the setup that lets a template
          # update roll the fleet), "$Latest" (always the newest version, even
          # one not yet promoted), or a numeric version for a hard pin.
          version = optional(string, "")
        }))

        # Attribute-based selection for this override -- one entry that
        # resolves to many pools. Mutually exclusive with instance_type.
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

          # Deny-list of instance types or families, with wildcards. At most
          # 400 entries. Mutually exclusive with allowed_instance_types.
          excluded_instance_types = optional(list(string), [])

          # Instance generations to include: "current" and/or "previous".
          instance_generations = optional(list(string), [])

          # CPU manufacturers to include: "intel", "amd",
          # "amazon-web-services" (Graviton), "apple".
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

          # Memory-to-vCPU ratio, in GiB per vCPU.
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

          # Number of accelerators (GPUs, FPGAs, inference chips).
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
          # "inferentia").
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
      })), [])

      # How capacity splits between On-Demand and Spot, and how each side
      # picks pools.
      instances_distribution = optional(object({
        # How On-Demand capacity picks pools: "lowest-price" or "prioritized"
        # (the override list order is a preference ranking). AWS default:
        # "lowest-price".
        on_demand_allocation_strategy = optional(string, "")

        # Instances of guaranteed On-Demand capacity before the percentage
        # split applies -- the "always-on core" of the fleet.
        on_demand_base_capacity = optional(number, 0)

        # Percentage of capacity ABOVE the base that is On-Demand, 0-100.
        # AWS default: 100 (all On-Demand). Explicit 0 means all-Spot above
        # the base -- the aggressive cost posture -- which is why this field
        # is optional: 0 must be distinguishable from unset.
        on_demand_percentage_above_base_capacity = optional(number)

        # How Spot capacity picks pools: "price-capacity-optimized" (the AWS
        # recommendation -- weighs price AND interruption risk),
        # "capacity-optimized", "capacity-optimized-prioritized",
        # "lowest-price" (cheapest but interruption-prone), or "diversified".
        spot_allocation_strategy = optional(string, "")

        # Number of Spot pools to spread across. Only valid with the
        # "lowest-price" strategy. AWS default: 2.
        spot_instance_pools = optional(number, 0)

        # Maximum Spot price per instance-hour, as a decimal string. AWS
        # default (unset): the On-Demand price -- the AWS recommendation.
        spot_max_price = optional(string, "")
      }))
    }))

    # The floor the group never shrinks below. 0 is valid -- a group that
    # scales to zero when idle.
    min_size = optional(number, 0)

    # The ceiling the group never grows above. Scaling policies and
    # instance refresh honor it strictly (unless a policy explicitly allows
    # a predictive-scaling buffer).
    max_size = optional(number, 0)

    # The capacity the group actively maintains. Leave 0 to start at
    # min_size and let scaling policies take over -- the declarative-fleet
    # default, since a literal desired count here fights the autoscaler on
    # every apply.
    desired_capacity = optional(number, 0)

    # What min/max/desired count: "units" (instances, the default), "vcpu",
    # or "memory-mib". The vCPU/memory units only make sense with
    # attribute-based instance requirements, where instance sizes vary.
    desired_capacity_type = optional(string, "")

    # Proactively replace Spot instances that AWS signals as
    # at-elevated-risk of interruption, before the two-minute notice.
    # Recommended for every Spot-bearing group.
    capacity_rebalance = optional(bool, false)

    # Seconds between scaling activities initiated by simple scaling
    # policies. 0 keeps the AWS default (300). Step and target-tracking
    # policies ignore this and use instance warmup instead.
    default_cooldown_seconds = optional(number, 0)

    # Seconds a newly launched instance is expected to take before its
    # metrics are representative. Used by target tracking, instance
    # refresh, and rebalancing as the default warmup. Setting it (even to
    # 0) noticeably improves scaling accuracy for fast-booting services.
    default_instance_warmup_seconds = optional(number, 0)

    # How instance health is judged:
    # - "EC2" (AWS default): instance status checks only.
    # - "ELB": additionally trust the load balancer's target health checks
    #   -- an instance failing its target group health check is replaced.
    #   The right choice whenever target_groups is set; without it a
    #   wedged-but-running process is never replaced.
    health_check_type = optional(string, "")

    # Seconds after launch before health checks can mark an instance
    # unhealthy -- boot-and-warm time. 0 keeps the provider default (300).
    # Ignored until the instance reaches the InService state.
    health_check_grace_period_seconds = optional(number, 0)

    # Target groups whose traffic this group serves. Instances are
    # registered on launch and deregistered (drained) on termination.
    # Reference AwsLbTargetGroup target_group_arn outputs or pass literal
    # ARNs. Pair with health_check_type = "ELB".
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    target_groups = optional(list(string), [])

    # Which instances are terminated first on scale-in, evaluated in order:
    # "Default", "OldestInstance", "NewestInstance",
    # "OldestLaunchTemplate", "OldestLaunchConfiguration",
    # "ClosestToNextInstanceHour", "AllocationStrategy", or the ARN of a
    # custom termination Lambda. "OldestLaunchTemplate" pairs naturally
    # with template-version rollouts; "AllocationStrategy" keeps a mixed
    # fleet on its preferred pools.
    termination_policies = optional(list(string), [])

    # Maximum seconds any instance lives before being replaced, 86400 (1
    # day) to 31536000 (1 year); 0 disables. Continuous fleet hygiene:
    # guarantees patched AMIs and clean processes without a manual rotate.
    max_instance_lifetime_seconds = optional(number, 0)

    # Protect instances from scale-in by default (scaling policies cannot
    # pick them; explicit terminations still work). For fleets whose
    # members hold long-lived work -- pair with lifecycle hooks for
    # graceful drain.
    protect_from_scale_in = optional(bool, false)

    # The placement group launched instances join (cluster/spread/
    # partition). A literal name -- placement groups have no Planton kind
    # yet.
    placement_group = optional(string, "")

    # The service-linked IAM role the Auto Scaling service itself assumes.
    # A literal ARN: service-linked roles are created and owned by AWS
    # (not user IAM roles), and the account default
    # AWSServiceRoleForAutoScaling is used when unset -- which is almost
    # always right. Set only for a custom-suffix role (e.g. per-team KMS
    # grants).
    service_linked_role_arn = optional(string, "")

    # CloudWatch group-level metrics to enable (e.g. "GroupMinSize",
    # "GroupMaxSize", "GroupDesiredCapacity", "GroupInServiceInstances",
    # "GroupPendingInstances", "GroupTerminatingInstances",
    # "GroupTotalInstances", and the warm-pool variants). Free of charge --
    # enabling them is almost always worth it for fleet observability.
    enabled_metrics = optional(list(string), [])

    # Auto Scaling processes to suspend, for maintenance windows or
    # incident response: "Launch", "Terminate", "AddToLoadBalancer",
    # "AlarmNotification", "AZRebalance", "HealthCheck", "InstanceRefresh",
    # "ReplaceUnhealthy", "ScheduledActions". Suspending "Launch" and
    # "Terminate" freezes the fleet entirely.
    suspended_processes = optional(list(string), [])

    # Rolling replacement of instances when the launch template (or other
    # watched attributes) changes -- the mechanism that turns a template
    # update into a zero-downtime fleet rollout.
    instance_refresh = optional(object({
      # The refresh strategy. "Rolling" is the only strategy AWS currently
      # supports. Required.
      strategy = string

      # Additional attribute changes that trigger a refresh beyond the
      # launch template (e.g. "tag"). Leave empty for template-only
      # triggers.
      triggers = optional(list(string), [])

      # Fine-grained rollout behavior. Unset keeps AWS defaults (90% min
      # healthy, no surge, no rollback).
      preferences = optional(object({
        # Percentage of desired capacity that must stay InService during the
        # refresh, 0-100. AWS default: 90. Lower = faster, riskier waves --
        # explicit 0 replaces the whole fleet at once, which is why this field
        # is optional: 0 must be distinguishable from unset.
        min_healthy_percentage = optional(number)

        # Upper bound on capacity during the refresh as a percentage of
        # desired, 100-200. Values above 100 let the refresh SURGE (launch
        # before terminate) -- with 110/100 min/max the fleet never dips below
        # full strength. 0 keeps the AWS default (100, no surge).
        max_healthy_percentage = optional(number, 0)

        # Seconds a fresh instance warms before counting toward min-healthy.
        # 0 keeps the group's default_instance_warmup (or health check grace
        # period).
        instance_warmup_seconds = optional(number, 0)

        # Percentage milestones (ascending, each 1-100) where the refresh
        # pauses for checkpoint_delay_seconds -- a staged canary rollout:
        # [10, 50, 100] proves 10% before committing half the fleet.
        checkpoint_percentages = optional(list(number), [])

        # Seconds to wait at each checkpoint before the next wave. AWS
        # default: 3600 (1 hour).
        checkpoint_delay_seconds = optional(number, 0)

        # Roll the fleet back to its previous configuration if the refresh
        # fails (or a watch alarm fires). The safety net that makes
        # template-driven rollouts trustworthy.
        auto_rollback = optional(bool, false)

        # CloudWatch alarms watched during the refresh: any alarm firing
        # fails the refresh (and rolls back when auto_rollback is set).
        # Reference AwsCloudwatchAlarm alarm_name outputs or pass literal
        # alarm names.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        alarms = optional(list(string), [])

        # What happens to instances protected from scale-in: "Ignore" (AWS
        # default -- leave them on the old config), "Refresh" (replace them
        # too), or "Wait" (block until protection is removed).
        scale_in_protected_instances = optional(string, "")

        # What happens to Standby instances: "Ignore" (AWS default),
        # "Terminate", or "Wait".
        standby_instances = optional(string, "")

        # Skip instances that already match the target configuration instead
        # of replacing everything -- resumes interrupted rollouts cheaply.
        skip_matching = optional(bool, false)
      }))
    }))

    # A pool of pre-initialized (stopped, running, or hibernated)
    # instances that dramatically cuts scale-out latency for slow-booting
    # workloads.
    warm_pool = optional(object({
      # The state pooled instances wait in: "Stopped" (AWS default --
      # near-zero compute cost, seconds to start), "Running" (instant but
      # full price), or "Hibernated" (RAM restored from disk -- fast JVM/
      # cache warmup without running cost).
      pool_state = optional(string, "")

      # Minimum number of instances always kept in the pool.
      min_size = optional(number, 0)

      # Ceiling on pool size. Unset keeps the AWS default: the gap between
      # the group's max_size and desired capacity. Explicit 0 is
      # meaningful (no prepared capacity beyond min_size), which is why
      # this field is optional.
      max_group_prepared_capacity = optional(number)

      # Return scaled-in instances to the pool instead of terminating them
      # -- reuse the warm boot instead of paying for it again.
      reuse_on_scale_in = optional(bool, false)
    }))

    # Group-wide health bounds for REPLACEMENT operations (instance
    # refresh, health replacement): the percentage of capacity that must
    # stay in service and the surge allowed above desired.
    instance_maintenance_policy = optional(object({
      # Percentage of desired capacity that must stay InService during
      # replacements, 0-100.
      min_healthy_percentage = optional(number, 0)

      # Upper bound on capacity during replacements as a percentage of
      # desired, 100-200. Set min 100 / max 110 for launch-before-terminate
      # on every replacement.
      max_healthy_percentage = optional(number, 0)
    }))

    # How capacity distributes across availability zones:
    # "balanced-best-effort" (AWS default -- launch in another zone when
    # one is impaired), "balanced-only" (strict balance; launches wait
    # for the impaired zone), or "reservations-then-balanced" (fill
    # targeted Capacity Reservations first, then balance -- pairs with
    # capacity_reservation).
    capacity_distribution_strategy = optional(string, "")

    # Delete the group without waiting for instances to terminate
    # gracefully. Reach for it only when tearing down a wedged group --
    # instances are orphaned mid-flight.
    force_delete = optional(bool, false)

    # How long the IaC engine waits for the group to reach its capacity on
    # create/update, as a duration string (e.g. "10m"). "0" skips the wait
    # entirely. Unset keeps the provider default (10m). An
    # engine-behavior knob (both engines honor it identically), not an AWS
    # API field.
    wait_for_capacity_timeout = optional(string, "")

    # Scaling policies attached to the group. Target tracking is the right
    # default for most services; step/simple react to specific CloudWatch
    # alarms; predictive scaling pre-provisions for forecast load.
    scaling_policies = optional(list(object({
      # Policy name, unique within the group. Required.
      name = string

      # The policy engine. Required.
      # - "TargetTrackingScaling": hold a metric at a target value -- the
      #   right default for services (CPU at 60%, requests-per-target).
      # - "StepScaling": react to a CloudWatch alarm with stepped
      #   adjustments.
      # - "SimpleScaling": the legacy single-step react-and-cooldown model.
      # - "PredictiveScaling": forecast load and pre-provision capacity
      #   (daily/weekly patterns).
      policy_type = string

      # Seconds a new instance warms before its metrics count. 0 keeps the
      # group default. Target tracking and step scaling only.
      estimated_instance_warmup_seconds = optional(number, 0)

      # Configuration for "TargetTrackingScaling".
      target_tracking = optional(object({
        # The value to hold the metric at (e.g. 60.0 for 60% CPU). Required.
        target_value = number

        # Track a predefined group metric: "ASGAverageCPUUtilization",
        # "ASGAverageNetworkIn", "ASGAverageNetworkOut", or
        # "ALBRequestCountPerTarget". Mutually exclusive with
        # customized_metric.
        predefined_metric_type = optional(string, "")

        # Identifies the ALB target group when predefined_metric_type is
        # "ALBRequestCountPerTarget", in the form
        # "app/<lb-name>/<lb-id>/targetgroup/<tg-name>/<tg-id>" (the load
        # balancer's arn_suffix + "/" + the target group's arn_suffix).
        resource_label = optional(string, "")

        # Track a custom CloudWatch metric instead of a predefined one.
        # Mutually exclusive with predefined_metric_type.
        customized_metric = optional(object({
          # The metric name (single-metric form). Mutually exclusive with
          # metrics.
          metric_name = optional(string, "")

          # The metric namespace (e.g. "MyApp/Queue").
          namespace = optional(string, "")

          # The statistic: "Average", "Minimum", "Maximum", "SampleCount", or
          # "Sum".
          statistic = optional(string, "")

          # The metric unit (e.g. "Percent", "Count").
          unit = optional(string, "")

          # Dimensions identifying the metric stream.
          dimensions = optional(list(object({
            # Dimension name (e.g. "QueueName"). Required.
            name = string

            # Dimension value (e.g. "orders"). Required.
            value = string
          })), [])

          # Metric granularity in seconds: 10, 30, or 60. High-resolution
          # metrics (10/30) let target tracking react in seconds.
          period_seconds = optional(number, 0)

          # Metric-math form: a set of query expressions combined into the
          # tracked value (e.g. backlog-per-instance = queue depth / instance
          # count). Mutually exclusive with the single-metric fields.
          metrics = optional(list(object({
            # Short identifier, unique within the query set, referenced by
            # expressions (e.g. "m1", "e1"). Required.
            id = string

            # A metric-math expression over other query ids (e.g. "m1 / m2").
            # Mutually exclusive with metric_stat.
            expression = optional(string, "")

            # A raw metric to fetch. Mutually exclusive with expression.
            metric_stat = optional(object({
              # The metric name. Required.
              metric_name = string

              # The metric namespace. Required.
              namespace = string

              # The statistic to fetch (e.g. "Average", "Sum"). Required.
              stat = string

              # The metric unit.
              unit = optional(string, "")

              # Dimensions identifying the metric stream.
              dimensions = optional(list(object({
                # Dimension name (e.g. "QueueName"). Required.
                name = string

                # Dimension value (e.g. "orders"). Required.
                value = string
              })), [])

              # Granularity in seconds: 10, 30, or 60.
              period_seconds = optional(number, 0)
            }))

            # Human-readable label for the query.
            label = optional(string, "")

            # Whether this entry is the value target tracking consumes. Exactly
            # one entry in the set should return data (AWS default: true --
            # explicitly set false on intermediate entries).
            return_data = optional(bool)
          })), [])
        }))

        # Never scale IN from this policy -- it only adds capacity. For
        # pairing a conservative scale-out tracker with a separate, slower
        # scale-in mechanism.
        disable_scale_in = optional(bool, false)
      }))

      # Configuration for "StepScaling".
      step_scaling = optional(object({
        # How the adjustment numbers are interpreted: "ChangeInCapacity" (add/
        # remove N instances), "ExactCapacity" (set capacity to N), or
        # "PercentChangeInCapacity" (grow/shrink by N%). Required.
        adjustment_type = string

        # How the metric is aggregated across the breach evaluation:
        # "Average" (AWS default), "Minimum", or "Maximum".
        metric_aggregation_type = optional(string, "")

        # With "PercentChangeInCapacity", the minimum number of instances any
        # single step changes -- keeps percentage scaling meaningful on small
        # fleets.
        min_adjustment_magnitude = optional(number, 0)

        # The steps, keyed by breach distance. At least one. Bounds are
        # decimal strings relative to the alarm threshold; an empty bound is
        # open-ended (negative infinity for the first lower bound, positive
        # infinity for the last upper bound).
        step_adjustments = list(object({
          # The capacity change this step applies (interpreted per the policy's
          # adjustment_type; negative shrinks; 0 is only meaningful with
          # "ExactCapacity").
          scaling_adjustment = optional(number, 0)

          # Lower bound of the breach range this step covers, relative to the
          # alarm threshold, as a decimal string (e.g. "0", "10.5"). Empty =
          # negative infinity.
          metric_interval_lower_bound = optional(string, "")

          # Upper bound of the breach range, relative to the alarm threshold,
          # as a decimal string. Empty = positive infinity.
          metric_interval_upper_bound = optional(string, "")
        }))
      }))

      # Configuration for "SimpleScaling".
      simple_scaling = optional(object({
        # How scaling_adjustment is interpreted: "ChangeInCapacity",
        # "ExactCapacity", or "PercentChangeInCapacity". Required.
        adjustment_type = string

        # The capacity change per breach (negative shrinks; 0 is only
        # meaningful with "ExactCapacity" -- scale to zero).
        scaling_adjustment = optional(number, 0)

        # Seconds after a scaling activity before this policy may fire again.
        # 0 keeps the group's default_cooldown.
        cooldown_seconds = optional(number, 0)

        # With "PercentChangeInCapacity", the minimum number of instances any
        # adjustment changes.
        min_adjustment_magnitude = optional(number, 0)
      }))

      # Configuration for "PredictiveScaling".
      predictive_scaling = optional(object({
        # The value to hold the scaling metric at (e.g. 60.0 for 60% CPU).
        # Required.
        target_value = number

        # The predefined load/scaling metric PAIR to forecast:
        # "ASGCPUUtilization", "ASGNetworkIn", "ASGNetworkOut", or
        # "ALBRequestCount". The one-liner for the common cases -- mutually
        # exclusive with the split and customized metric fields below.
        predefined_metric_pair_type = optional(string, "")

        # Identifies the ALB target group when the metric pair is
        # "ALBRequestCount" (load balancer arn_suffix + "/" + target group
        # arn_suffix).
        resource_label = optional(string, "")

        # "ForecastOnly" (AWS default -- observe the forecast before trusting
        # it) or "ForecastAndScale" (act on it).
        mode = optional(string, "")

        # Seconds ahead of the forecasted need that instances launch --
        # boot-and-warm lead time.
        scheduling_buffer_time_seconds = optional(number, 0)

        # What happens when the forecast exceeds max_size:
        # "HonorMaxCapacity" (AWS default) or "IncreaseMaxCapacity" (grow
        # max_size by max_capacity_buffer percent).
        max_capacity_breach_behavior = optional(string, "")

        # Percentage buffer above forecasted capacity when
        # max_capacity_breach_behavior is "IncreaseMaxCapacity", 0-100.
        max_capacity_buffer = optional(number, 0)

        # SPLIT form, load side: the predefined TOTAL metric the forecast is
        # trained on ("ASGTotalCPUUtilization", "ASGTotalNetworkIn",
        # "ASGTotalNetworkOut", "ALBTargetGroupRequestCount"). Pair with
        # predefined_scaling_metric (or customized_scaling_metric_queries).
        predefined_load_metric = optional(object({
          # The predefined metric name. Required.
          metric_type = string

          # Identifies the ALB target group for the ALB-based metric types
          # (load balancer arn_suffix + "/" + target group arn_suffix).
          resource_label = optional(string, "")
        }))

        # SPLIT form, scaling side: the predefined AVERAGE metric capacity is
        # sized against ("ASGAverageCPUUtilization", "ASGAverageNetworkIn",
        # "ASGAverageNetworkOut", "ALBRequestCountPerTarget").
        predefined_scaling_metric = optional(object({
          # The predefined metric name. Required.
          metric_type = string

          # Identifies the ALB target group for the ALB-based metric types
          # (load balancer arn_suffix + "/" + target group arn_suffix).
          resource_label = optional(string, "")
        }))

        # CUSTOMIZED form, load side: a metric-math query set (up to 10
        # entries, exactly one returning data) producing the total-load
        # signal the forecast is trained on. Mutually exclusive with
        # predefined_load_metric.
        customized_load_metric_queries = optional(list(object({
          # Short identifier, unique within the query set, referenced by
          # expressions (e.g. "m1", "e1"). Required.
          id = string

          # A metric-math expression over other query ids (e.g. "m1 / m2").
          # Mutually exclusive with metric_stat.
          expression = optional(string, "")

          # A raw metric to fetch. Mutually exclusive with expression.
          metric_stat = optional(object({
            # The metric name. Required.
            metric_name = string

            # The metric namespace. Required.
            namespace = string

            # The statistic to fetch (e.g. "Average", "Sum"). Required.
            stat = string

            # The metric unit.
            unit = optional(string, "")

            # Dimensions identifying the metric stream.
            dimensions = optional(list(object({
              # Dimension name (e.g. "QueueName"). Required.
              name = string

              # Dimension value (e.g. "orders"). Required.
              value = string
            })), [])

            # Granularity in seconds: 10, 30, or 60.
            period_seconds = optional(number, 0)
          }))

          # Human-readable label for the query.
          label = optional(string, "")

          # Whether this entry is the value target tracking consumes. Exactly
          # one entry in the set should return data (AWS default: true --
          # explicitly set false on intermediate entries).
          return_data = optional(bool)
        })), [])

        # CUSTOMIZED form, scaling side: a metric-math query set producing
        # the per-instance utilization signal capacity is sized against.
        # Mutually exclusive with predefined_scaling_metric.
        customized_scaling_metric_queries = optional(list(object({
          # Short identifier, unique within the query set, referenced by
          # expressions (e.g. "m1", "e1"). Required.
          id = string

          # A metric-math expression over other query ids (e.g. "m1 / m2").
          # Mutually exclusive with metric_stat.
          expression = optional(string, "")

          # A raw metric to fetch. Mutually exclusive with expression.
          metric_stat = optional(object({
            # The metric name. Required.
            metric_name = string

            # The metric namespace. Required.
            namespace = string

            # The statistic to fetch (e.g. "Average", "Sum"). Required.
            stat = string

            # The metric unit.
            unit = optional(string, "")

            # Dimensions identifying the metric stream.
            dimensions = optional(list(object({
              # Dimension name (e.g. "QueueName"). Required.
              name = string

              # Dimension value (e.g. "orders"). Required.
              value = string
            })), [])

            # Granularity in seconds: 10, 30, or 60.
            period_seconds = optional(number, 0)
          }))

          # Human-readable label for the query.
          label = optional(string, "")

          # Whether this entry is the value target tracking consumes. Exactly
          # one entry in the set should return data (AWS default: true --
          # explicitly set false on intermediate entries).
          return_data = optional(bool)
        })), [])

        # CUSTOMIZED form, capacity side: a metric-math query set reporting
        # the group's current capacity -- needed only when the scaling metric
        # is a custom signal whose relationship to instance count AWS cannot
        # infer.
        customized_capacity_metric_queries = optional(list(object({
          # Short identifier, unique within the query set, referenced by
          # expressions (e.g. "m1", "e1"). Required.
          id = string

          # A metric-math expression over other query ids (e.g. "m1 / m2").
          # Mutually exclusive with metric_stat.
          expression = optional(string, "")

          # A raw metric to fetch. Mutually exclusive with expression.
          metric_stat = optional(object({
            # The metric name. Required.
            metric_name = string

            # The metric namespace. Required.
            namespace = string

            # The statistic to fetch (e.g. "Average", "Sum"). Required.
            stat = string

            # The metric unit.
            unit = optional(string, "")

            # Dimensions identifying the metric stream.
            dimensions = optional(list(object({
              # Dimension name (e.g. "QueueName"). Required.
              name = string

              # Dimension value (e.g. "orders"). Required.
              value = string
            })), [])

            # Granularity in seconds: 10, 30, or 60.
            period_seconds = optional(number, 0)
          }))

          # Human-readable label for the query.
          label = optional(string, "")

          # Whether this entry is the value target tracking consumes. Exactly
          # one entry in the set should return data (AWS default: true --
          # explicitly set false on intermediate entries).
          return_data = optional(bool)
        })), [])
      }))

      # Suspend this policy without deleting it: the policy and its
      # CloudWatch alarms stay configured but stop acting on the group.
      # The pause button for incident response or load tests -- deleting
      # the policy instead would discard alarm history and forecast state.
      disabled = optional(bool, false)
    })), [])

    # Time-based capacity changes (cron or one-shot): business-hours
    # scale-up, overnight scale-down, batch-window pre-provisioning.
    scheduled_actions = optional(list(object({
      # Action name, unique within the group. Required.
      name = string

      # Cron expression in UTC (or time_zone), e.g. "0 8 * * MON-FRI" for
      # business-hours scale-up. Leave empty for a one-shot action at
      # start_time.
      recurrence = optional(string, "")

      # IANA time zone for the recurrence (e.g. "America/New_York"). Unset
      # = UTC.
      time_zone = optional(string, "")

      # First (or only) trigger time, RFC3339 UTC (e.g.
      # "2026-08-01T08:00:00Z").
      start_time = optional(string, "")

      # Last trigger time for a recurring action, RFC3339 UTC.
      end_time = optional(string, "")

      # New min_size when the action fires. Absent = leave unchanged
      # (which is why these are optional: 0 is a meaningful new value).
      min_size = optional(number)

      # New max_size when the action fires. Absent = leave unchanged.
      max_size = optional(number)

      # New desired capacity when the action fires. Absent = leave
      # unchanged.
      desired_capacity = optional(number)
    })), [])

    # Pause points in the instance lifecycle: run custom logic (warm a
    # cache, drain work, pull logs) while an instance waits in a
    # launching or terminating state.
    lifecycle_hooks = optional(list(object({
      # Hook name, unique within the group. Required.
      name = string

      # The transition the hook pauses:
      # "autoscaling:EC2_INSTANCE_LAUNCHING" or
      # "autoscaling:EC2_INSTANCE_TERMINATING". Required.
      lifecycle_transition = string

      # What happens when the heartbeat times out without a completion
      # signal: "ABANDON" (AWS default for launch hooks -- roll the
      # instance back) or "CONTINUE" (proceed anyway).
      default_result = optional(string, "")

      # Seconds the instance waits in the transition state before
      # default_result applies, 30-7200. AWS default: 3600.
      heartbeat_timeout_seconds = optional(number, 0)

      # Where the pause notification is delivered: an SNS topic or SQS
      # queue ARN. Reference an AwsSnsTopic's topic_arn output or pass a
      # literal ARN. Unset relies on EventBridge rules watching lifecycle
      # events (the modern pattern).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      notification_target_arn = optional(string, "")

      # The IAM role Auto Scaling assumes to publish to the notification
      # target. Required by AWS when notification_target_arn is set.
      # Reference an AwsIamRole's role_arn output or pass a literal ARN.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      role_arn = optional(string, "")

      # Free-form JSON delivered with every notification -- routing context
      # for the consumer.
      notification_metadata = optional(string, "")

      # Attach this hook atomically AT GROUP CREATION instead of as a
      # separate post-creation resource. Without it, instances the group
      # launches in the seconds before the standalone hook attaches slip
      # through unhooked -- set it on launch-transition hooks that must
      # catch the very first instance. Trade-off: AWS makes creation-time
      # hooks immutable, so any change to a flagged hook REPLACES the whole
      # group; leave it off (the default) for hooks that need in-place
      # updates.
      apply_at_launch = optional(bool, false)
    })), [])

    # SNS notifications for fleet lifecycle events -- the simplest way to
    # observe launches, terminations, and their failures.
    notifications = optional(object({
      # The SNS topic events are published to. Reference an AwsSnsTopic's
      # topic_arn output or pass a literal ARN. Required.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      topic = string

      # The event types to publish. At least one of:
      # "autoscaling:EC2_INSTANCE_LAUNCH",
      # "autoscaling:EC2_INSTANCE_LAUNCH_ERROR",
      # "autoscaling:EC2_INSTANCE_TERMINATE",
      # "autoscaling:EC2_INSTANCE_TERMINATE_ERROR".
      event_types = list(string)
    }))

    # Launch into EC2 Capacity Reservations -- guaranteed capacity a
    # reserved fleet has already paid for. Leave unset for the AWS
    # default behavior (use an open reservation when one matches).
    capacity_reservation = optional(object({
      # How launches use reservations:
      # - "default": AWS account default (open reservations match
      #   automatically).
      # - "capacity-reservations-only": launch ONLY into the targeted
      #   reservations -- fail rather than fall back to on-demand pool
      #   capacity.
      # - "capacity-reservations-first": try the reservations, fall back
      #   to regular capacity when exhausted.
      # - "none": never consume reservations, even matching open ones.
      preference = optional(string, "")

      # Specific Capacity Reservation IDs to launch into (e.g. "cr-...").
      # Mutually exclusive with capacity_reservation_resource_group_arns.
      capacity_reservation_ids = optional(list(string), [])

      # Resource-group ARNs that collect Capacity Reservations -- target
      # the group instead of chasing individual reservation IDs. Mutually
      # exclusive with capacity_reservation_ids.
      capacity_reservation_resource_group_arns = optional(list(string), [])
    }))

    # Traffic sources this group registers its instances with -- the
    # generalized successor to load-balancer attachment that also covers
    # VPC Lattice target groups. For ALB/NLB target groups prefer
    # target_groups (typed references); use this for VPC Lattice (no
    # Planton kind yet -- pass the target group ARN) or Classic ELBs.
    # Mutually exclusive with target_groups (the provider rejects the
    # combination).
    traffic_sources = optional(list(object({
      # What identifies the source: a VPC Lattice target group ARN, an
      # ALB/NLB target group ARN, or a Classic ELB name. Reference another
      # resource's output or pass the literal value.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      identifier = string

      # The source type: "vpc-lattice", "elbv2" (ALB/NLB target group), or
      # "elb" (Classic). AWS infers it from the identifier when unset --
      # set it explicitly for Classic ELB names, which look like plain
      # strings.
      type = optional(string, "")
    })), [])

    # What happens to instances whose TERMINATING lifecycle hook ends in
    # ABANDON -- terminate anyway (AWS default) or retain the instance
    # for debugging. Only meaningful with a terminating-transition
    # lifecycle hook.
    instance_lifecycle_policy = optional(object({
      # When a terminating instance's lifecycle hook ends in ABANDON:
      # "terminate" (AWS default -- proceed with termination) or "retain"
      # (keep the instance out of the group but running, for post-mortem
      # debugging of whatever made the hook fail).
      terminate_hook_abandon = optional(string, "")
    }))

    # Keep IaC applies moving while a scaling activity is failing: read
    # the group state without waiting on the failed activity. For groups
    # whose scaling errors are handled by their own alarms rather than
    # the deploy pipeline.
    ignore_failed_scaling_activities = optional(bool, false)

    # When force_delete tears the group down, also force-delete its warm
    # pool without draining the pooled instances.
    force_delete_warm_pool = optional(bool, false)

    # Minimum number of instances that must pass ELB health checks
    # before a CREATE is considered successful. Requires an attached
    # load balancer (target_groups or traffic_sources) and health checks
    # that can pass during the wait. An engine-behavior wait (both
    # engines honor it identically), not an AWS API field -- like
    # wait_for_capacity_timeout.
    min_elb_capacity = optional(number, 0)

    # Exact number of ELB-healthy instances to wait for on create AND
    # every update (min_elb_capacity waits on create only). Takes
    # precedence over min_elb_capacity when both are set. An
    # engine-behavior wait, not an AWS API field.
    wait_for_elb_capacity = optional(number, 0)
  })
}
