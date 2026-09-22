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
  description = "GcpCloudRunWorkerPool specification"
  type = object({
    # The GCP project the worker pool is created in. Accepts a literal
    # project ID or a reference to a GcpProject resource. If omitted, the
    # provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Region the worker pool runs in, e.g. "us-central1". Immutable.
    region = string

    # Name of the worker pool in GCP. Immutable. Defaults to metadata.name.
    # 1-63 characters: lowercase letters, digits, and hyphens, starting with
    # a letter and ending alphanumeric.
    worker_pool_name = optional(string, "")

    # Human-readable description, shown in the console.
    description = optional(string, "")

    # Labels on the worker pool object, shared with Google's billing system
    # for cost breakdowns. Keys in the `run.googleapis.com`,
    # `cloud.googleapis.com`, `serving.knative.dev`, and
    # `autoscaling.knative.dev` namespaces are rejected by the API.
    labels = optional(map(string), {})

    # Annotations on the worker pool object: unstructured metadata for
    # external tools, never behavioral for Cloud Run. Same namespace
    # restrictions as labels. For metadata stamped on each REVISION use
    # revision_annotations.
    annotations = optional(map(string), {})

    # The containers that make up one instance. The first container
    # conventionally does the work; additional containers are sidecars
    # (collectors, proxies) sharing the instance's localhost and volumes,
    # ordered by depends_on. No container exposes a port -- a worker pool
    # receives no requests. Deployment pipelines inject the built image into
    # containers whose image is left BLANK (the image-slot contract);
    # authored images are untouched.
    containers = list(object({
      # Name of the container. Required when the pool runs more than one
      # container (depends_on refers to these names). If omitted for a single
      # container, Cloud Run assigns one.
      name = optional(string, "")

      # Container image URL, e.g. "us-docker.pkg.dev/project/repo/worker:1.0.0".
      # Pin a digest or immutable tag for repeatable deploys -- Cloud Run
      # resolves the image to a digest at revision creation. Private images
      # are pulled only from Artifact Registry (or the legacy Container
      # Registry) the Cloud Run service agent can read; public Docker Hub and
      # GHCR images deploy directly.
      image = string

      # Entrypoint array -- overrides the image's ENTRYPOINT. Not executed in
      # a shell; variable references are not expanded.
      command = optional(list(string), [])

      # Arguments to the entrypoint -- overrides the image's CMD.
      args = optional(list(string), [])

      # Environment variables. Each entry carries a literal value or a Secret
      # Manager reference resolved at instance start.
      env = optional(list(object({
        # Variable name, e.g. "QUEUE_SUBSCRIPTION". Must not start with a digit.
        name = string

        # Literal value. Fine for configuration; never place credentials here --
        # use value_from_secret so the material stays in Secret Manager.
        value = optional(string, "")

        # Secret Manager reference resolved into the variable at instance start.
        value_from_secret = optional(object({
          # The secret: a short name for a secret in the same project
          # ("db-password") or a full resource name (projects/*/secrets/*) for
          # cross-project reads.
          secret = string

          # Secret version to resolve: a version number or "latest". GCP requires
          # an explicit version for env vars -- "latest" is the common choice, at
          # the cost of new instances silently picking up rotations.
          version = optional(string, "")
        }))
      })), [])

      # CPU and memory for this container. If omitted, Cloud Run defaults
      # apply (1 CPU, 512Mi). CPU is always allocated on a worker pool.
      resources = optional(object({
        # CPU limit: "1", "2", "4", "6", "8" or a fraction like "0.5"/"500m".
        # 4 CPU needs at least 2Gi of memory; 6 or more need 4Gi. GPU workers
        # need at least "4".
        cpu = optional(string, "")

        # Memory limit with unit suffix, e.g. "512Mi", "2Gi". Minimums scale
        # with CPU; GPU workers need at least "16Gi".
        memory = optional(string, "")
      }))

      # Volumes (declared at the pool level) mounted into this container's
      # filesystem.
      volume_mounts = optional(list(object({
        # Name of a volume declared in spec.volumes.
        name = string

        # Absolute path in the container to mount at. Cloud SQL volumes must
        # mount at "/cloudsql".
        mount_path = string

        # Path WITHIN the volume to mount instead of its root. Relative; empty
        # mounts the volume root.
        sub_path = optional(string, "")
      })), [])

      # Working directory for the entrypoint. If omitted, the image's WORKDIR
      # is used.
      working_dir = optional(string, "")

      # Probe that gates instance start: depends_on waiters stay blocked until
      # this succeeds, and an instance that never passes is killed. HTTP, TCP,
      # or gRPC. If omitted, Cloud Run considers the container started once
      # its process is running.
      startup_probe = optional(object({
        # Seconds to wait after container start before the first probe (0-240).
        initial_delay_seconds = optional(number)

        # Seconds after which a single probe attempt times out (1-240; GCP
        # default 1). Must not exceed period_seconds.
        timeout_seconds = optional(number)

        # Seconds between probe attempts (1-240; GCP default 10).
        period_seconds = optional(number)

        # Consecutive failures after which startup is considered failed and the
        # instance is killed (GCP default 3).
        failure_threshold = optional(number)

        # HTTP GET against a path on a port the container listens on; 2xx is
        # success. A worker has no serving port, so name the port explicitly.
        http_get = optional(object({
          # Path to probe, e.g. "/healthz". Defaults to "/".
          path = optional(string, "")

          # Port to probe. A worker pool has no serving port to fall back to, so
          # set the port the worker's health listener binds.
          port = optional(number)

          # Custom header sent with the probe request (e.g. an auth header for a
          # protected health endpoint). At most one header: the pinned Pulumi SDK
          # models a worker-pool probe's headers as a single header, so both
          # engines accept the same manifests; the list widens when the SDK does.
          http_headers = optional(list(object({
            # Header name.
            name = string

            # Header value.
            value = optional(string, "")
          })), [])
        }))

        # TCP connect to a port; a successful connection is success.
        tcp_socket = optional(object({
          # Port to connect to.
          port = optional(number)
        }))

        # Standard gRPC health-check protocol (grpc.health.v1.Health/Check).
        grpc = optional(object({
          # Port the gRPC health service listens on.
          port = optional(number)

          # Service name passed to the health check, letting one server report
          # per-service health. If empty, overall server health is checked.
          service = optional(string, "")
        }))
      }))

      # Probe that monitors a running instance: on failure_threshold
      # consecutive failures the container is restarted. HTTP and gRPC only --
      # Cloud Run rejects TCP liveness probes. If omitted, instances are never
      # health-restarted.
      liveness_probe = optional(object({
        # Seconds to wait after container start before the first probe
        # (0-3600).
        initial_delay_seconds = optional(number)

        # Seconds after which a single probe attempt times out (1-3600; GCP
        # default 1). Must not exceed period_seconds.
        timeout_seconds = optional(number)

        # Seconds between probe attempts (1-3600; GCP default 10).
        period_seconds = optional(number)

        # Consecutive failures after which the container is restarted (GCP
        # default 3).
        failure_threshold = optional(number)

        # HTTP GET against a path on a port the container listens on; 2xx is
        # success.
        http_get = optional(object({
          # Path to probe, e.g. "/healthz". Defaults to "/".
          path = optional(string, "")

          # Port to probe. A worker pool has no serving port to fall back to, so
          # set the port the worker's health listener binds.
          port = optional(number)

          # Custom header sent with the probe request (e.g. an auth header for a
          # protected health endpoint). At most one header: the pinned Pulumi SDK
          # models a worker-pool probe's headers as a single header, so both
          # engines accept the same manifests; the list widens when the SDK does.
          http_headers = optional(list(object({
            # Header name.
            name = string

            # Header value.
            value = optional(string, "")
          })), [])
        }))

        # Standard gRPC health-check protocol (grpc.health.v1.Health/Check).
        grpc = optional(object({
          # Port the gRPC health service listens on.
          port = optional(number)

          # Service name passed to the health check, letting one server report
          # per-service health. If empty, overall server health is checked.
          service = optional(string, "")
        }))
      }))

      # Names of containers this one waits for: this container starts only
      # after the listed containers pass their startup probes. Immutable --
      # changing the order replaces the worker pool.
      depends_on = optional(list(string), [])
    }))

    # Named volumes instances can mount: Cloud SQL sockets, Secret Manager
    # material, scratch space, GCS buckets (FUSE), and NFS shares. A volume
    # is inert until a container mounts it by name.
    volumes = optional(list(object({
      # Volume name referenced by volume_mounts entries.
      name = string

      # Cloud SQL Unix sockets, one per instance, under the mount path
      # (mount at "/cloudsql"; connect via
      # "/cloudsql/<project:region:instance>"). GCP manages the proxying.
      cloud_sql_instance = optional(object({
        # Cloud SQL instance connection names (project:region:instance).
        # Accepts literal values or references to GcpCloudSql resources.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        instances = list(string)
      }))

      # Secret Manager secret versions exposed as files.
      secret = optional(object({
        # The secret: a short name for a secret in the same project or a full
        # resource name (projects/*/secrets/*).
        secret = string

        # Default Unix permission mode for projected files, in decimal (e.g. 292
        # = 0444). If unset, GCP defaults to 0444.
        default_mode = optional(number)

        # Which versions land at which relative paths. If empty, the "latest"
        # version is projected at a file named after the secret.
        items = optional(list(object({
          # Relative path of the file under the volume's mount path.
          path = string

          # Secret version to project: a version number or "latest".
          version = optional(string, "")

          # Unix permission mode for this file, in decimal. Overrides
          # default_mode.
          mode = optional(number)
        })), [])
      }))

      # Ephemeral scratch space, in-memory (counts against the instance's
      # memory limit) or disk-backed.
      empty_dir = optional(object({
        # Backing medium: MEMORY (default -- tmpfs; usage counts against the
        # containers' memory limits) or DISK.
        medium = optional(string, "")

        # Capacity limit with unit suffix, e.g. "512Mi", "2Gi".
        size_limit = optional(string, "")
      }))

      # A GCS bucket mounted via Cloud Storage FUSE; object storage
      # semantics apply (no POSIX locking; renames are copies).
      gcs = optional(object({
        # The bucket to mount. Accepts a literal bucket name or a reference to a
        # GcpGcsBucket resource.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        bucket = string

        # Mount read-only. Recommended unless the worker genuinely writes --
        # concurrent writers through FUSE are easy to get wrong.
        read_only = optional(bool, false)

        # Flags passed to the gcsfuse command mounting this volume, without
        # leading dashes (e.g. "implicit-dirs", "only-dir=media").
        mount_options = optional(list(string), [])
      }))

      # An NFS share (e.g. Filestore) mounted into the instance. Needs VPC
      # access to reach the server.
      nfs = optional(object({
        # Hostname or IP of the NFS server.
        server = string

        # Exported path on the server, e.g. "/share1".
        path = string

        # Mount read-only.
        read_only = optional(bool, false)
      }))
    })), [])

    # Email of the IAM service account the instances run as -- the identity
    # whose permissions the code exercises. Accepts a literal email or a
    # GcpServiceAccount reference. If omitted, the project's Compute Engine
    # default service account is used -- fine for experiments, too broad for
    # production; give real workers a dedicated least-privilege identity.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    service_account = optional(string, "")

    # How many instances run. MANUAL (Google's default) pins the total to
    # manual_instance_count; AUTOMATIC lets Cloud Run move between
    # min_instance_count and max_instance_count on a signal you drive.
    # If omitted, Google runs the pool in MANUAL mode at its default count.
    scaling = optional(object({
      # MANUAL (Google's default): the pool runs exactly manual_instance_count
      # instances. AUTOMATIC: Cloud Run moves the count between
      # min_instance_count and max_instance_count on a signal you drive.
      scaling_mode = optional(string, "")

      # Total instances across all revisions in MANUAL mode. 0 parks the pool
      # (no instances, no bill) without deleting it.
      manual_instance_count = optional(number)

      # Lower bound on instances in AUTOMATIC mode, distributed across
      # revisions by instance_splits.
      min_instance_count = optional(number)

      # Upper bound on instances in AUTOMATIC mode -- the cost circuit
      # breaker.
      max_instance_count = optional(number)
    }))

    # How instances are split across revisions. If empty, every instance
    # runs the latest ready revision -- the right default. Populate for a
    # gradual rollout: percentages must sum to 100 (enforced by the API at
    # deploy time).
    instance_splits = optional(list(object({
      # Assign to the latest ready revision (LATEST) or a named revision
      # (REVISION).
      type = string

      # Revision name for REVISION splits (see spec.revision for deterministic
      # naming).
      revision = optional(string, "")

      # Percent of instances for this revision (0-100). All percents in the
      # list must sum to 100 -- enforced by the API at deploy time. Unset
      # means 0.
      percent = optional(number)
    })), [])

    # Customer-managed encryption key (CMEK) that encrypts the deployed
    # container images: a full crypto key ID
    # (projects/*/locations/*/keyRings/*/cryptoKeys/*) or a GcpKmsKey
    # reference. The key must be in the worker pool's region and the Cloud
    # Run service agent needs encrypter/decrypter on it.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    encryption_key = optional(string, "")

    # What Cloud Run does with running instances if the CMEK is revoked:
    #   PREVENT_NEW -- no new instances start; running ones keep going
    #   SHUTDOWN    -- every instance is stopped after
    #                  encryption_key_shutdown_duration
    # Only meaningful with encryption_key.
    encryption_key_revocation_action = optional(string, "")

    # Grace period before instances are shut down after a key revocation
    # under SHUTDOWN, as a seconds duration in whole hours (e.g. "3600s").
    encryption_key_shutdown_duration = optional(string, "")

    # Explicit name for the next revision, prefixed with the worker pool
    # name (e.g. "orders-worker-v42"). If omitted (recommended), Cloud Run
    # generates one. Pin names only when instance_splits routes by
    # revision; each template change then REQUIRES a new value here.
    revision = optional(string, "")

    # Labels stamped on every REVISION the template creates, as opposed to
    # `labels` on the worker pool object. Same namespace restrictions.
    revision_labels = optional(map(string), {})

    # Annotations stamped on every REVISION the template creates, as
    # opposed to `annotations` on the worker pool object.
    revision_annotations = optional(map(string), {})

    # Private networking for OUTBOUND traffic: Direct VPC egress
    # (network_interfaces -- recommended) or a Serverless VPC Access
    # connector. A worker pool that pulls from Memorystore, Cloud SQL private
    # IP, or an internal service needs this.
    vpc_access = optional(object({
      # Serverless VPC Access connector to route egress through. Full resource
      # name (projects/*/locations/*/connectors/*) or a reference to a
      # GcpServerlessVpcConnector resource.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      connector = optional(string, "")

      # Direct VPC egress: instances get IPs in the subnetwork and reach VPC
      # resources with no connector infrastructure. The subnetwork needs free
      # address space for the instance fleet.
      network_interfaces = optional(list(object({
        # The VPC network. Accepts a literal network name or a reference to a
        # GcpVpcNetwork resource. May be omitted when subnetwork is set (the
        # network is inferred).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        network = optional(string, "")

        # The subnetwork instances draw IPs from. Accepts a literal subnetwork
        # name or a reference to a GcpSubnetwork resource. Must be in the pool's
        # region.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        subnetwork = optional(string, "")

        # Network tags applied to the instances -- how VPC firewall rules select
        # the pool's egress traffic.
        tags = optional(list(string), [])
      })), [])

      # Which egress traffic uses the VPC path: everything (ALL_TRAFFIC) or
      # only RFC1918/private destinations (PRIVATE_RANGES_ONLY -- the
      # default; public egress keeps Cloud Run's own path).
      egress = optional(string, "")
    }))

    # Hardware requirement for GPU workers. Setting an accelerator (e.g.
    # "nvidia-l4") gives every instance one GPU; the containers' resource
    # limits must then meet Cloud Run's GPU minimums (4 CPU / 16Gi).
    node_selector = optional(object({
      # GPU accelerator type each instance gets, e.g. "nvidia-l4". GPU pools
      # need at least "4" CPU / "16Gi" memory and regional GPU quota.
      accelerator = string
    }))

    # Opts a GPU worker pool out of zonal redundancy: instances may be
    # served from a single zone, lowering GPU capacity cost for zonal-failure
    # risk. Only meaningful with node_selector.
    gpu_zonal_redundancy_disabled = optional(bool, false)

    # Launch-stage gate. Set BETA (or ALPHA) only when the spec uses preview
    # Cloud Run features the default GA stage rejects; the value is a
    # declaration, not a feature switch. The provider enum admits more
    # values, but Cloud Run supports only ALPHA, BETA, and GA -- this list
    # encodes the API truth.
    launch_stage = optional(string, "")

    # Binary Authorization: only images that pass the policy's attestation
    # checks may deploy. The project default policy or a named one.
    binary_authorization = optional(object({
      # Evaluate deploys against the project's default Binary Authorization
      # policy.
      use_default = optional(bool, false)

      # Evaluate deploys against a specific platform policy
      # (projects/*/platforms/cloudRun/policies/*).
      policy = optional(string, "")

      # Justification recorded when a break-glass deploy bypasses the policy.
      breakglass_justification = optional(string, "")
    }))

    # Prevents the worker pool from being destroyed while true. Defaults to
    # true (Google's posture): a destroy fails until this is set to false.
    # Both engines send the value explicitly so a manifest that never
    # mentions it behaves the same everywhere.
    deletion_protection = optional(bool)

    # What happens to the worker pool when this resource is destroyed
    # (after deletion_protection allows the destroy):
    #   "" / "DELETE" -- the worker pool is deleted (default)
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the worker pool leaves management but keeps running
    deletion_policy = optional(string, "")
  })
}
