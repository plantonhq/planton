# GCP Planton Runner Terraform Module
#
# Provisions a standing Planton runner appliance on Cloud Run: an
# always-on, outbound-only worker that executes deploy and cloud
# operations from inside your project's network perimeter -- the piece
# that makes private endpoints (most notably private GKE control planes)
# deployable and operable. The project, the optional VPC placement, and
# the optional runtime service account are referenced resources -- the
# module never creates or mutates them.
#
# ENROLLMENT IS TOKEN-FIRST: the service ships the runner TOKEN (via
# Secret Manager), never an identity. The runner joins the control plane
# on first boot, registers itself, and receives its own individually
# revocable identity; instance replacement re-joins with the same token
# (its lineage re-admits the runner it originally admitted).

# ── API enablement ──────────────────────────────────────────────────────
#
# Enable the Cloud Run Admin API first so a fresh project works on the
# first deploy. disable_on_destroy stays false: tearing down one runner
# must never disable the API for everything else in the project.
resource "google_project_service" "run" {
  project                    = local.project_id
  service                    = "run.googleapis.com"
  disable_dependent_services = true
  disable_on_destroy         = false
}

# ── Runtime identity ────────────────────────────────────────────────────
#
# The service account the runner runs as -- the seam keyless cloud access
# rides. Created permissionless when no service_account is referenced, so
# the identity seam always exists: permissions can be granted later
# without replacing the runner. Deliberately NEVER the project's Compute
# Engine default service account -- it typically carries broad project
# access the runner should not inherit silently.
resource "google_service_account" "runtime" {
  count = local.create_service_account ? 1 : 0

  project      = local.project_id
  account_id   = local.runner_name
  display_name = "Planton runner '${local.runner_name}'"
  description  = "Runtime identity for Planton runner '${local.runner_name}' -- grant it the roles keyless operations need"

  lifecycle {
    # FAIL LOUDLY past GCP's account_id budget: a silent truncation could
    # collide two runners' identities. Longer names keep working by
    # composing a GcpServiceAccount resource and referencing it. Twin:
    # the Pulumi module's runtimeServiceAccount guard.
    precondition {
      condition     = length(local.runner_name) <= 30
      error_message = "GCP caps service account ids at 30 characters -- reference your own service account via spec.service_account (or use a shorter name)."
    }
  }
}

