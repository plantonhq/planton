# The Cloud Build API. disable_on_destroy is false: tearing down one
# connection must never disable the API for every other build in the
# project.
resource "google_project_service" "cloudbuild_api" {
  project = local.project_id
  service = "cloudbuild.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The connection to one code host. At most one host block is set; every
# credential is a Secret Manager secret version name, never the secret.
resource "google_cloudbuildv2_connection" "this" {
  project         = local.project_id
  location        = var.spec.location
  name            = local.connection_id
  annotations     = length(var.spec.annotations) > 0 ? var.spec.annotations : null
  disabled        = var.spec.disabled ? true : null
  deletion_policy = local.deletion_policy

  dynamic "github_config" {
    for_each = var.spec.github_config != null ? [var.spec.github_config] : []
    content {
      app_installation_id = github_config.value.app_installation_id != 0 ? github_config.value.app_installation_id : null

      dynamic "authorizer_credential" {
        for_each = github_config.value.authorizer_credential != null ? [github_config.value.authorizer_credential] : []
        content {
          oauth_token_secret_version = authorizer_credential.value.oauth_token_secret_version != "" ? authorizer_credential.value.oauth_token_secret_version : null
        }
      }
    }
  }

  dynamic "github_enterprise_config" {
    for_each = var.spec.github_enterprise_config != null ? [var.spec.github_enterprise_config] : []
    content {
      host_uri                      = github_enterprise_config.value.host_uri
      app_id                        = github_enterprise_config.value.app_id != 0 ? github_enterprise_config.value.app_id : null
      app_installation_id           = github_enterprise_config.value.app_installation_id != 0 ? github_enterprise_config.value.app_installation_id : null
      app_slug                      = github_enterprise_config.value.app_slug != "" ? github_enterprise_config.value.app_slug : null
      private_key_secret_version    = github_enterprise_config.value.private_key_secret_version != "" ? github_enterprise_config.value.private_key_secret_version : null
      webhook_secret_secret_version = github_enterprise_config.value.webhook_secret_secret_version != "" ? github_enterprise_config.value.webhook_secret_secret_version : null
      ssl_ca                        = github_enterprise_config.value.ssl_ca != "" ? github_enterprise_config.value.ssl_ca : null

      dynamic "service_directory_config" {
        for_each = github_enterprise_config.value.service_directory_config != null ? [github_enterprise_config.value.service_directory_config] : []
        content {
          service = service_directory_config.value.service
        }
      }
    }
  }

  dynamic "gitlab_config" {
    for_each = var.spec.gitlab_config != null ? [var.spec.gitlab_config] : []
    content {
      host_uri                      = gitlab_config.value.host_uri != "" ? gitlab_config.value.host_uri : null
      webhook_secret_secret_version = gitlab_config.value.webhook_secret_secret_version
      ssl_ca                        = gitlab_config.value.ssl_ca != "" ? gitlab_config.value.ssl_ca : null

      authorizer_credential {
        user_token_secret_version = gitlab_config.value.authorizer_credential.user_token_secret_version
      }

      read_authorizer_credential {
        user_token_secret_version = gitlab_config.value.read_authorizer_credential.user_token_secret_version
      }

      dynamic "service_directory_config" {
        for_each = gitlab_config.value.service_directory_config != null ? [gitlab_config.value.service_directory_config] : []
        content {
          service = service_directory_config.value.service
        }
      }
    }
  }

  dynamic "bitbucket_cloud_config" {
    for_each = var.spec.bitbucket_cloud_config != null ? [var.spec.bitbucket_cloud_config] : []
    content {
      workspace                     = bitbucket_cloud_config.value.workspace
      webhook_secret_secret_version = bitbucket_cloud_config.value.webhook_secret_secret_version

      authorizer_credential {
        user_token_secret_version = bitbucket_cloud_config.value.authorizer_credential.user_token_secret_version
      }

      read_authorizer_credential {
        user_token_secret_version = bitbucket_cloud_config.value.read_authorizer_credential.user_token_secret_version
      }
    }
  }

  dynamic "bitbucket_data_center_config" {
    for_each = var.spec.bitbucket_data_center_config != null ? [var.spec.bitbucket_data_center_config] : []
    content {
      host_uri                      = bitbucket_data_center_config.value.host_uri
      webhook_secret_secret_version = bitbucket_data_center_config.value.webhook_secret_secret_version
      ssl_ca                        = bitbucket_data_center_config.value.ssl_ca != "" ? bitbucket_data_center_config.value.ssl_ca : null

      authorizer_credential {
        user_token_secret_version = bitbucket_data_center_config.value.authorizer_credential.user_token_secret_version
      }

      read_authorizer_credential {
        user_token_secret_version = bitbucket_data_center_config.value.read_authorizer_credential.user_token_secret_version
      }

      dynamic "service_directory_config" {
        for_each = bitbucket_data_center_config.value.service_directory_config != null ? [bitbucket_data_center_config.value.service_directory_config] : []
        content {
          service = service_directory_config.value.service
        }
      }
    }
  }

  depends_on = [google_project_service.cloudbuild_api]
}
