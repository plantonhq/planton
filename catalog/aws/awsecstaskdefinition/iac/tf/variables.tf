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
  description = "AwsEcsTaskDefinition specification"
  type = object({
    # The AWS region the task definition is registered in. A task definition
    # is a regional object: a service can only run revisions registered in
    # its own region.
    # Example: "us-west-2", "eu-west-1".
    region = string

    # The containers the task runs. Most tasks run one application container;
    # add sidecars (a log router, an OpenTelemetry collector, a proxy) as
    # additional entries and order their startup with depends_on. At least
    # one container must be essential -- when an essential container exits,
    # the whole task stops. Deployment pipelines inject the built image into
    # containers whose image is left BLANK (the image-slot contract);
    # authored images are untouched.
    containers = list(object({
      # The container's name, unique within the task definition. Referenced by
      # an ECS service's load_balancers.container_name, by sibling containers'
      # depends_on, and used as the CloudWatch log stream prefix under the
      # shared task log group.
      name = string

      # The container image, as a full reference: "<repository>:<tag>" or
      # "<repository>@<digest>". Private ECR images require execution_role to
      # carry pull permissions; other private registries use
      # repository_credentials.
      # Example: "123456789012.dkr.ecr.us-west-2.amazonaws.com/api:1.4.2".
      image = string

      # Whether this container is essential. When an essential container
      # exits, ECS stops the whole task; non-essential sidecars can exit or
      # fail without killing the application container. AWS default: true.
      # Optional so an explicit "this sidecar is not essential" (false) is
      # distinguishable from unset.
      essential = optional(bool)

      # CPU units reserved for this container (1024 = 1 vCPU). On Fargate this
      # subdivides the task-level cpu between containers (optional -- unset
      # containers share what is left); on EC2 it drives bin-packing.
      cpu = optional(number, 0)

      # Hard memory limit for this container, in MiB -- the container is
      # killed when it exceeds it. On Fargate the task-level memory already
      # caps the task; set per-container limits to fence sidecars off from
      # the application's share.
      memory = optional(number, 0)

      # Soft memory reservation, in MiB: the scheduler reserves this much but
      # lets the container burst up to memory (or the task limit). Set
      # reservation at the expected footprint and the limit at the tolerable
      # ceiling.
      memory_reservation = optional(number, 0)

      # Ports the container exposes. On awsvpc networking the container port
      # IS the host port (each task has its own ENI). Name a port to make it
      # referenceable from Service Connect.
      port_mappings = optional(list(object({
        # The port the application listens on inside the container.
        # Example: 8080.
        container_port = optional(number, 0)

        # Layer-4 protocol: "tcp" (default) or "udp".
        protocol = optional(string, "")

        # A name for this port, referenced by Service Connect
        # (service_connect.services[].port_name) and by sibling tooling. Name
        # the port whenever the service participates in Service Connect.
        name = optional(string, "")

        # The application protocol ECS uses for Service Connect telemetry and
        # routing: "http", "http2", or "grpc". Only meaningful on named ports.
        app_protocol = optional(string, "")
      })), [])

      # Entry point override (Docker ENTRYPOINT). Leave empty to use the
      # image's own.
      entry_point = optional(list(string), [])

      # Command override (Docker CMD) -- or the arguments to entry_point.
      # Leave empty to use the image's own.
      command = optional(list(string), [])

      # Working directory override for the command.
      working_directory = optional(string, "")

      # Plain-text environment variables (name -> value), written into the task
      # definition where anyone who can describe it reads them -- and kept in
      # every revision for good, since revisions are immutable. Configuration
      # only; a credential goes in secret_environment (or secrets).
      environment = optional(map(string), {})

      # Secret environment variables backed by secrets YOU already own (name ->
      # the ARN of an AWS Secrets Manager secret or SSM Parameter Store
      # parameter). The ECS agent resolves each reference at task start using
      # execution_role, so the value never appears in the task definition.
      # Append ":<json-key>::" to a Secrets Manager ARN to inject one key of a
      # JSON secret.
      secrets = optional(map(string), {})

      # Secret environment variables whose VALUES this component keeps in AWS
      # Secrets Manager for you (name -> value). Per entry it creates one secret
      # named "<family>/<container>/<name>", stores the value, and attaches a
      # resource policy that lets only execution_role read it; the container's
      # secrets list then carries that secret's ARN pinned to the stored
      # version, so the task definition holds a reference, never the value. A
      # changed value registers a new revision (a deploy is the rotation), and
      # destroying the task definition deletes the secrets with no recovery
      # window -- the value's source of truth is whoever supplied it here.
      # Requires execution_role.
      secret_environment = optional(map(string), {})

      # Environment files loaded from S3 (each entry an S3 object ARN of a
      # .env file). Applied before environment/secrets; later sources win.
      # execution_role must be able to read the objects.
      environment_files = optional(list(string), [])

      # Container-level health check ECS runs INSIDE the container (distinct
      # from any load-balancer target health check). Sibling containers can
      # gate their startup on it via depends_on condition "HEALTHY", and the
      # task reports unhealthy when it fails.
      health_check = optional(object({
        # The probe command, in Docker exec form. The first element is "CMD"
        # (exec directly) or "CMD-SHELL" (run through the shell).
        # Example: ["CMD-SHELL", "curl -f http://localhost:8080/healthz || exit 1"].
        command = list(string)

        # Seconds between probes. AWS default: 30.
        interval_seconds = optional(number, 0)

        # Seconds before an unanswered probe counts as a failure. AWS default: 5.
        timeout_seconds = optional(number, 0)

        # Consecutive failures before the container is marked unhealthy. AWS
        # default: 3.
        retries = optional(number, 0)

        # Grace period after container start during which failures do not count
        # -- give slow-booting apps room before probes matter. AWS default: 0.
        start_period_seconds = optional(number, 0)
      }))

      # Startup ordering against sibling containers: wait for a dependency to
      # START, become HEALTHY (requires its health_check), COMPLETE, or exit
      # with SUCCESS before this container starts. Shutdown runs in reverse.
      depends_on = optional(list(object({
        # The name of the sibling container this one waits for.
        container_name = string

        # The state to wait for: "START" (the dependency has started),
        # "HEALTHY" (its health_check passes -- the dependency must define one),
        # "COMPLETE" (it exited, any code), or "SUCCESS" (it exited 0).
        condition = optional(string, "")
      })), [])

      # Mounts of the task's named volumes into this container's filesystem.
      mount_points = optional(list(object({
        # The name of a volume declared in spec.volumes.
        source_volume = string

        # The path inside the container where the volume mounts.
        # Example: "/var/data".
        container_path = string

        # Mount read-only.
        read_only = optional(bool, false)
      })), [])

      # Per-container log configuration override. When set, this container
      # logs with exactly this driver/options instead of the task-level
      # logging default -- the escape hatch for shipping to Splunk, Fluentd,
      # or a FireLens router.
      log_configuration = optional(object({
        # The log driver: "awslogs" (CloudWatch), "awsfirelens" (route through a
        # FireLens sibling container), "splunk", "fluentd", "gelf", "syslog",
        # "journald", or "json-file" (EC2 only).
        log_driver = string

        # Driver-specific options (e.g. awslogs-group / awslogs-region /
        # awslogs-stream-prefix for awslogs; Name/host/port for awsfirelens
        # outputs).
        options = optional(map(string), {})

        # Driver options whose values come from Secrets Manager / SSM (name ->
        # ARN) -- e.g. a Splunk HEC token. Resolved by the agent at task start
        # via execution_role.
        secret_options = optional(map(string), {})
      }))

      # Marks this container as a FireLens log router (fluentbit or fluentd).
      # Sibling containers then use log_configuration with driver
      # "awsfirelens" to route their logs through it.
      firelens_configuration = optional(object({
        # The router type: "fluentbit" (the AWS-recommended lightweight router)
        # or "fluentd".
        type = optional(string, "")

        # Router options -- e.g. enable-ecs-log-metadata: "true", or
        # config-file-type/config-file-value for a custom parsing config.
        options = optional(map(string), {})
      }))

      # Credentials for pulling the image from a private NON-ECR registry:
      # the ARN of an AWS Secrets Manager secret holding {"username","password"}
      # -- a reference the ECS agent resolves at task start, never the
      # credential itself. ECR images need no credentials; grant
      # execution_role pull access instead.
      repository_credentials_secret_arn = optional(string, "")

      # Run the container process as this user ("uid", "uid:gid", or a
      # username present in the image). Fargate: uid-based values only.
      user = optional(string, "")

      # Mount the container's root filesystem read-only -- writable paths must
      # come from volumes. A strong hardening default for stateless services.
      readonly_root_filesystem = optional(bool, false)

      # Give the container elevated privileges on the host (EC2 launch type
      # only; not supported on Fargate). Reserved for host-integration agents.
      privileged = optional(bool, false)

      # Run an init process (PID 1) inside the container to reap zombie
      # processes -- maps to Docker's --init. Recommended for images whose
      # entrypoint spawns child processes.
      init_process_enabled = optional(bool, false)

      # Number of GPUs reserved for this container (EC2 GPU instance types
      # only; Fargate does not offer GPUs).
      gpu_count = optional(number, 0)

      # Resource limits (ulimits) for the container, e.g. raise "nofile" for
      # high-connection proxies. Fargate supports only "nofile" overrides
      # (default soft 1024 / hard 65535 there).
      ulimits = optional(list(object({
        # The limit name: "nofile" (open files -- the common override), "core",
        # "cpu", "data", "fsize", "locks", "memlock", "msgqueue", "nice",
        # "nproc", "rss", "rtprio", "rttime", "sigpending", or "stack".
        name = string

        # The soft limit, enforced but raisable by the process up to hard_limit.
        soft_limit = optional(number, 0)

        # The hard ceiling.
        hard_limit = optional(number, 0)
      })), [])

      # Docker labels applied to the container (key -> value) -- consumed by
      # on-host tooling and some log routers.
      docker_labels = optional(map(string), {})

      # Seconds ECS waits for this container's depends_on conditions before
      # giving up on starting it.
      start_timeout_seconds = optional(number, 0)

      # Seconds ECS waits after SIGTERM before SIGKILL at shutdown. Default
      # 30; raise it for workloads that need a longer graceful drain (capped
      # at 120 on Fargate).
      stop_timeout_seconds = optional(number, 0)

      # Restart the container in place when it exits, without replacing the
      # whole task -- faster recovery for a crashing sidecar than a full task
      # cycle.
      restart_policy = optional(object({
        # Enable in-place restarts for this container.
        enabled = optional(bool, false)

        # Exit codes that should NOT trigger a restart (e.g. 0 for a batch
        # sidecar that finished cleanly).
        ignored_exit_codes = optional(list(number), [])

        # Minimum seconds the container must run before a restart attempt is
        # made (60-1800). AWS default: 300.
        restart_attempt_period_seconds = optional(number, 0)
      }))
    }))

    # Launch types the task definition validates against: "FARGATE",
    # "EC2", "EXTERNAL" (ECS Anywhere), and/or "MANAGED_INSTANCES" (ECS
    # Managed Instances capacity). AWS registers the definition for those
    # environments and rejects incompatible settings at registration time
    # instead of at run time. When empty, both modules register for
    # ["FARGATE"] -- the serverless launch type that needs no instance
    # management.
    requires_compatibilities = optional(list(string), [])

    # Total CPU for the task, in CPU units (1024 = 1 vCPU). REQUIRED for
    # Fargate, where it selects the task size: 256, 512, 1024, 2048, 4096,
    # 8192, or 16384. Optional on EC2 (containers bin-pack by their own cpu).
    # Example: 512.
    cpu = optional(number, 0)

    # Total memory for the task, in MiB. REQUIRED for Fargate and constrained
    # by cpu (e.g. 256 CPU pairs with 512-2048 MiB, 1024 CPU with 2048-8192
    # MiB). Optional on EC2, where it caps the sum of container memory.
    # Example: 1024.
    memory = optional(number, 0)

    # Docker networking mode: "awsvpc" (each task gets its own ENI and
    # private IP -- required for Fargate and the modern default everywhere),
    # "bridge" (EC2 docker0 bridge), "host" (EC2 host network stack), or
    # "none". Default: "awsvpc".
    network_mode = optional(string, "")

    # The IAM role the ECS AGENT assumes to set the task up: pull private
    # images from ECR, fetch secrets for the secrets/repository_credentials
    # fields, and create/write CloudWatch log streams. Reference an
    # AwsIamRole's role_arn output or pass a literal role ARN. AWS REQUIRES
    # it at registration time for Fargate tasks that use the awslogs driver
    # -- which the default logging wiring does -- and in practice whenever
    # the task uses ECR images or secrets.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    execution_role = optional(string, "")

    # The IAM role the APPLICATION code assumes at runtime -- the task's
    # identity for every AWS API call the workload itself makes (S3, SQS,
    # DynamoDB, ...). Distinct from execution_role by design: the agent's
    # setup permissions and the app's runtime permissions should never be
    # one role. Omit when the app calls no AWS APIs.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    task_role = optional(string, "")

    # CPU architecture and OS the task runs on. Set cpu_architecture to
    # "ARM64" to run on Graviton (Fargate ARM pricing is ~20% below x86 for
    # the same vCPU/memory) -- the images must be built for arm64.
    runtime_platform = optional(object({
      # "X86_64" (default) or "ARM64" (Graviton -- cheaper per vCPU, images
      # must be multi-arch or arm64-built).
      cpu_architecture = optional(string, "")

      # OS family; "LINUX" (default) for almost everything. Windows families
      # ("WINDOWS_SERVER_2019_CORE", "WINDOWS_SERVER_2019_FULL",
      # "WINDOWS_SERVER_2022_CORE", "WINDOWS_SERVER_2022_FULL") are supported
      # on Fargate for Windows containers.
      operating_system_family = optional(string, "")
    }))

    # Ephemeral scratch storage shared by the task's containers, in GiB
    # (21-200). Fargate default: 20 GiB at no charge; set this only when the
    # workload needs more (image processing, builds, large temp files).
    ephemeral_storage_gib = optional(number, 0)

    # Named volumes containers mount via mount_points. Fargate supports EFS
    # and S3-backed volumes (durable, shared across tasks) and the implicit
    # ephemeral storage; host-path and Docker volumes are EC2-only. A
    # configure_at_launch volume declares only the NAME here -- the
    # AwsEcsService running the task supplies the backing (managed EBS) at
    # deployment time through its volume_configuration.
    volumes = optional(list(object({
      # The volume's name, referenced by containers' mount_points.
      name = string

      # Back the volume with an EFS file system.
      efs = optional(object({
        # The EFS file system backing the volume. Reference an AwsElasticFileSystem
        # resource or pass a literal file system ID (e.g. "fs-0123456789abcdef0").
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        file_system_id = string

        # The path within the file system to mount as the volume root. Ignored
        # when access_point_id is set (the access point defines the root).
        # Default: "/".
        root_directory = optional(string, "")

        # Mount through this EFS access point -- the recommended pattern:
        # the access point pins the POSIX identity and root path, so tasks
        # cannot wander the file system. Reference an AwsEfsAccessPoint resource
        # or pass a literal access point ID (e.g. "fsap-0123456789abcdef0").
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        access_point_id = optional(string, "")

        # Authorize the mount with the task's IAM role (execution/task role
        # must carry elasticfilesystem:ClientMount/ClientWrite). Requires
        # transit encryption, which the modules enable automatically.
        iam_authorization = optional(bool, false)

        # The port for the encrypted transit tunnel between host and EFS.
        # 0 (unset) lets AWS choose an ephemeral port -- the right answer
        # unless a host firewall requires a fixed one.
        transit_encryption_port = optional(number, 0)
      }))

      # Back the volume with a path on the container instance (EC2 launch
      # type only). Example: "/mnt/data".
      host_path = optional(string, "")

      # Defer the volume's configuration to launch time: the AwsEcsService
      # running this task supplies the backing through its
      # volume_configuration (managed EBS -- a fresh, service-owned EBS
      # volume per task) under this same volume name. Set it true and leave
      # every backing here unset; the service side names the volume in
      # spec.volume_configuration.name.
      configure_at_launch = optional(bool, false)

      # Back the volume with a Docker volume on the container instance (EC2
      # launch type only) -- named Docker-managed storage, optionally via a
      # volume driver plugin.
      docker = optional(object({
        # Provision the volume automatically if it does not already exist.
        # Only meaningful with scope "shared".
        autoprovision = optional(bool, false)

        # The Docker volume driver. Default: "local". Must match a driver
        # installed on the container instance.
        driver = optional(string, "")

        # Driver-specific options passed at volume creation.
        driver_opts = optional(map(string), {})

        # Custom metadata labels applied to the volume.
        labels = optional(map(string), {})

        # The volume's lifetime: "task" (AWS default -- destroyed when the task
        # stops) or "shared" (persists on the instance across tasks).
        scope = optional(string, "")
      }))

      # Back the volume with an S3 bucket mounted as a file system
      # (Mountpoint for Amazon S3) -- read-heavy shared data (models,
      # reference datasets) without provisioning EFS.
      s3files = optional(object({
        # The S3 bucket (by ARN) mounted as the volume. Reference an
        # AwsS3Bucket's bucket_arn output or pass a literal bucket ARN.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        file_system_arn = string

        # An S3 access point ARN to mount through instead of the bucket
        # directly -- scopes the mount to the access point's policy.
        access_point_arn = optional(string, "")

        # The path within the bucket mounted as the volume root. Default: "/".
        root_directory = optional(string, "")

        # The port for encrypted transit between host and mount target.
        # 0 (unset) lets AWS choose.
        transit_encryption_port = optional(number, 0)
      }))
    })), [])

    # Default CloudWatch logging for every container that does not declare
    # its own log_configuration. Enabled by default: the modules create ONE
    # log group named "/ecs/<family>" (30-day retention unless overridden)
    # and each container logs under its own name as the stream prefix --
    # so a task's containers land in one predictable place with zero
    # configuration.
    logging = optional(object({
      # Disable the default log wiring entirely. Containers without their own
      # log_configuration then produce no logs -- almost never what you want.
      disabled = optional(bool, false)

      # Use an existing CloudWatch log group instead of creating one.
      # Reference an AwsCloudwatchLogGroup's log_group_name output or pass a
      # literal group name. When unset, the modules create a group named
      # "/ecs/<family>" and manage its lifecycle with the task definition.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      log_group = optional(string, "")

      # Retention, in days, for the auto-created log group. Ignored when
      # log_group references an existing group (that group owns its own
      # retention).
      retention_days = optional(number)
    }))

    # Keep old revisions registered when this resource is destroyed, instead
    # of deregistering every revision of the family. Useful when other
    # consumers (a scheduled task, a manual RunTask) may still reference
    # older revisions.
    skip_destroy = optional(bool, false)

    # Enable AWS Fault Injection Service (FIS) actions against this task's
    # containers -- the opt-in that lets chaos experiments (CPU stress,
    # network latency, process kill) target the task. Off by default; only
    # enable on definitions you deliberately run experiments against.
    enable_fault_injection = optional(bool, false)

    # IPC namespace sharing for the task's containers: "host" (share the
    # instance's IPC namespace -- weakest isolation), "task" (containers
    # share one namespace within the task), or "none" (each container
    # isolated). EC2 tasks only -- Fargate rejects it at registration.
    # Unset keeps Docker's default (private namespace per container).
    ipc_mode = optional(string, "")

    # Process namespace sharing: "host" (containers see the instance's
    # processes -- weakest isolation) or "task" (containers within the task
    # share one PID namespace -- what lets a sidecar observe the app's
    # processes). On Fargate only "task" is supported (platform 1.4+).
    # Unset keeps Docker's default (private namespace per container).
    pid_mode = optional(string, "")

    # Task-level placement constraints, evaluated at RunTask/service
    # scheduling time on EC2 container instances (at most 10). EC2 tasks
    # only -- Fargate rejects them at registration. Service-level
    # constraints live on the AwsEcsService; declare here only what is
    # intrinsic to the task itself (e.g. an instance-attribute expression
    # every consumer must honor).
    placement_constraints = optional(list(object({
      # The constraint type. "memberOf" (restrict placement to instances
      # matching the expression) is the only type AWS supports at the task
      # definition level.
      type = optional(string, "")

      # A cluster query language expression, e.g.
      # "attribute:ecs.instance-type =~ t2.*" or
      # "attribute:ecs.availability-zone in [us-west-2a, us-west-2b]".
      expression = string
    })), [])
  })
}
