# Secret values the deployment carries (spec.deployment_spec.secret_env[].value)
# are kept in Secret Manager, one secret per variable, so the agent
# references a secret it owns and never holds the value. The naming,
# placement, and grant rules match the Pulumi module's shared envsecrets
# helper:
#   - id "agentengine_<location>_<metadata.name>_<variable>" ('.' becomes
#     '-'); the Planton name stands in for the agent's own id, which Google
#     assigns at create;
#   - replicated only in the agent's location;
#   - secretAccessor on that secret alone for the agent's identity (the
#     spec's service_account, else the project's Vertex AI Reasoning Engine
#     service agent). AGENT_IDENTITY is refused with stored values by the
#     spec, because that identity exists only after the create that reads
#     the secrets.
# The variable reads the exact version, so a new value redeploys the agent.

locals {
  env_secrets = {
    for env in try(var.spec.spec.deployment_spec.secret_env, []) : env.name => {
      secret_id = join("_", ["agentengine", var.spec.location, var.metadata.name, replace(env.name, ".", "-")])
      value     = env.value
    } if env.value != null
  }

  env_secret_ids = [for key, secret in local.env_secrets : secret.secret_id]

  env_secret_service_account = try(var.spec.spec.service_account, "") != "" ? var.spec.spec.service_account : null

  env_secret_member = local.env_secret_service_account != null ? "serviceAccount:${local.env_secret_service_account}" : (
    length(data.google_project.env_secrets) > 0
    ? "serviceAccount:service-${data.google_project.env_secrets[0].number}@gcp-sa-aiplatform-re.iam.gserviceaccount.com"
    : null
  )
}

# The project number names the Reasoning Engine service agent, the identity
# the agent runs as when the spec names no service account.
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
        location = var.spec.location
      }
    }
  }

  lifecycle {
    precondition {
      condition     = length(distinct(local.env_secret_ids)) == length(local.env_secret_ids)
      error_message = "Two secret_env variables would be stored as the same Secret Manager secret (a '.' in a name becomes '-'): ${join(", ", local.env_secret_ids)} -- rename one of them."
    }
    precondition {
      condition     = length(each.value.secret_id) <= 255
      error_message = "The Secret Manager id ${each.value.secret_id} is over the 255 characters Secret Manager allows -- shorten the variable or resource name."
    }
  }

  depends_on = [google_project_service.secretmanager_api]
}

resource "google_secret_manager_secret_version" "env" {
  for_each = local.env_secrets

  secret      = google_secret_manager_secret.env[each.key].id
  secret_data = each.value.value

  # A new value creates its version before the old one goes, so the running
  # agent keeps a readable version until its update lands.
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
