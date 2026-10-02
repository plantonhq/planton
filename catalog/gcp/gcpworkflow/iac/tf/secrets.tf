# Secret values the workflow reads (spec.secret_env_vars) are kept in Secret
# Manager, one secret per variable. Cloud Workflows has no secret field, so
# the workflow receives each version's resource name as the variable and
# reads the value through the Secret Manager connector at run time; the
# value itself never reaches the workflow. The naming, placement, and grant
# rules match the Pulumi module's shared envsecrets helper:
#   - id "workflow_<region>_<workflow name>_<key>";
#   - replicated only in the workflow's region (the spec requires region
#     with secret_env_vars);
#   - secretAccessor on that secret alone for the runtime identity (the
#     spec's service_account by email, else the project's Compute Engine
#     default service account).
# A new value adds a version, which changes the variable and deploys a new
# revision.

locals {
  env_secrets = {
    for name, value in var.spec.secret_env_vars : name => {
      secret_id = join("_", ["workflow", var.spec.region, local.workflow_name, name])
      value     = value
    }
  }

  # The grant names the account's email; service_account may also carry the
  # projects/{project}/serviceAccounts/ prefix.
  env_secret_service_account = var.spec.service_account != "" ? replace(var.spec.service_account, "/^projects/[^/]+/serviceAccounts//", "") : null

  env_secret_member = local.env_secret_service_account != null ? "serviceAccount:${local.env_secret_service_account}" : (
    length(data.google_project.env_secrets) > 0
    ? "serviceAccount:${data.google_project.env_secrets[0].number}-compute@developer.gserviceaccount.com"
    : null
  )

  # What the workflow receives as its environment: the literals, plus each
  # stored secret's version resource name under its variable.
  user_env_vars = merge(
    var.spec.user_env_vars,
    { for name, secret in local.env_secrets : name => google_secret_manager_secret_version.env[name].name },
  )
}

# The project number names the Compute Engine default service account, the
# identity the workflow runs as when the spec names none.
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
      error_message = "The Secret Manager id ${each.value.secret_id} is over the 255 characters Secret Manager allows -- shorten the variable or workflow name."
    }
  }

  depends_on = [google_project_service.secretmanager_api]
}

resource "google_secret_manager_secret_version" "env" {
  for_each = local.env_secrets

  secret      = google_secret_manager_secret.env[each.key].id
  secret_data = each.value.value

  # A new value creates its version before the old one goes, so executions
  # on the running revision keep a readable version until the new revision
  # is deployed.
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
