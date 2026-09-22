variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "GcpVertexAiModelGardenDeployment specification"
  type = object({
    # The GCP project the endpoint and model live in: a literal project ID
    # or a GcpProject reference. If omitted, the provider's default project
    # is used. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The Vertex AI location (region) the model is deployed in, e.g.
    # "us-central1". Accelerator availability differs by region. Immutable.
    location = string

    # A Model Garden model, as
    # publishers/{publisher}/models/{model}@{version}, e.g.
    # "publishers/google/models/gemma@gemma-1.1-2b-it" or
    # "publishers/hf-google/models/gemma-2-2b-it@001" for a Hugging Face
    # model Model Garden lists. Exactly one of this or
    # hugging_face_model_id.
    publisher_model_name = optional(string, "")

    # A Hugging Face model by its hub ID, e.g. "Qwen/Qwen3-0.6B" or
    # "google/gemma-2-2b-it"; gated models also need accept_eula and,
    # when the gate requires it, a hugging_face_access_token. Exactly one
    # of this or publisher_model_name.
    hugging_face_model_id = optional(string, "")

    # License acceptance, Hugging Face access, and the serving-container
    # override.
    model_config = optional(object({
      # Accept the model's End User License Agreement. Gated models (Gemma,
      # Llama, and every Hugging Face model with a license gate) refuse to
      # deploy until this is true.
      accept_eula = optional(bool, false)

      # Hugging Face read token used to pull the artifacts of a gated Hugging
      # Face model. Stored as a secret; never logged.
      hugging_face_access_token = optional(string, "")

      # Deploy from Google's cached copy of the Hugging Face model instead of
      # downloading from Hugging Face -- for VPC Service Controls perimeters
      # with limited internet egress.
      hugging_face_cache_enabled = optional(bool, false)

      # Display name of the uploaded Model resource; Google picks one when
      # empty.
      model_display_name = optional(string, "")

      # Override the model's serving container.
      container_spec = optional(object({
        # Container image URI in Artifact Registry (or Container Registry), e.g.
        # us-docker.pkg.dev/vertex-ai/prediction/tf2-cpu.2-13:latest.
        image_uri = string

        # Entrypoint override (the image's ENTRYPOINT is replaced).
        command = optional(list(string), [])

        # Arguments to the entrypoint (the image's CMD is replaced).
        args = optional(list(string), [])

        # Environment variables for the container.
        env = optional(list(object({
          # Variable name.
          name = string

          # Literal value. Variables may reference each other as $(VAR_NAME);
          # the reference is expanded regardless of declaration order.
          value = string
        })), [])

        # HTTP ports the container listens on; Vertex AI sends predictions and
        # health checks to the FIRST one (defaults to 8080 when empty).
        ports = optional(list(object({
          # The port number (1-65535).
          container_port = optional(number)
        })), [])

        # gRPC ports the container listens on, for models served over gRPC.
        grpc_ports = optional(list(object({
          # The port number (1-65535).
          container_port = optional(number)
        })), [])

        # HTTP path Vertex AI forwards prediction requests to, e.g. "/predict";
        # Google's default is its AIP_PREDICT_ROUTE convention.
        predict_route = optional(string, "")

        # HTTP path Vertex AI GETs to check the container's health, e.g.
        # "/health"; Google's default is its AIP_HEALTH_ROUTE convention.
        health_route = optional(string, "")

        # How long the deployment may take before it fails, as a duration string
        # (e.g. "1800s"); Google caps it at two hours.
        deployment_timeout = optional(string, "")

        # VM memory reserved as /dev/shm for the model, in megabytes -- large
        # models loaded across GPUs need it.
        shared_memory_size_mb = optional(number)

        # Probe that gates the container being considered started; liveness
        # and health probes wait for it.
        startup_probe = optional(object({
          # Seconds after container start before the first probe (0-3600).
          initial_delay_seconds = optional(number)

          # Seconds after which one probe attempt times out (1-3600; Google's
          # default 1). Must not exceed period_seconds.
          timeout_seconds = optional(number)

          # Seconds between probe attempts (1-3600; Google's default 10).
          period_seconds = optional(number)

          # Consecutive successes after a failure before the probe counts as
          # healthy again (Google's default 1).
          success_threshold = optional(number)

          # Consecutive failures before the container counts as unhealthy
          # (Google's default 3).
          failure_threshold = optional(number)

          # Run a command inside the container.
          exec = optional(object({
            # The command and its arguments, run without a shell (no variable
            # expansion, no pipes).
            command = list(string)
          }))

          # Call the standard gRPC health service.
          grpc = optional(object({
            # Port of the gRPC service (1-65535).
            port = optional(number)

            # Service name placed in the check request; empty checks the server's
            # overall health.
            service = optional(string, "")
          }))

          # HTTP GET against a path.
          http_get = optional(object({
            # Path to probe, e.g. "/health".
            path = optional(string, "")

            # Port to probe (1-65535).
            port = optional(number)

            # Host header to send; defaults to the container's IP.
            host = optional(string, "")

            # HTTP or HTTPS (Google's default HTTP).
            scheme = optional(string, "")

            # Custom headers sent with the probe request.
            http_headers = optional(list(object({
              # Header name.
              name = string

              # Header value.
              value = optional(string, "")
            })), [])
          }))

          # Open a TCP connection.
          tcp_socket = optional(object({
            # Port to connect to (1-65535).
            port = optional(number)

            # Host to connect to; defaults to the container's IP.
            host = optional(string, "")
          }))
        }))

        # Probe that restarts the container when it fails repeatedly.
        liveness_probe = optional(object({
          # Seconds after container start before the first probe (0-3600).
          initial_delay_seconds = optional(number)

          # Seconds after which one probe attempt times out (1-3600; Google's
          # default 1). Must not exceed period_seconds.
          timeout_seconds = optional(number)

          # Seconds between probe attempts (1-3600; Google's default 10).
          period_seconds = optional(number)

          # Consecutive successes after a failure before the probe counts as
          # healthy again (Google's default 1).
          success_threshold = optional(number)

          # Consecutive failures before the container counts as unhealthy
          # (Google's default 3).
          failure_threshold = optional(number)

          # Run a command inside the container.
          exec = optional(object({
            # The command and its arguments, run without a shell (no variable
            # expansion, no pipes).
            command = list(string)
          }))

          # Call the standard gRPC health service.
          grpc = optional(object({
            # Port of the gRPC service (1-65535).
            port = optional(number)

            # Service name placed in the check request; empty checks the server's
            # overall health.
            service = optional(string, "")
          }))

          # HTTP GET against a path.
          http_get = optional(object({
            # Path to probe, e.g. "/health".
            path = optional(string, "")

            # Port to probe (1-65535).
            port = optional(number)

            # Host header to send; defaults to the container's IP.
            host = optional(string, "")

            # HTTP or HTTPS (Google's default HTTP).
            scheme = optional(string, "")

            # Custom headers sent with the probe request.
            http_headers = optional(list(object({
              # Header name.
              name = string

              # Header value.
              value = optional(string, "")
            })), [])
          }))

          # Open a TCP connection.
          tcp_socket = optional(object({
            # Port to connect to (1-65535).
            port = optional(number)

            # Host to connect to; defaults to the container's IP.
            host = optional(string, "")
          }))
        }))

        # Probe that decides whether the container receives traffic.
        health_probe = optional(object({
          # Seconds after container start before the first probe (0-3600).
          initial_delay_seconds = optional(number)

          # Seconds after which one probe attempt times out (1-3600; Google's
          # default 1). Must not exceed period_seconds.
          timeout_seconds = optional(number)

          # Seconds between probe attempts (1-3600; Google's default 10).
          period_seconds = optional(number)

          # Consecutive successes after a failure before the probe counts as
          # healthy again (Google's default 1).
          success_threshold = optional(number)

          # Consecutive failures before the container counts as unhealthy
          # (Google's default 3).
          failure_threshold = optional(number)

          # Run a command inside the container.
          exec = optional(object({
            # The command and its arguments, run without a shell (no variable
            # expansion, no pipes).
            command = list(string)
          }))

          # Call the standard gRPC health service.
          grpc = optional(object({
            # Port of the gRPC service (1-65535).
            port = optional(number)

            # Service name placed in the check request; empty checks the server's
            # overall health.
            service = optional(string, "")
          }))

          # HTTP GET against a path.
          http_get = optional(object({
            # Path to probe, e.g. "/health".
            path = optional(string, "")

            # Port to probe (1-65535).
            port = optional(number)

            # Host header to send; defaults to the container's IP.
            host = optional(string, "")

            # HTTP or HTTPS (Google's default HTTP).
            scheme = optional(string, "")

            # Custom headers sent with the probe request.
            http_headers = optional(list(object({
              # Header name.
              name = string

              # Header value.
              value = optional(string, "")
            })), [])
          }))

          # Open a TCP connection.
          tcp_socket = optional(object({
            # Port to connect to (1-65535).
            port = optional(number)

            # Host to connect to; defaults to the container's IP.
            host = optional(string, "")
          }))
        }))
      }))
    }))

    # The compute the model is deployed on. Omit to accept Model Garden's
    # recommended shape for the model.
    deploy_config = optional(object({
      # Machine shape and replica bounds.
      dedicated_resources = optional(object({
        # The machine each replica runs on.
        machine_spec = optional(object({
          # Compute Engine machine type, e.g. "g2-standard-12" (with an L4) or
          # "a2-highgpu-1g" (with an A100). Google picks a default for the model
          # when the whole deploy_config is omitted.
          machine_type = optional(string, "")

          # Accelerator attached to each replica, e.g. "NVIDIA_L4",
          # "NVIDIA_TESLA_A100", "NVIDIA_H100_80GB", "TPU_V5_LITEPOD". Set together
          # with accelerator_count.
          accelerator_type = optional(string, "")

          # Accelerators per replica.
          accelerator_count = optional(number)

          # TPU topology for TPU machine types, e.g. "2x2x1".
          tpu_topology = optional(string, "")

          # Nodes per replica for multi-host GPU deployments.
          multihost_gpu_node_count = optional(number)

          # Consume Compute Engine reservations for the replicas.
          reservation_affinity = optional(object({
            # NO_RESERVATION never consumes a reservation; ANY_RESERVATION consumes
            # any matching one; SPECIFIC_RESERVATION consumes the reservation named
            # by key and values.
            reservation_affinity_type = string

            # For SPECIFIC_RESERVATION: the label key
            # "compute.googleapis.com/reservation-name".
            key = optional(string, "")

            # For SPECIFIC_RESERVATION: the full resource name(s) of the reservation
            # or reservation block.
            values = optional(list(string), [])
          }))
        }))

        # Replicas always kept serving (at least 1 -- a deployed model never
        # scales to zero, so this is the committed spend).
        min_replica_count = optional(number, 0)

        # Ceiling the replica count may autoscale to; Google defaults it to
        # min_replica_count. Sent only when set.
        max_replica_count = optional(number)

        # Replicas that must be available for the deployment to succeed; the
        # rest are retried. Google defaults it to min_replica_count.
        required_replica_count = optional(number)

        # Run the replicas on Spot VMs -- much cheaper, preemptible at any
        # time.
        spot = optional(bool, false)

        # Metrics that drive autoscaling between the replica bounds.
        autoscaling_metric_specs = optional(list(object({
          # The metric:
          # "aiplatform.googleapis.com/prediction/online/accelerator/duty_cycle"
          # or "aiplatform.googleapis.com/prediction/online/cpu/utilization".
          metric_name = string

          # Target utilization percentage (1-100; Google's default 60).
          target = optional(number)
        })), [])
      }))

      # Enable Model Garden's fast-tryout serving path for the model when it
      # supports one.
      fast_tryout_enabled = optional(bool, false)

      # Google-managed system labels for Model Garden tracking.
      system_labels = optional(map(string), {})
    }))

    # The endpoint the model is served from.
    endpoint_config = optional(object({
      # Display name of the created endpoint; Google picks one when empty.
      endpoint_display_name = optional(string, "")

      # Give the endpoint its own DNS name
      # ({endpoint}.{region}-{project}.prediction.vertexai.goog) isolated from
      # the shared regional host. Once on, the shared host no longer serves
      # it.
      dedicated_endpoint_enabled = optional(bool, false)

      # Expose the endpoint through Private Service Connect.
      private_service_connect_config = optional(object({
        # Publish the endpoint through Private Service Connect.
        enable_private_service_connect = optional(bool, false)

        # Consumer projects allowed to create forwarding rules to the endpoint's
        # service attachment: GcpProject references or literal project IDs.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        project_allowlist = optional(list(string), [])

        # One consumer network Vertex AI creates the PSC endpoint in
        # automatically (Google's resource takes exactly one; more consumers
        # build their own forwarding rules against the service attachment).
        psc_automation_config = optional(object({
          # The consumer project the endpoint is created in: a GcpProject
          # reference or a literal project ID.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          project_id = string

          # The consumer VPC network the endpoint is created in, as
          # projects/{project}/global/networks/{name}. A GcpVpcNetwork reference
          # resolves to its network_id output.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          network = string
        }))
      }))
    }))

    # What happens to the deployment when this resource is destroyed:
    #   "" / "DELETE" -- the model is undeployed and the endpoint and model
    #                    are deleted
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the deployment leaves management and keeps serving
    #                    (and billing)
    deletion_policy = optional(string, "")
  })
}
