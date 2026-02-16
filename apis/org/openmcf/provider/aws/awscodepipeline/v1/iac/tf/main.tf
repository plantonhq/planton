# --- CodePipeline ---

resource "aws_codepipeline" "this" {
  name           = var.metadata.id
  role_arn       = var.spec.role_arn.value
  pipeline_type  = var.spec.pipeline_type != "" ? var.spec.pipeline_type : "V2"
  execution_mode = var.spec.execution_mode != "" ? var.spec.execution_mode : "SUPERSEDED"
  tags           = local.tags

  # Single-region artifact store (no region field)
  dynamic "artifact_store" {
    for_each = local.is_single_region ? var.spec.artifact_stores : []
    content {
      location = artifact_store.value.location.value
      type     = "S3"

      dynamic "encryption_key" {
        for_each = artifact_store.value.encryption_key_id != null ? [artifact_store.value.encryption_key_id] : []
        content {
          id   = encryption_key.value.value
          type = "KMS"
        }
      }
    }
  }

  # Cross-region artifact stores (with region field)
  dynamic "artifact_store" {
    for_each = local.is_single_region ? [] : var.spec.artifact_stores
    content {
      location = artifact_store.value.location.value
      type     = "S3"
      region   = artifact_store.value.region

      dynamic "encryption_key" {
        for_each = artifact_store.value.encryption_key_id != null ? [artifact_store.value.encryption_key_id] : []
        content {
          id   = encryption_key.value.value
          type = "KMS"
        }
      }
    }
  }

  # Stages
  dynamic "stage" {
    for_each = var.spec.stages
    content {
      name = stage.value.name

      dynamic "action" {
        for_each = stage.value.actions
        content {
          name             = action.value.name
          category         = action.value.category
          owner            = action.value.owner
          provider         = action.value.provider
          version          = action.value.version
          configuration    = length(action.value.configuration) > 0 ? action.value.configuration : null
          input_artifacts  = length(action.value.input_artifacts) > 0 ? action.value.input_artifacts : null
          output_artifacts = length(action.value.output_artifacts) > 0 ? action.value.output_artifacts : null
          namespace        = action.value.namespace != "" ? action.value.namespace : null
          region           = action.value.region != "" ? action.value.region : null
          role_arn         = try(action.value.role_arn.value, null)
          run_order        = action.value.run_order > 0 ? action.value.run_order : null
        }
      }
    }
  }

  # Triggers (V2 only)
  dynamic "trigger" {
    for_each = local.has_triggers ? var.spec.triggers : []
    content {
      provider_type = trigger.value.provider_type

      git_configuration {
        source_action_name = trigger.value.git_configuration.source_action_name

        dynamic "push" {
          for_each = try(trigger.value.git_configuration.push, [])
          content {
            dynamic "branches" {
              for_each = push.value.branches != null ? [push.value.branches] : []
              content {
                includes = branches.value.includes
                excludes = branches.value.excludes
              }
            }
            dynamic "file_paths" {
              for_each = push.value.file_paths != null ? [push.value.file_paths] : []
              content {
                includes = file_paths.value.includes
                excludes = file_paths.value.excludes
              }
            }
            dynamic "tags" {
              for_each = push.value.tags != null ? [push.value.tags] : []
              content {
                includes = tags.value.includes
                excludes = tags.value.excludes
              }
            }
          }
        }

        dynamic "pull_request" {
          for_each = try(trigger.value.git_configuration.pull_request, [])
          content {
            events = try(pull_request.value.events, null)
            dynamic "branches" {
              for_each = pull_request.value.branches != null ? [pull_request.value.branches] : []
              content {
                includes = branches.value.includes
                excludes = branches.value.excludes
              }
            }
            dynamic "file_paths" {
              for_each = pull_request.value.file_paths != null ? [pull_request.value.file_paths] : []
              content {
                includes = file_paths.value.includes
                excludes = file_paths.value.excludes
              }
            }
          }
        }
      }
    }
  }

  # Variables (V2 only)
  dynamic "variable" {
    for_each = local.has_variables ? var.spec.variables : []
    content {
      name          = variable.value.name
      default_value = variable.value.default_value != "" ? variable.value.default_value : null
      description   = variable.value.description != "" ? variable.value.description : null
    }
  }
}

# --- Outputs ---

output "pipeline_arn" {
  description = "ARN of the CodePipeline pipeline"
  value       = aws_codepipeline.this.arn
}

output "pipeline_name" {
  description = "Name of the CodePipeline pipeline"
  value       = aws_codepipeline.this.name
}
