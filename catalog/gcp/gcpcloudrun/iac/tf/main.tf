# Enable the Cloud Run Admin API before creating the service so a fresh
# project works first try. disable_on_destroy=false: turning an API off on
# teardown is a project-wide blast radius no single resource should own.
resource "google_project_service" "run_api" {
  project = local.project_id
  service = "run.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The Cloud Run service: a stable serving endpoint plus a revision template.
# Every apply that changes the template stamps out a new immutable revision;
# the traffic block controls how requests split across revisions.
resource "google_cloud_run_v2_service" "main" {
  name        = local.service_name
  location    = var.spec.region
  project     = local.project_id
  description = local.description

  # NAMES arriving from the spec's enums match the API values exactly;
  # null lets GCP apply its own default (ALL / GA).
  ingress      = local.ingress
  launch_stage = local.launch_stage

  # Turns the IAM run.routes.invoke check off entirely — the org-policy
  # alternative to granting allUsers (the IAM member below). The proto
  # rejects setting both.
  invoker_iam_disabled = var.spec.invoker_iam_disabled

  # Extra token audiences authenticated callers may use instead of the
  # run.app URL.
  custom_audiences = length(var.spec.custom_audiences) > 0 ? var.spec.custom_audiences : null

  labels = local.final_labels

  # Service-object annotations for external tools (system namespaces are
  # API-rejected). Revision-level annotations live in the template below.
  annotations = length(var.spec.annotations) > 0 ? var.spec.annotations : null

  # Identity-Aware Proxy in front of the service — Google's managed login
  # wall. Omitted when false so the provider default applies cleanly.
  iap_enabled = var.spec.iap_enabled ? true : null

  # Turns off the default *.run.app URL, leaving custom-domain / LB paths
  # as the only front doors.
  default_uri_disabled = var.spec.default_uri_disabled ? true : null

  # Deletion guard, honest by default: the spec defaults this to true, so a
  # destroy fails until the manifest explicitly opts out.
  deletion_protection = var.spec.deletion_protection

  # Terraform-side destroy stance: PREVENT fails destroys, ABANDON removes
  # the service from management without deleting it in GCP.
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # Deploy-from-source: Cloud Build produces the serving image (the Cloud
  # Run functions build path) instead of a prebuilt container image.
  dynamic "build_config" {
    for_each = var.spec.build_config != null ? [var.spec.build_config] : []
    content {
      source_location          = build_config.value.source_location != "" ? build_config.value.source_location : null
      function_target          = build_config.value.function_target != "" ? build_config.value.function_target : null
      image_uri                = build_config.value.image_uri != "" ? build_config.value.image_uri : null
      base_image               = build_config.value.base_image != "" ? build_config.value.base_image : null
      enable_automatic_updates = build_config.value.enable_automatic_updates ? true : null
      environment_variables    = length(build_config.value.environment_variables) > 0 ? build_config.value.environment_variables : null
      worker_pool              = build_config.value.worker_pool != "" ? build_config.value.worker_pool : null
      service_account          = build_config.value.service_account != "" ? build_config.value.service_account : null
    }
  }

  # Multi-region service: one identity serving from several regions. The
  # proto guarantees region is "global" whenever this block is present.
  dynamic "multi_region_settings" {
    for_each = var.spec.multi_region_settings != null ? [var.spec.multi_region_settings] : []
    content {
      regions = multi_region_settings.value.regions
    }
  }

  # Binary Authorization deploy gate: the project default policy XOR a
  # named platform policy (the proto rejects both).
  dynamic "binary_authorization" {
    for_each = var.spec.binary_authorization != null ? [var.spec.binary_authorization] : []
    content {
      use_default              = binary_authorization.value.use_default ? true : null
      policy                   = binary_authorization.value.policy != "" ? binary_authorization.value.policy : null
      breakglass_justification = binary_authorization.value.breakglass_justification != "" ? binary_authorization.value.breakglass_justification : null
    }
  }

  # Service-wide scaling posture (distinct from the per-revision bounds in
  # the template): MANUAL pins the total instance count across revisions.
  dynamic "scaling" {
    for_each = var.spec.service_scaling != null ? [var.spec.service_scaling] : []
    content {
      scaling_mode          = scaling.value.scaling_mode != "" ? scaling.value.scaling_mode : null
      manual_instance_count = scaling.value.manual_instance_count
      min_instance_count    = scaling.value.min_instance_count
      max_instance_count    = scaling.value.max_instance_count
    }
  }

  template {
    # Explicit revision naming makes declarative blue/green possible; null
    # (the norm) lets Cloud Run generate names.
    revision = local.revision

    # Revision-level metadata, stamped on every revision the template
    # creates (distinct from the service-object labels/annotations above).
    labels      = length(var.spec.revision_labels) > 0 ? var.spec.revision_labels : null
    annotations = length(var.spec.revision_annotations) > 0 ? var.spec.revision_annotations : null

    # The runtime identity whose permissions the code exercises. Null uses
    # the project's Compute Engine default service account.
    service_account = local.service_account

    execution_environment            = local.execution_environment
    max_instance_request_concurrency = var.spec.max_instance_request_concurrency
    timeout                          = local.timeout
    session_affinity                 = var.spec.session_affinity
    encryption_key                   = local.encryption_key
    gpu_zonal_redundancy_disabled    = var.spec.gpu_zonal_redundancy_disabled ? true : null

    # Escape hatch disabling ALL probes (startup and liveness) for
    # workloads whose serving model breaks the probe contract.
    health_check_disabled = var.spec.health_check_disabled ? true : null

    # Per-revision instance bounds; an omitted block scales 0..default cap.
    dynamic "scaling" {
      for_each = var.spec.scaling != null ? [var.spec.scaling] : []
      content {
        min_instance_count = scaling.value.min_instance_count
        max_instance_count = scaling.value.max_instance_count
      }
    }

    # GPU hardware requirement (e.g. "nvidia-l4").
    dynamic "node_selector" {
      for_each = var.spec.node_selector != null ? [var.spec.node_selector] : []
      content {
        accelerator = node_selector.value.accelerator
      }
    }

    # Outbound VPC networking: a Serverless VPC Access connector XOR direct
    # VPC egress network_interfaces — the proto guarantees exactly one.
    dynamic "vpc_access" {
      for_each = local.has_vpc_access ? [1] : []
      content {
        connector = local.vpc_connector
        egress    = local.vpc_egress

        dynamic "network_interfaces" {
          for_each = local.vpc_interfaces
          content {
            network    = network_interfaces.value.network != "" ? network_interfaces.value.network : null
            subnetwork = network_interfaces.value.subnetwork != "" ? network_interfaces.value.subnetwork : null
            tags       = length(network_interfaces.value.tags) > 0 ? network_interfaces.value.tags : null
          }
        }
      }
    }

    # Named volumes; each spec entry carries exactly one source arm.
    dynamic "volumes" {
      for_each = var.spec.volumes
      content {
        name = volumes.value.name

        # Cloud SQL Unix sockets — GCP manages the proxying; connect via
        # /cloudsql/<project:region:instance> under the mount path.
        dynamic "cloud_sql_instance" {
          for_each = volumes.value.cloud_sql_instance != null ? [volumes.value.cloud_sql_instance] : []
          content {
            instances = cloud_sql_instance.value.instances
          }
        }

        dynamic "secret" {
          for_each = volumes.value.secret != null ? [volumes.value.secret] : []
          content {
            secret       = secret.value.secret
            default_mode = secret.value.default_mode

            dynamic "items" {
              for_each = secret.value.items
              content {
                path    = items.value.path
                version = items.value.version != "" ? items.value.version : null
                mode    = items.value.mode
              }
            }
          }
        }

        dynamic "empty_dir" {
          for_each = volumes.value.empty_dir != null ? [volumes.value.empty_dir] : []
          content {
            medium     = empty_dir.value.medium != "" ? empty_dir.value.medium : null
            size_limit = empty_dir.value.size_limit != "" ? empty_dir.value.size_limit : null
          }
        }

        # GCS FUSE mounts require the GEN2 execution environment.
        dynamic "gcs" {
          for_each = volumes.value.gcs != null ? [volumes.value.gcs] : []
          content {
            bucket        = gcs.value.bucket
            read_only     = gcs.value.read_only
            mount_options = length(gcs.value.mount_options) > 0 ? gcs.value.mount_options : null
          }
        }

        dynamic "nfs" {
          for_each = volumes.value.nfs != null ? [volumes.value.nfs] : []
          content {
            server    = nfs.value.server
            path      = nfs.value.path
            read_only = nfs.value.read_only
          }
        }
      }
    }

    # The serving container plus any sidecars; containers share localhost
    # and volumes, and depends_on orders their startup.
    dynamic "containers" {
      for_each = var.spec.containers
      content {
        name        = containers.value.name != "" ? containers.value.name : null
        image       = containers.value.image
        command     = length(containers.value.command) > 0 ? containers.value.command : null
        args        = length(containers.value.args) > 0 ? containers.value.args : null
        working_dir = containers.value.working_dir != "" ? containers.value.working_dir : null
        depends_on  = length(containers.value.depends_on) > 0 ? containers.value.depends_on : null

        # Base image for automatic base-image updates on source deploys
        # (pairs with build_config.enable_automatic_updates).
        base_image_uri = containers.value.base_image_uri != "" ? containers.value.base_image_uri : null

        # Environment: a literal, a Secret Manager secret the author owns,
        # or a secret value secrets.tf stored (one of the three --
        # proto-enforced).
        dynamic "env" {
          for_each = containers.value.env
          content {
            name  = env.value.name
            value = env.value.value_from_secret == null && env.value.secret_value == "" ? env.value.value : null

            dynamic "value_source" {
              for_each = env.value.secret_value != "" ? [{
                secret  = google_secret_manager_secret.env["${containers.key}/${env.value.name}"].secret_id
                version = google_secret_manager_secret_version.env["${containers.key}/${env.value.name}"].version
                }] : env.value.value_from_secret != null ? [{
                secret  = env.value.value_from_secret.secret
                version = env.value.value_from_secret.version != "" ? env.value.value_from_secret.version : null
              }] : []
              content {
                secret_key_ref {
                  secret  = value_source.value.secret
                  version = value_source.value.version
                }
              }
            }
          }
        }

        # The single traffic-serving port; name "h2c" enables end-to-end
        # HTTP/2 (required for gRPC streaming).
        dynamic "ports" {
          for_each = containers.value.ports != null ? [containers.value.ports] : []
          content {
            container_port = ports.value.container_port
            name           = ports.value.name != "" ? ports.value.name : null
          }
        }

        # CPU/memory limits land in the API's limits map; the allocation
        # levers (cpu_idle, startup_cpu_boost) ride alongside.
        dynamic "resources" {
          for_each = containers.value.resources != null ? [containers.value.resources] : []
          content {
            # Null (not an empty map) when neither limit is set, so the
            # provider computes defaults instead of diffing on {}.
            limits = (resources.value.cpu != "" || resources.value.memory != "") ? merge(
              resources.value.cpu != "" ? { cpu = resources.value.cpu } : {},
              resources.value.memory != "" ? { memory = resources.value.memory } : {},
            ) : null
            cpu_idle          = resources.value.cpu_idle
            startup_cpu_boost = resources.value.startup_cpu_boost
          }
        }

        dynamic "volume_mounts" {
          for_each = containers.value.volume_mounts
          content {
            name       = volume_mounts.value.name
            mount_path = volume_mounts.value.mount_path
            sub_path   = volume_mounts.value.sub_path != "" ? volume_mounts.value.sub_path : null
          }
        }

        # Startup probe: gates traffic and depends_on waiters until the
        # container is ready. Exactly one handler arm arrives.
        dynamic "startup_probe" {
          for_each = containers.value.startup_probe != null ? [containers.value.startup_probe] : []
          content {
            initial_delay_seconds = startup_probe.value.initial_delay_seconds
            timeout_seconds       = startup_probe.value.timeout_seconds
            period_seconds        = startup_probe.value.period_seconds
            failure_threshold     = startup_probe.value.failure_threshold

            dynamic "http_get" {
              for_each = startup_probe.value.http_get != null ? [startup_probe.value.http_get] : []
              content {
                path = http_get.value.path != "" ? http_get.value.path : null
                port = http_get.value.port

                dynamic "http_headers" {
                  for_each = http_get.value.http_headers
                  content {
                    name  = http_headers.value.name
                    value = http_headers.value.value
                  }
                }
              }
            }

            dynamic "tcp_socket" {
              for_each = startup_probe.value.tcp_socket != null ? [startup_probe.value.tcp_socket] : []
              content {
                port = tcp_socket.value.port
              }
            }

            dynamic "grpc" {
              for_each = startup_probe.value.grpc != null ? [startup_probe.value.grpc] : []
              content {
                port    = grpc.value.port
                service = grpc.value.service != "" ? grpc.value.service : null
              }
            }
          }
        }

        # Liveness probe: restarts an unhealthy container. HTTP/gRPC only —
        # the proto rejects TCP liveness (Cloud Run does not support it).
        dynamic "liveness_probe" {
          for_each = containers.value.liveness_probe != null ? [containers.value.liveness_probe] : []
          content {
            initial_delay_seconds = liveness_probe.value.initial_delay_seconds
            timeout_seconds       = liveness_probe.value.timeout_seconds
            period_seconds        = liveness_probe.value.period_seconds
            failure_threshold     = liveness_probe.value.failure_threshold

            dynamic "http_get" {
              for_each = liveness_probe.value.http_get != null ? [liveness_probe.value.http_get] : []
              content {
                path = http_get.value.path != "" ? http_get.value.path : null
                port = http_get.value.port

                dynamic "http_headers" {
                  for_each = http_get.value.http_headers
                  content {
                    name  = http_headers.value.name
                    value = http_headers.value.value
                  }
                }
              }
            }

            dynamic "grpc" {
              for_each = liveness_probe.value.grpc != null ? [liveness_probe.value.grpc] : []
              content {
                port    = grpc.value.port
                service = grpc.value.service != "" ? grpc.value.service : null
              }
            }
          }
        }

        # Readiness probe: pulls a failing instance from serving (and
        # re-admits it on recovery) without restarting it. HTTP/gRPC only,
        # no initial delay — the API's readiness shape differs from the
        # other probes deliberately. The provider's success_threshold
        # argument is never sent: the GA API's probe message carries no
        # such field and silently drops it, leaving a perpetual re-plan
        # diff (live-verified; the spec has no such field).
        dynamic "readiness_probe" {
          for_each = containers.value.readiness_probe != null ? [containers.value.readiness_probe] : []
          content {
            timeout_seconds   = readiness_probe.value.timeout_seconds
            period_seconds    = readiness_probe.value.period_seconds
            failure_threshold = readiness_probe.value.failure_threshold

            dynamic "http_get" {
              for_each = readiness_probe.value.http_get != null ? [readiness_probe.value.http_get] : []
              content {
                path = http_get.value.path != "" ? http_get.value.path : null
                port = http_get.value.port
              }
            }

            dynamic "grpc" {
              for_each = readiness_probe.value.grpc != null ? [readiness_probe.value.grpc] : []
              content {
                port    = grpc.value.port
                service = grpc.value.service != "" ? grpc.value.service : null
              }
            }
          }
        }
      }
    }
  }

  # Traffic split across revisions. An empty spec list means "100% to the
  # latest ready revision" — achieved by omitting the block entirely so the
  # provider applies GCP's default without recording a diff-prone split.
  dynamic "traffic" {
    for_each = var.spec.traffic
    content {
      type     = traffic.value.type
      revision = traffic.value.revision != "" ? traffic.value.revision : null
      percent  = traffic.value.percent
      tag      = traffic.value.tag != "" ? traffic.value.tag : null
    }
  }

  # Cloud Run checks the runtime identity's access to every secret a
  # revision reads when it creates the revision, so the grants land first.
  depends_on = [google_project_service.run_api, google_secret_manager_secret_iam_member.env]
}

# Public access: grants roles/run.invoker to allUsers when the spec says so.
# This is the additive-IAM path (access THROUGH the invoker check); the
# invoker_iam_disabled field above is the org-policy alternative that turns
# the check off instead. Destroying the grant restores authenticated-only.
resource "google_cloud_run_v2_service_iam_member" "public_invoker" {
  count = var.spec.allow_unauthenticated ? 1 : 0

  project  = google_cloud_run_v2_service.main.project
  location = google_cloud_run_v2_service.main.location
  name     = google_cloud_run_v2_service.main.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
