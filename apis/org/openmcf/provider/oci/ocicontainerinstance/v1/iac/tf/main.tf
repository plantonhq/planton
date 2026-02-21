resource "oci_container_instances_container_instance" "this" {
  compartment_id      = var.spec.compartment_id.value
  availability_domain = var.spec.availability_domain
  shape               = var.spec.shape
  display_name        = local.display_name
  freeform_tags       = local.freeform_tags

  container_restart_policy             = var.spec.container_restart_policy != "" ? lookup(local.restart_policy_map, var.spec.container_restart_policy, var.spec.container_restart_policy) : null
  fault_domain                         = var.spec.fault_domain != "" ? var.spec.fault_domain : null
  graceful_shutdown_timeout_in_seconds = var.spec.graceful_shutdown_timeout_in_seconds > 0 ? tostring(var.spec.graceful_shutdown_timeout_in_seconds) : null

  shape_config {
    ocpus         = var.spec.shape_config.ocpus
    memory_in_gbs = var.spec.shape_config.memory_in_gbs > 0 ? var.spec.shape_config.memory_in_gbs : null
  }

  dynamic "containers" {
    for_each = var.spec.containers
    content {
      image_url                      = containers.value.image_url
      display_name                   = containers.value.display_name != "" ? containers.value.display_name : null
      command                        = length(containers.value.command) > 0 ? containers.value.command : null
      arguments                      = length(containers.value.arguments) > 0 ? containers.value.arguments : null
      environment_variables          = length(containers.value.environment_variables) > 0 ? containers.value.environment_variables : null
      working_directory              = containers.value.working_directory != "" ? containers.value.working_directory : null
      is_resource_principal_disabled = containers.value.is_resource_principal_disabled ? true : null

      dynamic "resource_config" {
        for_each = containers.value.resource_config != null ? [containers.value.resource_config] : []
        content {
          memory_limit_in_gbs = resource_config.value.memory_limit_in_gbs > 0 ? resource_config.value.memory_limit_in_gbs : null
          vcpus_limit         = resource_config.value.vcpus_limit > 0 ? resource_config.value.vcpus_limit : null
        }
      }

      dynamic "health_checks" {
        for_each = containers.value.health_checks
        content {
          health_check_type        = lookup(local.health_check_type_map, health_checks.value.health_check_type, health_checks.value.health_check_type)
          port                     = health_checks.value.port
          name                     = health_checks.value.name != "" ? health_checks.value.name : null
          path                     = health_checks.value.path != "" ? health_checks.value.path : null
          failure_action           = health_checks.value.failure_action != "" ? lookup(local.failure_action_map, health_checks.value.failure_action, health_checks.value.failure_action) : null
          failure_threshold        = health_checks.value.failure_threshold > 0 ? health_checks.value.failure_threshold : null
          success_threshold        = health_checks.value.success_threshold > 0 ? health_checks.value.success_threshold : null
          initial_delay_in_seconds = health_checks.value.initial_delay_in_seconds > 0 ? health_checks.value.initial_delay_in_seconds : null
          interval_in_seconds      = health_checks.value.interval_in_seconds > 0 ? health_checks.value.interval_in_seconds : null
          timeout_in_seconds       = health_checks.value.timeout_in_seconds > 0 ? health_checks.value.timeout_in_seconds : null

          dynamic "headers" {
            for_each = health_checks.value.headers
            content {
              name  = headers.value.name
              value = headers.value.value
            }
          }
        }
      }

      dynamic "security_context" {
        for_each = containers.value.security_context != null ? [containers.value.security_context] : []
        content {
          security_context_type        = "LINUX"
          is_non_root_user_check_enabled = security_context.value.is_non_root_user_check_enabled ? true : null
          is_root_file_system_readonly   = security_context.value.is_root_file_system_readonly ? true : null
          run_as_user                    = security_context.value.run_as_user > 0 ? security_context.value.run_as_user : null
          run_as_group                   = security_context.value.run_as_group > 0 ? security_context.value.run_as_group : null

          dynamic "capabilities" {
            for_each = security_context.value.capabilities != null ? [security_context.value.capabilities] : []
            content {
              add_capabilities  = length(capabilities.value.add_capabilities) > 0 ? capabilities.value.add_capabilities : null
              drop_capabilities = length(capabilities.value.drop_capabilities) > 0 ? capabilities.value.drop_capabilities : null
            }
          }
        }
      }

      dynamic "volume_mounts" {
        for_each = containers.value.volume_mounts
        content {
          mount_path  = volume_mounts.value.mount_path
          volume_name = volume_mounts.value.volume_name
          is_read_only = volume_mounts.value.is_read_only ? true : null
          partition    = volume_mounts.value.partition > 0 ? volume_mounts.value.partition : null
          sub_path     = volume_mounts.value.sub_path != "" ? volume_mounts.value.sub_path : null
        }
      }
    }
  }

  dynamic "vnics" {
    for_each = var.spec.vnics
    content {
      subnet_id              = vnics.value.subnet_id.value
      display_name           = vnics.value.display_name != "" ? vnics.value.display_name : null
      hostname_label         = vnics.value.hostname_label != "" ? vnics.value.hostname_label : null
      is_public_ip_assigned  = vnics.value.is_public_ip_assigned
      nsg_ids                = length(vnics.value.nsg_ids) > 0 ? [for nsg in vnics.value.nsg_ids : nsg.value] : null
      private_ip             = vnics.value.private_ip != "" ? vnics.value.private_ip : null
      skip_source_dest_check = vnics.value.skip_source_dest_check ? true : null
    }
  }

  dynamic "dns_config" {
    for_each = var.spec.dns_config != null ? [var.spec.dns_config] : []
    content {
      nameservers = length(dns_config.value.nameservers) > 0 ? dns_config.value.nameservers : null
      options     = length(dns_config.value.options) > 0 ? dns_config.value.options : null
      searches    = length(dns_config.value.searches) > 0 ? dns_config.value.searches : null
    }
  }

  dynamic "image_pull_secrets" {
    for_each = var.spec.image_pull_secrets
    content {
      registry_endpoint = image_pull_secrets.value.registry_endpoint
      secret_type       = lookup(local.secret_type_map, image_pull_secrets.value.secret_type, image_pull_secrets.value.secret_type)
      username          = image_pull_secrets.value.username != "" ? image_pull_secrets.value.username : null
      password          = image_pull_secrets.value.password != "" ? image_pull_secrets.value.password : null
      secret_id         = image_pull_secrets.value.secret_id != null ? image_pull_secrets.value.secret_id.value : null
    }
  }

  dynamic "volumes" {
    for_each = var.spec.volumes
    content {
      name          = volumes.value.name
      volume_type   = lookup(local.volume_type_map, volumes.value.volume_type, volumes.value.volume_type)
      backing_store = volumes.value.backing_store != "" ? volumes.value.backing_store : null

      dynamic "configs" {
        for_each = volumes.value.configs
        content {
          data      = configs.value.data != "" ? configs.value.data : null
          file_name = configs.value.file_name != "" ? configs.value.file_name : null
          path      = configs.value.path != "" ? configs.value.path : null
        }
      }
    }
  }
}
