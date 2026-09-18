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
  description = "GcpCloudRun specification"
  type = object({
    # The GCP project the service is created in. Accepts a literal project ID
    # or a reference to a GcpProject resource. If omitted, the provider's
    # default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Region the service is deployed in, e.g. "us-central1". Immutable.
    # Cloud Run services are regional; the one exception is multi-region
    # services (multi_region_settings), which are created through the
    # special "global" region with the actual serving regions listed in
    # multi_region_settings.regions.
    region = string

    # Name of the Cloud Run service in GCP. Immutable. If not specified,
    # defaults to metadata.name. Must be 1-63 characters: lowercase letters,
    # digits, and hyphens; starting with a letter.
    service_name = optional(string, "")

    # Human-readable description of the service, shown in the console.
    description = optional(string, "")

    # Labels applied to the service object. User labels are shared with
    # Google's billing system, so they can filter or break down billing
    # charges by team, component, or environment. Keys and values in the
    # `run.googleapis.com`, `cloud.googleapis.com`, `serving.knative.dev`,
    # and `autoscaling.knative.dev` namespaces are rejected by the API.
    labels = optional(map(string), {})

    # The containers that make up one instance of the service. The first
    # container conventionally serves requests; additional containers are
    # sidecars (log collectors, auth proxies, service meshes) that share the
    # instance's network namespace (localhost) and volumes. Exactly one
    # container may expose a port. Use depends_on for startup ordering.
    # Deployment pipelines inject the built image into containers whose image
    # is left BLANK (the image-slot contract); authored images are untouched.
    containers = list(object({
      # Name of the container. Required when the service runs more than one
      # container (depends_on refers to these names). If omitted for a
      # single-container service, Cloud Run assigns one.
      name = optional(string, "")

      # Container image URL, e.g. "us-docker.pkg.dev/project/repo/app:1.0.0".
      # Pin a digest or immutable tag for repeatable deploys — Cloud Run
      # resolves the image to a digest at revision creation, so a moving tag
      # only takes effect on the NEXT deploy anyway.
      #
      # Cloud Run pulls PRIVATE images only from Artifact Registry (or the
      # legacy Container Registry) in a project its service agent can read;
      # public Docker Hub and GHCR images deploy directly. There is no field
      # here for a registry login and Google accepts none. For a private image
      # in another registry, push it to Artifact Registry, or declare a
      # GcpArtifactRegistryRepo in REMOTE_REPOSITORY mode that proxies that
      # registry and point this URL at it.
      image = string

      # Entrypoint array — overrides the image's ENTRYPOINT. Not executed in a
      # shell; variable references are not expanded.
      command = optional(list(string), [])

      # Arguments to the entrypoint — overrides the image's CMD.
      args = optional(list(string), [])

      # Environment variables. Each entry carries either a literal value or a
      # Secret Manager reference resolved at instance start.
      env = optional(list(object({
        # Variable name, e.g. "DATABASE_URL". Must not start with a digit.
        name = string

        # Literal value. Fine for configuration; never place credentials here —
        # use value_from_secret so the material stays in Secret Manager.
        value = optional(string, "")

        # Secret Manager reference resolved into the variable at instance start.
        value_from_secret = optional(object({
          # The secret: a short name for a secret in the same project ("my-secret")
          # or a full resource name (projects/*/secrets/*) for cross-project reads.
          secret = string

          # Secret version to resolve: a version number or "latest". If unset,
          # GCP requires an explicit version for env vars — "latest" is the common
          # choice, at the cost of new instances silently picking up rotations.
          version = optional(string, "")
        }))
      })), [])

      # The port this container listens on for requests. At most ONE container
      # in the service may declare a port (that container receives traffic).
      # If no container declares one, Cloud Run injects PORT=8080 into the
      # first container.
      ports = optional(object({
        # Port number the container listens on. Cloud Run injects it as $PORT.
        container_port = optional(number)

        # Protocol selector: "http1" (default) or "h2c" for end-to-end HTTP/2
        # (required for serving gRPC streams).
        name = optional(string, "")
      }))

      # CPU and memory for this container, plus the CPU-allocation levers.
      # If omitted, Cloud Run defaults apply (1 CPU, 512Mi).
      resources = optional(object({
        # CPU limit, e.g. "1", "2", "4", "8" or a fraction like "0.5"/"500m"
        # (fractions below 1 require max_instance_request_concurrency <= 1 for
        # some features). GPU services need at least "4".
        cpu = optional(string, "")

        # Memory limit with unit suffix, e.g. "512Mi", "2Gi". Minimums scale
        # with CPU (1 CPU >= 128Mi, 4 CPU >= 2Gi, GPU >= 16Gi).
        memory = optional(string, "")

        # Keep CPU allocated only while requests are in flight (true — the
        # default, request-based billing) or allocate it for the instance's
        # whole lifetime (false — instance-based billing, required for real
        # background work between requests).
        cpu_idle = optional(bool)

        # Temporarily boost CPU during instance startup, cutting cold-start
        # latency for JIT-heavy runtimes (JVM, .NET) at no idle cost.
        startup_cpu_boost = optional(bool, false)
      }))

      # Volumes (declared at the service level) mounted into this container's
      # filesystem.
      volume_mounts = optional(list(object({
        # Name of a volume declared in spec.volumes.
        name = string

        # Absolute path in the container to mount at. Cloud SQL volumes must
        # mount at "/cloudsql".
        mount_path = string

        # Path WITHIN the volume to mount instead of its root — e.g. mount only
        # one secret item or one bucket directory. Relative path; empty mounts
        # the volume root.
        sub_path = optional(string, "")
      })), [])

      # Working directory for the entrypoint. If omitted, the image's WORKDIR
      # is used.
      working_dir = optional(string, "")

      # Probe that gates instance start: the container receives no traffic
      # (and depends_on waiters stay blocked) until this succeeds. Supports
      # HTTP, TCP, and gRPC checks. If omitted, Cloud Run performs a default
      # TCP check on the container port.
      startup_probe = optional(object({
        # Seconds to wait after container start before the first probe (0-240).
        initial_delay_seconds = optional(number)

        # Seconds after which a single probe attempt times out (1-240; GCP
        # default 1). Must not exceed period_seconds.
        timeout_seconds = optional(number)

        # Seconds between probe attempts (1-240; GCP default 10). Bounded
        # further by the 240-second startup window.
        period_seconds = optional(number)

        # Consecutive failures after which startup is considered failed and the
        # instance is killed (GCP default 3).
        failure_threshold = optional(number)

        # HTTP GET against a path on the container; 2xx is success.
        http_get = optional(object({
          # Path to probe, e.g. "/healthz". Defaults to "/".
          path = optional(string, "")

          # Port to probe. If unset, the container's serving port is used.
          port = optional(number)

          # Custom headers sent with the probe request (e.g. an auth header for a
          # protected health endpoint).
          http_headers = optional(list(object({
            # Header name.
            name = string

            # Header value.
            value = optional(string, "")
          })), [])
        }))

        # TCP connect to a port; a successful connection is success. Startup
        # is the only probe type Cloud Run accepts TCP on.
        tcp_socket = optional(object({
          # Port to connect to. If unset, the container's serving port is used.
          port = optional(number)
        }))

        # Standard gRPC health-check protocol (grpc.health.v1.Health/Check).
        grpc = optional(object({
          # Port the gRPC health service listens on. If unset, the container's
          # serving port is used.
          port = optional(number)

          # Service name passed to the health check, letting one server report
          # per-service health. If empty, overall server health is checked.
          service = optional(string, "")
        }))
      }))

      # Probe that monitors a running instance: on failure_threshold
      # consecutive failures the container is restarted. Supports HTTP and
      # gRPC checks only — the message has no TCP arm because Cloud Run
      # rejects TCP liveness probes. If omitted, instances are never
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

        # HTTP GET against a path on the container; 2xx is success.
        http_get = optional(object({
          # Path to probe, e.g. "/healthz". Defaults to "/".
          path = optional(string, "")

          # Port to probe. If unset, the container's serving port is used.
          port = optional(number)

          # Custom headers sent with the probe request (e.g. an auth header for a
          # protected health endpoint).
          http_headers = optional(list(object({
            # Header name.
            name = string

            # Header value.
            value = optional(string, "")
          })), [])
        }))

        # Standard gRPC health-check protocol (grpc.health.v1.Health/Check).
        grpc = optional(object({
          # Port the gRPC health service listens on. If unset, the container's
          # serving port is used.
          port = optional(number)

          # Service name passed to the health check, letting one server report
          # per-service health. If empty, overall server health is checked.
          service = optional(string, "")
        }))
      }))

      # Names of containers this one waits for: this container starts only
      # after the listed containers pass their startup probes. The mechanism
      # that makes proxy/collector sidecars start before (and stop after) the
      # serving container.
      depends_on = optional(list(string), [])

      # Base image for automatic base-image updates on source-deployed
      # services (see spec.build_config): Google patches the OS/runtime
      # layers of the running image without a redeploy. A serverless
      # runtimes image URI, e.g.
      # "us-central1-docker.pkg.dev/serverless-runtimes/google-24-full/runtimes/nodejs24".
      base_image_uri = optional(string, "")

      # Probe that gates whether this container receives traffic: instances
      # whose readiness check fails are pulled from serving without being
      # restarted (unlike liveness, which restarts). Supports HTTP and gRPC
      # checks only, and unlike the other probes has no initial delay — it
      # starts with the container and runs for the instance's lifetime.
      readiness_probe = optional(object({
        # Seconds after which a single probe attempt times out (GCP default 1).
        # Must not exceed period_seconds.
        timeout_seconds = optional(number)

        # Seconds between probe attempts (GCP default 10).
        period_seconds = optional(number)

        # Consecutive failures after which the instance is pulled from serving
        # (GCP default 3). The instance is NOT restarted — that is the liveness
        # probe's job.
        failure_threshold = optional(number)

        # HTTP GET against a path on the container; 2xx is success.
        http_get = optional(object({
          # Path to probe, e.g. "/ready". Defaults to "/".
          path = optional(string, "")

          # Port to probe. If unset, the container's serving port is used.
          port = optional(number)
        }))

        # Standard gRPC health-check protocol (grpc.health.v1.Health/Check).
        grpc = optional(object({
          # Port the gRPC health service listens on. If unset, the container's
          # serving port is used.
          port = optional(number)

          # Service name passed to the health check, letting one server report
          # per-service health. If empty, overall server health is checked.
          service = optional(string, "")
        }))
      }))

      # Marks this container as the sandbox supervisor: the one process in
      # the instance allowed to launch the isolated sandboxes declared in
      # spec.sandbox_templates (through the Cloud Run sandbox CLI/API). The
      # pattern for agent workloads that run untrusted, model-generated code:
      # the supervisor orchestrates, each sandbox executes in isolation.
      sandbox_launcher = optional(bool, false)
    }))

    # Named volumes instances can mount: Cloud SQL sockets, Secret Manager
    # material, in-memory or disk scratch space, GCS buckets (FUSE), and NFS
    # shares. A volume is inert until a container mounts it by name via
    # volume_mounts.
    volumes = optional(list(object({
      # Volume name referenced by volume_mounts entries.
      name = string

      # Cloud SQL Unix sockets, one per instance, under the mount path
      # (mount at "/cloudsql"; connect via
      # "/cloudsql/<project:region:instance>"). GCP manages the proxying —
      # no sidecar, no VPC needed.
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
        # = 0444). If unset, GCP defaults to 0444 (read-only for all).
        default_mode = optional(number)

        # Which versions land at which relative paths. If empty, the "latest"
        # version is projected at a file named after the secret.
        items = optional(list(object({
          # Relative path of the file under the volume's mount path.
          path = string

          # Secret version to project: a version number or "latest".
          version = optional(string, "")

          # Unix permission mode for this file, in decimal. Overrides default_mode.
          mode = optional(number)
        })), [])
      }))

      # Ephemeral scratch space, in-memory (counts against the instance's
      # memory limit) or disk-backed.
      empty_dir = optional(object({
        # Backing medium: MEMORY (default — tmpfs; usage counts against the
        # containers' memory limits) or DISK.
        medium = optional(string, "")

        # Capacity limit with unit suffix, e.g. "512Mi", "2Gi". For MEMORY
        # volumes, leave unset to let GCP cap it sensibly relative to instance
        # memory.
        size_limit = optional(string, "")
      }))

      # A GCS bucket mounted via Cloud Storage FUSE. Requires the GEN2
      # execution environment; object storage semantics apply (no POSIX
      # locking; renames are copies).
      gcs = optional(object({
        # The bucket to mount. Accepts a literal bucket name or a reference to a
        # GcpGcsBucket resource.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        bucket = string

        # Mount read-only. Recommended unless the service genuinely writes —
        # concurrent writers through FUSE are easy to get wrong.
        read_only = optional(bool, false)

        # Flags passed to the gcsfuse command mounting this volume, without
        # leading dashes (e.g. "implicit-dirs", "only-dir=media",
        # "file-cache-max-size-mb=512"). The tuning surface for FUSE caching
        # and directory semantics.
        mount_options = optional(list(string), [])
      }))

      # An NFS share (e.g. Filestore) mounted into the instance. Requires
      # the GEN2 execution environment and VPC access to reach the server.
      nfs = optional(object({
        # Hostname or IP of the NFS server, e.g. a Filestore instance's IP.
        server = string

        # Exported path on the server, e.g. "/share1".
        path = string

        # Mount read-only.
        read_only = optional(bool, false)
      }))
    })), [])

    # Email of the IAM service account the revisions run as — the identity
    # whose permissions the code exercises when calling other GCP APIs.
    # Accepts a literal email or a reference to a GcpServiceAccount resource.
    # If omitted, the project's Compute Engine default service account is
    # used — fine for experiments, too broad for production; give real
    # services a dedicated least-privilege identity.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    service_account = optional(string, "")

    # Per-revision instance bounds: how far a single revision scales in and
    # out. If omitted, the service scales zero-to-100. min_instance_count > 0
    # keeps instances warm to eliminate cold starts — at idle cost.
    scaling = optional(object({
      # Instances kept warm even with zero traffic. 0 (default) scales to
      # zero — cheapest, with cold starts; 1+ eliminates cold starts at idle
      # cost.
      min_instance_count = optional(number)

      # Upper bound on instances for this revision — the cost/overload
      # circuit breaker. If unset, GCP's default cap (100) applies.
      max_instance_count = optional(number)
    }))

    # Service-level scaling posture across ALL revisions. The MANUAL mode
    # pins total instance count regardless of traffic — an emergency brake
    # or a load-test lever. Distinct from the per-revision bounds in
    # `scaling`.
    service_scaling = optional(object({
      # AUTOMATIC (default): Cloud Run scales with traffic. MANUAL: the
      # service runs exactly manual_instance_count instances regardless of
      # traffic — a load-test lever or emergency brake.
      scaling_mode = optional(string, "")

      # Exact total instance count in MANUAL mode.
      manual_instance_count = optional(number)

      # Service-level minimum instances, distributed across serving revisions
      # (unlike scaling.min_instance_count, which is per revision). Useful
      # during gradual rollouts so warm capacity follows the traffic split.
      min_instance_count = optional(number)

      # Service-level maximum across ALL revisions combined — the total-cost
      # circuit breaker during rollouts, when two revisions serve at once and
      # per-revision caps (scaling.max_instance_count) would otherwise double
      # the worst case.
      max_instance_count = optional(number)
    }))

    # Maximum concurrent requests one instance serves before Cloud Run scales
    # out (1-1000). If unset, GCP defaults to 80 for instances with >= 1 CPU
    # and 1 below that. Lower values isolate requests (memory-heavy or
    # CPU-bound work); higher values improve utilization for I/O-bound
    # services.
    max_instance_request_concurrency = optional(number)

    # Maximum time a request may run before Cloud Run cancels it, in seconds
    # (1-3600). If unset, GCP defaults to 300 (5 minutes). Also bounds
    # startup: an instance that cannot serve within this window is killed.
    timeout_seconds = optional(number)

    # Sandbox generation the revisions execute in. GEN2 (recommended) offers
    # full Linux compatibility (required for GCS/NFS volumes and network file
    # systems) at slightly slower cold starts; GEN1 starts faster with a
    # gVisor-restricted syscall surface. If unset, GCP picks per workload.
    execution_environment = optional(string, "")

    # Routes requests from the same client to the same instance on a
    # best-effort basis (session affinity cookie). Useful for local caches;
    # never a correctness guarantee — instances still scale in.
    session_affinity = optional(bool, false)

    # Customer-managed encryption key (CMEK) used to encrypt the deployed
    # container images. Accepts a full crypto key ID
    # (projects/*/locations/*/keyRings/*/cryptoKeys/*) or a reference to a
    # GcpKmsKey resource. The key must be in the same region as the service,
    # and the Cloud Run service agent needs encrypter/decrypter on it.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    encryption_key = optional(string, "")

    # Explicit name for the next revision. Must be prefixed with the service
    # name (e.g. "my-api-v42"). If omitted (recommended), Cloud Run
    # auto-generates revision names. Pin revision names only when the traffic
    # block routes by revision name — deterministic names make declarative
    # blue/green possible, but each template change then REQUIRES a new
    # unique value here.
    revision = optional(string, "")

    # Private networking for OUTBOUND traffic: route egress into a VPC either
    # through Direct VPC egress (network_interfaces — no extra infrastructure,
    # recommended) or a Serverless VPC Access connector. Inbound restriction
    # is `ingress`, not this.
    vpc_access = optional(object({
      # Serverless VPC Access connector to route egress through — the legacy
      # mechanism, still required for some org constraints. Full resource name
      # (projects/*/locations/*/connectors/*) or a reference to a
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
        # name or a reference to a GcpSubnetwork resource. Must be in the
        # service's region.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        subnetwork = optional(string, "")

        # Network tags applied to the instances — how VPC firewall rules select
        # Cloud Run egress traffic.
        tags = optional(list(string), [])
      })), [])

      # Which egress traffic uses the VPC path: everything (ALL_TRAFFIC) or
      # only RFC1918/private destinations (PRIVATE_RANGES_ONLY — the default;
      # public egress keeps Cloud Run's own path).
      egress = optional(string, "")
    }))

    # Hardware requirements for GPU inference workloads. Setting an
    # accelerator (e.g. "nvidia-l4") gives every instance one GPU; the
    # containers' resource limits must then meet Cloud Run's GPU minimums
    # (4 CPU / 16Gi recommended).
    node_selector = optional(object({
      # GPU accelerator type each instance gets, e.g. "nvidia-l4". GPU
      # services need GEN2-class resources (at least "4" CPU / "16Gi" memory)
      # and regional GPU quota.
      accelerator = string
    }))

    # Opts a GPU service out of zonal redundancy: instances may be served
    # from a single zone, which lowers the price of GPU capacity in exchange
    # for zonal-failure risk. Only meaningful when node_selector is set.
    gpu_zonal_redundancy_disabled = optional(bool, false)

    # Which network paths may reach the service: everything, only internal
    # traffic, or internal plus Cloud Load Balancing. Services behind the
    # composed HTTPS load balancer should use INTERNAL_LOAD_BALANCER so the
    # default run.app URL stops accepting public traffic.
    ingress = optional(string, "")

    # If true, grants roles/run.invoker to allUsers so unauthenticated
    # callers can reach the service — the standard shape for public HTTP
    # APIs and websites. If false, callers must present an IAM-authorized
    # identity token.
    allow_unauthenticated = optional(bool, false)

    # Disables the IAM run.routes.invoke permission check entirely. Unlike
    # allow_unauthenticated (which grants access THROUGH IAM), this switches
    # the check off — required by some org policies that forbid allUsers
    # grants while still serving public traffic. Set at most one of the two.
    invoker_iam_disabled = optional(bool, false)

    # Additional audience values accepted in the OAuth/OIDC tokens of
    # authenticated callers (beyond the default service URL). Lets callers
    # mint one token for a stable custom audience instead of the run.app URL.
    custom_audiences = optional(list(string), [])

    # How traffic is split across revisions. If empty, 100% of traffic goes
    # to the latest ready revision — the right default for most services.
    # Populate for gradual rollouts: percentages must sum to 100 (enforced
    # by the API at deploy time), and tagged entries get stable
    # <tag>---<host> preview URLs that receive no percentage-based traffic
    # unless assigned some.
    traffic = optional(list(object({
      # Route to the latest ready revision (LATEST) or a specific named
      # revision (REVISION).
      type = string

      # Revision name for REVISION targets (see spec.revision for
      # deterministic naming).
      revision = optional(string, "")

      # Percent of traffic for this target (0-100). All percents in the
      # traffic block must sum to 100 — GCP enforces this at deploy time.
      percent = optional(number)

      # Tag that gives this target a stable preview URL
      # (https://<tag>---<service-host>) receiving only requests addressed to
      # it — how a canary is smoke-tested before percent-based traffic moves.
      tag = optional(string, "")
    })), [])

    # Launch-stage gate for the service. Set BETA (or ALPHA) only when the
    # spec uses preview Cloud Run features that GCP rejects at the default
    # GA stage; the value is a declaration, not a feature switch. The
    # provider enum admits more values (UNIMPLEMENTED, PRELAUNCH,
    # EARLY_ACCESS, DEPRECATED), but Cloud Run itself supports only
    # ALPHA/BETA/GA — this list encodes the API truth, deliberately.
    launch_stage = optional(string, "")

    # Binary Authorization: only container images that pass the policy's
    # attestation checks may deploy. Use the project default policy or name
    # a specific one.
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

    # Prevents the service from being destroyed while true. Defaults to true
    # (matching GCP's posture): a delete of this resource fails until this is
    # set to false. Deleting a service tears down its endpoint and every
    # revision; keep this on for anything real.
    deletion_protection = optional(bool)

    # Annotations on the SERVICE object: unstructured metadata read by
    # external tools (never queryable, never behavioral for Cloud Run
    # itself). Keys in the `run.googleapis.com`, `cloud.googleapis.com`,
    # `serving.knative.dev`, and `autoscaling.knative.dev` namespaces are
    # rejected by the API. For metadata stamped on each REVISION instead,
    # use revision_annotations.
    annotations = optional(map(string), {})

    # Labels stamped on every REVISION the template creates (visible on
    # revision objects and in billing breakdowns), as opposed to `labels`,
    # which lands on the service object itself. Same namespace
    # restrictions as service labels.
    revision_labels = optional(map(string), {})

    # Annotations stamped on every REVISION the template creates, as
    # opposed to `annotations`, which lands on the service object. Same
    # namespace restrictions.
    revision_annotations = optional(map(string), {})

    # Deploy-from-source: have Cloud Build produce the serving image from
    # source code (the Cloud Run functions build path) instead of deploying
    # a prebuilt image. The built artifact deploys as the service's
    # container image; containers[].base_image_uri pairs with
    # enable_automatic_updates for managed base-image patching.
    build_config = optional(object({
      # Cloud Storage URI of the source code to build, e.g.
      # "gs://my-bucket/source.zip". Uploading source is the caller's job;
      # this field tells the build where it lives.
      source_location = optional(string, "")

      # Name of the function (as defined in the source code) to execute.
      # Defaults to the resource name suffix; the API falls back to a
      # function named "function" when the named one is not found.
      function_target = optional(string, "")

      # Artifact Registry URI where the built image is stored, e.g.
      # "us-docker.pkg.dev/project/repo/image".
      image_uri = optional(string, "")

      # Base image used to build the function — a serverless runtimes image
      # (see containers[].base_image_uri for the serving-side pairing).
      base_image = optional(string, "")

      # Whether the function receives automatic base-image updates: Google
      # patches OS/runtime layers of the deployed image without a redeploy.
      enable_automatic_updates = optional(bool, false)

      # Build-time environment variables visible to the build process (NOT
      # to the running service — runtime env lives on the containers).
      environment_variables = optional(map(string), {})

      # Cloud Build Custom Worker Pool to run the build in, as
      # "projects/{project}/locations/{region}/workerPools/{workerPool}".
      # For builds that must run inside a private network perimeter.
      worker_pool = optional(string, "")

      # Service account the BUILD runs as, in the full resource form
      # "projects/{projectId}/serviceAccounts/{email}" (note: not a bare
      # email — this is the build-time identity, distinct from the runtime
      # service_account).
      service_account = optional(string, "")
    }))

    # Puts the service behind Identity-Aware Proxy: callers authenticate
    # with Google identities and IAP enforces access policy before requests
    # reach the service — Google's managed login wall, with no code changes.
    iap_enabled = optional(bool, false)

    # Disables the default *.run.app URL, leaving the service reachable only
    # through custom domains or the load-balancer path. Removes the
    # often-forgotten second front door on locked-down services.
    default_uri_disabled = optional(bool, false)

    # Disables ALL health checking (startup and liveness probes) for the
    # revisions. An escape hatch for workloads whose serving model breaks
    # the probe contract — not a performance optimization.
    health_check_disabled = optional(bool, false)

    # Multi-region service: one service identity serving from several
    # regions at once. Requires region = "global"; the actual serving
    # regions are listed here.
    multi_region_settings = optional(object({
      # The regions the multi-region service deploys to, e.g.
      # ["us-central1", "europe-west1"].
      regions = list(string)
    }))

    # What happens to the Cloud Run service when this resource is
    # destroyed:
    #   "" / "DELETE" -- the service is deleted (default)
    #   "PREVENT"     -- destroy operations fail while this is set
    #   "ABANDON"     -- the service is removed from management but left
    #                    running in GCP
    deletion_policy = optional(string, "")

    # Sandbox templates the instance's supervisor container (the one with
    # sandbox_launcher) may launch on demand: isolated, short-lived
    # containers for executing untrusted or model-generated code beside
    # the serving container without exposing it. Each template names the
    # image and startup shape a sandbox runs with; the supervisor picks a
    # template by name at launch time. Requires exactly one container with
    # sandbox_launcher set.
    sandbox_templates = optional(list(object({
      # Template name the supervisor launches by, a DNS label (RFC 1123):
      # lowercase letters, digits, hyphens; starts and ends alphanumeric.
      name = string

      # Container image the sandbox runs, e.g.
      # "us-docker.pkg.dev/project/repo/sandbox:1.0.0". A bare name without a
      # registry host is pulled from Docker Hub.
      image = string

      # Entrypoint array, not run through a shell. Empty uses the image's
      # ENTRYPOINT.
      command = optional(list(string), [])

      # Arguments to the entrypoint. Empty uses the image's CMD.
      args = optional(list(string), [])

      # Environment variables set in the sandbox: literal values only.
      env = optional(list(object({
        # Variable name, e.g. "PYTHONUNBUFFERED". Must not start with a digit.
        name = string

        # Literal value (up to 32768 characters). Never place credentials here.
        value = optional(string, "")
      })), [])

      # Volumes (declared in spec.volumes) mounted into the sandbox's
      # filesystem.
      volume_mounts = optional(list(object({
        # Name of a volume declared in spec.volumes.
        name = string

        # Absolute path in the container to mount at. Cloud SQL volumes must
        # mount at "/cloudsql".
        mount_path = string

        # Path WITHIN the volume to mount instead of its root — e.g. mount only
        # one secret item or one bucket directory. Relative path; empty mounts
        # the volume root.
        sub_path = optional(string, "")
      })), [])

      # Working directory for the entrypoint. Empty uses the image's WORKDIR.
      working_dir = optional(string, "")
    })), [])

    # Resource Manager tags bound to the service at creation, as a map of
    # tagKeys/{tag_key_id} to tagValues/{tag_value_id} — the tag bindings
    # that organization policies, IAM conditions, and cost reports key on.
    # Immutable: changing the map replaces the service (Cloud Run applies
    # tags only at create), so plan tag changes as a redeploy.
    resource_manager_tags = optional(map(string), {})
  })
}
