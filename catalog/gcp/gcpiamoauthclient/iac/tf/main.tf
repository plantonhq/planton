# Enable the IAM API, which serves the workforce OAuth surface.
# disable_on_destroy is false: tearing down one client must never disable
# IAM for everything else in the project.
resource "google_project_service" "iam_api" {
  project = local.project_id
  service = "iam.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# A WORKFORCE Identity Federation OAuth client — the only kind of OAuth
# client Google's APIs can create programmatically (consent-screen clients
# remain a console step; see the kind's README).
resource "google_iam_oauth_client" "this" {
  project = local.project_id

  oauth_client_id = local.oauth_client_id
  location        = local.location

  allowed_grant_types   = var.spec.allowed_grant_types
  allowed_scopes        = var.spec.allowed_scopes
  allowed_redirect_uris = var.spec.allowed_redirect_uris

  client_type  = var.spec.client_type != "" ? var.spec.client_type : null
  display_name = var.spec.display_name != "" ? var.spec.display_name : null
  description  = var.spec.description != "" ? var.spec.description : null
  disabled     = var.spec.disabled ? true : null

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  depends_on = [google_project_service.iam_api]
}

# Managed client secrets, one per spec.credentials entry; GCP generates
# each secret server-side. `disabled` is sent EXPLICITLY: GCP requires a
# credential to be DISABLED before it can be deleted, so a spec transition
# false -> true is exactly the pre-removal step and must reach the API.
#
# The kind-level deletion_policy fans out to every credential — the
# credentials have no life apart from the client, so one switch governs
# both objects.
resource "google_iam_oauth_client_credential" "this" {
  for_each = local.credentials

  project = local.project_id

  oauthclient                = google_iam_oauth_client.this.oauth_client_id
  location                   = local.location
  oauth_client_credential_id = each.value.credential_id

  display_name = each.value.display_name != "" ? each.value.display_name : null
  disabled     = each.value.disabled

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
