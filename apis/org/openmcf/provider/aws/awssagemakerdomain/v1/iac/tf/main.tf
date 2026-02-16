resource "aws_sagemaker_domain" "this" {
  domain_name             = var.metadata.id
  auth_mode               = var.spec.auth_mode
  vpc_id                  = var.spec.vpc_id
  subnet_ids              = var.spec.subnet_ids
  kms_key_id              = var.spec.kms_key_id != "" ? var.spec.kms_key_id : null
  app_network_access_type = var.spec.app_network_access_type
  tags                    = local.tags

  # --- domain_settings ---

  dynamic "domain_settings" {
    for_each = local.has_domain_settings ? [1] : []
    content {
      security_group_ids = length(var.spec.domain_security_group_ids) > 0 ? var.spec.domain_security_group_ids : null

      dynamic "docker_settings" {
        for_each = var.spec.docker_settings != null ? [var.spec.docker_settings] : []
        content {
          enable_docker_access      = docker_settings.value.enable_docker_access
          vpc_only_trusted_accounts = length(docker_settings.value.vpc_only_trusted_accounts) > 0 ? docker_settings.value.vpc_only_trusted_accounts : null
        }
      }
    }
  }

  # --- default_user_settings ---

  default_user_settings {
    execution_role  = var.spec.default_user_settings.execution_role
    security_groups = length(var.spec.default_user_settings.security_groups) > 0 ? var.spec.default_user_settings.security_groups : null

    default_landing_uri = var.spec.default_user_settings.default_landing_uri != "" ? var.spec.default_user_settings.default_landing_uri : null
    studio_web_portal   = var.spec.default_user_settings.studio_web_portal != "" ? var.spec.default_user_settings.studio_web_portal : null

    # --- jupyter_lab_app_settings ---

    dynamic "jupyter_lab_app_settings" {
      for_each = var.spec.default_user_settings.jupyter_lab_app_settings != null ? [var.spec.default_user_settings.jupyter_lab_app_settings] : []
      content {
        lifecycle_config_arns = length(jupyter_lab_app_settings.value.lifecycle_config_arns) > 0 ? jupyter_lab_app_settings.value.lifecycle_config_arns : null

        dynamic "default_resource_spec" {
          for_each = jupyter_lab_app_settings.value.default_resource_spec != null ? [jupyter_lab_app_settings.value.default_resource_spec] : []
          content {
            instance_type        = default_resource_spec.value.instance_type != "" ? default_resource_spec.value.instance_type : null
            lifecycle_config_arn = default_resource_spec.value.lifecycle_config_arn != "" ? default_resource_spec.value.lifecycle_config_arn : null
            sagemaker_image_arn  = default_resource_spec.value.sagemaker_image_arn != "" ? default_resource_spec.value.sagemaker_image_arn : null
          }
        }

        dynamic "custom_image" {
          for_each = jupyter_lab_app_settings.value.custom_image
          content {
            app_image_config_name = custom_image.value.app_image_config_name
            image_name            = custom_image.value.image_name
            image_version_number  = custom_image.value.image_version_number
          }
        }

        dynamic "code_repository" {
          for_each = jupyter_lab_app_settings.value.code_repository
          content {
            repository_url = code_repository.value.repository_url
          }
        }

        dynamic "app_lifecycle_management" {
          for_each = jupyter_lab_app_settings.value.app_lifecycle_management != null ? [jupyter_lab_app_settings.value.app_lifecycle_management] : []
          content {
            dynamic "idle_settings" {
              for_each = app_lifecycle_management.value.idle_settings != null ? [app_lifecycle_management.value.idle_settings] : []
              content {
                idle_timeout_in_minutes     = idle_settings.value.idle_timeout_in_minutes
                lifecycle_management        = idle_settings.value.lifecycle_management != "" ? idle_settings.value.lifecycle_management : null
                max_idle_timeout_in_minutes = idle_settings.value.max_idle_timeout_in_minutes
                min_idle_timeout_in_minutes = idle_settings.value.min_idle_timeout_in_minutes
              }
            }
          }
        }
      }
    }

    # --- kernel_gateway_app_settings ---

    dynamic "kernel_gateway_app_settings" {
      for_each = var.spec.default_user_settings.kernel_gateway_app_settings != null ? [var.spec.default_user_settings.kernel_gateway_app_settings] : []
      content {
        lifecycle_config_arns = length(kernel_gateway_app_settings.value.lifecycle_config_arns) > 0 ? kernel_gateway_app_settings.value.lifecycle_config_arns : null

        dynamic "default_resource_spec" {
          for_each = kernel_gateway_app_settings.value.default_resource_spec != null ? [kernel_gateway_app_settings.value.default_resource_spec] : []
          content {
            instance_type        = default_resource_spec.value.instance_type != "" ? default_resource_spec.value.instance_type : null
            lifecycle_config_arn = default_resource_spec.value.lifecycle_config_arn != "" ? default_resource_spec.value.lifecycle_config_arn : null
            sagemaker_image_arn  = default_resource_spec.value.sagemaker_image_arn != "" ? default_resource_spec.value.sagemaker_image_arn : null
          }
        }

        dynamic "custom_image" {
          for_each = kernel_gateway_app_settings.value.custom_image
          content {
            app_image_config_name = custom_image.value.app_image_config_name
            image_name            = custom_image.value.image_name
            image_version_number  = custom_image.value.image_version_number
          }
        }
      }
    }

    # --- sharing_settings ---

    dynamic "sharing_settings" {
      for_each = var.spec.default_user_settings.sharing_settings != null ? [var.spec.default_user_settings.sharing_settings] : []
      content {
        notebook_output_option = sharing_settings.value.notebook_output_option != "" ? sharing_settings.value.notebook_output_option : null
        s3_kms_key_id          = sharing_settings.value.s3_kms_key_id != "" ? sharing_settings.value.s3_kms_key_id : null
        s3_output_path         = sharing_settings.value.s3_output_path != "" ? sharing_settings.value.s3_output_path : null
      }
    }

    # --- space_storage_settings ---

    dynamic "space_storage_settings" {
      for_each = var.spec.default_user_settings.space_storage_settings != null ? [var.spec.default_user_settings.space_storage_settings] : []
      content {
        dynamic "default_ebs_storage_settings" {
          for_each = space_storage_settings.value.default_ebs_storage_settings != null ? [space_storage_settings.value.default_ebs_storage_settings] : []
          content {
            default_ebs_volume_size_in_gb = default_ebs_storage_settings.value.default_ebs_volume_size_in_gb
            maximum_ebs_volume_size_in_gb = default_ebs_storage_settings.value.maximum_ebs_volume_size_in_gb
          }
        }
      }
    }
  }
}
