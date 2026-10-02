# The auto-scaling group is a pure orchestrator: WHAT to launch lives in
# the referenced launch template, WHERE traffic comes from lives in the
# referenced target groups; this resource owns how many, where, and when.
# Only the group name is create-only in AWS -- everything else updates in
# place.
resource "aws_autoscaling_group" "this" {
  name = local.group_name

  vpc_zone_identifier = var.spec.subnets
  min_size            = var.spec.min_size
  max_size            = var.spec.max_size

  # Leaving desired_capacity unset lets scaling policies own the number: a
  # literal count here would fight the autoscaler on every apply.
  desired_capacity      = var.spec.desired_capacity > 0 ? var.spec.desired_capacity : null
  desired_capacity_type = var.spec.desired_capacity_type != "" ? var.spec.desired_capacity_type : null

  # Exactly one of launch_template / mixed_instances_policy is set (spec
  # validation enforces it) -- mirroring AWS's ExactlyOneOf on these fields.
  dynamic "launch_template" {
    for_each = var.spec.launch_template != null ? [var.spec.launch_template] : []
    content {
      id = launch_template.value.launch_template_id
      # An empty version keeps AWS's "$Default" behavior -- the setup that
      # lets a template update roll the fleet.
      version = launch_template.value.version != "" ? launch_template.value.version : null
    }
  }

  dynamic "mixed_instances_policy" {
    for_each = var.spec.mixed_instances_policy != null ? [var.spec.mixed_instances_policy] : []
    content {
      launch_template {
        launch_template_specification {
          launch_template_id = mixed_instances_policy.value.launch_template.launch_template_id
          version            = mixed_instances_policy.value.launch_template.version != "" ? mixed_instances_policy.value.launch_template.version : null
        }

        dynamic "override" {
          for_each = mixed_instances_policy.value.overrides
          content {
            instance_type = override.value.instance_type != "" ? override.value.instance_type : null
            # weighted_capacity is a string at AWS despite being numeric
            # (1-999); the spec keeps the honest int and converts here.
            weighted_capacity = override.value.weighted_capacity > 0 ? tostring(override.value.weighted_capacity) : null

            dynamic "launch_template_specification" {
              for_each = override.value.launch_template != null ? [override.value.launch_template] : []
              content {
                launch_template_id = launch_template_specification.value.launch_template_id
                version            = launch_template_specification.value.version != "" ? launch_template_specification.value.version : null
              }
            }

            # Attribute-based selection: memory_mib and vcpu_count are the
            # two AWS-required dimensions (spec-enforced); every other
            # field narrows the candidate set and is sent only when set so
            # AWS's own defaults keep applying.
            dynamic "instance_requirements" {
              for_each = override.value.instance_requirements != null ? [override.value.instance_requirements] : []
              content {
                memory_mib {
                  min = instance_requirements.value.memory_mib.min
                  max = instance_requirements.value.memory_mib.max > 0 ? instance_requirements.value.memory_mib.max : null
                }
                vcpu_count {
                  min = instance_requirements.value.vcpu_count.min
                  max = instance_requirements.value.vcpu_count.max > 0 ? instance_requirements.value.vcpu_count.max : null
                }

                allowed_instance_types  = length(instance_requirements.value.allowed_instance_types) > 0 ? instance_requirements.value.allowed_instance_types : null
                excluded_instance_types = length(instance_requirements.value.excluded_instance_types) > 0 ? instance_requirements.value.excluded_instance_types : null
                instance_generations    = length(instance_requirements.value.instance_generations) > 0 ? instance_requirements.value.instance_generations : null
                cpu_manufacturers       = length(instance_requirements.value.cpu_manufacturers) > 0 ? instance_requirements.value.cpu_manufacturers : null

                bare_metal                = instance_requirements.value.bare_metal != "" ? instance_requirements.value.bare_metal : null
                burstable_performance     = instance_requirements.value.burstable_performance != "" ? instance_requirements.value.burstable_performance : null
                require_hibernate_support = instance_requirements.value.require_hibernate_support ? true : null

                spot_max_price_percentage_over_lowest_price             = instance_requirements.value.spot_max_price_percentage_over_lowest_price > 0 ? instance_requirements.value.spot_max_price_percentage_over_lowest_price : null
                max_spot_price_as_percentage_of_optimal_on_demand_price = instance_requirements.value.max_spot_price_as_percentage_of_optimal_on_demand_price > 0 ? instance_requirements.value.max_spot_price_as_percentage_of_optimal_on_demand_price : null
                on_demand_max_price_percentage_over_lowest_price        = instance_requirements.value.on_demand_max_price_percentage_over_lowest_price > 0 ? instance_requirements.value.on_demand_max_price_percentage_over_lowest_price : null

                local_storage       = instance_requirements.value.local_storage != "" ? instance_requirements.value.local_storage : null
                local_storage_types = length(instance_requirements.value.local_storage_types) > 0 ? instance_requirements.value.local_storage_types : null

                dynamic "total_local_storage_gb" {
                  for_each = instance_requirements.value.total_local_storage_gb != null ? [instance_requirements.value.total_local_storage_gb] : []
                  content {
                    min = total_local_storage_gb.value.min > 0 ? total_local_storage_gb.value.min : null
                    max = total_local_storage_gb.value.max > 0 ? total_local_storage_gb.value.max : null
                  }
                }
                dynamic "memory_gib_per_vcpu" {
                  for_each = instance_requirements.value.memory_gib_per_vcpu != null ? [instance_requirements.value.memory_gib_per_vcpu] : []
                  content {
                    min = memory_gib_per_vcpu.value.min > 0 ? memory_gib_per_vcpu.value.min : null
                    max = memory_gib_per_vcpu.value.max > 0 ? memory_gib_per_vcpu.value.max : null
                  }
                }
                dynamic "network_interface_count" {
                  for_each = instance_requirements.value.network_interface_count != null ? [instance_requirements.value.network_interface_count] : []
                  content {
                    min = network_interface_count.value.min > 0 ? network_interface_count.value.min : null
                    max = network_interface_count.value.max > 0 ? network_interface_count.value.max : null
                  }
                }
                dynamic "network_bandwidth_gbps" {
                  for_each = instance_requirements.value.network_bandwidth_gbps != null ? [instance_requirements.value.network_bandwidth_gbps] : []
                  content {
                    min = network_bandwidth_gbps.value.min > 0 ? network_bandwidth_gbps.value.min : null
                    max = network_bandwidth_gbps.value.max > 0 ? network_bandwidth_gbps.value.max : null
                  }
                }
                dynamic "baseline_ebs_bandwidth_mbps" {
                  for_each = instance_requirements.value.baseline_ebs_bandwidth_mbps != null ? [instance_requirements.value.baseline_ebs_bandwidth_mbps] : []
                  content {
                    min = baseline_ebs_bandwidth_mbps.value.min > 0 ? baseline_ebs_bandwidth_mbps.value.min : null
                    max = baseline_ebs_bandwidth_mbps.value.max > 0 ? baseline_ebs_bandwidth_mbps.value.max : null
                  }
                }
                dynamic "accelerator_count" {
                  for_each = instance_requirements.value.accelerator_count != null ? [instance_requirements.value.accelerator_count] : []
                  content {
                    min = accelerator_count.value.min > 0 ? accelerator_count.value.min : null
                    max = accelerator_count.value.max > 0 ? accelerator_count.value.max : null
                  }
                }
                accelerator_manufacturers = length(instance_requirements.value.accelerator_manufacturers) > 0 ? instance_requirements.value.accelerator_manufacturers : null
                accelerator_names         = length(instance_requirements.value.accelerator_names) > 0 ? instance_requirements.value.accelerator_names : null
                accelerator_types         = length(instance_requirements.value.accelerator_types) > 0 ? instance_requirements.value.accelerator_types : null
                dynamic "accelerator_total_memory_mib" {
                  for_each = instance_requirements.value.accelerator_total_memory_mib != null ? [instance_requirements.value.accelerator_total_memory_mib] : []
                  content {
                    min = accelerator_total_memory_mib.value.min > 0 ? accelerator_total_memory_mib.value.min : null
                    max = accelerator_total_memory_mib.value.max > 0 ? accelerator_total_memory_mib.value.max : null
                  }
                }
              }
            }
          }
        }
      }

      dynamic "instances_distribution" {
        for_each = mixed_instances_policy.value.instances_distribution != null ? [mixed_instances_policy.value.instances_distribution] : []
        content {
          on_demand_allocation_strategy = instances_distribution.value.on_demand_allocation_strategy != "" ? instances_distribution.value.on_demand_allocation_strategy : null
          on_demand_base_capacity       = instances_distribution.value.on_demand_base_capacity > 0 ? instances_distribution.value.on_demand_base_capacity : null
          # Explicit 0 means all-Spot above the base -- the aggressive cost
          # posture -- so presence (null vs a value) decides what is sent.
          on_demand_percentage_above_base_capacity = instances_distribution.value.on_demand_percentage_above_base_capacity
          spot_allocation_strategy                 = instances_distribution.value.spot_allocation_strategy != "" ? instances_distribution.value.spot_allocation_strategy : null
          spot_instance_pools                      = instances_distribution.value.spot_instance_pools > 0 ? instances_distribution.value.spot_instance_pools : null
          spot_max_price                           = instances_distribution.value.spot_max_price != "" ? instances_distribution.value.spot_max_price : null
        }
      }
    }
  }

  capacity_rebalance = var.spec.capacity_rebalance ? true : null
  default_cooldown   = var.spec.default_cooldown_seconds > 0 ? var.spec.default_cooldown_seconds : null
  # default_instance_warmup: 0 is meaningful to AWS ("metrics count
  # immediately") but indistinguishable from unset in the spec; both
  # engines send it only when positive.
  default_instance_warmup = var.spec.default_instance_warmup_seconds > 0 ? var.spec.default_instance_warmup_seconds : null

  health_check_type         = var.spec.health_check_type != "" ? var.spec.health_check_type : null
  health_check_grace_period = var.spec.health_check_grace_period_seconds > 0 ? var.spec.health_check_grace_period_seconds : null

  target_group_arns = length(var.spec.target_groups) > 0 ? var.spec.target_groups : null

  termination_policies  = length(var.spec.termination_policies) > 0 ? var.spec.termination_policies : null
  max_instance_lifetime = var.spec.max_instance_lifetime_seconds > 0 ? var.spec.max_instance_lifetime_seconds : null
  protect_from_scale_in = var.spec.protect_from_scale_in ? true : null

  placement_group         = var.spec.placement_group != "" ? var.spec.placement_group : null
  service_linked_role_arn = var.spec.service_linked_role_arn != "" ? var.spec.service_linked_role_arn : null

  enabled_metrics     = length(var.spec.enabled_metrics) > 0 ? var.spec.enabled_metrics : null
  suspended_processes = length(var.spec.suspended_processes) > 0 ? var.spec.suspended_processes : null

  force_delete              = var.spec.force_delete ? true : null
  force_delete_warm_pool    = var.spec.force_delete_warm_pool ? true : null
  wait_for_capacity_timeout = var.spec.wait_for_capacity_timeout != "" ? var.spec.wait_for_capacity_timeout : null

  # ELB-health waits (engine behavior, like wait_for_capacity_timeout):
  # min_elb_capacity gates the CREATE; wait_for_elb_capacity gates create
  # AND every update, and takes precedence when both are set.
  min_elb_capacity      = var.spec.min_elb_capacity > 0 ? var.spec.min_elb_capacity : null
  wait_for_elb_capacity = var.spec.wait_for_elb_capacity > 0 ? var.spec.wait_for_elb_capacity : null

  # Keep IaC applies moving while a scaling activity is failing -- for
  # groups whose scaling errors are watched by their own alarms.
  ignore_failed_scaling_activities = var.spec.ignore_failed_scaling_activities ? true : null

  # Launch into EC2 Capacity Reservations: a preference shapes reservation
  # use; targets pin the group to specific reservations or a resource
  # group of them.
  dynamic "capacity_reservation_specification" {
    for_each = var.spec.capacity_reservation != null ? [var.spec.capacity_reservation] : []
    content {
      capacity_reservation_preference = capacity_reservation_specification.value.preference != "" ? capacity_reservation_specification.value.preference : null
      dynamic "capacity_reservation_target" {
        for_each = (length(capacity_reservation_specification.value.capacity_reservation_ids) > 0 || length(capacity_reservation_specification.value.capacity_reservation_resource_group_arns) > 0) ? [capacity_reservation_specification.value] : []
        content {
          capacity_reservation_ids                 = length(capacity_reservation_target.value.capacity_reservation_ids) > 0 ? capacity_reservation_target.value.capacity_reservation_ids : null
          capacity_reservation_resource_group_arns = length(capacity_reservation_target.value.capacity_reservation_resource_group_arns) > 0 ? capacity_reservation_target.value.capacity_reservation_resource_group_arns : null
        }
      }
    }
  }

  # Generalized traffic sources (VPC Lattice / Classic ELB); ALB/NLB
  # target groups use target_group_arns above -- the spec forbids mixing
  # the two, mirroring the provider's ConflictsWith.
  dynamic "traffic_source" {
    for_each = var.spec.traffic_sources
    content {
      identifier = traffic_source.value.identifier
      type       = traffic_source.value.type != "" ? traffic_source.value.type : null
    }
  }

  # The fate of instances whose TERMINATING hook ends in ABANDON:
  # "retain" keeps them running (out of the group) for post-mortems.
  dynamic "instance_lifecycle_policy" {
    for_each = var.spec.instance_lifecycle_policy != null ? [var.spec.instance_lifecycle_policy] : []
    content {
      dynamic "retention_triggers" {
        for_each = instance_lifecycle_policy.value.terminate_hook_abandon != "" ? [instance_lifecycle_policy.value.terminate_hook_abandon] : []
        content {
          terminate_hook_abandon = retention_triggers.value
        }
      }
    }
  }

  # Hooks flagged apply_at_launch attach atomically at group creation, so
  # even the very first instance is caught; AWS makes these creation-time
  # hooks immutable (changing one replaces the group). The rest are
  # standalone hook resources below, individually updatable.
  dynamic "initial_lifecycle_hook" {
    for_each = { for hook in var.spec.lifecycle_hooks : hook.name => hook if hook.apply_at_launch }
    content {
      name                    = initial_lifecycle_hook.value.name
      lifecycle_transition    = initial_lifecycle_hook.value.lifecycle_transition
      default_result          = initial_lifecycle_hook.value.default_result != "" ? initial_lifecycle_hook.value.default_result : null
      heartbeat_timeout       = initial_lifecycle_hook.value.heartbeat_timeout_seconds > 0 ? initial_lifecycle_hook.value.heartbeat_timeout_seconds : null
      notification_target_arn = initial_lifecycle_hook.value.notification_target_arn != "" ? initial_lifecycle_hook.value.notification_target_arn : null
      role_arn                = initial_lifecycle_hook.value.role_arn != "" ? initial_lifecycle_hook.value.role_arn : null
      notification_metadata   = initial_lifecycle_hook.value.notification_metadata != "" ? initial_lifecycle_hook.value.notification_metadata : null
    }
  }

  # Rolling replacement when the launch template (or another watched
  # attribute) changes -- the mechanism that turns a template update into a
  # zero-downtime fleet rollout.
  dynamic "instance_refresh" {
    for_each = var.spec.instance_refresh != null ? [var.spec.instance_refresh] : []
    content {
      strategy = instance_refresh.value.strategy
      triggers = length(instance_refresh.value.triggers) > 0 ? instance_refresh.value.triggers : null

      dynamic "preferences" {
        for_each = instance_refresh.value.preferences != null ? [instance_refresh.value.preferences] : []
        content {
          # Explicit 0 ("replace the whole fleet at once") is meaningful and
          # distinct from unset (AWS default 90): the contract delivers null
          # when unset, so the value passes straight through.
          min_healthy_percentage = preferences.value.min_healthy_percentage
          max_healthy_percentage = preferences.value.max_healthy_percentage > 0 ? preferences.value.max_healthy_percentage : null
          # The provider models warmup and checkpoint delay as strings
          # (nullable ints at AWS); the spec keeps honest ints and converts.
          instance_warmup        = preferences.value.instance_warmup_seconds > 0 ? tostring(preferences.value.instance_warmup_seconds) : null
          checkpoint_percentages = length(preferences.value.checkpoint_percentages) > 0 ? preferences.value.checkpoint_percentages : null
          checkpoint_delay       = preferences.value.checkpoint_delay_seconds > 0 ? tostring(preferences.value.checkpoint_delay_seconds) : null
          auto_rollback          = preferences.value.auto_rollback ? true : null

          dynamic "alarm_specification" {
            for_each = length(preferences.value.alarms) > 0 ? [preferences.value.alarms] : []
            content {
              alarms = alarm_specification.value
            }
          }

          scale_in_protected_instances = preferences.value.scale_in_protected_instances != "" ? preferences.value.scale_in_protected_instances : null
          standby_instances            = preferences.value.standby_instances != "" ? preferences.value.standby_instances : null
          skip_matching                = preferences.value.skip_matching ? true : null
        }
      }
    }
  }

  # Pre-initialized capacity: scale-out in seconds instead of boot time.
  dynamic "warm_pool" {
    for_each = var.spec.warm_pool != null ? [var.spec.warm_pool] : []
    content {
      pool_state = warm_pool.value.pool_state != "" ? warm_pool.value.pool_state : null
      min_size   = warm_pool.value.min_size > 0 ? warm_pool.value.min_size : null
      # Explicit 0 is meaningful (no prepared capacity beyond min_size), so
      # presence (null vs a value) decides what is sent.
      max_group_prepared_capacity = warm_pool.value.max_group_prepared_capacity

      dynamic "instance_reuse_policy" {
        for_each = warm_pool.value.reuse_on_scale_in ? [true] : []
        content {
          reuse_on_scale_in = true
        }
      }
    }
  }

  # Group-wide health bounds for every replacement operation.
  dynamic "instance_maintenance_policy" {
    for_each = var.spec.instance_maintenance_policy != null ? [var.spec.instance_maintenance_policy] : []
    content {
      min_healthy_percentage = instance_maintenance_policy.value.min_healthy_percentage
      max_healthy_percentage = instance_maintenance_policy.value.max_healthy_percentage
    }
  }

  dynamic "availability_zone_distribution" {
    for_each = var.spec.capacity_distribution_strategy != "" ? [var.spec.capacity_distribution_strategy] : []
    content {
      capacity_distribution_strategy = availability_zone_distribution.value
    }
  }

  # ASG tags are the native key/value/propagate-at-launch triple, not a
  # plain map: propagate_at_launch=true copies the identity tags onto every
  # launched instance, so fleet members never escape cost-allocation and
  # orphan-cleanup queries.
  dynamic "tag" {
    for_each = local.aws_tags
    content {
      key                 = tag.key
      value               = tag.value
      propagate_at_launch = true
    }
  }
}

