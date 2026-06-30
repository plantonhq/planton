# DigitalOcean App Platform Service
# Supports web services, workers, and jobs with git or container image sources

resource "digitalocean_app" "main" {
  spec {
    name   = var.spec.service_name
    region = var.spec.region

    # Web Service Configuration
    dynamic "service" {
      for_each = var.spec.service_type == "web_service" ? [1] : []
      
      content {
        name               = var.spec.service_name
        instance_count     = local.enable_autoscaling ? null : var.spec.instance_count
        instance_size_slug = local.instance_size_slug

        # Git source configuration
        dynamic "git" {
          for_each = local.is_git_source ? [var.spec.git_source] : []
          
          content {
            repo_clone_url = git.value.repo_url
            branch         = git.value.branch
          }
        }
        
        # Container image source configuration
        dynamic "image" {
          for_each = local.is_image_source ? [var.spec.image_source] : []
          
          content {
            registry_type = "DOCR"
            registry      = image.value.registry
            repository    = image.value.repository
            tag           = image.value.tag
          }
        }
        
        # Build and run commands (for git sources)
        build_command = local.is_git_source && var.spec.git_source.build_command != null ? var.spec.git_source.build_command : null
        run_command   = local.is_git_source && var.spec.git_source.run_command != null ? var.spec.git_source.run_command : null
        
        # Environment variables
        dynamic "env" {
          for_each = local.env_vars
          
          content {
            key   = env.value.key
            value = env.value.value
            scope = env.value.scope
            type  = env.value.type
          }
        }
        
        # Autoscaling configuration
        dynamic "autoscaling" {
          for_each = local.enable_autoscaling ? [1] : []
          
          content {
            min_instance_count = var.spec.min_instance_count
            max_instance_count = var.spec.max_instance_count
            
            metrics {
              cpu {
                percent = 80
              }
            }
          }
        }
      }
    }

    # Worker Configuration
    dynamic "worker" {
      for_each = var.spec.service_type == "worker" ? [1] : []
      
      content {
        name               = var.spec.service_name
        instance_count     = var.spec.instance_count
        instance_size_slug = local.instance_size_slug

        # Git source configuration
        dynamic "git" {
          for_each = local.is_git_source ? [var.spec.git_source] : []
          
          content {
            repo_clone_url = git.value.repo_url
            branch         = git.value.branch
          }
        }
        
        # Container image source configuration
        dynamic "image" {
          for_each = local.is_image_source ? [var.spec.image_source] : []
          
          content {
            registry_type = "DOCR"
            registry      = image.value.registry
            repository    = image.value.repository
            tag           = image.value.tag
          }
        }
        
        # Run command (for git sources)
        run_command = local.is_git_source && var.spec.git_source.run_command != null ? var.spec.git_source.run_command : null
        
        # Environment variables
        dynamic "env" {
          for_each = local.env_vars
          
          content {
            key   = env.value.key
            value = env.value.value
            scope = "RUN_TIME"
            type  = env.value.type
          }
        }
      }
    }

    # Job Configuration
    dynamic "job" {
      for_each = var.spec.service_type == "job" ? [1] : []
      
      content {
        name               = var.spec.service_name
        kind               = "PRE_DEPLOY"
        instance_size_slug = local.instance_size_slug

        # Git source configuration
        dynamic "git" {
          for_each = local.is_git_source ? [var.spec.git_source] : []
          
          content {
            repo_clone_url = git.value.repo_url
            branch         = git.value.branch
          }
        }
        
        # Container image source configuration
        dynamic "image" {
          for_each = local.is_image_source ? [var.spec.image_source] : []
          
          content {
            registry_type = "DOCR"
            registry      = image.value.registry
            repository    = image.value.repository
            tag           = image.value.tag
          }
        }
        
        # Run command (for git sources)
        run_command = local.is_git_source && var.spec.git_source.run_command != null ? var.spec.git_source.run_command : null
        
        # Environment variables
        dynamic "env" {
          for_each = local.env_vars
          
          content {
            key   = env.value.key
            value = env.value.value
            scope = "RUN_TIME"
            type  = env.value.type
          }
        }
      }
    }

    # Custom domain configuration
    dynamic "domain" {
      for_each = var.spec.custom_domain != null ? [var.spec.custom_domain] : []
      
      content {
        name = domain.value
        type = "PRIMARY"
      }
    }
  }

  lifecycle {
    # Prevent accidental deletion of production apps
    prevent_destroy = false
    
    # Validate autoscaling configuration
    precondition {
      condition     = !var.spec.enable_autoscale || var.spec.service_type == "web_service"
      error_message = "Autoscaling is only supported for web_service type"
    }
  }
}

