# The app registration: the resource of substance, and the reason this
# module needs the beta provider -- Google publishes google_firebase_apple_app
# only there (admitted in pkg/providerparity/admissions/google-beta.yaml).
# bundle_id is the app's identity in Firebase and forces replacement;
# deletion_policy DELETE posts :remove with immediate=true, removing the app
# PERMANENTLY at once (Firebase's 30-day recoverable window is skipped).
# The Firebase Management API itself is enabled by the GcpFirebaseProject
# the app lives in. What this module does NOT do, by Google's design: upload
# the APNs authentication key push needs on Apple platforms -- Firebase
# exposes no API for it; it is a console step.
resource "google_firebase_apple_app" "this" {
  provider = google-beta
  project  = local.project_id

  display_name = var.spec.display_name
  bundle_id    = var.spec.bundle_id

  app_store_id = local.app_store_id
  team_id      = local.team_id

  api_key_id = local.api_key_id

  deletion_policy = local.deletion_policy
}

# App Check, on the GA provider. The App Check API is enabled only when the
# spec composes at least one App Check resource, and never disabled on
# destroy -- the project's other apps may depend on it. Every App Check
# resource waits on the registration AND the API: the configurations are
# addressed by the app id the registration returns, and the App Check service
# must be reachable.
resource "google_project_service" "firebaseappcheck_api" {
  count = local.wants_app_check ? 1 : 0

  project = local.project_id
  service = "firebaseappcheck.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# App Attest attestation: a per-app singleton Google never deletes (the
# provider only forgets it on destroy), so it carries no deletion_policy.
# token_ttl is Optional+Computed: null lets Google assume its 1-hour default
# and read it back.
resource "google_firebase_app_check_app_attest_config" "this" {
  count = local.wants_app_attest ? 1 : 0

  project = local.project_id
  app_id  = google_firebase_apple_app.this.app_id

  token_ttl = var.spec.app_check.app_attest.token_ttl != "" ? var.spec.app_check.app_attest.token_ttl : null

  depends_on = [google_firebase_apple_app.this, google_project_service.firebaseappcheck_api]
}

# DeviceCheck attestation: the same singleton grain. The private key is a
# secret Google never returns -- the provider marks it Sensitive and reports
# only private_key_set on read.
resource "google_firebase_app_check_device_check_config" "this" {
  count = local.wants_device_check ? 1 : 0

  project = local.project_id
  app_id  = google_firebase_apple_app.this.app_id

  key_id      = var.spec.app_check.device_check.key_id
  private_key = var.spec.app_check.device_check.private_key

  token_ttl = var.spec.app_check.device_check.token_ttl != "" ? var.spec.app_check.device_check.token_ttl : null

  depends_on = [google_firebase_apple_app.this, google_project_service.firebaseappcheck_api]
}

# Debug tokens, keyed by display name (unique within the app by validation)
# so plans stay stable as list order changes. The token value is a secret:
# the provider marks it Sensitive, so it never prints in plans or state
# listings. The spec's deletion_policy governs each token.
resource "google_firebase_app_check_debug_token" "this" {
  for_each = local.debug_tokens

  project      = local.project_id
  app_id       = google_firebase_apple_app.this.app_id
  display_name = each.value.display_name
  token        = each.value.token

  deletion_policy = local.deletion_policy

  depends_on = [google_firebase_apple_app.this, google_project_service.firebaseappcheck_api]
}

# The app's configuration file (GoogleService-Info.plist). depends_on the
# registration so the read is DEFERRED to apply: a data source whose inputs
# are all known would be read at plan time, which needs credentials and
# would break the catalog's credential-free offline plans. After apply, the
# steady-state re-plan reads it with credentials and sees no diff. A build
# input that ships in the app bundle, not a secret.
data "google_firebase_apple_app_config" "this" {
  provider = google-beta
  project  = local.project_id
  app_id   = google_firebase_apple_app.this.app_id

  depends_on = [google_firebase_apple_app.this]
}
