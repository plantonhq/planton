resource "digitalocean_app" "main" {
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  spec {
    # The App Platform app is named from spec.app_name, never metadata.name:
    # the API caps app names at 32 characters and requires them to be unique
    # across the account, neither of which a Planton metadata name guarantees.
    name   = var.spec.app_name
    region = local.region

    function {
      name = var.spec.function_name
      # Omitted when unset so App Platform reads project.yml from the
      # repository root; an empty string would name a directory that does not
      # exist.
      source_dir = var.spec.source_directory != "" ? var.spec.source_directory : null

      dynamic "git" {
        for_each = var.spec.git != null ? [var.spec.git] : []
        content {
          repo_clone_url = git.value.repo_clone_url
          branch         = git.value.branch
        }
      }
      dynamic "github" {
        for_each = var.spec.github != null ? [var.spec.github] : []
        content {
          repo           = github.value.repo
          branch         = github.value.branch
          deploy_on_push = github.value.deploy_on_push
        }
      }
      dynamic "gitlab" {
        for_each = var.spec.gitlab != null ? [var.spec.gitlab] : []
        content {
          repo           = gitlab.value.repo
          branch         = gitlab.value.branch
          deploy_on_push = gitlab.value.deploy_on_push
        }
      }
      dynamic "bitbucket" {
        for_each = var.spec.bitbucket != null ? [var.spec.bitbucket] : []
        content {
          repo           = bitbucket.value.repo
          branch         = bitbucket.value.branch
          deploy_on_push = bitbucket.value.deploy_on_push
        }
      }

      dynamic "env" {
        for_each = var.spec.envs
        content {
          key   = env.value.key
          value = env.value.secret != "" ? env.value.secret : env.value.plaintext
          type  = env.value.secret != "" ? "SECRET" : "GENERAL"
          scope = (env.value.scope == "" || endswith(env.value.scope, "_unspecified")) ? "RUN_AND_BUILD_TIME" : upper(env.value.scope)
        }
      }

      dynamic "alert" {
        for_each = var.spec.alerts
        content {
          rule     = upper(alert.value.rule)
          operator = upper(alert.value.operator)
          window   = upper(alert.value.window)
          value    = alert.value.value
          disabled = alert.value.disabled
          dynamic "destinations" {
            for_each = alert.value.destinations != null ? [alert.value.destinations] : []
            content {
              emails = length(destinations.value.emails) > 0 ? destinations.value.emails : null
              dynamic "slack_webhooks" {
                for_each = destinations.value.slack_webhooks
                content {
                  channel = slack_webhooks.value.channel
                  url     = slack_webhooks.value.url
                }
              }
            }
          }
        }
      }

      dynamic "log_destination" {
        for_each = var.spec.log_destinations
        content {
          name = log_destination.value.name
          dynamic "papertrail" {
            for_each = log_destination.value.papertrail != null ? [log_destination.value.papertrail] : []
            content { endpoint = papertrail.value.endpoint }
          }
          dynamic "datadog" {
            for_each = log_destination.value.datadog != null ? [log_destination.value.datadog] : []
            content {
              api_key  = datadog.value.api_key
              endpoint = datadog.value.endpoint != "" ? datadog.value.endpoint : null
            }
          }
          dynamic "logtail" {
            for_each = log_destination.value.logtail != null ? [log_destination.value.logtail] : []
            content { token = logtail.value.token }
          }
          dynamic "open_search" {
            for_each = log_destination.value.open_search != null ? [log_destination.value.open_search] : []
            content {
              endpoint     = open_search.value.endpoint != "" ? open_search.value.endpoint : null
              index_name   = open_search.value.index_name != "" ? open_search.value.index_name : null
              cluster_name = open_search.value.cluster_name != "" ? open_search.value.cluster_name : null
              basic_auth {
                user     = try(open_search.value.basic_auth.user, null)
                password = try(open_search.value.basic_auth.password, null)
              }
            }
          }
        }
      }
    }
  }
}