# ── Token secret ────────────────────────────────────────────────────────
#
# The container never sees the token in its launch configuration: the
# service's env carries only a secret reference, and Cloud Run resolves
# the value at instance start through the runtime service account -- so
# reading the service definition (a common, low-sensitivity permission)
# reveals nothing.
#
# Automatic replication: the token is a single small value read at
# instance start; regional placement decisions belong to the service, not
# its bootstrap secret.
resource "google_secret_manager_secret" "token" {
  project   = local.project_id
  secret_id = local.token_secret_id
  labels    = local.gcp_labels

  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "token" {
  secret      = google_secret_manager_secret.token.id
  secret_data = var.spec.token
}

# The accessor grant lives ON the module-owned secret (never on the
# referenced service account) and names exactly one principal -- the
# least-privilege twin of the AWS execution role's one-secret inline
# policy. Without it the first instance start fails resolving the env
# reference, however broad the project's other grants are.
resource "google_secret_manager_secret_iam_member" "token_accessor" {
  project   = google_secret_manager_secret.token.project
  secret_id = google_secret_manager_secret.token.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${local.service_account_email}"
}

# ── Compute ─────────────────────────────────────────────────────────────
#
# The Cloud Run v2 service that keeps exactly one runner running.
resource "google_cloud_run_v2_service" "runner" {
  project  = local.project_id
  name     = local.runner_name
  location = var.spec.region
  labels   = local.gcp_labels

  # The appliance's managed teardown IS the destroy path, and the token
  # makes it re-mintable standing infrastructure -- the provider's default
  # deletion protection would turn every teardown into a two-step dance
  # for no protective gain.
  deletion_protection = false

  # Ingress setting is required by the API; nothing meaningful is
  # reachable either way -- the default authenticated-only posture (no
  # run.invoker grants exist) already refuses every caller.
  ingress = "INGRESS_TRAFFIC_ALL"

  template {
    # Exactly one runner per enrollment: a runner's identity is minted
    # for ONE live instance -- a second instance joining under the same
    # name would revoke the first's key (token lineage: re-admission
    # re-mints and revokes). Never enable scaling here without
    # redesigning enrollment for fleets. (Revision rollover briefly
    # overlaps two instances; the draining one's revoked key dies with
    # it, and a rollback self-heals by re-joining.)
    scaling {
      min_instance_count = 1
      max_instance_count = 1
    }

    service_account = local.service_account_email

    # GEN2 for full Linux compatibility -- IaC toolchains the runner
    # executes (tofu/pulumi child processes) assume a complete syscall
    # surface.
    execution_environment = "EXECUTION_ENVIRONMENT_GEN2"

    # Direct VPC egress: only private-range traffic rides the VPC (the
    # route to private endpoints); the runner's control-plane dial-out
    # keeps its normal internet path -- so a misconfigured VPC can never
    # sever the runner from the control plane that manages it.
    dynamic "vpc_access" {
      for_each = try(var.spec.vpc_access, null) != null ? [var.spec.vpc_access] : []
      content {
        egress = "PRIVATE_RANGES_ONLY"
        network_interfaces {
          network    = vpc_access.value.network
          subnetwork = vpc_access.value.subnetwork
          tags       = try(vpc_access.value.tags, [])
        }
      }
    }

    containers {
      image = "${var.spec.image_repository}:${var.spec.runner_version}"
      # args, NEVER command: Cloud Run's `command` REPLACES the image
      # entrypoint (the runner binary), so command=["start"] makes the
      # platform exec a binary literally named "start" -- the instance dies
      # with "Application exec likely failed" before one log line (proven
      # live). The image's entrypoint is the bare runner binary; the
      # subcommand rides args.
      args = ["start"]

      # h2c: the runner's server speaks plaintext HTTP/2 (gRPC) behind
      # Cloud Run's TLS edge.
      ports {
        name           = "h2c"
        container_port = 50051
      }

      # The runner's environment contract: the token via a Secret Manager
      # reference (resolved at instance start through the runtime service
      # account -- never plaintext here), the name it registers itself
      # under, and the control-plane endpoint when one is declared
      # (omitted, the runner's built-in hosted default applies). No
      # EXECUTION_MODE: only the control plane knows whether its work
      # queue is reachable, so the runner derives its mode from the
      # identity the join returns -- a mode knob here would silently strip
      # capability.
      env {
        name = "PLANTON_RUNNER_TOKEN"
        value_source {
          secret_key_ref {
            secret = google_secret_manager_secret.token.secret_id
            # "latest" deliberately: token rotation needs no service
            # update -- running instances keep serving on their minted
            # identity (the token is only read at join), and the next
            # instance start joins with the new value.
            version = "latest"
          }
        }
      }
      env {
        name  = "PLANTON_RUNNER_NAME"
        value = local.registration_name
      }
      dynamic "env" {
        for_each = try(var.spec.control_plane_endpoint, "") != "" ? [var.spec.control_plane_endpoint] : []
        content {
          name  = "PLANTON_RUNNER_ENDPOINT"
          value = env.value
        }
      }
      # NO explicit PORT env: Cloud Run reserves the name and rejects the
      # whole create with a 400 when a template sets it (proven live). The
      # platform injects PORT itself from the ports block's container_port
      # above, which is how the runner's server learns its port here.
      env {
        name  = "LOG_LEVEL"
        value = "info"
      }

      resources {
        limits = {
          cpu    = var.spec.cpu
          memory = var.spec.memory
        }
        # CPU stays allocated between requests: the runner is a PULL-based
        # worker -- it polls its work queue and executes long-running IaC
        # operations with no inbound request in flight, so
        # throttled-between-requests CPU (the Cloud Run default) would
        # freeze it mid-operation.
        cpu_idle = false
        # Faster cold start for the one instance replacement ever in
        # flight.
        startup_cpu_boost = true
      }

      # The health server answers independently of control-plane
      # reachability, so a runner whose control plane is momentarily
      # unreachable still starts -- its readiness contract is the work
      # queue, not the probe.
      startup_probe {
        initial_delay_seconds = 5
        period_seconds        = 10
        failure_threshold     = 3
        timeout_seconds       = 5
        http_get {
          path = "/healthz"
          port = 8093
        }
      }
    }
  }

  # The first instance start resolves the token reference, so the version
  # and the accessor grant must exist before the service -- a missing edge
  # here fails at instance START, not at plan.
  depends_on = [
    google_project_service.run,
    google_secret_manager_secret_version.token,
    google_secret_manager_secret_iam_member.token_accessor,
  ]
}
