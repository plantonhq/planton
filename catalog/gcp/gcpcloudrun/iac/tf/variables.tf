variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Specification for the GCP Cloud Run service"
  type = object({
    # The GCP project the service is created in. The CLI's tfvars converter
    # resolves StringValueOrRef fields to their literal string before the
    # module runs, so this arrives as a plain string. Empty falls back to
    # the provider's default project.
    project_id = optional(string, "")

    # Region the service is deployed in, e.g. "us-central1". Immutable.
    region = string

    # Service name in GCP. Empty means "use metadata.name". Immutable.
    service_name = optional(string, "")

    # Human-readable description shown in the console.
    description = optional(string, "")

    # User labels on the service object; merged beneath the platform
    # attribution labels.
    labels = optional(map(string), {})

    # The instance's containers: the serving container plus any sidecars.
    containers = list(object({
      name    = optional(string, "")
      image   = string
      command = optional(list(string), [])
      args    = optional(list(string), [])

      # Environment variables: a literal value XOR a Secret Manager
      # reference (the proto guarantees never both).
      env = optional(list(object({
        name  = string
        value = optional(string, "")
        value_from_secret = optional(object({
          secret  = string
          version = optional(string, "")
        }), null)
      })), [])

      # The single traffic-serving port (at most one container sets it).
      # name selects the protocol: "http1" (default) or "h2c" for HTTP/2.
      ports = optional(object({
        container_port = optional(number, null)
        name           = optional(string, "")
      }), null)

      # CPU/memory limits plus the two CPU-allocation levers.
      resources = optional(object({
        cpu               = optional(string, "")
        memory            = optional(string, "")
        cpu_idle          = optional(bool, null)
        startup_cpu_boost = optional(bool, false)
      }), null)

      volume_mounts = optional(list(object({
        name       = string
        mount_path = string
        sub_path   = optional(string, "")
      })), [])

      working_dir = optional(string, "")

      # Startup probe gates instance start; liveness probe restarts
      # unhealthy containers. Each carries exactly one handler arm
      # (http_get / tcp_socket / grpc) — the proto oneof guarantees it.
      startup_probe = optional(object({
        initial_delay_seconds = optional(number, null)
        timeout_seconds       = optional(number, null)
        period_seconds        = optional(number, null)
        failure_threshold     = optional(number, null)
        http_get = optional(object({
          path = optional(string, "")
          port = optional(number, null)
          http_headers = optional(list(object({
            name  = string
            value = optional(string, "")
          })), [])
        }), null)
        tcp_socket = optional(object({
          port = optional(number, null)
        }), null)
        grpc = optional(object({
          port    = optional(number, null)
          service = optional(string, "")
        }), null)
      }), null)

      liveness_probe = optional(object({
        initial_delay_seconds = optional(number, null)
        timeout_seconds       = optional(number, null)
        period_seconds        = optional(number, null)
        failure_threshold     = optional(number, null)
        http_get = optional(object({
          path = optional(string, "")
          port = optional(number, null)
          http_headers = optional(list(object({
            name  = string
            value = optional(string, "")
          })), [])
        }), null)
        grpc = optional(object({
          port    = optional(number, null)
          service = optional(string, "")
        }), null)
      }), null)

      # Names of containers this one waits for (startup ordering).
      depends_on = optional(list(string), [])

      # Base image for automatic base-image updates on source deploys.
      base_image_uri = optional(string, "")

      # Readiness probe: gates traffic without restarting. HTTP/gRPC only,
      # no initial delay. No success threshold — the GA API's probe
      # message carries no such field (silently dropped; live-verified).
      readiness_probe = optional(object({
        timeout_seconds   = optional(number, null)
        period_seconds    = optional(number, null)
        failure_threshold = optional(number, null)
        http_get = optional(object({
          path = optional(string, "")
          port = optional(number, null)
        }), null)
        grpc = optional(object({
          port    = optional(number, null)
          service = optional(string, "")
        }), null)
      }), null)

      # Marks the sandbox supervisor: the container allowed to launch the
      # sandbox_templates declared on the service.
      sandbox_launcher = optional(bool, false)
    }))

    # Named volumes; each carries exactly one source arm (proto oneof).
    volumes = optional(list(object({
      name = string
      cloud_sql_instance = optional(object({
        # Connection names (project:region:instance) — plain strings after
        # ref resolution.
        instances = list(string)
      }), null)
      secret = optional(object({
        secret       = string
        default_mode = optional(number, null)
        items = optional(list(object({
          path    = string
          version = optional(string, "")
          mode    = optional(number, null)
        })), [])
      }), null)
      empty_dir = optional(object({
        medium     = optional(string, "")
        size_limit = optional(string, "")
      }), null)
      gcs = optional(object({
        # Bucket name — plain string after ref resolution.
        bucket        = string
        read_only     = optional(bool, false)
        mount_options = optional(list(string), [])
      }), null)
      nfs = optional(object({
        server    = string
        path      = string
        read_only = optional(bool, false)
      }), null)
    })), [])

    # Runtime identity email (plain string after ref resolution). Empty
    # uses the project's Compute Engine default service account.
    service_account = optional(string, "")

    # Per-revision instance bounds.
    scaling = optional(object({
      min_instance_count = optional(number, null)
      max_instance_count = optional(number, null)
    }), null)

    # Service-wide scaling posture across all revisions.
    service_scaling = optional(object({
      scaling_mode          = optional(string, "")
      manual_instance_count = optional(number, null)
      min_instance_count    = optional(number, null)
      max_instance_count    = optional(number, null)
    }), null)

    # Per-instance request concurrency (1-1000); null lets GCP default.
    max_instance_request_concurrency = optional(number, null)

    # Request timeout in seconds (1-3600); null lets GCP default to 300.
    timeout_seconds = optional(number, null)

    # Sandbox generation — the proto enum NAME arrives as a string
    # ("EXECUTION_ENVIRONMENT_GEN2"); the provider takes the same values.
    execution_environment = optional(string, "")

    # Best-effort session affinity cookie.
    session_affinity = optional(bool, false)

    # CMEK crypto key ID (plain string after ref resolution).
    encryption_key = optional(string, "")

    # Explicit next-revision name; empty lets Cloud Run generate names.
    revision = optional(string, "")

    # Outbound VPC networking: a connector XOR direct-VPC
    # network_interfaces (the proto CEL guarantees exactly one mechanism).
    vpc_access = optional(object({
      connector = optional(string, "")
      network_interfaces = optional(list(object({
        network    = optional(string, "")
        subnetwork = optional(string, "")
        tags       = optional(list(string), [])
      })), [])
      egress = optional(string, "")
    }), null)

    # GPU hardware requirement.
    node_selector = optional(object({
      accelerator = string
    }), null)

    # Single-zone GPU serving opt-in (cheaper GPU capacity).
    gpu_zonal_redundancy_disabled = optional(bool, false)

    # Ingress posture — the proto enum NAME arrives as a string
    # ("INGRESS_TRAFFIC_ALL"); the provider takes the same values.
    ingress = optional(string, "")

    # Grants roles/run.invoker to allUsers when true.
    allow_unauthenticated = optional(bool, false)

    # Switches the IAM invoker check off entirely (org-policy alternative
    # to the allUsers grant — never combined with it).
    invoker_iam_disabled = optional(bool, false)

    # Extra accepted token audiences for authenticated callers.
    custom_audiences = optional(list(string), [])

    # Traffic split across revisions; empty routes 100% to latest ready.
    traffic = optional(list(object({
      type     = string
      revision = optional(string, "")
      percent  = optional(number, null)
      tag      = optional(string, "")
    })), [])

    # Launch-stage declaration for preview features ("", ALPHA, BETA, GA).
    launch_stage = optional(string, "")

    # Binary Authorization deploy gate.
    binary_authorization = optional(object({
      use_default              = optional(bool, false)
      policy                   = optional(string, "")
      breakglass_justification = optional(string, "")
    }), null)

    # Deletion guard; defaults to true (a destroy fails until disabled).
    deletion_protection = optional(bool, true)

    # Service-object annotations for external tools (system namespaces
    # rejected by the API).
    annotations = optional(map(string), {})

    # Labels / annotations stamped on every revision the template creates.
    revision_labels      = optional(map(string), {})
    revision_annotations = optional(map(string), {})

    # Deploy-from-source build configuration (Cloud Run functions path).
    build_config = optional(object({
      source_location          = optional(string, "")
      function_target          = optional(string, "")
      image_uri                = optional(string, "")
      base_image               = optional(string, "")
      enable_automatic_updates = optional(bool, false)
      environment_variables    = optional(map(string), {})
      worker_pool              = optional(string, "")
      service_account          = optional(string, "")
    }), null)

    # Identity-Aware Proxy in front of the service.
    iap_enabled = optional(bool, false)

    # Disables the default *.run.app URL.
    default_uri_disabled = optional(bool, false)

    # Disables ALL health checking (startup and liveness probes).
    health_check_disabled = optional(bool, false)

    # Multi-region service (requires region = "global").
    multi_region_settings = optional(object({
      regions = list(string)
    }), null)

    # Destroy stance: "", DELETE (default), PREVENT, or ABANDON.
    deletion_policy = optional(string, "")

    # Sandbox templates the supervisor container may launch: isolated
    # containers for untrusted or model-generated code. Literal env only.
    sandbox_templates = optional(list(object({
      name    = string
      image   = string
      command = optional(list(string), [])
      args    = optional(list(string), [])
      env = optional(list(object({
        name  = string
        value = optional(string, "")
      })), [])
      volume_mounts = optional(list(object({
        name       = string
        mount_path = string
        sub_path   = optional(string, "")
      })), [])
      working_dir = optional(string, "")
    })), [])

    # Resource Manager tags bound at creation (tagKeys/* -> tagValues/*).
    # Immutable: a change replaces the service.
    resource_manager_tags = optional(map(string), {})
  })

  validation {
    condition     = length(var.spec.sandbox_templates) == 0 || length([for c in var.spec.containers : c if c.sandbox_launcher]) == 1
    error_message = "sandbox_templates need a supervisor: set sandbox_launcher on exactly one container."
  }
}
