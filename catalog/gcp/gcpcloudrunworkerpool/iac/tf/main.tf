# Enable the Cloud Run Admin API before creating the pool so a fresh
# project works first try. disable_on_destroy=false: turning an API off on
# teardown is a project-wide blast radius no single resource should own.
resource "google_project_service" "run_api" {
  project = local.project_id
  service = "run.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The Cloud Run worker pool: a Cloud Run service with the request path
# removed -- the same revision template, but no port, no ingress, no
# invoker IAM, and instances scaled to a count (or between bounds by the
# owner's signal) instead of by traffic. Every apply that changes the
# template stamps out a new immutable revision; instance_splits decides how
# many instances each revision runs.
#
# Immutable: location, name, and a container's depends_on. Everything else
# rolls a new revision in place.
resource "google_cloud_run_v2_worker_pool" "main" {
  name        = local.worker_pool_name
  location    = var.spec.region
  project     = local.project_id
  description = local.description

  # Null lets GCP apply its own default (GA).
  launch_stage = local.launch_stage

  labels = local.final_labels

  # Pool-object annotations for external tools (system namespaces are
  # API-rejected). Revision-level annotations live in the template below.
  annotations = length(var.spec.annotations) > 0 ? var.spec.annotations : null

  # Deletion guard, honest by default: the spec defaults this to true, so a
  # destroy fails until the manifest explicitly opts out.
  deletion_protection = var.spec.deletion_protection

  # Terraform-side destroy stance: PREVENT fails destroys, ABANDON removes
  # the pool from management without deleting it in GCP.
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

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

  # Instance count: MANUAL pins the total (Google's default mode); AUTOMATIC
  # moves between bounds. Each lever is sent only when set so Google's
  # defaults stay in charge otherwise.
  dynamic "scaling" {
    for_each = var.spec.scaling != null ? [var.spec.scaling] : []
    content {
      scaling_mode          = scaling.value.scaling_mode != "" ? scaling.value.scaling_mode : null
      manual_instance_count = scaling.value.manual_instance_count
      min_instance_count    = scaling.value.min_instance_count
      max_instance_count    = scaling.value.max_instance_count
    }
  }

  # Instance split across revisions. An empty spec list means "every
  # instance on the latest ready revision" -- achieved by omitting the block
  # so the provider applies GCP's default without a diff-prone split.
  dynamic "instance_splits" {
    for_each = var.spec.instance_splits
    content {
      type     = instance_splits.value.type
      revision = instance_splits.value.revision != "" ? instance_splits.value.revision : null
      percent  = instance_splits.value.percent
    }
  }

  template {
    # Explicit revision naming makes declarative rollouts by revision
    # possible; null (the norm) lets Cloud Run generate names.
    revision = local.revision

    # Revision-level metadata, stamped on every revision the template
    # creates (distinct from the pool-object labels/annotations above).
    labels      = length(var.spec.revision_labels) > 0 ? var.spec.revision_labels : null
    annotations = length(var.spec.revision_annotations) > 0 ? var.spec.revision_annotations : null

    # The runtime identity whose permissions the code exercises. Null uses
    # the project's Compute Engine default service account.
    service_account = local.service_account

    # CMEK for the deployed images, and what happens to running instances
    # if the key is revoked (the proto ties the two levers to the key).
    encryption_key                   = local.encryption_key
    encryption_key_revocation_action = local.encryption_key_revocation_action
    encryption_key_shutdown_duration = local.encryption_key_shutdown_duration

    # Single-zone GPU serving opt-in (cheaper GPU capacity for zonal risk).
    gpu_zonal_redundancy_disabled = var.spec.gpu_zonal_redundancy_disabled ? true : null

    # GPU hardware requirement (e.g. "nvidia-l4").
    dynamic "node_selector" {
      for_each = var.spec.node_selector != null ? [var.spec.node_selector] : []
      content {
        accelerator = node_selector.value.accelerator
      }
    }

    # Outbound VPC networking: a Serverless VPC Access connector XOR direct
    # VPC egress network_interfaces -- the proto guarantees exactly one.
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

        # Cloud SQL Unix sockets -- GCP manages the proxying; connect via
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

    # The worker container plus any sidecars; containers share localhost
    # and volumes, and depends_on orders their startup. No container
    # exposes a port -- a worker pool receives no requests.
    dynamic "containers" {
      for_each = var.spec.containers
      content {
        name        = containers.value.name != "" ? containers.value.name : null
        image       = containers.value.image
        command     = length(containers.value.command) > 0 ? containers.value.command : null
        args        = length(containers.value.args) > 0 ? containers.value.args : null
        working_dir = containers.value.working_dir != "" ? containers.value.working_dir : null
        depends_on  = length(containers.value.depends_on) > 0 ? containers.value.depends_on : null

        # A literal, a Secret Manager secret the author owns, or a secret
        # value secrets.tf stored (one of the three -- proto-enforced).
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

        # CPU/memory limits land in the API's limits map. A worker pool has
        # no idle-CPU or startup-boost lever: CPU is always allocated.
        dynamic "resources" {
          for_each = containers.value.resources != null ? [containers.value.resources] : []
          content {
            # Null (not an empty map) when neither limit is set, so the
            # provider computes defaults instead of diffing on {}.
            limits = (resources.value.cpu != "" || resources.value.memory != "") ? merge(
              resources.value.cpu != "" ? { cpu = resources.value.cpu } : {},
              resources.value.memory != "" ? { memory = resources.value.memory } : {},
            ) : null
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

        # Startup probe: gates depends_on waiters until the container is
        # ready. Exactly one handler arm arrives.
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

                # At most one header (spec-capped): the pinned Pulumi SDK
                # models a worker-pool probe's headers as one header, so
                # both engines accept the same manifests.
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

        # Liveness probe: restarts an unhealthy container. HTTP/gRPC only --
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
      }
    }
  }

  depends_on = [google_project_service.run_api, google_secret_manager_secret_iam_member.env]
}
