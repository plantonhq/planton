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
  description = "GcpCloudRunJob specification"
  type = object({
    # The GCP project the job is created in. Accepts a literal project ID
    # or a reference to a GcpProject resource. If omitted, the provider's
    # default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Region the job is deployed in, e.g. "us-central1". Immutable.
    region = string

    # Name of the Cloud Run job in GCP. Immutable. If not specified,
    # defaults to metadata.name. Must be 1-63 characters: lowercase letters,
    # digits, and hyphens; starting with a letter.
    job_name = optional(string, "")

    # Labels applied to the job object. User labels are shared with
    # Google's billing system. Keys and values in the `run.googleapis.com`,
    # `cloud.googleapis.com`, `serving.knative.dev`, and
    # `autoscaling.knative.dev` namespaces are rejected by the API.
    labels = optional(map(string), {})

    # Unstructured metadata preserved by external tools. Not queryable.
    # System namespaces (`run.googleapis.com`, etc.) are rejected on create.
    annotations = optional(map(string), {})

    # The task template every execution runs: containers, volumes,
    # networking, hardware, and per-task limits. Required.
    template = object({
      # The containers that make up one task. The first container is the main
      # worker; additional containers are sidecars sharing the task's network
      # namespace (localhost) and volumes. Use depends_on for startup ordering.
      containers = list(object({
        # Name of the container. Required when the task runs more than one
        # container (depends_on refers to these names).
        name = optional(string, "")

        # Container image URL, e.g. "us-docker.pkg.dev/project/repo/worker:1.0.0".
        #
        # Cloud Run pulls PRIVATE images only from Artifact Registry (or the
        # legacy Container Registry) in a project its service agent can read;
        # public Docker Hub and GHCR images run directly. There is no field here
        # for a registry login and Google accepts none. For a private image in
        # another registry, push it to Artifact Registry, or declare a
        # GcpArtifactRegistryRepo in REMOTE_REPOSITORY mode that proxies that
        # registry and point this URL at it.
        image = string

        # Entrypoint array — overrides the image's ENTRYPOINT.
        command = optional(list(string), [])

        # Arguments to the entrypoint — overrides the image's CMD.
        args = optional(list(string), [])

        # Environment variables. Each entry carries either a literal value or a
        # Secret Manager reference resolved at task start.
        env = optional(list(object({
          # Variable name, e.g. "BATCH_SIZE". Must not start with a digit.
          name = string

          # Literal value, written into the job's task template where anyone who
          # can view the job reads it. Fine for configuration; never a credential
          # -- a credential goes in secret_value (or value_from_secret).
          value = optional(string, "")

          # A Secret Manager secret you already own, resolved at task start.
          value_from_secret = optional(object({
            # The secret: a short name or full resource name (projects/*/secrets/*).
            secret = string

            # Secret version: a version number or "latest".
            version = optional(string, "")
          }))

          # A secret value this component keeps in Secret Manager for you. It
          # creates one secret for this variable, replicated only in the job's
          # region, stores the value as a version, grants the job's runtime
          # identity secretAccessor on that secret alone, and points the variable at
          # that exact version -- the task template carries a reference, never the
          # value. A changed value adds a version and updates the template, so
          # rotation is a deploy; destroying the job removes the secret.
          secret_value = optional(string, "")
        })), [])

        # CPU and memory limits for this container.
        resources = optional(object({
          # CPU limit, e.g. "1", "2", "4", "8" or a fraction like "0.5"/"500m".
          # GPU tasks need at least "4".
          cpu = optional(string, "")

          # Memory limit with unit suffix, e.g. "512Mi", "2Gi", "16Gi".
          memory = optional(string, "")
        }))

        # Volumes (declared at the template level) mounted into this container.
        volume_mounts = optional(list(object({
          # Name of a volume declared in template.volumes.
          name = string

          # Absolute path in the container. Cloud SQL volumes must mount at "/cloudsql".
          mount_path = string

          # Path WITHIN the volume to mount instead of its root — e.g. mount only
          # one secret item or one bucket directory. Relative path; empty mounts
          # the volume root.
          sub_path = optional(string, "")
        })), [])

        # Working directory for the entrypoint.
        working_dir = optional(string, "")

        # Names of containers this one waits for before starting.
        depends_on = optional(list(string), [])

        # The port this container listens on — for jobs this exists to give
        # probes a target (there is no request traffic). If unset and a probe
        # needs a port, the probe names one explicitly.
        ports = optional(object({
          # Port number the container listens on.
          container_port = optional(number)

          # Protocol selector: "http1" (default) or "h2c".
          name = optional(string, "")
        }))

        # Probe that gates task start: the task's work does not begin (and
        # depends_on waiters stay blocked) until this succeeds; a container
        # that never passes is shut down and the task retried per max_retries.
        # Supports HTTP, TCP, and gRPC checks — the ONLY probe type jobs have
        # (no liveness/readiness; those are serving concerns).
        startup_probe = optional(object({
          # Seconds to wait after container start before the first probe (0-240).
          initial_delay_seconds = optional(number)

          # Seconds after which a single probe attempt times out (1-240; GCP
          # default 1). Must not exceed period_seconds.
          timeout_seconds = optional(number)

          # Seconds between probe attempts (1-240; GCP default 10).
          period_seconds = optional(number)

          # Consecutive failures after which the startup probe fails and the
          # container is shut down (GCP default 3).
          failure_threshold = optional(number)

          # HTTP GET against a path on the container; 2xx is success. The job
          # code must expose the health endpoint itself.
          http_get = optional(object({
            # Path to probe, e.g. "/healthz". Defaults to "/".
            path = optional(string, "")

            # Port to probe. If unset, the container's declared port is used.
            port = optional(number)

            # Custom headers sent with the probe request.
            http_headers = optional(list(object({
              # Header name.
              name = string

              # Header value.
              value = optional(string, "")
            })), [])
          }))

          # TCP connect to a port; a successful connection is success. The
          # simplest probe — no code changes needed beyond listening.
          tcp_socket = optional(object({
            # Port to connect to. If unset, the container's declared port is used.
            port = optional(number)
          }))

          # Standard gRPC health-check protocol (grpc.health.v1.Health/Check),
          # which the job code must implement.
          grpc = optional(object({
            # Port the gRPC health service listens on. If unset, the container's
            # declared port is used.
            port = optional(number)

            # Service name passed to the health check. If empty, overall server
            # health is checked.
            service = optional(string, "")
          }))
        }))
      }))

      # Named volumes tasks can mount: Cloud SQL sockets, Secret Manager
      # material, scratch space, GCS buckets (FUSE), and NFS shares.
      volumes = optional(list(object({
        # Volume name referenced by volume_mounts entries.
        name = string

        # Cloud SQL Unix sockets under the mount path.
        cloud_sql_instance = optional(object({
          # Cloud SQL instance connection names (project:region:instance).
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          instances = list(string)
        }))

        # Secret Manager secret versions exposed as files.
        secret = optional(object({
          # The secret: a short name or full resource name.
          secret = string

          # Default Unix permission mode for projected files, in decimal (e.g. 292 = 0444).
          default_mode = optional(number)

          # Which versions land at which relative paths.
          items = optional(list(object({
            # Relative path under the volume's mount path.
            path = string

            # Secret version: a version number or "latest".
            version = optional(string, "")

            # Unix permission mode for this file, in decimal.
            mode = optional(number)
          })), [])
        }))

        # Ephemeral scratch space, in-memory or disk-backed.
        empty_dir = optional(object({
          # Backing medium: MEMORY (default) or DISK.
          medium = optional(string, "")

          # Capacity limit with unit suffix, e.g. "512Mi", "2Gi".
          size_limit = optional(string, "")
        }))

        # A GCS bucket mounted via Cloud Storage FUSE (requires GEN2).
        gcs = optional(object({
          # The bucket to mount. Accepts a literal name or a GcpGcsBucket reference.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          bucket = string

          # Mount read-only.
          read_only = optional(bool, false)

          # Flags passed to the gcsfuse command mounting this volume, without
          # leading dashes (e.g. "implicit-dirs", "only-dir=media",
          # "file-cache-max-size-mb=512").
          mount_options = optional(list(string), [])
        }))

        # An NFS share mounted into the task (requires GEN2 and VPC reachability).
        nfs = optional(object({
          # Hostname or IP of the NFS server.
          server = string

          # Exported path on the server, e.g. "/share1".
          path = string

          # Mount read-only.
          read_only = optional(bool, false)
        }))
      })), [])

      # Email of the IAM service account each task runs as. Accepts a literal
      # email or a reference to a GcpServiceAccount resource. If omitted, the
      # project's Compute Engine default service account is used.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      service_account = optional(string, "")

      # Sandbox generation tasks execute in. GEN2 (recommended) offers full
      # Linux compatibility (required for GCS/NFS volumes); GEN1 starts faster
      # with a gVisor-restricted syscall surface.
      execution_environment = optional(string, "")

      # Customer-managed encryption key (CMEK) encrypting deployed container
      # images. Accepts a full crypto key ID or a reference to a GcpKmsKey
      # resource. The key must be in the same region as the job.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      encryption_key = optional(string, "")

      # Maximum time one task attempt may run before GCP marks it failed, in
      # seconds (1-86400). If unset, GCP defaults to 600 (10 minutes). Each
      # retry gets a fresh timeout budget.
      timeout_seconds = optional(number)

      # Retries per task before marking it failed (>= 0). GCP defaults to 3.
      # Set 0 for fail-fast batch work with no retries.
      max_retries = optional(number)

      # Private networking for OUTBOUND traffic from tasks.
      vpc_access = optional(object({
        # Serverless VPC Access connector (legacy mechanism). Full resource name
        # (projects/*/locations/*/connectors/*) or a reference to a
        # GcpServerlessVpcConnector resource.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        connector = optional(string, "")

        # Direct VPC egress: tasks get IPs in the subnetwork.
        network_interfaces = optional(list(object({
          # The VPC network. Accepts a literal name or a GcpVpcNetwork reference.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          network = optional(string, "")

          # The subnetwork tasks draw IPs from. Must be in the job's region.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          subnetwork = optional(string, "")

          # Network tags applied to tasks — how VPC firewall rules select egress.
          tags = optional(list(string), [])
        })), [])

        # Which egress uses the VPC path: ALL_TRAFFIC or PRIVATE_RANGES_ONLY.
        egress = optional(string, "")
      }))

      # Hardware requirements for GPU batch workloads. Setting an accelerator
      # (e.g. "nvidia-l4") gives each task one GPU; container resource limits
      # must meet Cloud Run's GPU minimums (4 CPU / 16Gi recommended).
      node_selector = optional(object({
        # GPU accelerator type each task gets, e.g. "nvidia-l4".
        accelerator = string
      }))
    })

    # Desired number of tasks each execution runs (>= 1). Setting 1 means
    # success of that single task signals execution success. If unset, GCP
    # defaults to 1.
    task_count = optional(number)

    # Maximum tasks running concurrently during an execution (>= 0). Must be
    # <= task_count when both are set. If 0 or unset, GCP uses the maximum
    # possible for that run.
    parallelism = optional(number)

    # Launch-stage gate for preview Cloud Run features. Set BETA (or ALPHA)
    # only when the spec uses features GCP rejects at the default GA stage.
    # The provider enum admits more values (UNIMPLEMENTED, PRELAUNCH,
    # EARLY_ACCESS, DEPRECATED), but Cloud Run itself supports only
    # ALPHA/BETA/GA — this list encodes the API truth, deliberately.
    launch_stage = optional(string, "")

    # Binary Authorization: only container images that pass the policy's
    # attestation checks may deploy.
    binary_authorization = optional(object({
      # Evaluate deploys against the project's default Binary Authorization policy.
      use_default = optional(bool, false)

      # Evaluate deploys against a specific platform policy.
      policy = optional(string, "")

      # Justification recorded when a break-glass deploy bypasses the policy.
      breakglass_justification = optional(string, "")
    }))

    # Opts a GPU job out of zonal redundancy: tasks may run from a single
    # zone for cheaper GPU capacity. Only meaningful when
    # template.node_selector is set.
    gpu_zonal_redundancy_disabled = optional(bool, false)

    # Prevents the job from being destroyed while true. Defaults to true
    # (matching GCP's posture): a delete fails until this is set to false.
    deletion_protection = optional(bool)

    # Labels stamped on every EXECUTION the job creates (visible on
    # execution objects and in billing breakdowns), as opposed to `labels`,
    # which lands on the job object itself. Same namespace restrictions as
    # job labels.
    execution_labels = optional(map(string), {})

    # Annotations stamped on every EXECUTION the job creates, as opposed to
    # `annotations`, which lands on the job object. Same namespace
    # restrictions.
    execution_annotations = optional(map(string), {})

    # Declarative run-on-deploy: a unique suffix string that triggers a new
    # execution when the job is created or updated, with the job counted
    # READY as soon as the execution successfully STARTS. Changing the
    # token on a later update triggers another run. The combined length of
    # job name and token must stay under 63 characters. Set at most one of
    # the two token fields.
    start_execution_token = optional(string, "")

    # Declarative run-on-deploy like start_execution_token, but the job is
    # counted READY only when the triggered execution successfully
    # COMPLETES — deploy-and-verify semantics for migrations and one-shot
    # setup work. The combined length of job name and token must stay under
    # 63 characters. Set at most one of the two token fields.
    run_execution_token = optional(string, "")

    # What happens to the Cloud Run job when this resource is destroyed:
    #   "" / "DELETE" -- the job is deleted (default)
    #   "PREVENT"     -- destroy operations fail while this is set
    #   "ABANDON"     -- the job is removed from management but left
    #                    running in GCP
    deletion_policy = optional(string, "")

    # Resource Manager tags bound to the job at creation, as a map of
    # tagKeys/{tag_key_id} to tagValues/{tag_value_id} — the tag bindings
    # that organization policies, IAM conditions, and cost reports key on.
    # Immutable: changing the map replaces the job (Cloud Run applies tags
    # only at create), so plan tag changes as a recreate.
    resource_manager_tags = optional(map(string), {})
  })
}
