# Enable the Vertex AI API first so a fresh project works on the first
# deploy. disable_on_destroy is false: tearing down one deployment must
# never disable the API for everything else in the project.
resource "google_project_service" "aiplatform_api" {
  project = local.project_id
  service = "aiplatform.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# A Model Garden or Hugging Face model deployed to a Vertex AI endpoint in
# one step. Every argument is immutable on Google's side, so any change
# replaces the whole deployment (undeploy, delete the endpoint, redeploy);
# deletion_policy is the only lever that updates in place. Optional strings
# and booleans are sent only when set, and Optional+Computed replica counts
# only when set, so Model Garden's per-model defaults stay in charge -- the
# Pulumi module's posture.
resource "google_vertex_ai_endpoint_with_model_garden_deployment" "this" {
  project  = local.project_id
  location = var.spec.location

  # Exactly one model source (proto-enforced).
  publisher_model_name  = local.publisher_model_name
  hugging_face_model_id = local.hugging_face_model_id

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
  deletion_policy = local.deletion_policy

  dynamic "model_config" {
    for_each = local.model_config != null ? [local.model_config] : []
    content {
      accept_eula                = model_config.value.accept_eula ? true : null
      hugging_face_access_token  = model_config.value.hugging_face_access_token != "" ? model_config.value.hugging_face_access_token : null
      hugging_face_cache_enabled = model_config.value.hugging_face_cache_enabled ? true : null
      model_display_name         = model_config.value.model_display_name != "" ? model_config.value.model_display_name : null

      dynamic "container_spec" {
        for_each = model_config.value.container_spec != null ? [model_config.value.container_spec] : []
        content {
          image_uri             = container_spec.value.image_uri
          command               = length(container_spec.value.command) > 0 ? container_spec.value.command : null
          args                  = length(container_spec.value.args) > 0 ? container_spec.value.args : null
          predict_route         = container_spec.value.predict_route != "" ? container_spec.value.predict_route : null
          health_route          = container_spec.value.health_route != "" ? container_spec.value.health_route : null
          deployment_timeout    = container_spec.value.deployment_timeout != "" ? container_spec.value.deployment_timeout : null
          shared_memory_size_mb = container_spec.value.shared_memory_size_mb != null ? tostring(container_spec.value.shared_memory_size_mb) : null

          dynamic "env" {
            for_each = container_spec.value.env
            content {
              name  = env.value.name
              value = env.value.value
            }
          }

          dynamic "ports" {
            for_each = container_spec.value.ports
            content {
              container_port = ports.value.container_port
            }
          }

          dynamic "grpc_ports" {
            for_each = container_spec.value.grpc_ports
            content {
              container_port = grpc_ports.value.container_port
            }
          }

          # The three probes share one spec shape; each block below is the
          # same rendering of that shape.
          dynamic "startup_probe" {
            for_each = container_spec.value.startup_probe != null ? [container_spec.value.startup_probe] : []
            content {
              initial_delay_seconds = startup_probe.value.initial_delay_seconds
              timeout_seconds       = startup_probe.value.timeout_seconds
              period_seconds        = startup_probe.value.period_seconds
              success_threshold     = startup_probe.value.success_threshold
              failure_threshold     = startup_probe.value.failure_threshold

              dynamic "exec" {
                for_each = startup_probe.value.exec != null ? [startup_probe.value.exec] : []
                content {
                  command = exec.value.command
                }
              }
              dynamic "grpc" {
                for_each = startup_probe.value.grpc != null ? [startup_probe.value.grpc] : []
                content {
                  port    = grpc.value.port
                  service = grpc.value.service != "" ? grpc.value.service : null
                }
              }
              dynamic "http_get" {
                for_each = startup_probe.value.http_get != null ? [startup_probe.value.http_get] : []
                content {
                  path   = http_get.value.path != "" ? http_get.value.path : null
                  port   = http_get.value.port
                  host   = http_get.value.host != "" ? http_get.value.host : null
                  scheme = http_get.value.scheme != "" ? http_get.value.scheme : null
                  dynamic "http_headers" {
                    for_each = http_get.value.http_headers
                    content {
                      name  = http_headers.value.name
                      value = http_headers.value.value != "" ? http_headers.value.value : null
                    }
                  }
                }
              }
              dynamic "tcp_socket" {
                for_each = startup_probe.value.tcp_socket != null ? [startup_probe.value.tcp_socket] : []
                content {
                  port = tcp_socket.value.port
                  host = tcp_socket.value.host != "" ? tcp_socket.value.host : null
                }
              }
            }
          }

          dynamic "liveness_probe" {
            for_each = container_spec.value.liveness_probe != null ? [container_spec.value.liveness_probe] : []
            content {
              initial_delay_seconds = liveness_probe.value.initial_delay_seconds
              timeout_seconds       = liveness_probe.value.timeout_seconds
              period_seconds        = liveness_probe.value.period_seconds
              success_threshold     = liveness_probe.value.success_threshold
              failure_threshold     = liveness_probe.value.failure_threshold

              dynamic "exec" {
                for_each = liveness_probe.value.exec != null ? [liveness_probe.value.exec] : []
                content {
                  command = exec.value.command
                }
              }
              dynamic "grpc" {
                for_each = liveness_probe.value.grpc != null ? [liveness_probe.value.grpc] : []
                content {
                  port    = grpc.value.port
                  service = grpc.value.service != "" ? grpc.value.service : null
                }
              }
              dynamic "http_get" {
                for_each = liveness_probe.value.http_get != null ? [liveness_probe.value.http_get] : []
                content {
                  path   = http_get.value.path != "" ? http_get.value.path : null
                  port   = http_get.value.port
                  host   = http_get.value.host != "" ? http_get.value.host : null
                  scheme = http_get.value.scheme != "" ? http_get.value.scheme : null
                  dynamic "http_headers" {
                    for_each = http_get.value.http_headers
                    content {
                      name  = http_headers.value.name
                      value = http_headers.value.value != "" ? http_headers.value.value : null
                    }
                  }
                }
              }
              dynamic "tcp_socket" {
                for_each = liveness_probe.value.tcp_socket != null ? [liveness_probe.value.tcp_socket] : []
                content {
                  port = tcp_socket.value.port
                  host = tcp_socket.value.host != "" ? tcp_socket.value.host : null
                }
              }
            }
          }

          dynamic "health_probe" {
            for_each = container_spec.value.health_probe != null ? [container_spec.value.health_probe] : []
            content {
              initial_delay_seconds = health_probe.value.initial_delay_seconds
              timeout_seconds       = health_probe.value.timeout_seconds
              period_seconds        = health_probe.value.period_seconds
              success_threshold     = health_probe.value.success_threshold
              failure_threshold     = health_probe.value.failure_threshold

              dynamic "exec" {
                for_each = health_probe.value.exec != null ? [health_probe.value.exec] : []
                content {
                  command = exec.value.command
                }
              }
              dynamic "grpc" {
                for_each = health_probe.value.grpc != null ? [health_probe.value.grpc] : []
                content {
                  port    = grpc.value.port
                  service = grpc.value.service != "" ? grpc.value.service : null
                }
              }
              dynamic "http_get" {
                for_each = health_probe.value.http_get != null ? [health_probe.value.http_get] : []
                content {
                  path   = http_get.value.path != "" ? http_get.value.path : null
                  port   = http_get.value.port
                  host   = http_get.value.host != "" ? http_get.value.host : null
                  scheme = http_get.value.scheme != "" ? http_get.value.scheme : null
                  dynamic "http_headers" {
                    for_each = http_get.value.http_headers
                    content {
                      name  = http_headers.value.name
                      value = http_headers.value.value != "" ? http_headers.value.value : null
                    }
                  }
                }
              }
              dynamic "tcp_socket" {
                for_each = health_probe.value.tcp_socket != null ? [health_probe.value.tcp_socket] : []
                content {
                  port = tcp_socket.value.port
                  host = tcp_socket.value.host != "" ? tcp_socket.value.host : null
                }
              }
            }
          }
        }
      }
    }
  }

  dynamic "deploy_config" {
    for_each = local.deploy_config != null ? [local.deploy_config] : []
    content {
      fast_tryout_enabled = deploy_config.value.fast_tryout_enabled ? true : null
      system_labels       = length(deploy_config.value.system_labels) > 0 ? deploy_config.value.system_labels : null

      dynamic "dedicated_resources" {
        for_each = deploy_config.value.dedicated_resources != null ? [deploy_config.value.dedicated_resources] : []
        content {
          min_replica_count      = dedicated_resources.value.min_replica_count
          max_replica_count      = dedicated_resources.value.max_replica_count
          required_replica_count = dedicated_resources.value.required_replica_count
          spot                   = dedicated_resources.value.spot ? true : null

          # machine_spec is required by the API even when every field is
          # left to Google; an empty block is the honest form.
          machine_spec {
            machine_type             = dedicated_resources.value.machine_spec != null && dedicated_resources.value.machine_spec.machine_type != "" ? dedicated_resources.value.machine_spec.machine_type : null
            accelerator_type         = dedicated_resources.value.machine_spec != null && dedicated_resources.value.machine_spec.accelerator_type != "" ? dedicated_resources.value.machine_spec.accelerator_type : null
            accelerator_count        = dedicated_resources.value.machine_spec != null ? dedicated_resources.value.machine_spec.accelerator_count : null
            tpu_topology             = dedicated_resources.value.machine_spec != null && dedicated_resources.value.machine_spec.tpu_topology != "" ? dedicated_resources.value.machine_spec.tpu_topology : null
            multihost_gpu_node_count = dedicated_resources.value.machine_spec != null ? dedicated_resources.value.machine_spec.multihost_gpu_node_count : null

            dynamic "reservation_affinity" {
              for_each = dedicated_resources.value.machine_spec != null && dedicated_resources.value.machine_spec.reservation_affinity != null ? [dedicated_resources.value.machine_spec.reservation_affinity] : []
              content {
                reservation_affinity_type = reservation_affinity.value.reservation_affinity_type
                key                       = reservation_affinity.value.key != "" ? reservation_affinity.value.key : null
                values                    = length(reservation_affinity.value.values) > 0 ? reservation_affinity.value.values : null
              }
            }
          }

          dynamic "autoscaling_metric_specs" {
            for_each = dedicated_resources.value.autoscaling_metric_specs
            content {
              metric_name = autoscaling_metric_specs.value.metric_name
              target      = autoscaling_metric_specs.value.target
            }
          }
        }
      }
    }
  }

  dynamic "endpoint_config" {
    for_each = local.endpoint_config != null ? [local.endpoint_config] : []
    content {
      endpoint_display_name      = endpoint_config.value.endpoint_display_name != "" ? endpoint_config.value.endpoint_display_name : null
      dedicated_endpoint_enabled = endpoint_config.value.dedicated_endpoint_enabled ? true : null

      dynamic "private_service_connect_config" {
        for_each = endpoint_config.value.private_service_connect_config != null ? [endpoint_config.value.private_service_connect_config] : []
        content {
          enable_private_service_connect = private_service_connect_config.value.enable_private_service_connect
          project_allowlist              = length(private_service_connect_config.value.project_allowlist) > 0 ? private_service_connect_config.value.project_allowlist : null

          dynamic "psc_automation_configs" {
            for_each = private_service_connect_config.value.psc_automation_config != null ? [private_service_connect_config.value.psc_automation_config] : []
            content {
              project_id = psc_automation_configs.value.project_id
              network    = psc_automation_configs.value.network
            }
          }
        }
      }
    }
  }

  depends_on = [google_project_service.aiplatform_api]
}
