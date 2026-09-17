# API enablement is module plumbing. firebase.googleapis.com is what
# :addFirebase talks to; fcm.googleapis.com is what a control plane sends
# push through -- declared explicitly, never assumed from :addFirebase's side
# effects. disable_on_destroy is false: destroying this resource must never
# switch off APIs other resources in the project depend on.
resource "google_project_service" "firebase_api" {
  project = local.project_id
  service = "firebase.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

resource "google_project_service" "fcm_api" {
  project = local.project_id
  service = "fcm.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# Firebase enablement: a ONE-WAY project singleton, and the reason this
# module needs the beta provider -- Google publishes google_firebase_project
# only there (admitted in pkg/providerparity/admissions/google-beta.yaml).
# The provider's create GETs the project first and ADOPTS an already-enabled
# project (no error, no second :addFirebase); its delete drops the resource
# from state and leaves the project enabled -- Google offers no way to
# remove Firebase from a project. That is why this resource carries no
# deletion_policy: the spec's deletion_policy governs only the composed
# resources below.
resource "google_firebase_project" "this" {
  provider = google-beta
  project  = local.project_id

  depends_on = [google_project_service.firebase_api, google_project_service.fcm_api]
}

# The default Cloud Storage for Firebase bucket, when the spec asks for one.
# Created at most once per project (the API path is
# projects/{project}/defaultBucket) and needs the pay-as-you-go plan. Beta
# provider, admitted alongside the enablement. The storage API is enabled
# only when the bucket is wanted.
resource "google_project_service" "firebasestorage_api" {
  count = local.wants_default_bucket ? 1 : 0

  project = local.project_id
  service = "firebasestorage.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

resource "google_firebase_storage_default_bucket" "this" {
  count    = local.wants_default_bucket ? 1 : 0
  provider = google-beta

  project  = local.project_id
  location = var.spec.default_storage_location

  deletion_policy = local.deletion_policy

  depends_on = [google_firebase_project.this, google_project_service.firebasestorage_api]
}

# App Check: per-service enforcement and per-resource overrides, on the GA
# provider. The App Check API is enabled only when the spec composes at
# least one configuration. An empty enforcement_mode is OFF -- Google's
# unset state -- and is sent as null so the API records exactly that.
resource "google_project_service" "firebaseappcheck_api" {
  count = local.wants_app_check ? 1 : 0

  project = local.project_id
  service = "firebaseappcheck.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

resource "google_firebase_app_check_service_config" "this" {
  for_each = local.app_check_service_configs

  project    = local.project_id
  service_id = each.value.service_id

  enforcement_mode = each.value.enforcement_mode != "" ? each.value.enforcement_mode : null
  deletion_policy  = local.deletion_policy

  depends_on = [google_firebase_project.this, google_project_service.firebaseappcheck_api]
}

resource "google_firebase_app_check_resource_policy" "this" {
  for_each = local.app_check_resource_policies

  project         = local.project_id
  service_id      = each.value.service_id
  target_resource = each.value.target_resource

  enforcement_mode = each.value.enforcement_mode != "" ? each.value.enforcement_mode : null
  deletion_policy  = local.deletion_policy

  depends_on = [google_firebase_project.this, google_project_service.firebaseappcheck_api]
}

# The Admin SDK configuration -- the project-level values the Firebase SDKs
# are initialised with. depends_on the enablement so the read is DEFERRED to
# apply: a data source whose inputs are all known would be read at plan
# time, which needs credentials and would break the catalog's
# credential-free offline plans. After apply, the steady-state re-plan reads
# it with credentials and sees no diff. Every value is conditionally present
# on Google's side (RTDB, default bucket, finalized location), so each
# output degrades to "" rather than failing.
data "google_firebase_admin_sdk_config" "this" {
  provider = google-beta
  project  = google_firebase_project.this.project

  depends_on = [google_firebase_project.this]
}
