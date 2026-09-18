# Enable the APIs a Gen 2 function deploy exercises, so a fresh project
# works first try: Cloud Functions (the control plane), Cloud Build (turns
# source into a container), Cloud Run (serves the function), Artifact
# Registry (stores the built image), and Eventarc (delivers event
# triggers). disable_on_destroy=false: turning an API off on teardown is a
# project-wide blast radius no single resource should own.
resource "google_project_service" "apis" {
  for_each = toset([
    "cloudfunctions.googleapis.com",
    "cloudbuild.googleapis.com",
    "run.googleapis.com",
    "artifactregistry.googleapis.com",
    "eventarc.googleapis.com",
  ])

  project = local.project_id
  service = each.value

  disable_dependent_services = false
  disable_on_destroy         = false
}

# The Cloud Functions (Gen 2) function. Cloud Build containerizes the
# source with buildpacks; Cloud Run serves it. The build_config owns how
# source becomes a container, the service_config owns the serving shape,
# and the event_trigger (when present) wires a CloudEvent source through
# Eventarc.
resource "google_cloudfunctions2_function" "function" {
  name        = local.function_name
  project     = local.project_id
  location    = var.spec.region
  description = var.spec.description != "" ? var.spec.description : null
  labels      = local.final_labels

  # CMEK: encrypts the built container image and source artifacts. Requires
  # a customer-managed docker_repository and encrypter/decrypter grants for
  # the Cloud Functions + Artifact Registry service agents.
  kms_key_name = var.spec.kms_key_name != "" ? var.spec.kms_key_name : null

  # DELETE (provider default) removes the function on destroy; PREVENT
  # fails the destroy; ABANDON leaves it serving and consuming events.
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  build_config {
    runtime     = var.spec.build_config.runtime
    entry_point = var.spec.build_config.entry_point

    source {
      # Exactly one source arm — enforced pre-deploy by the spec's CEL rule.
      dynamic "storage_source" {
        for_each = var.spec.build_config.source.storage_source != null ? [var.spec.build_config.source.storage_source] : []
        content {
          bucket     = storage_source.value.bucket
          object     = storage_source.value.object
          generation = storage_source.value.generation
        }
      }

      dynamic "repo_source" {
        for_each = var.spec.build_config.source.repo_source != null ? [var.spec.build_config.source.repo_source] : []
        content {
          repo_name    = repo_source.value.repo_name
          branch_name  = repo_source.value.branch_name != "" ? repo_source.value.branch_name : null
          tag_name     = repo_source.value.tag_name != "" ? repo_source.value.tag_name : null
          commit_sha   = repo_source.value.commit_sha != "" ? repo_source.value.commit_sha : null
          dir          = repo_source.value.dir != "" ? repo_source.value.dir : null
          invert_regex = repo_source.value.invert_regex
          project_id   = repo_source.value.project_id != "" ? repo_source.value.project_id : null
        }
      }
    }

    environment_variables = length(var.spec.build_config.build_environment_variables) > 0 ? var.spec.build_config.build_environment_variables : null

    # Build identity: the fully-qualified service account resource name
    # (projects/*/serviceAccounts/*) Cloud Build runs as.
    service_account   = var.spec.build_config.service_account != "" ? var.spec.build_config.service_account : null
    worker_pool       = var.spec.build_config.worker_pool != "" ? var.spec.build_config.worker_pool : null
    docker_repository = var.spec.build_config.docker_repository != "" ? var.spec.build_config.docker_repository : null

    # Runtime base-image patching: AUTOMATIC is the proto zero value AND
    # the API default, so it sends nothing (indistinguishable from unset —
    # and the API behaves identically either way). Only the non-default
    # ON_DEPLOY choice sends a block; the Pulumi module does the same.
    dynamic "on_deploy_update_policy" {
      for_each = local.build_update_policy == "ON_DEPLOY" ? [1] : []
      content {}
    }
  }

  dynamic "service_config" {
    for_each = local.service_config != null ? [local.service_config] : []
    content {
      available_memory = local.available_memory
      available_cpu    = local.available_cpu
      timeout_seconds  = service_config.value.timeout_seconds

      # GCP defaults concurrency to 1 (every request its own instance);
      # values above 1 require at least 1 CPU.
      # Zero means unset for these counts (the spec fields have no presence):
      # the API rejects a concurrency of 0 and fills its own default when the
      # argument is omitted.
      max_instance_request_concurrency = service_config.value.max_instance_request_concurrency != 0 ? service_config.value.max_instance_request_concurrency : null

      min_instance_count = try(service_config.value.scaling.min_instance_count, 0) != 0 ? service_config.value.scaling.min_instance_count : null
      max_instance_count = try(service_config.value.scaling.max_instance_count, 0) != 0 ? service_config.value.scaling.max_instance_count : null

      # Runtime identity: bare service-account email.
      service_account_email = service_config.value.service_account_email != "" ? service_config.value.service_account_email : null

      environment_variables = length(service_config.value.environment_variables) > 0 ? service_config.value.environment_variables : null

      # Secret Manager references resolved at instance start — material
      # never appears in configuration or state.
      dynamic "secret_environment_variables" {
        for_each = service_config.value.secret_environment_variables
        content {
          key     = secret_environment_variables.value.key
          secret  = secret_environment_variables.value.secret
          version = secret_environment_variables.value.version != "" ? secret_environment_variables.value.version : "latest"
          # The API requires an explicit project on every entry; default to
          # the function's own (or ambient) project when unset.
          project_id = secret_environment_variables.value.project_id != "" ? secret_environment_variables.value.project_id : local.effective_secret_project
        }
      }

      dynamic "secret_volumes" {
        for_each = service_config.value.secret_volumes
        content {
          mount_path = secret_volumes.value.mount_path
          secret     = secret_volumes.value.secret
          project_id = secret_volumes.value.project_id != "" ? secret_volumes.value.project_id : local.effective_secret_project

          dynamic "versions" {
            for_each = secret_volumes.value.versions
            content {
              version = versions.value.version
              path    = versions.value.path
            }
          }
        }
      }

      vpc_connector                 = local.vpc_connector
      vpc_connector_egress_settings = local.vpc_egress
      ingress_settings              = local.ingress_settings

      # Direct VPC egress: the connectorless path (mutually exclusive
      # with vpc_connector — the spec's CEL enforces it pre-deploy).
      dynamic "direct_vpc_network_interface" {
        for_each = local.direct_vpc != null ? [local.direct_vpc] : []
        content {
          network    = direct_vpc_network_interface.value.network != "" ? direct_vpc_network_interface.value.network : null
          subnetwork = direct_vpc_network_interface.value.subnetwork != "" ? direct_vpc_network_interface.value.subnetwork : null
          tags       = length(direct_vpc_network_interface.value.tags) > 0 ? direct_vpc_network_interface.value.tags : null
        }
      }

      # Sent explicitly whenever the interface is present (see locals.tf).
      direct_vpc_egress = local.direct_vpc_egress

      # true (the API default) sends 100% of traffic to the latest ready
      # revision; false holds traffic for manual canary/rollback on the
      # underlying Cloud Run service.
      all_traffic_on_latest_revision = service_config.value.all_traffic_on_latest_revision

      binary_authorization_policy = service_config.value.binary_authorization_policy != "" ? service_config.value.binary_authorization_policy : null
    }
  }

  dynamic "event_trigger" {
    for_each = !local.is_http_trigger && var.spec.trigger != null && var.spec.trigger.event_trigger != null ? [var.spec.trigger.event_trigger] : []
    content {
      event_type   = event_trigger.value.event_type
      pubsub_topic = event_trigger.value.pubsub_topic != "" ? event_trigger.value.pubsub_topic : null

      # If unset, GCP uses the function's region; multi-region sources
      # (Storage multi-region buckets) use "us"/"eu".
      trigger_region = event_trigger.value.trigger_region != "" ? event_trigger.value.trigger_region : null

      retry_policy          = event_trigger.value.retry_policy != "" ? event_trigger.value.retry_policy : null
      service_account_email = event_trigger.value.service_account_email != "" ? event_trigger.value.service_account_email : null

      dynamic "event_filters" {
        for_each = event_trigger.value.event_filters
        content {
          attribute = event_filters.value.attribute
          value     = event_filters.value.value
          operator  = event_filters.value.operator != "" ? event_filters.value.operator : null
        }
      }
    }
  }

  depends_on = [google_project_service.apis]
}

# Resolves the ambient project for secret references that omit their own —
# the one case that needs a live lookup (secret entries need an explicit
# project id in the API payload). Count-gated so every plan that names its
# project (or carries no ambient-project secrets) runs credential-free;
# the Pulumi module gates its client-config lookup identically.
data "google_project" "function" {
  count = local.needs_project_lookup ? 1 : 0
}

# Public invocation for HTTP functions: Gen 2 functions are served by Cloud
# Run, so "allow unauthenticated" is run.invoker for allUsers on the
# UNDERLYING Cloud Run service (which shares the function's name), not a
# Cloud Functions IAM binding.
resource "google_cloud_run_service_iam_member" "public_invoker" {
  count = local.is_http_trigger && try(local.service_config.allow_unauthenticated, false) ? 1 : 0

  project  = local.project_id
  location = var.spec.region
  service  = google_cloudfunctions2_function.function.name

  role   = "roles/run.invoker"
  member = "allUsers"
}