# ---------------------------------------------------------------------------
# Folded sub-resources. Each is an AWS sub-resource of exactly ONE group
# with no referenceable identity of its own -- which is why they live in
# this spec instead of being standalone kinds. Managing them as separate
# provider resources (keyed by name) makes adding or removing one an
# in-place update that never replaces the group.
# ---------------------------------------------------------------------------

# Scaling policies. policy_type decides which configuration block applies
# (spec validation enforces the discriminated union), mirroring how
# PutScalingPolicy interprets its input.
resource "aws_autoscaling_policy" "this" {
  for_each = { for policy in var.spec.scaling_policies : policy.name => policy }

  autoscaling_group_name = aws_autoscaling_group.this.name
  name                   = each.value.name
  policy_type            = each.value.policy_type

  # The pause button: a disabled policy stays configured (alarms, history,
  # forecast state) but stops acting on the group. AWS defaults to
  # enabled, so only an explicit disable is sent.
  enabled = each.value.disabled ? false : null

  estimated_instance_warmup = each.value.estimated_instance_warmup_seconds > 0 ? each.value.estimated_instance_warmup_seconds : null

  # SimpleScaling / StepScaling shared knobs.
  adjustment_type = (
    each.value.simple_scaling != null ? each.value.simple_scaling.adjustment_type :
    each.value.step_scaling != null ? each.value.step_scaling.adjustment_type : null
  )
  min_adjustment_magnitude = (
    each.value.simple_scaling != null && try(each.value.simple_scaling.min_adjustment_magnitude, 0) > 0 ? each.value.simple_scaling.min_adjustment_magnitude :
    each.value.step_scaling != null && try(each.value.step_scaling.min_adjustment_magnitude, 0) > 0 ? each.value.step_scaling.min_adjustment_magnitude : null
  )

  # SimpleScaling only.
  scaling_adjustment = each.value.simple_scaling != null ? each.value.simple_scaling.scaling_adjustment : null
  cooldown           = each.value.simple_scaling != null && try(each.value.simple_scaling.cooldown_seconds, 0) > 0 ? each.value.simple_scaling.cooldown_seconds : null

  # StepScaling only.
  metric_aggregation_type = each.value.step_scaling != null && try(each.value.step_scaling.metric_aggregation_type, "") != "" ? each.value.step_scaling.metric_aggregation_type : null

  dynamic "step_adjustment" {
    for_each = each.value.step_scaling != null ? each.value.step_scaling.step_adjustments : []
    content {
      scaling_adjustment          = step_adjustment.value.scaling_adjustment
      metric_interval_lower_bound = step_adjustment.value.metric_interval_lower_bound != "" ? step_adjustment.value.metric_interval_lower_bound : null
      metric_interval_upper_bound = step_adjustment.value.metric_interval_upper_bound != "" ? step_adjustment.value.metric_interval_upper_bound : null
    }
  }

  # TargetTrackingScaling: hold a predefined or customized metric (single
  # or metric-math form) at a target value.
  dynamic "target_tracking_configuration" {
    for_each = each.value.target_tracking != null ? [each.value.target_tracking] : []
    content {
      target_value     = target_tracking_configuration.value.target_value
      disable_scale_in = target_tracking_configuration.value.disable_scale_in ? true : null

      dynamic "predefined_metric_specification" {
        for_each = target_tracking_configuration.value.predefined_metric_type != "" ? [target_tracking_configuration.value] : []
        content {
          predefined_metric_type = predefined_metric_specification.value.predefined_metric_type
          resource_label         = predefined_metric_specification.value.resource_label != "" ? predefined_metric_specification.value.resource_label : null
        }
      }

      dynamic "customized_metric_specification" {
        for_each = target_tracking_configuration.value.customized_metric != null ? [target_tracking_configuration.value.customized_metric] : []
        content {
          metric_name = customized_metric_specification.value.metric_name != "" ? customized_metric_specification.value.metric_name : null
          namespace   = customized_metric_specification.value.namespace != "" ? customized_metric_specification.value.namespace : null
          statistic   = customized_metric_specification.value.statistic != "" ? customized_metric_specification.value.statistic : null
          unit        = customized_metric_specification.value.unit != "" ? customized_metric_specification.value.unit : null
          period      = customized_metric_specification.value.period_seconds > 0 ? customized_metric_specification.value.period_seconds : null

          dynamic "metric_dimension" {
            for_each = customized_metric_specification.value.dimensions
            content {
              name  = metric_dimension.value.name
              value = metric_dimension.value.value
            }
          }

          dynamic "metrics" {
            for_each = customized_metric_specification.value.metrics
            content {
              id         = metrics.value.id
              expression = metrics.value.expression != "" ? metrics.value.expression : null
              label      = metrics.value.label != "" ? metrics.value.label : null
              # return_data defaults to true at AWS; only an explicit value
              # is sent, so intermediate entries carry an explicit false.
              return_data = metrics.value.return_data

              dynamic "metric_stat" {
                for_each = metrics.value.metric_stat != null ? [metrics.value.metric_stat] : []
                content {
                  stat   = metric_stat.value.stat
                  unit   = metric_stat.value.unit != "" ? metric_stat.value.unit : null
                  period = metric_stat.value.period_seconds > 0 ? metric_stat.value.period_seconds : null

                  metric {
                    metric_name = metric_stat.value.metric_name
                    namespace   = metric_stat.value.namespace

                    dynamic "dimensions" {
                      for_each = metric_stat.value.dimensions
                      content {
                        name  = dimensions.value.name
                        value = dimensions.value.value
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }

  # PredictiveScaling: forecast-driven pre-provisioning. The metrics come
  # in three forms (spec validation enforces the choice): one predefined
  # PAIR, SPLIT predefined load + scaling metrics, or fully CUSTOMIZED
  # metric-math query sets.
  dynamic "predictive_scaling_configuration" {
    for_each = each.value.predictive_scaling != null ? [each.value.predictive_scaling] : []
    content {
      mode = predictive_scaling_configuration.value.mode != "" ? predictive_scaling_configuration.value.mode : null
      # The provider models buffer time and capacity buffer as strings
      # (nullable ints at AWS); the spec keeps honest ints and converts.
      scheduling_buffer_time       = predictive_scaling_configuration.value.scheduling_buffer_time_seconds > 0 ? tostring(predictive_scaling_configuration.value.scheduling_buffer_time_seconds) : null
      max_capacity_breach_behavior = predictive_scaling_configuration.value.max_capacity_breach_behavior != "" ? predictive_scaling_configuration.value.max_capacity_breach_behavior : null
      max_capacity_buffer          = predictive_scaling_configuration.value.max_capacity_buffer > 0 ? tostring(predictive_scaling_configuration.value.max_capacity_buffer) : null

      metric_specification {
        target_value = predictive_scaling_configuration.value.target_value

        dynamic "predefined_metric_pair_specification" {
          for_each = predictive_scaling_configuration.value.predefined_metric_pair_type != "" ? [predictive_scaling_configuration.value] : []
          content {
            predefined_metric_type = predefined_metric_pair_specification.value.predefined_metric_pair_type
            resource_label         = predefined_metric_pair_specification.value.resource_label != "" ? predefined_metric_pair_specification.value.resource_label : null
          }
        }

        dynamic "predefined_load_metric_specification" {
          for_each = predictive_scaling_configuration.value.predefined_load_metric != null ? [predictive_scaling_configuration.value.predefined_load_metric] : []
          content {
            predefined_metric_type = predefined_load_metric_specification.value.metric_type
            resource_label         = predefined_load_metric_specification.value.resource_label != "" ? predefined_load_metric_specification.value.resource_label : null
          }
        }

        dynamic "predefined_scaling_metric_specification" {
          for_each = predictive_scaling_configuration.value.predefined_scaling_metric != null ? [predictive_scaling_configuration.value.predefined_scaling_metric] : []
          content {
            predefined_metric_type = predefined_scaling_metric_specification.value.metric_type
            resource_label         = predefined_scaling_metric_specification.value.resource_label != "" ? predefined_scaling_metric_specification.value.resource_label : null
          }
        }

        dynamic "customized_load_metric_specification" {
          for_each = length(predictive_scaling_configuration.value.customized_load_metric_queries) > 0 ? [predictive_scaling_configuration.value.customized_load_metric_queries] : []
          content {
            dynamic "metric_data_queries" {
              for_each = customized_load_metric_specification.value
              content {
                id          = metric_data_queries.value.id
                expression  = metric_data_queries.value.expression != "" ? metric_data_queries.value.expression : null
                label       = metric_data_queries.value.label != "" ? metric_data_queries.value.label : null
                return_data = metric_data_queries.value.return_data

                dynamic "metric_stat" {
                  for_each = metric_data_queries.value.metric_stat != null ? [metric_data_queries.value.metric_stat] : []
                  content {
                    stat = metric_stat.value.stat
                    unit = metric_stat.value.unit != "" ? metric_stat.value.unit : null

                    metric {
                      metric_name = metric_stat.value.metric_name
                      namespace   = metric_stat.value.namespace

                      dynamic "dimensions" {
                        for_each = metric_stat.value.dimensions
                        content {
                          name  = dimensions.value.name
                          value = dimensions.value.value
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }

        dynamic "customized_scaling_metric_specification" {
          for_each = length(predictive_scaling_configuration.value.customized_scaling_metric_queries) > 0 ? [predictive_scaling_configuration.value.customized_scaling_metric_queries] : []
          content {
            dynamic "metric_data_queries" {
              for_each = customized_scaling_metric_specification.value
              content {
                id          = metric_data_queries.value.id
                expression  = metric_data_queries.value.expression != "" ? metric_data_queries.value.expression : null
                label       = metric_data_queries.value.label != "" ? metric_data_queries.value.label : null
                return_data = metric_data_queries.value.return_data

                dynamic "metric_stat" {
                  for_each = metric_data_queries.value.metric_stat != null ? [metric_data_queries.value.metric_stat] : []
                  content {
                    stat = metric_stat.value.stat
                    unit = metric_stat.value.unit != "" ? metric_stat.value.unit : null

                    metric {
                      metric_name = metric_stat.value.metric_name
                      namespace   = metric_stat.value.namespace

                      dynamic "dimensions" {
                        for_each = metric_stat.value.dimensions
                        content {
                          name  = dimensions.value.name
                          value = dimensions.value.value
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }

        dynamic "customized_capacity_metric_specification" {
          for_each = length(predictive_scaling_configuration.value.customized_capacity_metric_queries) > 0 ? [predictive_scaling_configuration.value.customized_capacity_metric_queries] : []
          content {
            dynamic "metric_data_queries" {
              for_each = customized_capacity_metric_specification.value
              content {
                id          = metric_data_queries.value.id
                expression  = metric_data_queries.value.expression != "" ? metric_data_queries.value.expression : null
                label       = metric_data_queries.value.label != "" ? metric_data_queries.value.label : null
                return_data = metric_data_queries.value.return_data

                dynamic "metric_stat" {
                  for_each = metric_data_queries.value.metric_stat != null ? [metric_data_queries.value.metric_stat] : []
                  content {
                    stat = metric_stat.value.stat
                    unit = metric_stat.value.unit != "" ? metric_stat.value.unit : null

                    metric {
                      metric_name = metric_stat.value.metric_name
                      namespace   = metric_stat.value.namespace

                      dynamic "dimensions" {
                        for_each = metric_stat.value.dimensions
                        content {
                          name  = dimensions.value.name
                          value = dimensions.value.value
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}

# Scheduled actions: time-based capacity changes. Absent capacity values
# mean "leave unchanged", which AWS expresses as -1 -- the reason the spec
# models them as optional ints.
resource "aws_autoscaling_schedule" "this" {
  for_each = { for action in var.spec.scheduled_actions : action.name => action }

  autoscaling_group_name = aws_autoscaling_group.this.name
  scheduled_action_name  = each.value.name

  recurrence = each.value.recurrence != "" ? each.value.recurrence : null
  time_zone  = each.value.time_zone != "" ? each.value.time_zone : null
  start_time = each.value.start_time != "" ? each.value.start_time : null
  end_time   = each.value.end_time != "" ? each.value.end_time : null

  min_size         = each.value.min_size == null ? -1 : each.value.min_size
  max_size         = each.value.max_size == null ? -1 : each.value.max_size
  desired_capacity = each.value.desired_capacity == null ? -1 : each.value.desired_capacity
}

# Lifecycle hooks: pause points in the instance lifecycle. Hooks WITHOUT
# apply_at_launch are managed as standalone hook resources so they stay
# individually updatable; flagged hooks render inline on the group above
# (attached atomically at creation, at the cost of immutability).
resource "aws_autoscaling_lifecycle_hook" "this" {
  for_each = { for hook in var.spec.lifecycle_hooks : hook.name => hook if !hook.apply_at_launch }

  autoscaling_group_name = aws_autoscaling_group.this.name
  name                   = each.value.name
  lifecycle_transition   = each.value.lifecycle_transition

  default_result          = each.value.default_result != "" ? each.value.default_result : null
  heartbeat_timeout       = each.value.heartbeat_timeout_seconds > 0 ? each.value.heartbeat_timeout_seconds : null
  notification_target_arn = each.value.notification_target_arn != "" ? each.value.notification_target_arn : null
  role_arn                = each.value.role_arn != "" ? each.value.role_arn : null
  notification_metadata   = each.value.notification_metadata != "" ? each.value.notification_metadata : null
}

# SNS notifications for fleet lifecycle events.
resource "aws_autoscaling_notification" "this" {
  count = var.spec.notifications != null ? 1 : 0

  group_names   = [aws_autoscaling_group.this.name]
  notifications = var.spec.notifications.event_types
  topic_arn     = var.spec.notifications.topic
}
