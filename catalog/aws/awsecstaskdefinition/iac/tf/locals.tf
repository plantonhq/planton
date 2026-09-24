locals {
  # The task-definition FAMILY is the manifest name; every apply that changes
  # anything registers a new immutable revision of this family.
  family = var.metadata.name

  # Resource-identity tags, matching the Pulumi module key-for-key.
  aws_tags = {
    "Name"                     = local.family
    "planton.ai/resource"      = "true"
    "planton.ai/organization"  = var.metadata.org
    "planton.ai/environment"   = var.metadata.env
    "planton.ai/resource-kind" = "AwsEcsTaskDefinition"
    "planton.ai/resource-id"   = var.metadata.id
  }

  # Foreign-key fields are already flattened to primitive strings by the tofu
  # generator (the orchestrator resolves any value_from before the module
  # runs), so refs are consumed directly.

  # Empty means FARGATE -- the serverless default this catalog leads with.
  requires_compatibilities = length(var.spec.requires_compatibilities) > 0 ? var.spec.requires_compatibilities : ["FARGATE"]

  # awsvpc is both the Fargate requirement and the modern EC2 posture, so it
  # is the default rather than AWS's launch-type-dependent one.
  network_mode = var.spec.network_mode != "" ? var.spec.network_mode : "awsvpc"

  # Zero task-level sizing means unset -- valid only for EC2/EXTERNAL, which
  # the spec's CEL already guarantees. AWS takes the values as strings.
  task_cpu    = var.spec.cpu > 0 ? tostring(var.spec.cpu) : null
  task_memory = var.spec.memory > 0 ? tostring(var.spec.memory) : null

  execution_role_arn = var.spec.execution_role != "" ? var.spec.execution_role : null
  task_role_arn      = var.spec.task_role != "" ? var.spec.task_role : null

  # ---------------------------------------------------------------------
  # Default logging: one shared log group, one stream prefix per container.
  # The group NAME is decided here so the container-definitions document can
  # embed it as a plain string; the aws_ecs_task_definition resource then
  # depends on the group so it exists before any task launches (a missing
  # group fails at task START, not at registration).
  # ---------------------------------------------------------------------
  logging_disabled     = var.spec.logging != null ? var.spec.logging.disabled : false
  referenced_log_group = var.spec.logging != null ? var.spec.logging.log_group : ""
  create_log_group     = !local.logging_disabled && local.referenced_log_group == ""
  log_group_name       = local.logging_disabled ? "" : (local.referenced_log_group != "" ? local.referenced_log_group : "/ecs/${local.family}")
  # retention_days is proto-optional (null when unset); the platform default
  # is 30 days for the auto-created group.
  log_retention_days = var.spec.logging != null ? coalesce(var.spec.logging.retention_days, 30) : 30

  # ---------------------------------------------------------------------
  # Container definitions, rendered from the structured spec into the JSON
  # document the ECS API takes. Optional attributes are emitted only when
  # set so the registered document stays minimal and deterministic; maps
  # become sorted lists via the keys() iteration order (Terraform sorts map
  # keys), keeping revisions stable across applies.
  # ---------------------------------------------------------------------
  container_definitions = [
    for container in var.spec.containers : merge(
      {
        name  = container.name
        image = container.image
        # AWS defaults essential to true when omitted; emitting it
        # explicitly keeps the registered document identical to intent.
        # essential is a proto-optional tri-state: null means unset (true).
        essential = container.essential == null ? true : container.essential
      },
      container.cpu > 0 ? { cpu = container.cpu } : {},
      container.memory > 0 ? { memory = container.memory } : {},
      container.memory_reservation > 0 ? { memoryReservation = container.memory_reservation } : {},
      length(container.port_mappings) > 0 ? {
        portMappings = [
          for port_mapping in container.port_mappings : merge(
            {
              containerPort = port_mapping.container_port
              protocol      = port_mapping.protocol != "" ? port_mapping.protocol : "tcp"
            },
            port_mapping.name != "" ? { name = port_mapping.name } : {},
            port_mapping.app_protocol != "" ? { appProtocol = port_mapping.app_protocol } : {},
          )
        ]
      } : {},
      length(container.entry_point) > 0 ? { entryPoint = container.entry_point } : {},
      length(container.command) > 0 ? { command = container.command } : {},
      container.working_directory != "" ? { workingDirectory = container.working_directory } : {},
      length(container.environment) > 0 ? {
        environment = [
          for env_name in sort(keys(container.environment)) : {
            name  = env_name
            value = container.environment[env_name]
          }
        ]
      } : {},
      # Secrets are name -> ARN pairs -- the author's own, plus the pinned
      # ARNs of the secret_environment values secrets.tf stored; the agent
      # resolves them at task start via the execution role, so no secret
      # material passes through here.
      length(merge(container.secrets, local.stored_secret_value_from[container.name])) > 0 ? {
        secrets = [
          for secret_name in sort(keys(merge(container.secrets, local.stored_secret_value_from[container.name]))) : {
            name      = secret_name
            valueFrom = merge(container.secrets, local.stored_secret_value_from[container.name])[secret_name]
          }
        ]
      } : {},
      length(container.environment_files) > 0 ? {
        environmentFiles = [
          for s3_arn in container.environment_files : {
            value = s3_arn
            type  = "s3"
          }
        ]
      } : {},
      container.health_check != null ? {
        healthCheck = merge(
          { command = container.health_check.command },
          container.health_check.interval_seconds > 0 ? { interval = container.health_check.interval_seconds } : {},
          container.health_check.timeout_seconds > 0 ? { timeout = container.health_check.timeout_seconds } : {},
          container.health_check.retries > 0 ? { retries = container.health_check.retries } : {},
          container.health_check.start_period_seconds > 0 ? { startPeriod = container.health_check.start_period_seconds } : {},
        )
      } : {},
      length(container.depends_on) > 0 ? {
        dependsOn = [
          for dependency in container.depends_on : {
            containerName = dependency.container_name
            condition     = dependency.condition
          }
        ]
      } : {},
      length(container.mount_points) > 0 ? {
        mountPoints = [
          for mount_point in container.mount_points : {
            sourceVolume  = mount_point.source_volume
            containerPath = mount_point.container_path
            readOnly      = mount_point.read_only
          }
        ]
      } : {},
      # Log configuration precedence: the container's own block wins;
      # otherwise the task-level default wires awslogs into the shared group
      # with the container's name as the stream prefix.
      container.log_configuration != null ? {
        logConfiguration = merge(
          { logDriver = container.log_configuration.log_driver },
          length(container.log_configuration.options) > 0 ? { options = container.log_configuration.options } : {},
          length(container.log_configuration.secret_options) > 0 ? {
            secretOptions = [
              for option_name in sort(keys(container.log_configuration.secret_options)) : {
                name      = option_name
                valueFrom = container.log_configuration.secret_options[option_name]
              }
            ]
          } : {},
        )
        } : (local.log_group_name != "" ? {
          logConfiguration = {
            logDriver = "awslogs"
            options = {
              awslogs-group         = local.log_group_name
              awslogs-region        = var.spec.region
              awslogs-stream-prefix = container.name
            }
          }
      } : {}),
      container.firelens_configuration != null ? {
        firelensConfiguration = merge(
          { type = container.firelens_configuration.type },
          length(container.firelens_configuration.options) > 0 ? { options = container.firelens_configuration.options } : {},
        )
      } : {},
      container.repository_credentials_secret_arn != "" ? {
        repositoryCredentials = {
          credentialsParameter = container.repository_credentials_secret_arn
        }
      } : {},
      container.user != "" ? { user = container.user } : {},
      container.readonly_root_filesystem ? { readonlyRootFilesystem = true } : {},
      container.privileged ? { privileged = true } : {},
      # initProcessEnabled lives under linuxParameters in the ECS API.
      container.init_process_enabled ? {
        linuxParameters = { initProcessEnabled = true }
      } : {},
      container.gpu_count > 0 ? {
        resourceRequirements = [
          {
            type  = "GPU"
            value = tostring(container.gpu_count)
          }
        ]
      } : {},
      length(container.ulimits) > 0 ? {
        ulimits = [
          for ulimit in container.ulimits : {
            name      = ulimit.name
            softLimit = ulimit.soft_limit
            hardLimit = ulimit.hard_limit
          }
        ]
      } : {},
      length(container.docker_labels) > 0 ? { dockerLabels = container.docker_labels } : {},
      container.start_timeout_seconds > 0 ? { startTimeout = container.start_timeout_seconds } : {},
      container.stop_timeout_seconds > 0 ? { stopTimeout = container.stop_timeout_seconds } : {},
      container.restart_policy != null ? (container.restart_policy.enabled ? {
        restartPolicy = merge(
          { enabled = true },
          length(container.restart_policy.ignored_exit_codes) > 0 ? { ignoredExitCodes = container.restart_policy.ignored_exit_codes } : {},
          container.restart_policy.restart_attempt_period_seconds > 0 ? { restartAttemptPeriod = container.restart_policy.restart_attempt_period_seconds } : {},
        )
      } : {}) : {},
    )
  ]
}
