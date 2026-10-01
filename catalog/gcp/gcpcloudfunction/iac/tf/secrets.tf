# Secret values the service config carries
# (service_config.secret_environment_variables[].value) are kept in Secret
# Manager, one secret per variable, so the function references a secret it
# owns and never holds the value. The naming, placement, and grant rules
# match the Pulumi module's shared envsecrets helper:
#   - id "function_<region>_<function name>_<key>";
#   - replicated only in the function's region;
#   - secretAccessor on that secret alone for the runtime identity (the
#     spec's service_account_email, else the project's Compute Engine
#     default service account).
# The variable reads the exact version, so a new value redeploys the
# function.

locals {
  env_secrets = {
    for env in try(var.spec.service_config.secret_environment_variables, []) : env.key => {
      secret_id = join("_", ["function", var.spec.region, local.function_name, env.key])
      value     = env.value
    } if env.value != ""
  }

  env_secret_service_account = try(var.spec.service_config.service_account_email, "") != "" ? var.spec.service_config.service_account_email : null

  env_secret_member = local.env_secret_service_account != null ? "serviceAccount:${local.env_secret_service_account}" : (
    length(data.google_project.env_secrets) > 0
    ? "serviceAccount:${data.google_project.env_secrets[0].number}-compute@developer.gserviceaccount.com"
    : null
  )
}

# The project number names the Compute Engine default service account, the
# identity the function runs as when the spec names none.
data "google_project" "env_secrets" {
  count      = length(local.env_secrets) > 0 && local.env_secret_service_account == null ? 1 : 0
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
      condition     = length(each.value.secret_id) <= 255
      error_message = "The Secret Manager id ${each.value.secret_id} is over the 255 characters Secret Manager allows -- shorten the variable or function name."
    }
  }

  depends_on = [google_project_service.secretmanager_api]
}

resource "google_secret_manager_secret_version" "env" {
  for_each = local.env_secrets

  secret      = google_secret_manager_secret.env[each.key].id
  secret_data = each.value.value

  # A new value creates its version before the old one goes, so the running
  # function keeps a readable version until its update lands.
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
