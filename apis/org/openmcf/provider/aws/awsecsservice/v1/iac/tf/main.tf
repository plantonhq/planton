# AWS ECS Service Terraform Module
# Auto-release test: Single Terraform module change triggers v{semver}-terraform-awsecsservice-{YYYYMMDD}.{N}

# CloudWatch Log Group for ECS service logging
resource "aws_cloudwatch_log_group" "this" {
  count             = local.logging_enabled ? 1 : 0
  name              = "/ecs/${local.resource_name}"
  retention_in_days = 30
  tags              = local.tags
}

# Data source to get ALB information for outputs
data "aws_lb" "selected" {
  count = local.has_alb_config ? 1 : 0
  arn   = local.safe_alb_arn
}

# Resolve the VPC from the first service subnet. The ALB target group needs a VPC id
# but the spec only carries subnet ids, so we derive it -- matching the Pulumi module's
# ec2.LookupSubnet(firstSubnet).VpcId.
data "aws_subnet" "first" {
  count = local.alb_enabled && local.container_port != null ? 1 : 0
  id    = local.safe_subnet_ids[0]
}

# Resolve the real ALB listener ARN by load balancer + port. A listener ARN cannot be
# synthesized from the load balancer ARN, so we look it up -- matching the Pulumi
# module's lb.LookupListener(LoadBalancerArn, Port). The ALB module creates the listener
# on this port (443 when SSL is on, 80 otherwise) before this service's node runs.
data "aws_lb_listener" "selected" {
  count             = local.alb_enabled && local.container_port != null ? 1 : 0
  load_balancer_arn = local.safe_alb_arn
  port              = local.alb_listener_port
}

# Current AWS region data source
data "aws_region" "current" {}

# ECS Task Definition
resource "aws_ecs_task_definition" "this" {
  family                   = local.resource_name
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = tostring(local.cpu)
  memory                   = tostring(local.memory)

  container_definitions = jsonencode([
    {
      name      = local.resource_name
      image     = local.container_image
      essential = true
      
      # Port mappings (only if container port is specified)
      portMappings = local.container_port != null ? [{
        containerPort = local.container_port
        protocol      = "tcp"
      }] : []
      
      # Logging configuration
      logConfiguration = local.logging_enabled ? {
        logDriver = "awslogs"
        options = {
          awslogs-group         = aws_cloudwatch_log_group.this[0].name
          awslogs-region        = data.aws_region.current.name
          awslogs-stream-prefix = "ecs"
        }
      } : null
      
      # Environment variables
      environment = [
        for k, v in try(var.spec.container.env.variables, {}) : {
          name  = k
          value = v
        }
      ]
      
      # Secrets (if any)
      secrets = [
        for k, v in try(var.spec.container.env.secrets, {}) : {
          name      = k
          valueFrom = v
        }
      ]
      
      # Environment files from S3 (if any)
      environmentFiles = [
        for s3_uri in try(var.spec.container.env.s3_files, []) : {
          value = s3_uri
          type  = "s3"
        }
      ]
    }
  ])

  # IAM roles
  execution_role_arn = local.safe_task_execution_role_arn
  task_role_arn      = local.safe_task_role_arn
  
  tags = local.tags
}

# ECS Service
resource "aws_ecs_service" "this" {
  name            = local.resource_name
  cluster         = local.safe_cluster_arn
  task_definition = aws_ecs_task_definition.this.arn
  desired_count   = local.desired_count
  launch_type     = "FARGATE"

  # Network configuration
  network_configuration {
    subnets          = local.safe_subnet_ids
    security_groups  = local.safe_security_group_ids
    assign_public_ip = false
  }

  # Load balancer configuration (only if ALB is enabled and container port is specified)
  dynamic "load_balancer" {
    for_each = local.alb_enabled && local.container_port != null ? [1] : []
    content {
      target_group_arn = aws_lb_target_group.this[0].arn
      container_name   = local.resource_name
      container_port   = local.container_port
    }
  }

  # Health check grace period (only when ALB is enabled)
  health_check_grace_period_seconds = local.alb_enabled && local.container_port != null ? local.health_check_grace_period_seconds : null

  # desired_count is runtime state: when autoscaling is enabled the appautoscaling
  # target owns it, and operators may scale out of band. Terraform sets the initial
  # count from spec.container.replicas and then leaves it alone. ignore_changes must
  # be a static literal list (OpenTofu forbids a conditional expression here), so this
  # always ignores -- matching the house convention for autoscaling-managed counts
  # (e.g. gcp-gke-node-pool's node_count).
  lifecycle {
    ignore_changes = [desired_count]
  }

  depends_on = [aws_ecs_task_definition.this]
  tags       = local.tags
}

