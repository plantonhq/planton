# Enable the Certificate Manager API so a fresh project can host issuance
# configs. disable_on_destroy is false: tearing down one config must never
# disable the API for everything else in the project.
resource "google_project_service" "certificatemanager_api" {
  project = local.project_id
  service = "certificatemanager.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The Certificate Manager service agent
# (service-<project_number>@gcp-sa-certificatemanager.iam.gserviceaccount.com)
# requests every certificate from the pool. Google creates it only on request
# (services.generateServiceIdentity), never when the API is turned on, and a
# grant naming it before then is refused as an invalid member. The resource
# is published only in the beta provider (admitted in
# pkg/providerparity/admissions/google-beta.yaml); creating an agent that
# already exists returns it, and destroy is a no-op, so the agent outlives
# this config and serves every other config in the project.
resource "google_project_service_identity" "certificatemanager_agent" {
  provider = google-beta

  project = local.project_id
  service = "certificatemanager.googleapis.com"

  depends_on = [google_project_service.certificatemanager_api]
}

# Google checks at create that the agent holds
# roles/privateca.certificateRequester on the pool, and refuses the config
# without it ("Config validation failed"). The grant lands in the pool's own
# project and location, parsed from its full name, so a pool in another
# project works. It is a member grant: two configs over one pool write the
# same binding, and destroying either removes it until the other is applied
# again.
resource "google_privateca_ca_pool_iam_member" "certificatemanager_agent_requests_certificates" {
  # The pool's full name: the provider's IAM parser accepts the full name or
  # {project}/{location}/{pool}, never the bare pool id.
  ca_pool = var.spec.ca_pool
  role    = "roles/privateca.certificateRequester"
  member  = google_project_service_identity.certificatemanager_agent.member
}

# One certificate issuance config: how Google-managed certificates that name
# it are issued from a Certificate Authority Service pool. Every argument
# except labels and deletion_policy forces replacement.
resource "google_certificate_manager_certificate_issuance_config" "this" {
  name                       = local.issuance_config_name
  project                    = local.project_id
  location                   = local.location
  description                = var.spec.description != "" ? var.spec.description : null
  labels                     = local.final_labels
  key_algorithm              = var.spec.key_algorithm
  lifetime                   = var.spec.lifetime
  rotation_window_percentage = var.spec.rotation_window_percentage

  # Empty defers to the provider default (DELETE).
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # The spec lifts the provider's two single-field wrappers into ca_pool.
  certificate_authority_config {
    certificate_authority_service_config {
      ca_pool = var.spec.ca_pool
    }
  }

  depends_on = [
    google_project_service.certificatemanager_api,
    google_privateca_ca_pool_iam_member.certificatemanager_agent_requests_certificates,
  ]
}
