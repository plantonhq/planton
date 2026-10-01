# The Cloud Build API. disable_on_destroy is false: tearing down one
# trigger must never disable the API for every other build in the project.
resource "google_project_service" "cloudbuild_api" {
  project = local.project_id
  service = "cloudbuild.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The trigger. Location is immutable; everything else, the name included,
# updates in place. Exactly one of build, filename, or git_file_source,
# and at most one event source (none is a manual trigger). Optional
# arguments are sent only when set, so Google's defaults apply.
resource "google_cloudbuild_trigger" "this" {
  project            = local.project_id
  location           = local.location
  name               = local.trigger_name
  description        = var.spec.description != "" ? var.spec.description : null
  disabled           = var.spec.disabled ? true : null
  tags               = length(var.spec.tags) > 0 ? var.spec.tags : null
  substitutions      = length(var.spec.substitutions) > 0 ? var.spec.substitutions : null
  service_account    = var.spec.service_account != "" ? var.spec.service_account : null
  filter             = var.spec.filter != "" ? var.spec.filter : null
  ignored_files      = length(var.spec.ignored_files) > 0 ? var.spec.ignored_files : null
  included_files     = length(var.spec.included_files) > 0 ? var.spec.included_files : null
  include_build_logs = var.spec.include_build_logs != "" ? var.spec.include_build_logs : null
  filename           = var.spec.filename != "" ? var.spec.filename : null
  deletion_policy    = local.deletion_policy

  dynamic "approval_config" {
    for_each = var.spec.approval_config != null ? [var.spec.approval_config] : []
    content {
      approval_required = approval_config.value.approval_required ? true : null
    }
  }

  # --- Event sources ---

  dynamic "repository_event_config" {
    for_each = var.spec.repository_event_config != null ? [var.spec.repository_event_config] : []
    content {
      repository = repository_event_config.value.repository != "" ? repository_event_config.value.repository : null

      dynamic "pull_request" {
        for_each = repository_event_config.value.pull_request != null ? [repository_event_config.value.pull_request] : []
        content {
          branch          = pull_request.value.branch != "" ? pull_request.value.branch : null
          comment_control = pull_request.value.comment_control != "" ? pull_request.value.comment_control : null
          invert_regex    = pull_request.value.invert_regex ? true : null
        }
      }

      dynamic "push" {
        for_each = repository_event_config.value.push != null ? [repository_event_config.value.push] : []
        content {
          branch       = push.value.branch != "" ? push.value.branch : null
          tag          = push.value.tag != "" ? push.value.tag : null
          invert_regex = push.value.invert_regex ? true : null
        }
      }
    }
  }

  dynamic "github" {
    for_each = var.spec.github != null ? [var.spec.github] : []
    content {
      owner                           = github.value.owner != "" ? github.value.owner : null
      name                            = github.value.name != "" ? github.value.name : null
      enterprise_config_resource_name = github.value.enterprise_config_resource_name != "" ? github.value.enterprise_config_resource_name : null

      dynamic "pull_request" {
        for_each = github.value.pull_request != null ? [github.value.pull_request] : []
        content {
          branch          = pull_request.value.branch
          comment_control = pull_request.value.comment_control != "" ? pull_request.value.comment_control : null
          invert_regex    = pull_request.value.invert_regex ? true : null
        }
      }

      dynamic "push" {
        for_each = github.value.push != null ? [github.value.push] : []
        content {
          branch       = push.value.branch != "" ? push.value.branch : null
          tag          = push.value.tag != "" ? push.value.tag : null
          invert_regex = push.value.invert_regex ? true : null
        }
      }
    }
  }

  dynamic "bitbucket_server_trigger_config" {
    for_each = var.spec.bitbucket_server_trigger_config != null ? [var.spec.bitbucket_server_trigger_config] : []
    content {
      bitbucket_server_config_resource = bitbucket_server_trigger_config.value.bitbucket_server_config_resource
      project_key                      = bitbucket_server_trigger_config.value.project_key
      repo_slug                        = bitbucket_server_trigger_config.value.repo_slug

      dynamic "pull_request" {
        for_each = bitbucket_server_trigger_config.value.pull_request != null ? [bitbucket_server_trigger_config.value.pull_request] : []
        content {
          branch          = pull_request.value.branch
          comment_control = pull_request.value.comment_control != "" ? pull_request.value.comment_control : null
          invert_regex    = pull_request.value.invert_regex ? true : null
        }
      }

      dynamic "push" {
        for_each = bitbucket_server_trigger_config.value.push != null ? [bitbucket_server_trigger_config.value.push] : []
        content {
          branch       = push.value.branch != "" ? push.value.branch : null
          tag          = push.value.tag != "" ? push.value.tag : null
          invert_regex = push.value.invert_regex ? true : null
        }
      }
    }
  }

  dynamic "developer_connect_event_config" {
    for_each = var.spec.developer_connect_event_config != null ? [var.spec.developer_connect_event_config] : []
    content {
      git_repository_link = developer_connect_event_config.value.git_repository_link

      dynamic "pull_request" {
        for_each = developer_connect_event_config.value.pull_request != null ? [developer_connect_event_config.value.pull_request] : []
        content {
          branch          = pull_request.value.branch != "" ? pull_request.value.branch : null
          comment_control = pull_request.value.comment_control != "" ? pull_request.value.comment_control : null
          invert_regex    = pull_request.value.invert_regex ? true : null
        }
      }

      dynamic "push" {
        for_each = developer_connect_event_config.value.push != null ? [developer_connect_event_config.value.push] : []
        content {
          branch       = push.value.branch != "" ? push.value.branch : null
          tag          = push.value.tag != "" ? push.value.tag : null
          invert_regex = push.value.invert_regex ? true : null
        }
      }
    }
  }

  dynamic "trigger_template" {
    for_each = var.spec.trigger_template != null ? [var.spec.trigger_template] : []
    content {
      project_id   = trigger_template.value.project_id != "" ? trigger_template.value.project_id : null
      repo_name    = trigger_template.value.repo_name != "" ? trigger_template.value.repo_name : null
      dir          = trigger_template.value.dir != "" ? trigger_template.value.dir : null
      branch_name  = trigger_template.value.branch_name != "" ? trigger_template.value.branch_name : null
      tag_name     = trigger_template.value.tag_name != "" ? trigger_template.value.tag_name : null
      commit_sha   = trigger_template.value.commit_sha != "" ? trigger_template.value.commit_sha : null
      invert_regex = trigger_template.value.invert_regex ? true : null
    }
  }

  dynamic "pubsub_config" {
    for_each = var.spec.pubsub_config != null ? [var.spec.pubsub_config] : []
    content {
      topic                 = pubsub_config.value.topic
      service_account_email = pubsub_config.value.service_account_email != "" ? pubsub_config.value.service_account_email : null
    }
  }

  dynamic "webhook_config" {
    for_each = var.spec.webhook_config != null ? [var.spec.webhook_config] : []
    content {
      secret = webhook_config.value.secret
    }
  }

  dynamic "source_to_build" {
    for_each = var.spec.source_to_build != null ? [var.spec.source_to_build] : []
    content {
      repository               = source_to_build.value.repository != "" ? source_to_build.value.repository : null
      uri                      = source_to_build.value.uri != "" ? source_to_build.value.uri : null
      ref                      = source_to_build.value.ref
      repo_type                = source_to_build.value.repo_type
      github_enterprise_config = source_to_build.value.github_enterprise_config != "" ? source_to_build.value.github_enterprise_config : null
      bitbucket_server_config  = source_to_build.value.bitbucket_server_config != "" ? source_to_build.value.bitbucket_server_config : null
    }
  }

  # --- Build configuration ---

  dynamic "git_file_source" {
    for_each = var.spec.git_file_source != null ? [var.spec.git_file_source] : []
    content {
      path                     = git_file_source.value.path
      repo_type                = git_file_source.value.repo_type
      repository               = git_file_source.value.repository != "" ? git_file_source.value.repository : null
      uri                      = git_file_source.value.uri != "" ? git_file_source.value.uri : null
      revision                 = git_file_source.value.revision != "" ? git_file_source.value.revision : null
      github_enterprise_config = git_file_source.value.github_enterprise_config != "" ? git_file_source.value.github_enterprise_config : null
      bitbucket_server_config  = git_file_source.value.bitbucket_server_config != "" ? git_file_source.value.bitbucket_server_config : null
    }
  }

  dynamic "build" {
    for_each = var.spec.build != null ? [var.spec.build] : []
    content {
      timeout       = build.value.timeout != "" ? build.value.timeout : null
      queue_ttl     = build.value.queue_ttl != "" ? build.value.queue_ttl : null
      images        = length(build.value.images) > 0 ? build.value.images : null
      logs_bucket   = build.value.logs_bucket != "" ? build.value.logs_bucket : null
      substitutions = length(build.value.substitutions) > 0 ? build.value.substitutions : null
      tags          = length(build.value.tags) > 0 ? build.value.tags : null

      # The spec's steps, in order (Google's Build.steps).
      dynamic "step" {
        for_each = build.value.steps
        content {
          name             = step.value.name
          id               = step.value.id != "" ? step.value.id : null
          args             = length(step.value.args) > 0 ? step.value.args : null
          entrypoint       = step.value.entrypoint != "" ? step.value.entrypoint : null
          script           = step.value.script != "" ? step.value.script : null
          dir              = step.value.dir != "" ? step.value.dir : null
          env              = length(step.value.env) > 0 ? step.value.env : null
          secret_env       = length(step.value.secret_env) > 0 ? step.value.secret_env : null
          wait_for         = length(step.value.wait_for) > 0 ? step.value.wait_for : null
          timeout          = step.value.timeout != "" ? step.value.timeout : null
          allow_failure    = step.value.allow_failure ? true : null
          allow_exit_codes = length(step.value.allow_exit_codes) > 0 ? step.value.allow_exit_codes : null

          dynamic "volumes" {
            for_each = step.value.volumes
            content {
              name = volumes.value.name
              path = volumes.value.path
            }
          }
        }
      }

      dynamic "artifacts" {
        for_each = build.value.artifacts != null ? [build.value.artifacts] : []
        content {
          images = length(artifacts.value.images) > 0 ? artifacts.value.images : null

          dynamic "objects" {
            for_each = artifacts.value.objects != null ? [artifacts.value.objects] : []
            content {
              location = objects.value.location != "" ? objects.value.location : null
              paths    = length(objects.value.paths) > 0 ? objects.value.paths : null
            }
          }

          dynamic "maven_artifacts" {
            for_each = artifacts.value.maven_artifacts
            content {
              repository  = maven_artifacts.value.repository != "" ? maven_artifacts.value.repository : null
              path        = maven_artifacts.value.path != "" ? maven_artifacts.value.path : null
              artifact_id = maven_artifacts.value.artifact_id != "" ? maven_artifacts.value.artifact_id : null
              group_id    = maven_artifacts.value.group_id != "" ? maven_artifacts.value.group_id : null
              version     = maven_artifacts.value.version != "" ? maven_artifacts.value.version : null
            }
          }

          dynamic "npm_packages" {
            for_each = artifacts.value.npm_packages
            content {
              repository   = npm_packages.value.repository != "" ? npm_packages.value.repository : null
              package_path = npm_packages.value.package_path != "" ? npm_packages.value.package_path : null
            }
          }

          dynamic "python_packages" {
            for_each = artifacts.value.python_packages
            content {
              repository = python_packages.value.repository != "" ? python_packages.value.repository : null
              paths      = length(python_packages.value.paths) > 0 ? python_packages.value.paths : null
            }
          }
        }
      }

      # dynamic_substitutions and substitution_option are fixed by Google
      # for triggered builds and never sent.
      dynamic "options" {
        for_each = build.value.options != null ? [build.value.options] : []
        content {
          machine_type            = options.value.machine_type != "" ? options.value.machine_type : null
          disk_size_gb            = options.value.disk_size_gb != 0 ? options.value.disk_size_gb : null
          worker_pool             = options.value.worker_pool != "" ? options.value.worker_pool : null
          logging                 = options.value.logging != "" ? options.value.logging : null
          log_streaming_option    = options.value.log_streaming_option != "" ? options.value.log_streaming_option : null
          requested_verify_option = options.value.requested_verify_option != "" ? options.value.requested_verify_option : null
          source_provenance_hash  = length(options.value.source_provenance_hash) > 0 ? options.value.source_provenance_hash : null
          env                     = length(options.value.env) > 0 ? options.value.env : null
          secret_env              = length(options.value.secret_env) > 0 ? options.value.secret_env : null

          dynamic "volumes" {
            for_each = options.value.volumes
            content {
              name = volumes.value.name != "" ? volumes.value.name : null
              path = volumes.value.path != "" ? volumes.value.path : null
            }
          }
        }
      }

      dynamic "source" {
        for_each = build.value.source != null ? [build.value.source] : []
        content {
          dynamic "repo_source" {
            for_each = source.value.repo_source != null ? [source.value.repo_source] : []
            content {
              project_id    = repo_source.value.project_id != "" ? repo_source.value.project_id : null
              repo_name     = repo_source.value.repo_name
              dir           = repo_source.value.dir != "" ? repo_source.value.dir : null
              branch_name   = repo_source.value.branch_name != "" ? repo_source.value.branch_name : null
              tag_name      = repo_source.value.tag_name != "" ? repo_source.value.tag_name : null
              commit_sha    = repo_source.value.commit_sha != "" ? repo_source.value.commit_sha : null
              invert_regex  = repo_source.value.invert_regex ? true : null
              substitutions = length(repo_source.value.substitutions) > 0 ? repo_source.value.substitutions : null
            }
          }

          dynamic "storage_source" {
            for_each = source.value.storage_source != null ? [source.value.storage_source] : []
            content {
              bucket     = storage_source.value.bucket
              object     = storage_source.value.object
              generation = storage_source.value.generation != "" ? storage_source.value.generation : null
            }
          }
        }
      }

      dynamic "available_secrets" {
        for_each = build.value.available_secrets != null ? [build.value.available_secrets] : []
        content {
          dynamic "secret_manager" {
            for_each = available_secrets.value.secret_manager
            content {
              env          = secret_manager.value.env
              version_name = secret_manager.value.version_name
            }
          }
        }
      }

      # The spec's secrets (Google's Build.secrets).
      dynamic "secret" {
        for_each = build.value.secrets
        content {
          kms_key_name = secret.value.kms_key_name
          secret_env   = length(secret.value.secret_env) > 0 ? secret.value.secret_env : null
        }
      }
    }
  }

  depends_on = [google_project_service.cloudbuild_api]
}