# ALB Target Group (only if ALB is enabled and container port is specified).
# Fargate tasks use awsvpc networking, so they register with the load balancer by IP
# (target_type = "ip"), which also requires the VPC id -- derived from the service's
# first subnet above. Matches the Pulumi module (TargetType "ip", VpcId from the subnet).
resource "aws_lb_target_group" "this" {
  count       = local.alb_enabled && local.container_port != null ? 1 : 0
  name        = substr("tg-${local.resource_name}", 0, 32)
  port        = local.container_port
  protocol    = "HTTP"
  target_type = "ip"
  vpc_id      = data.aws_subnet.first[0].vpc_id

  # Health check configuration
  health_check {
    enabled             = true
    path                = coalesce(try(var.spec.alb.health_check.path, null), "/")
    interval            = coalesce(try(var.spec.alb.health_check.interval, null), 30)
    timeout             = coalesce(try(var.spec.alb.health_check.timeout, null), 5)
    healthy_threshold   = coalesce(try(var.spec.alb.health_check.healthy_threshold, null), 5)
    unhealthy_threshold = coalesce(try(var.spec.alb.health_check.unhealthy_threshold, null), 2)
    protocol            = coalesce(try(var.spec.alb.health_check.protocol, null), "HTTP")
    port                = coalesce(try(var.spec.alb.health_check.port, null), "traffic-port")
  }

  tags = local.tags
}

# ALB Listener Rule (only if ALB is enabled and container port is specified)
resource "aws_lb_listener_rule" "this" {
  count        = local.alb_enabled && local.container_port != null ? 1 : 0
  listener_arn = data.aws_lb_listener.selected[0].arn
  priority     = local.alb_listener_priority

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.this[0].arn
  }

  # Path-based routing condition
  dynamic "condition" {
    for_each = local.alb_routing_type == "path" && local.alb_path != null ? [1] : []
    content {
      path_pattern {
        values = [local.alb_path]
      }
    }
  }

  # Hostname-based routing condition
  dynamic "condition" {
    for_each = local.alb_routing_type == "hostname" && local.alb_hostname != null ? [1] : []
    content {
      host_header {
        values = [local.alb_hostname]
      }
    }
  }

  depends_on = [aws_lb_target_group.this]
}

# Auto Scaling Target (only if autoscaling is enabled)
resource "aws_appautoscaling_target" "ecs_target" {
  count              = local.autoscaling_enabled ? 1 : 0
  max_capacity       = local.autoscaling_max_tasks
  min_capacity       = local.autoscaling_min_tasks
  resource_id        = "service/${local.cluster_name}/${local.resource_name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"

  depends_on = [aws_ecs_service.this]
}

# Auto Scaling Policy - CPU-based (only if autoscaling is enabled and target_cpu_percent is set)
resource "aws_appautoscaling_policy" "ecs_cpu_policy" {
  count              = local.autoscaling_enabled && local.autoscaling_target_cpu_percent != null ? 1 : 0
  name               = "${local.resource_name}-cpu-scaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs_target[0].resource_id
  scalable_dimension = aws_appautoscaling_target.ecs_target[0].scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs_target[0].service_namespace

  target_tracking_scaling_policy_configuration {
    target_value       = local.autoscaling_target_cpu_percent
    scale_in_cooldown  = 300
    scale_out_cooldown = 60

    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
  }
}

# Auto Scaling Policy - Memory-based (only if autoscaling is enabled and target_memory_percent is set)
resource "aws_appautoscaling_policy" "ecs_memory_policy" {
  count              = local.autoscaling_enabled && local.autoscaling_target_memory_percent != null ? 1 : 0
  name               = "${local.resource_name}-memory-scaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs_target[0].resource_id
  scalable_dimension = aws_appautoscaling_target.ecs_target[0].scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs_target[0].service_namespace

  target_tracking_scaling_policy_configuration {
    target_value       = local.autoscaling_target_memory_percent
    scale_in_cooldown  = 300
    scale_out_cooldown = 60

    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageMemoryUtilization"
    }
  }
}

