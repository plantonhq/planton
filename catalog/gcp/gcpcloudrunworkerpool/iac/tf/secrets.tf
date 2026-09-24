# Secret values the environment carries (env[].secret_value) are kept in
# Secret Manager, one secret per variable, so the revision template references
# a secret the worker pool owns and never holds the value. The naming,
# placement, and grant rules match the Pulumi module's shared cloudrunenv
# helper:
#   - id "runpool_<region>_<pool>_<container>_<variable>" ('.' becomes '-',
#     an unnamed container is "c<index>");
#   - replicated only in the pool's region;
#   - secretAccessor on that secret alone for the runtime identity (the
#     pool's service_account, else the project's Compute Engine default).
# The variable reads the exact version, so a new value stamps a new revision.

locals {
  # Keyed "<container index>/<variable>" -- the address the env block in
  # main.tf looks each variable up by.
  env_secrets = merge([
    for container_index, container in var.spec.containers : {
      for env in container.env : "${container_index}/${env.name}" => {
        secret_id = join("_", [
          "runpool",
          var.spec.region,
          local.worker_pool_name,
          container.name != "" ? container.name : "c${container_index}",
          replace(env.name, ".", "-"),
        ])
        value = env.secret_value
      } if env.secret_value != ""
    }
  ]...)

  env_secret_ids = [for key, secret in local.env_secrets : secret.secret_id]

  env_secret_member = local.service_account != null ? "serviceAccount:${local.service_account}" : (
    length(data.google_project.env_secrets) > 0
    ? "serviceAccount:${data.google_project.env_secrets[0].number}-compute@developer.gserviceaccount.com"
    : null
  )
}

# The project number names the Compute Engine default service account, the
# identity Cloud Run runs instances as when the pool names none.
data "google_project" "env_secrets" {
  count      = length(local.env_secrets) > 0 && local.service_account == null ? 1 : 0
  project_id = local.project_id
}

resource "google_project_service" "secretmanager_api" {
  count   = length(local.env_secrets) > 0 ? 1 : 0
  project = local.project_id
  service = "secretmanager.googleapis.com"

  disable_on_destroy = false
}

resource "google_secret_manager_secret" "env" {
  for_each = local.env_secrets

  project   = local.project_id
  secret_id = each.value.secret_id
  labels    = local.final_labels

  replication {
    user_managed {
      replicas {
        location = var.spec.region
      }
    }
  }

  lifecycle {
    precondition {
      condition     = length(distinct(local.env_secret_ids)) == length(local.env_secret_ids)
      error_message = "Two variables in one container would be stored as the same Secret Manager secret (a '.' in a name becomes '-'): ${join(", ", local.env_secret_ids)} -- rename one of them."
    }
    precondition {
      condition     = length(each.value.secret_id) <= 255
      error_message = "The Secret Manager id ${each.value.secret_id} is over the 255 characters Secret Manager allows -- shorten the variable or container name."
    }
  }

  depends_on = [google_project_service.secretmanager_api]
}

resource "google_secret_manager_secret_version" "env" {
  for_each = local.env_secrets

  secret      = google_secret_manager_secret.env[each.key].id
  secret_data = each.value.value

  # A new value creates its version before the old one goes, so running
  # instances keep a readable version until the new revision exists.
  lifecycle {
    create_before_destroy = true
  }
}

resource "google_secret_manager_secret_iam_member" "env" {
  for_each = local.env_secrets

  project   = google_secret_manager_secret.env[each.key].project
  secret_id = google_secret_manager_secret.env[each.key].secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = local.env_secret_member
}
