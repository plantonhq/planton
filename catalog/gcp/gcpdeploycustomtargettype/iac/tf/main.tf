# The Cloud Deploy API. disable_on_destroy is false: tearing down one
# custom target type must never disable the API for every pipeline in the
# project.
resource "google_project_service" "clouddeploy_api" {
  project = local.project_id
  service = "clouddeploy.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The custom target type. Location and ID are immutable; the description,
# labels, annotations, and the render and deploy definition update in
# place. Empty optional strings, lists, and maps are not sent.
resource "google_clouddeploy_custom_target_type" "this" {
  project         = local.project_id
  location        = var.spec.location
  name            = local.custom_target_type_id
  description     = var.spec.description != "" ? var.spec.description : null
  labels          = local.final_labels
  annotations     = length(var.spec.annotations) > 0 ? var.spec.annotations : null
  deletion_policy = local.deletion_policy

  dynamic "custom_actions" {
    for_each = var.spec.custom_actions != null ? [var.spec.custom_actions] : []
    content {
      deploy_action = custom_actions.value.deploy_action
      render_action = custom_actions.value.render_action != "" ? custom_actions.value.render_action : null

      # Exactly one source per module (the spec enforces it).
      dynamic "include_skaffold_modules" {
        for_each = custom_actions.value.include_skaffold_modules
        content {
          configs = length(include_skaffold_modules.value.configs) > 0 ? include_skaffold_modules.value.configs : null

          dynamic "git" {
            for_each = include_skaffold_modules.value.git != null ? [include_skaffold_modules.value.git] : []
            content {
              repo = git.value.repo
              path = git.value.path != "" ? git.value.path : null
              ref  = git.value.ref != "" ? git.value.ref : null
            }
          }

          dynamic "google_cloud_build_repo" {
            for_each = include_skaffold_modules.value.google_cloud_build_repo != null ? [include_skaffold_modules.value.google_cloud_build_repo] : []
            content {
              repository = google_cloud_build_repo.value.repository
              path       = google_cloud_build_repo.value.path != "" ? google_cloud_build_repo.value.path : null
              ref        = google_cloud_build_repo.value.ref != "" ? google_cloud_build_repo.value.ref : null
            }
          }

          dynamic "google_cloud_storage" {
            for_each = include_skaffold_modules.value.google_cloud_storage != null ? [include_skaffold_modules.value.google_cloud_storage] : []
            content {
              source = google_cloud_storage.value.source
              path   = google_cloud_storage.value.path != "" ? google_cloud_storage.value.path : null
            }
          }
        }
      }
    }
  }

  dynamic "tasks" {
    for_each = var.spec.tasks != null ? [var.spec.tasks] : []
    content {
      deploy {
        dynamic "container" {
          for_each = tasks.value.deploy.container != null ? [tasks.value.deploy.container] : []
          content {
            image   = container.value.image
            command = length(container.value.command) > 0 ? container.value.command : null
            args    = length(container.value.args) > 0 ? container.value.args : null
            env     = length(container.value.env) > 0 ? container.value.env : null
          }
        }
      }

      dynamic "render" {
        for_each = tasks.value.render != null ? [tasks.value.render] : []
        content {
          dynamic "container" {
            for_each = render.value.container != null ? [render.value.container] : []
            content {
              image   = container.value.image
              command = length(container.value.command) > 0 ? container.value.command : null
              args    = length(container.value.args) > 0 ? container.value.args : null
              env     = length(container.value.env) > 0 ? container.value.env : null
            }
          }
        }
      }
    }
  }

  depends_on = [google_project_service.clouddeploy_api]
}
