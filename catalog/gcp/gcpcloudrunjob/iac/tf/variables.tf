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
  description = "Specification for the GCP Cloud Run job"
  type = object({
    project_id  = optional(string, "")
    region      = string
    job_name    = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})

    template = object({
      containers = list(object({
        name    = optional(string, "")
        image   = string
        command = optional(list(string), [])
        args    = optional(list(string), [])
        env = optional(list(object({
          name  = string
          value = optional(string, "")
          value_from_secret = optional(object({
            secret  = string
            version = optional(string, "")
          }), null)
        })), [])
        resources = optional(object({
          cpu    = optional(string, "")
          memory = optional(string, "")
        }), null)
        volume_mounts = optional(list(object({
          name       = string
          mount_path = string
          sub_path   = optional(string, "")
        })), [])
        working_dir = optional(string, "")
        depends_on  = optional(list(string), [])

        # Probe target port ("h2c"/"http1" protocol selector in name).
        ports = optional(object({
          container_port = optional(number, null)
          name           = optional(string, "")
        }), null)

        # Startup probe — the only probe type jobs have. Exactly one
        # handler arm arrives (http_get / tcp_socket / grpc).
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
      }))

      volumes = optional(list(object({
        name = string
        cloud_sql_instance = optional(object({
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

      service_account       = optional(string, "")
      execution_environment = optional(string, "")
      encryption_key        = optional(string, "")
      timeout_seconds       = optional(number, null)
      max_retries           = optional(number, null)

      vpc_access = optional(object({
        connector = optional(string, "")
        network_interfaces = optional(list(object({
          network    = optional(string, "")
          subnetwork = optional(string, "")
          tags       = optional(list(string), [])
        })), [])
        egress = optional(string, "")
      }), null)

      node_selector = optional(object({
        accelerator = string
      }), null)
    })

    task_count   = optional(number, null)
    parallelism  = optional(number, null)
    launch_stage = optional(string, "")

    binary_authorization = optional(object({
      use_default              = optional(bool, false)
      policy                   = optional(string, "")
      breakglass_justification = optional(string, "")
    }), null)

    gpu_zonal_redundancy_disabled = optional(bool, false)
    deletion_protection           = optional(bool, true)

    # Labels / annotations stamped on every execution the job creates.
    execution_labels      = optional(map(string), {})
    execution_annotations = optional(map(string), {})

    # Declarative run-on-deploy tokens (mutually exclusive). start_* counts
    # the job ready when the execution starts; run_* when it completes.
    start_execution_token = optional(string, "")
    run_execution_token   = optional(string, "")

    # Destroy stance: "", DELETE (default), PREVENT, or ABANDON.
    deletion_policy = optional(string, "")

    # Resource Manager tags bound at creation (tagKeys/* -> tagValues/*).
    # Immutable: a change replaces the job.
    resource_manager_tags = optional(map(string), {})
  })
}
