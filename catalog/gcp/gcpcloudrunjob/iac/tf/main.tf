# Enable the Cloud Run Admin API before creating the job so a fresh project
# works first try. disable_on_destroy=false: turning an API off on teardown
# is a project-wide blast radius no single resource should own.
resource "google_project_service" "run_api" {
  project = local.project_id
  service = "run.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The Cloud Run job: a run-to-completion workload definition. Each manual or
# scheduled trigger creates an execution that stamps out task_count tasks with
# up to parallelism running concurrently. This resource owns the definition;
# individual executions are separate API objects.
resource "google_cloud_run_v2_job" "main" {
  name         = local.job_name
  location     = var.spec.region
  project      = local.project_id
  launch_stage = local.launch_stage

  labels      = local.final_labels
  annotations = length(var.spec.annotations) > 0 ? var.spec.annotations : null

  deletion_protection = var.spec.deletion_protection

  # Terraform-side destroy stance: PREVENT fails destroys, ABANDON removes
  # the job from management without deleting it in GCP.
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # Declarative run-on-deploy tokens (mutually exclusive — proto-enforced):
  # start_* counts the job ready when the triggered execution STARTS;
  # run_* counts it ready when the execution COMPLETES.
  start_execution_token = var.spec.start_execution_token != "" ? var.spec.start_execution_token : null
  run_execution_token   = var.spec.run_execution_token != "" ? var.spec.run_execution_token : null

  dynamic "binary_authorization" {
    for_each = var.spec.binary_authorization != null ? [var.spec.binary_authorization] : []
    content {
      use_default              = binary_authorization.value.use_default ? true : null
      policy                   = binary_authorization.value.policy != "" ? binary_authorization.value.policy : null
      breakglass_justification = binary_authorization.value.breakglass_justification != "" ? binary_authorization.value.breakglass_justification : null
    }
  }

  template {
    task_count  = var.spec.task_count
    parallelism = var.spec.parallelism

    # Execution-level metadata, stamped on every execution the job creates
    # (distinct from the job-object labels/annotations above).
    labels      = length(var.spec.execution_labels) > 0 ? var.spec.execution_labels : null
    annotations = length(var.spec.execution_annotations) > 0 ? var.spec.execution_annotations : null

    # The inner template describes one task: containers, volumes, identity,
    # networking, and per-task limits.
    template {
      service_account               = local.service_account
      execution_environment         = local.execution_environment
      encryption_key                = local.encryption_key
      timeout                       = local.timeout
      max_retries                   = var.spec.template.max_retries
      gpu_zonal_redundancy_disabled = var.spec.gpu_zonal_redundancy_disabled ? true : null

      dynamic "node_selector" {
        for_each = var.spec.template.node_selector != null ? [var.spec.template.node_selector] : []
        content {
          accelerator = node_selector.value.accelerator
        }
      }

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

      dynamic "volumes" {
        for_each = var.spec.template.volumes
        content {
          name = volumes.value.name

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

      dynamic "containers" {
        for_each = var.spec.template.containers
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

          dynamic "resources" {
            for_each = containers.value.resources != null ? [containers.value.resources] : []
            content {
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

          # The port a probe targets by default — jobs serve no traffic,
          # so this exists purely for the startup probe.
          dynamic "ports" {
            for_each = containers.value.ports != null ? [containers.value.ports] : []
            content {
              container_port = ports.value.container_port
              name           = ports.value.name != "" ? ports.value.name : null
            }
          }

          # Startup probe: gates task start; a container that never passes
          # is shut down and the task retried per max_retries. The only
          # probe type jobs have.
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
        }
      }
    }
  }

  # The runtime identity must already read every secret the task template
  # references, so the grants land first.
  depends_on = [google_project_service.run_api, google_secret_manager_secret_iam_member.env]
}
