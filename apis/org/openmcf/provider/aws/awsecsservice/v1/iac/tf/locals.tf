locals {
  # Basic naming and tags
  resource_name = coalesce(try(var.metadata.name, null), "aws-ecs-service")
  tags          = merge({
    "Name" = local.resource_name
  }, try(var.metadata.labels, {}))

  # Foreign-key fields are already flattened to primitive strings by the tofu
  # generator (the orchestrator resolves any value_from before the module runs),
  # so refs are consumed directly. Optional scalars arrive as their zero value
  # (the generated optional(...) default), so non-zero defaults are applied here.

  # Cluster ARN (required) and the cluster name extracted from it.
  safe_cluster_arn = var.spec.cluster_arn
  cluster_name     = length(split("/", local.safe_cluster_arn)) > 1 ? element(split("/", local.safe_cluster_arn), 1) : ""

  # Container settings. port 0 means "no port" (background worker); normalize to
  # null so the != null guards below behave. replicas defaults to 1 task.
  container_port = try(var.spec.container.port, 0) > 0 ? var.spec.container.port : null
  desired_count  = try(var.spec.container.replicas, 0) > 0 ? var.spec.container.replicas : 1
  cpu            = var.spec.container.cpu
  memory         = var.spec.container.memory

  # Container image construction
  image_repo      = try(var.spec.container.image.repo, null)
  image_tag       = try(var.spec.container.image.tag, null)
  container_image = local.image_repo != null && local.image_repo != "" ? (
    local.image_tag != null && local.image_tag != "" ? "${local.image_repo}:${local.image_tag}" : local.image_repo
  ) : null

  # Logging defaults on (the optional logging object is null when omitted).
  logging_enabled = try(var.spec.container.logging.enabled, true)

  # Networking (subnets required; security groups optional).
  safe_subnet_ids         = try(var.spec.network.subnets, [])
  safe_security_group_ids = try(var.spec.network.security_groups, [])

  # IAM roles: the iam object is optional and each arn defaults to "". An empty
  # arn means "let ECS use the default role", so normalize "" to null.
  task_execution_role_arn_raw  = try(var.spec.iam.task_execution_role_arn, "")
  safe_task_execution_role_arn = local.task_execution_role_arn_raw != "" ? local.task_execution_role_arn_raw : null
  task_role_arn_raw            = try(var.spec.iam.task_role_arn, "")
  safe_task_role_arn           = local.task_role_arn_raw != "" ? local.task_role_arn_raw : null

  # ALB configuration (the alb object is optional; arn flattens to a string).
  alb_enabled           = try(var.spec.alb.enabled, false)
  safe_alb_arn          = try(var.spec.alb.arn, "") != "" ? var.spec.alb.arn : null
  alb_listener_port     = try(var.spec.alb.listener_port, 80)
  alb_listener_priority = try(var.spec.alb.listener_priority, 0) > 0 ? var.spec.alb.listener_priority : 100
  alb_routing_type      = lower(try(var.spec.alb.routing_type, ""))
  alb_path              = try(var.spec.alb.path, "") != "" ? var.spec.alb.path : null
  alb_hostname          = try(var.spec.alb.hostname, "") != "" ? var.spec.alb.hostname : null

  # Boolean flags for conditional logic
  has_container_port = local.container_port != null
  has_alb_config     = local.alb_enabled && local.safe_alb_arn != null

  # Health check grace period (only applied when ALB is enabled).
  health_check_grace_period_seconds = try(var.spec.health_check_grace_period_seconds, 0) > 0 ? var.spec.health_check_grace_period_seconds : 60

  # Auto scaling configuration (the autoscaling object is optional). The target
  # percents are null when unset so their policies are not created.
  autoscaling_enabled               = try(var.spec.autoscaling.enabled, false)
  autoscaling_min_tasks             = try(var.spec.autoscaling.min_tasks, 0) > 0 ? var.spec.autoscaling.min_tasks : 1
  autoscaling_max_tasks             = try(var.spec.autoscaling.max_tasks, 0) > 0 ? var.spec.autoscaling.max_tasks : 10
  autoscaling_target_cpu_percent    = try(var.spec.autoscaling.target_cpu_percent, 0) > 0 ? var.spec.autoscaling.target_cpu_percent : null
  autoscaling_target_memory_percent = try(var.spec.autoscaling.target_memory_percent, 0) > 0 ? var.spec.autoscaling.target_memory_percent : null
}
