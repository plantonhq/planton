# The app registration: the resource of substance, and the reason this
# module needs the beta provider -- Google publishes google_firebase_web_app
# only there (admitted in pkg/providerparity/admissions/google-beta.yaml).
# A web app has no identity beyond its display name, so nothing forces
# replacement; deletion_policy DELETE posts :remove with immediate=true,
# removing the app PERMANENTLY at once (Firebase's 30-day recoverable window
# is skipped). The Firebase Management API itself is enabled by the
# GcpFirebaseProject the app lives in.
resource "google_firebase_web_app" "this" {
  provider = google-beta
  project  = local.project_id

  display_name = var.spec.display_name

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

# reCAPTCHA v3 attestation: a per-app singleton Google never deletes (the
# provider only forgets it on destroy), so it carries no deletion_policy.
# The site secret is a secret Google never returns -- the provider marks it
# Sensitive and reports only site_secret_set on read. token_ttl is
# Optional+Computed: null lets Google assume its 1-hour default.
resource "google_firebase_app_check_recaptcha_v3_config" "this" {
  count = local.wants_recaptcha_v3 ? 1 : 0

  project = local.project_id
  app_id  = google_firebase_web_app.this.app_id

  site_secret = var.spec.app_check.recaptcha_v3.site_secret

  token_ttl = var.spec.app_check.recaptcha_v3.token_ttl != "" ? var.spec.app_check.recaptcha_v3.token_ttl : null

  depends_on = [google_firebase_web_app.this, google_project_service.firebaseappcheck_api]
}

# reCAPTCHA Enterprise attestation: the same singleton grain. The site key
# is the PUBLIC half of the reCAPTCHA key -- the value the page embeds.
resource "google_firebase_app_check_recaptcha_enterprise_config" "this" {
  count = local.wants_recaptcha_enterprise ? 1 : 0

  project = local.project_id
  app_id  = google_firebase_web_app.this.app_id

  site_key = var.spec.app_check.recaptcha_enterprise.site_key

  token_ttl = var.spec.app_check.recaptcha_enterprise.token_ttl != "" ? var.spec.app_check.recaptcha_enterprise.token_ttl : null

  depends_on = [google_firebase_web_app.this, google_project_service.firebaseappcheck_api]
}

# Debug tokens, keyed by display name (unique within the app by validation)
# so plans stay stable as list order changes. The token value is a secret:
# the provider marks it Sensitive, so it never prints in plans or state
# listings. The spec's deletion_policy governs each token.
resource "google_firebase_app_check_debug_token" "this" {
  for_each = local.debug_tokens

  project      = local.project_id
  app_id       = google_firebase_web_app.this.app_id
  display_name = each.value.display_name
  token        = each.value.token

  deletion_policy = local.deletion_policy

  depends_on = [google_firebase_web_app.this, google_project_service.firebaseappcheck_api]
}

# The app's firebaseConfig. The web lookup names its input web_app_id (the
# one naming divergence in the family). depends_on the registration so the
# read is DEFERRED to apply: a data source whose inputs are all known would
# be read at plan time, which needs credentials and would break the
# catalog's credential-free offline plans. After apply, the steady-state
# re-plan reads it with credentials and sees no diff. Every value is a
# client identifier that ships in the page; the conditionally present ones
# (RTDB, default bucket, finalized location, linked Analytics) degrade to
# "" in outputs.tf.
data "google_firebase_web_app_config" "this" {
  provider   = google-beta
  project    = local.project_id
  web_app_id = google_firebase_web_app.this.app_id

  depends_on = [google_firebase_web_app.this]
}
