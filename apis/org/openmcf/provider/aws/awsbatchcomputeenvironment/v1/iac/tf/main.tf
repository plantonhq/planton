# --- Scheduling Policy (optional) ---

resource "aws_batch_scheduling_policy" "this" {
  count = var.spec.scheduling_policy != null ? 1 : 0
  name  = "${var.metadata.id}-scheduling-policy"
  tags  = local.tags

  fair_share_policy {
    compute_reservation = try(var.spec.scheduling_policy.compute_reservation, null)
    share_decay_seconds = try(var.spec.scheduling_policy.share_decay_seconds, null)

    dynamic "share_distribution" {
      for_each = try(var.spec.scheduling_policy.share_distributions, [])
      content {
        share_identifier = share_distribution.value.share_identifier
        weight_factor    = share_distribution.value.weight_factor
      }
    }
  }
}

# --- Compute Environment ---

resource "aws_batch_compute_environment" "this" {
  compute_environment_name = var.metadata.id
  type                     = "MANAGED"
  state                    = var.spec.state
  service_role             = try(var.spec.service_role.value, null)
  tags                     = local.tags

  compute_resources {
    type      = var.spec.compute_resources.type
    max_vcpus = var.spec.compute_resources.max_vcpus
    min_vcpus = local.is_ec2 ? var.spec.compute_resources.min_vcpus : null
    subnets   = [for s in var.spec.compute_resources.subnet_ids : s.value]

    security_group_ids  = [for sg in var.spec.compute_resources.security_group_ids : sg.value]
    instance_type       = local.is_ec2 ? var.spec.compute_resources.instance_types : null
    allocation_strategy = local.is_ec2 && var.spec.compute_resources.allocation_strategy != "" ? var.spec.compute_resources.allocation_strategy : null
    instance_role       = local.is_ec2 ? try(var.spec.compute_resources.instance_role.value, null) : null
    ec2_key_pair        = local.is_ec2 && var.spec.compute_resources.ec2_key_pair != "" ? var.spec.compute_resources.ec2_key_pair : null
    bid_percentage      = local.is_spot ? var.spec.compute_resources.bid_percentage : null
    spot_iam_fleet_role = local.is_spot ? try(var.spec.compute_resources.spot_iam_fleet_role.value, null) : null

    desired_vcpus = local.is_ec2 && var.spec.compute_resources.desired_vcpus > 0 ? var.spec.compute_resources.desired_vcpus : null

    dynamic "launch_template" {
      for_each = var.spec.compute_resources.launch_template != null ? [var.spec.compute_resources.launch_template] : []
      content {
        launch_template_id   = launch_template.value.launch_template_id != "" ? launch_template.value.launch_template_id : null
        launch_template_name = launch_template.value.launch_template_name != "" ? launch_template.value.launch_template_name : null
        version              = launch_template.value.version != "" ? launch_template.value.version : null
      }
    }

    dynamic "ec2_configuration" {
      for_each = var.spec.compute_resources.ec2_configurations
      content {
        image_type        = ec2_configuration.value.image_type != "" ? ec2_configuration.value.image_type : null
        image_id_override = ec2_configuration.value.image_id_override != "" ? ec2_configuration.value.image_id_override : null
      }
    }

    tags = try(var.spec.compute_resources.resource_tags, {})
  }

  dynamic "update_policy" {
    for_each = var.spec.update_policy != null ? [var.spec.update_policy] : []
    content {
      terminate_jobs_on_update      = update_policy.value.terminate_jobs_on_update
      job_execution_timeout_minutes = update_policy.value.job_execution_timeout_minutes
    }
  }
}

# --- Job Queues ---

resource "aws_batch_job_queue" "this" {
  for_each = { for idx, jq in var.spec.job_queues : jq.name => jq }

  name                 = each.value.name
  state                = each.value.state
  priority             = each.value.priority
  scheduling_policy_arn = var.spec.scheduling_policy != null ? aws_batch_scheduling_policy.this[0].arn : null
  tags                 = local.tags

  compute_environment_order {
    compute_environment = aws_batch_compute_environment.this.arn
    order               = 1
  }

  dynamic "job_state_time_limit_action" {
    for_each = each.value.job_state_time_limit_actions
    content {
      action           = job_state_time_limit_action.value.action
      max_time_seconds = job_state_time_limit_action.value.max_time_seconds
      reason           = job_state_time_limit_action.value.reason
      state            = job_state_time_limit_action.value.state
    }
  }
}

# --- Outputs ---

output "compute_environment_arn" {
  description = "ARN of the compute environment"
  value       = aws_batch_compute_environment.this.arn
}

output "compute_environment_name" {
  description = "Name of the compute environment"
  value       = aws_batch_compute_environment.this.compute_environment_name
}

output "ecs_cluster_arn" {
  description = "ARN of the underlying ECS cluster"
  value       = aws_batch_compute_environment.this.ecs_cluster_arn
}

output "status" {
  description = "Status of the compute environment"
  value       = aws_batch_compute_environment.this.status
}

output "job_queue_arns" {
  description = "Map of queue name to ARN"
  value       = { for name, jq in aws_batch_job_queue.this : name => jq.arn }
}

output "scheduling_policy_arn" {
  description = "ARN of the scheduling policy (if created)"
  value       = var.spec.scheduling_policy != null ? aws_batch_scheduling_policy.this[0].arn : ""
}
