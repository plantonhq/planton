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
  description = "AwsEcsService specification"
  type = object({
    # The AWS region the service is created in. Must match the cluster's and
    # task definition's region.
    # Example: "us-west-2", "eu-west-1".
    region = string

    # The ECS cluster the service runs in. Reference an AwsEcsCluster's
    # cluster_arn output or pass a literal cluster ARN.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    cluster_arn = string

    # The task definition revision the service runs. Reference an
    # AwsEcsTaskDefinition's task_definition_arn output (the recommended
    # wiring -- each new revision changes the output and rolls the service)
    # or pass a literal "family:revision" / full ARN.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    task_definition = string

    # The number of task copies to keep running. Explicit 0 deploys the
    # service with nothing running (the wiring exists; scale up later).
    # When autoscaling is configured, this only seeds the initial count --
    # both modules then leave the live count to the scaler (and to
    # operators) rather than fighting it on every apply.
    desired_count = optional(number)

    # The launch type: "FARGATE" (serverless -- the default), "EC2"
    # (container instances you manage), or "EXTERNAL" (ECS Anywhere).
    # Mutually exclusive with capacity_provider_strategy: name a type OR
    # blend providers, not both.
    launch_type = optional(string, "")

    # Blend capacity across providers instead of naming one launch type --
    # the cost-optimization lever. Example: FARGATE base 1 / weight 1 +
    # FARGATE_SPOT weight 4 keeps one guaranteed on-demand task and runs
    # ~80% of scaled capacity on Spot. EC2 clusters list their
    # auto-scaling-group-backed capacity providers by name.
    capacity_provider_strategy = optional(list(object({
      # The capacity provider: "FARGATE", "FARGATE_SPOT", or the name of an
      # EC2 capacity provider attached to the cluster.
      capacity_provider = string

      # Minimum number of tasks guaranteed on this provider before weights
      # apply. Only one entry of the strategy may set a non-zero base.
      base = optional(number, 0)

      # Relative share of tasks beyond the bases. Example: weight 1 on
      # FARGATE + weight 4 on FARGATE_SPOT scales 1:4 on-demand:Spot.
      weight = optional(number, 0)
    })), [])

    # The Fargate platform version (e.g. "1.4.0" or "LATEST"). Fargate
    # only; leave unset to track LATEST.
    platform_version = optional(string, "")

    # Task placement across the cluster: "REPLICA" (the default -- run
    # desired_count copies wherever they fit) or "DAEMON" (exactly one task
    # per container instance -- EC2 only, for host agents like log shippers
    # and monitors; desired_count and autoscaling do not apply).
    scheduling_strategy = optional(string, "")

    # VPC networking for the task ENIs. Required for tasks whose definition
    # uses "awsvpc" networking -- which is every Fargate task and the modern
    # EC2 posture. Omit only for EC2 bridge/host-mode task definitions.
    network = optional(object({
      # The subnets task ENIs are placed in -- private subnets for production
      # services, at least two AZs for availability. Reference AwsSubnet
      # subnet_id outputs or pass literal subnet IDs.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      subnets = list(string)

      # Security groups applied to each task ENI. Reference AwsSecurityGroup
      # security_group_id outputs or pass literal group IDs. Unset falls back
      # to the VPC's default security group -- not a production posture.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      security_groups = optional(list(string), [])

      # Assign each task ENI a public IPv4 address. Only for tasks in public
      # subnets that must reach the internet without a NAT gateway; keep
      # false for private-subnet services.
      assign_public_ip = optional(bool, false)
    }))

    # Load balancer wiring: which container/port registers into which
    # target group. The target group is a first-class AwsLbTargetGroup the
    # listener (rule) routes into -- the service only registers task IPs
    # there. Multiple entries register multiple container ports (e.g. an
    # app port behind the public listener and a metrics port behind an
    # internal one). AWS requires each referenced target group to already
    # be associated with a load balancer listener at service creation.
    load_balancers = optional(list(object({
      # The target group tasks register into. Reference an AwsLbTargetGroup's
      # target_group_arn output (the group must use target type "ip" for
      # awsvpc tasks) or pass a literal ARN.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      target_group_arn = string

      # The container (by task-definition container name) that receives the
      # traffic.
      container_name = string

      # The container port that receives the traffic.
      container_port = optional(number, 0)

      # Blue/green target-group pair for the BLUE_GREEN deployment strategy:
      # ECS shifts the production listener rule between this entry's target
      # group and the alternate as deployments bake. Requires
      # deployment_configuration.strategy = "BLUE_GREEN".
      advanced_configuration = optional(object({
        # The second target group of the blue/green pair. Reference an
        # AwsLbTargetGroup's target_group_arn output.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        alternate_target_group_arn = string

        # The production listener rule whose forward action ECS swaps between
        # the two target groups. Reference an AwsLbListenerRule's rule_arn
        # output.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        production_listener_rule = string

        # An optional test listener rule pointed at the green tasks before the
        # production swap -- how smoke traffic reaches the new version during
        # the bake.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        test_listener_rule = optional(string, "")

        # The IAM role ECS assumes to modify the listener rules during the
        # swap. Reference an AwsIamRole's role_arn output.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        role_arn = string
      }))
    })), [])

    # Seconds ECS ignores load-balancer health-check failures after a task
    # starts, so slow-booting apps are not killed mid-startup. Only valid
    # with load_balancers (no platform default, since most services set it
    # only when fronting traffic). Recommended: 60-120 for typical apps.
    health_check_grace_period_seconds = optional(number)

    # Upper bound on running tasks during a deployment, as a percentage of
    # desired_count. 200 (the AWS default) starts a full replacement set
    # before draining the old one; 100 forces in-place replacement (needed
    # when capacity is tight).
    deployment_maximum_percent = optional(number)

    # Lower bound on healthy running tasks during a deployment, as a
    # percentage of desired_count. 100 (the AWS default) never dips below
    # desired capacity; lower values trade headroom for faster rollouts.
    deployment_minimum_healthy_percent = optional(number)

    # The deployment circuit breaker: stop a rollout whose tasks keep
    # failing to reach steady state, and optionally roll back to the last
    # healthy deployment. The zero-configuration deployment guard -- enable
    # both for every production service.
    deployment_circuit_breaker = optional(object({
      # Enable the circuit breaker.
      enable = optional(bool, false)

      # Roll back to the last steady-state deployment when the breaker
      # trips, instead of leaving the service stuck in a failing rollout.
      rollback = optional(bool, false)
    }))

    # CloudWatch alarms that gate deployments: while a deployment is in
    # progress, if any referenced alarm fires, ECS marks the deployment
    # failed and (optionally) rolls back. Catches regressions the circuit
    # breaker cannot see -- error rates, latency, business metrics.
    alarms = optional(object({
      # The alarms to watch, by NAME (the CloudWatch alarm API keys on names,
      # not ARNs). Reference AwsCloudwatchAlarm alarm_name outputs or pass
      # literal alarm names.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      alarm_names = list(string)

      # Enable alarm-gated deployments.
      enable = optional(bool, false)

      # Roll back to the last steady-state deployment when an alarm fires
      # mid-deployment.
      rollback = optional(bool, false)
    }))

    # Advanced deployment behavior: the ROLLING/BLUE_GREEN strategy choice,
    # bake time, canary/linear traffic shifting, and deployment lifecycle
    # hooks. Leave unset for plain rolling deployments.
    deployment_configuration = optional(object({
      # "ROLLING" or "BLUE_GREEN". BLUE_GREEN requires load_balancers entries
      # with advanced_configuration (the target-group pair to swap between).
      strategy = optional(string, "")

      # Minutes the new (green) version serves production traffic before the
      # old one is drained -- the window in which alarms or a manual check
      # can still trigger an instant rollback. 0-1440.
      bake_time_in_minutes = optional(number)

      # Canary traffic shifting: send a fixed percentage to green, bake, then
      # shift the rest. Mutually exclusive with linear_configuration.
      canary_configuration = optional(object({
        # The percentage of traffic shifted in the first step (0.1-100).
        canary_percent = optional(number, 0)

        # Minutes the canary bakes before the remaining traffic shifts. 0-1440.
        canary_bake_time_in_minutes = optional(number)
      }))

      # Linear traffic shifting: shift in equal percentage steps with a bake
      # between steps. Mutually exclusive with canary_configuration.
      linear_configuration = optional(object({
        # The percentage shifted per step (3-100).
        step_percent = optional(number, 0)

        # Minutes each step bakes before the next. 0-1440.
        step_bake_time_in_minutes = optional(number)
      }))

      # Lambda hooks invoked at chosen stages of the deployment lifecycle --
      # run integration tests against the green stack before traffic shifts,
      # notify, or veto.
      lifecycle_hooks = optional(list(object({
        # The Lambda function ECS invokes. A literal function ARN.
        hook_target_arn = string

        # The IAM role ECS assumes to invoke the hook. Reference an AwsIamRole's
        # role_arn output.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        role_arn = string

        # The deployment stages the hook fires at.
        lifecycle_stages = list(string)

        # Free-form JSON passed to the hook invocation -- deployment context
        # the function needs (environment names, test-suite selectors).
        hook_details = optional(string, "")
      })), [])
    }))

    # Who orchestrates deployments: "ECS" (the default -- rolling and native
    # blue/green), "CODE_DEPLOY" (AWS CodeDeploy drives blue/green), or
    # "EXTERNAL" (a third-party controller owns deployments).
    deployment_controller = optional(string, "")

    # Service Connect: ECS-managed service-to-service networking on top of
    # AWS Cloud Map -- sidecar-free discovery ("call http://orders") with
    # per-request telemetry, retries, and optional TLS between services.
    service_connect = optional(object({
      # Enable Service Connect for this service. A service with enabled =
      # true and no services entries is a CLIENT: it can call other Service
      # Connect services in the namespace but exposes nothing itself.
      enabled = optional(bool, false)

      # The Cloud Map namespace (name or ARN) the mesh lives in. Unset falls
      # back to the cluster's default Service Connect namespace. A literal
      # value -- Planton has no Cloud Map kind yet.
      namespace = optional(string, "")

      # The ports this service EXPOSES to the mesh. Each entry publishes one
      # named task-definition port under a discovery name siblings call.
      services = optional(list(object({
        # The name of a port mapping in the task definition (the port_mappings
        # entry must set name). This is the join key between the service and
        # its task definition.
        port_name = string

        # The discovery name published to the namespace. Default: port_name.
        discovery_name = optional(string, "")

        # The port and DNS name CLIENTS use to call this service.
        client_alias = optional(object({
          # The port clients connect to (often the same as the container port).
          port = optional(number, 0)

          # The DNS name clients use (e.g. "orders" or "orders.internal").
          # Default: the discovery name.
          dns_name = optional(string, "")

          # Route requests carrying a matching header to this service's TEST
          # revision during a blue/green deployment -- how testers reach the new
          # version through the mesh before traffic shifts.
          test_traffic_rules = optional(list(object({
            # The header match selecting test traffic.
            header = optional(object({
              # The header name to match (e.g. "x-canary-test").
              name = string

              # The header value match.
              value = object({
                # The exact header value that selects test traffic.
                exact = string
              })
            }))
          })), [])
        }))

        # Override the ingress port the proxy listens on for this service --
        # for interposing network appliances; rarely needed.
        ingress_port_override = optional(number)

        # Proxy timeouts for calls TO this service.
        timeout = optional(object({
          # Seconds an idle connection stays open. 0 disables the idle timeout.
          idle_timeout_seconds = optional(number, 0)

          # Seconds a single request may take end to end. 0 disables the
          # per-request timeout.
          per_request_timeout_seconds = optional(number, 0)
        }))

        # TLS between mesh services, with certificates issued by a Private CA.
        tls = optional(object({
          # The AWS Private Certificate Authority that issues the service's
          # certificates. A literal PCA ARN.
          aws_pca_authority_arn = string

          # The KMS key that protects the private key material. Reference an
          # AwsKmsKey's key_arn output or pass a literal key ARN.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          kms_key = optional(string, "")

          # The IAM role ECS assumes to request certificates. Reference an
          # AwsIamRole's role_arn output.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          role_arn = optional(string, "")
        }))
      })), [])

      # Where the Service Connect proxy's own logs go (the injected agent's
      # logs, not the application's).
      log_configuration = optional(object({
        # The log driver: "awslogs" (CloudWatch), "awsfirelens", "splunk",
        # "fluentd", "gelf", "syslog", "journald", or "json-file".
        log_driver = string

        # Driver-specific options (e.g. awslogs-group / awslogs-region /
        # awslogs-stream-prefix).
        options = optional(map(string), {})

        # Driver options whose values come from Secrets Manager / SSM (name ->
        # ARN), resolved by the ECS agent at task start.
        secret_options = optional(map(string), {})
      }))

      # Per-request access logging by the Service Connect proxy -- one line
      # per mesh request (caller, target, status, latency), emitted to the
      # proxy's log destination above. The mesh-level equivalent of load
      # balancer access logs.
      access_log_configuration = optional(object({
        # The log line format: "TEXT" or "JSON" (structured -- the right choice
        # when logs feed a query engine).
        format = optional(string, "")

        # Include request query parameters in the log lines: "ENABLED" or
        # "DISABLED". Unset keeps AWS's default (disabled -- query strings
        # often carry sensitive values).
        include_query_parameters = optional(string, "")
      }))
    }))

    # Legacy AWS Cloud Map service discovery registration (DNS-based). For
    # new meshes prefer service_connect; use this when other consumers
    # already resolve the Cloud Map DNS name.
    service_registries = optional(object({
      # The Cloud Map service to register into. A literal registry ARN --
      # Planton has no Cloud Map kind yet.
      registry_arn = string

      # For SRV records on bridge/host-mode tasks: the container name (from
      # the task definition) whose address is published.
      container_name = optional(string, "")

      # For SRV records on bridge/host-mode tasks: the container port
      # published alongside container_name.
      container_port = optional(number)

      # For SRV records on awsvpc tasks: the port published with the task IP.
      port = optional(number)
    }))

    # A per-deployment managed EBS volume attached to each task, configured
    # at deployment time against a volume the task definition declares with
    # configure_at_launch (AwsEcsTaskDefinition spec.volumes) -- name must
    # match that volume's name. How ECS tasks get real block storage beyond
    # ephemeral scratch space.
    volume_configuration = optional(object({
      # The volume name -- must match a volume name in the task definition.
      name = string

      # The EBS volume ECS creates and attaches per task.
      managed_ebs_volume = object({
        # The IAM role ECS assumes to create, attach, and delete the volumes
        # (needs the AmazonECSInfrastructureRolePolicyForVolumes managed
        # policy). Reference an AwsIamRole's role_arn output.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        role_arn = string

        # Volume size in GiB. Required unless snapshot_id is set (the snapshot
        # then defines the minimum size).
        size_in_gb = optional(number, 0)

        # Volume type: "gp3" (the sensible default), "gp2", "io1", "io2",
        # "st1", "sc1", or "standard".
        volume_type = optional(string, "")

        # Provisioned IOPS -- required for io1/io2, optional for gp3.
        iops = optional(number, 0)

        # Throughput in MiB/s, gp3 only (125-1000).
        throughput = optional(number, 0)

        # Encrypt the volume at rest. AWS default: true (and account-level
        # EBS-encryption-by-default may enforce it regardless). Optional so an
        # explicit false is distinguishable from unset.
        encrypted = optional(bool)

        # The KMS key for encryption. Reference an AwsKmsKey's key_arn output
        # or pass a literal key ARN; unset with encryption uses the AWS-managed
        # aws/ebs key.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        kms_key_id = optional(string, "")

        # Create each volume from this snapshot instead of empty.
        snapshot_id = optional(string, "")

        # The filesystem ECS formats the volume with: "xfs" (default), "ext4",
        # "ext3", or "ntfs" (Windows tasks).
        file_system_type = optional(string, "")

        # Tags applied to each created EBS volume at creation time -- without
        # this, per-task volumes carry no cost-allocation tags at all.
        tag_specifications = optional(list(object({
          # The resource type being tagged; "volume" is the only type EBS task
          # volumes support.
          resource_type = optional(string, "")

          # The tags applied to each created volume.
          tags = optional(map(string), {})

          # Also propagate tags from "SERVICE" or "TASK_DEFINITION" ("NONE"
          # disables propagation) onto the created volumes.
          propagate_tags = optional(string, "")
        })), [])

        # How fast a snapshot-restored volume hydrates, in MiB/s. Only
        # meaningful with snapshot_id; 0 (unset) uses the EBS default lazy
        # loading.
        volume_initialization_rate = optional(number, 0)
      })
    }))

    # Task placement strategies, applied in order (EC2 launch type only --
    # Fargate places tasks itself). Example: spread across AZs, then
    # binpack on memory.
    ordered_placement_strategy = optional(list(object({
      # "spread" (distribute across the field's values), "binpack" (fill the
      # least-remaining field first -- densest packing), or "random".
      type = optional(string, "")

      # What to spread over or binpack on: "attribute:ecs.availability-zone",
      # "instanceId" for spread; "cpu" or "memory" for binpack. Unused for
      # random.
      field = optional(string, "")
    })), [])

    # Task placement constraints (EC2 launch type only). Example: memberOf
    # "attribute:ecs.instance-type =~ m5.*".
    placement_constraints = optional(list(object({
      # "memberOf" (instances matching a cluster query expression) or
      # "distinctInstance" (never co-locate two of this service's tasks).
      type = optional(string, "")

      # The cluster query for memberOf, e.g.
      # "attribute:ecs.instance-type =~ m5.*". Unused for distinctInstance.
      expression = optional(string, "")
    })), [])

    # Automatic redistribution of tasks across availability zones when AZs
    # become unbalanced (after an AZ event, tasks pile into the surviving
    # zones and stay there): "ENABLED" or "DISABLED". Unset lets AWS decide
    # -- new services default to ENABLED where supported.
    availability_zone_rebalancing = optional(string, "")

    # Propagate tags to tasks from "SERVICE" or "TASK_DEFINITION" ("NONE"
    # disables propagation). Task-level tags are how per-task cost
    # allocation works.
    propagate_tags = optional(string, "")

    # Let ECS add its managed cluster/service tags to tasks for cost and
    # usage attribution.
    enable_ecs_managed_tags = optional(bool, false)

    # Enable ECS Exec: interactive shells into running containers through
    # SSM ("kubectl exec" for ECS). The task role needs the SSM messages
    # permissions; the cluster's execute_command_configuration governs
    # session audit logging.
    enable_execute_command = optional(bool, false)

    # Force-delete the service even while it still has running tasks --
    # destroy skips the scale-to-zero-first dance. Appropriate for
    # ephemeral environments; leave false where a stuck deletion should be
    # investigated instead.
    force_delete = optional(bool, false)

    # Target-tracking autoscaling of desired_count via Application Auto
    # Scaling. Folded into the service because the scaler's identity IS
    # this service (one scalable target per service); the CloudWatch alarms
    # the policies create are managed by AWS. Not applicable to DAEMON
    # services.
    autoscaling = optional(object({
      # The floor. 0 allows scale-to-zero (with a metric that can get there,
      # e.g. a custom metric -- CPU tracking never reaches zero tasks).
      min_tasks = optional(number, 0)

      # The ceiling -- also the cost guardrail.
      max_tasks = optional(number, 0)

      # Track average CPU utilization across tasks (e.g. target 70 scales to
      # hold CPU near 70%). The bread-and-butter policy for compute-bound
      # services.
      cpu = optional(object({
        # The utilization percentage to hold (1-100). 70-75 is the usual
        # production sweet spot: headroom for spikes without paying for idle.
        target_percent = optional(number, 0)

        # Seconds to wait after a scale-in before another may follow. Longer
        # cooldowns damp flapping. AWS default: 300.
        scale_in_cooldown_seconds = optional(number)

        # Seconds to wait after a scale-out before another may follow. Keep it
        # short -- under-capacity hurts more than a brief overshoot. AWS
        # default: 60.
        scale_out_cooldown_seconds = optional(number)

        # Only ever scale out on this policy; never remove capacity. For
        # services where a human decides when to scale in.
        disable_scale_in = optional(bool, false)
      }))

      # Track average memory utilization across tasks. Note memory rarely
      # shrinks under reduced load -- prefer CPU or request tracking for
      # scale-in behavior.
      memory = optional(object({
        # The utilization percentage to hold (1-100). 70-75 is the usual
        # production sweet spot: headroom for spikes without paying for idle.
        target_percent = optional(number, 0)

        # Seconds to wait after a scale-in before another may follow. Longer
        # cooldowns damp flapping. AWS default: 300.
        scale_in_cooldown_seconds = optional(number)

        # Seconds to wait after a scale-out before another may follow. Keep it
        # short -- under-capacity hurts more than a brief overshoot. AWS
        # default: 60.
        scale_out_cooldown_seconds = optional(number)

        # Only ever scale out on this policy; never remove capacity. For
        # services where a human decides when to scale in.
        disable_scale_in = optional(bool, false)
      }))

      # Track requests-per-target on the load balancer -- the most direct
      # signal for request-serving services, reacting before CPU climbs.
      requests_per_target = optional(object({
        # Requests per target per minute to hold (e.g. 1000). Derive it from a
        # load test: the RPS one task handles at your latency budget, times 60.
        target_requests_per_target = optional(number, 0)

        # The load balancer's ARN suffix (e.g. "app/my-alb/50dc6c495c0c9188")
        # -- the CloudWatch LoadBalancer dimension. Reference an AwsAlb's
        # arn_suffix output.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        load_balancer_arn_suffix = string

        # The target group's ARN suffix (e.g. "targetgroup/api/943f017f100becff")
        # -- the CloudWatch TargetGroup dimension. Reference an AwsLbTargetGroup's
        # arn_suffix output; use the group this service registers into.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        target_group_arn_suffix = string

        # Scale-in cooldown, seconds. AWS default: 300.
        scale_in_cooldown_seconds = optional(number)

        # Scale-out cooldown, seconds. AWS default: 60.
        scale_out_cooldown_seconds = optional(number)

        # Only ever scale out on this policy; never remove capacity.
        disable_scale_in = optional(bool, false)
      }))
    }))

    # VPC Lattice target-group attachments: register this service's tasks
    # into VPC Lattice target groups so Lattice services route to them --
    # the application-network alternative to a load balancer for
    # cross-VPC/cross-account traffic. Each entry names the target group,
    # the port name (from the task definition's port_mappings) to register,
    # and the infrastructure role ECS assumes to manage the registration.
    vpc_lattice_configurations = optional(list(object({
      # The IAM role ECS assumes to register and deregister task targets with
      # VPC Lattice (the ECS service principal must be able to assume it, and
      # it needs the Lattice target-management permissions). Reference an
      # AwsIamRole's role_arn output or pass a literal ARN.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      role_arn = string

      # The VPC Lattice target group (by ARN) the tasks register into. A
      # literal ARN -- Planton has no VPC Lattice kinds yet.
      target_group_arn = string

      # The name of a port mapping in the task definition (the port_mappings
      # entry must set name) whose port is registered with the target group.
      # 1-64 characters: lowercase letters, digits, underscores, hyphens.
      port_name = string
    })), [])
  })
}
