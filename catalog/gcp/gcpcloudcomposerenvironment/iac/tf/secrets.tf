# Secret values Airflow reads (spec.software_config.secret_env_variables) are
# kept in Secret Manager, one secret per variable. Composer has no secret
# field for environment variables, so Airflow receives each version's
# resource name as the variable and DAG code reads the value with the Secret
# Manager client; the value itself never reaches the environment. The
# naming, placement, and grant rules match the Pulumi module's shared
# envsecrets helper:
#   - id "composer_<region>_<environment name>_<key>";
#   - replicated only in the environment's region;
#   - secretAccessor on that secret alone for the node service account (the
#     spec's node_config.service_account by email, else the project's
#     Compute Engine default service account a Composer 2 environment falls
#     back to).
# A new value adds a version, which changes the variable and runs an
# environment update.

locals {
  env_secrets = {
    for name, value in try(var.spec.software_config.secret_env_variables, {}) : name => {
      secret_id = join("_", ["composer", var.spec.region, local.environment_name, name])
      value     = value
    }
  }

  # The grant names the account's email; the node service account may also
  # carry the projects/{project}/serviceAccounts/ prefix.
  env_secret_service_account = try(var.spec.node_config.service_account, "") != "" ? replace(var.spec.node_config.service_account, "/^projects/[^/]+/serviceAccounts//", "") : null

  env_secret_member = local.env_secret_service_account != null ? "serviceAccount:${local.env_secret_service_account}" : (
    length(data.google_project.env_secrets) > 0
    ? "serviceAccount:${data.google_project.env_secrets[0].number}-compute@developer.gserviceaccount.com"
    : null
  )

  # What Airflow receives as its environment: the literals, plus each stored
  # secret's version resource name under its variable.
  env_variables = merge(
    try(var.spec.software_config.env_variables, {}),
    { for name, secret in local.env_secrets : name => google_secret_manager_secret_version.env[name].name },
  )
}

# The project number names the Compute Engine default service account, the
# identity a Composer 2 environment's nodes run as when the spec names none.
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
      error_message = "The Secret Manager id ${each.value.secret_id} is over the 255 characters Secret Manager allows -- shorten the variable or environment name."
    }
  }

  depends_on = [google_project_service.secretmanager_api]
}

resource "google_secret_manager_secret_version" "env" {
  for_each = local.env_secrets

  secret      = google_secret_manager_secret.env[each.key].id
  secret_data = each.value.value

  # A new value creates its version before the old one goes, so running
  # tasks keep a readable version until the environment update lands.
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
