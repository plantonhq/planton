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
  description = "AwsBatchJobDefinition specification"
  type = object({
    # The AWS region the job definition is registered in. A queue can only
    # run definitions registered in its own region.
    # Example: "us-west-2", "eu-west-1".
    region = string

    # The ECS-based container the job runs: image, command, sizing,
    # identities, logging, and storage. Exactly one of container or eks is
    # set -- this arm targets EC2/Fargate compute environments.
    container = optional(object({
      # The container image, as a full reference: "<repository>:<tag>" or
      # "<repository>@<digest>". Up to 255 characters. Private ECR images
      # require execution_role (Fargate) or the instance role (EC2) to carry
      # pull permissions; other private registries use
      # repository_credentials_secret_arn. The image architecture must match
      # the compute it lands on (ARM images need ARM compute).
      # Example: "123456789012.dkr.ecr.us-west-2.amazonaws.com/etl:1.4.2".
      image = string

      # Command override (Docker CMD) -- or the arguments to the image's
      # ENTRYPOINT. Supports "Ref::<key>" placeholders resolved from
      # spec.parameters and per-job SubmitJob overrides.
      # Example: ["python", "process.py", "Ref::input_path"].
      command = optional(list(string), [])

      # vCPUs reserved for the job (the VCPU resource requirement). EC2 jobs
      # take whole numbers; Fargate jobs take the Fargate sizes: 0.25, 0.5, 1,
      # 2, 4, 8, or 16, each paired with a valid memory_mib range.
      vcpus = optional(number, 0)

      # Memory hard limit in MiB (the MEMORY resource requirement) -- the job
      # is killed when it exceeds it. Fargate pairs memory with vcpus (e.g.
      # 0.25 vCPU allows 512-2048 MiB in 1024-MiB steps).
      memory_mib = optional(number, 0)

      # GPUs reserved for the job (the GPU resource requirement) -- pinned
      # exclusively to this job's container. EC2 GPU instance types only
      # (use an ECS_AL2_NVIDIA image type on the compute environment);
      # Fargate offers no GPUs.
      gpus = optional(number, 0)

      # The IAM role the JOB'S CODE assumes at runtime -- its identity for
      # every AWS API call the workload makes (S3, DynamoDB, ...). Distinct
      # from execution_role by design: setup permissions and runtime
      # permissions should never be one role. Reference an AwsIamRole's
      # role_arn output or pass a literal role ARN.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      job_role = optional(string, "")

      # The IAM role the ECS AGENT assumes to set the job up: pull private
      # images, resolve secrets, write logs. REQUIRED for Fargate job
      # definitions; on EC2 the instance profile usually covers it. Reference
      # an AwsIamRole's role_arn output or pass a literal role ARN.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      execution_role = optional(string, "")

      # Plain-text environment variables (name -> value). For anything
      # sensitive use secrets instead -- environment values are visible to
      # anyone who can describe the job definition. Names must not start with
      # "AWS_BATCH" (reserved by the service).
      environment = optional(map(string), {})

      # Secret environment variables (name -> the ARN of an AWS Secrets
      # Manager secret or SSM Parameter Store parameter). The agent resolves
      # each reference at job start -- via execution_role on Fargate, the
      # instance role on EC2 -- so the value never appears in the job
      # definition. Append ":<json-key>::" to a Secrets Manager ARN to inject
      # one key of a JSON secret.
      secrets = optional(map(string), {})

      # Log driver configuration override. When unset, Batch sends container
      # logs to CloudWatch under the /aws/batch/job log group with zero
      # configuration -- set this only to change drivers or options.
      log_configuration = optional(object({
        # The log driver: "awslogs" (CloudWatch -- the default even without this
        # block), "splunk", "fluentd", "gelf", "syslog", "journald", or
        # "json-file". The compute environment's instances must have the driver
        # available (Fargate supports awslogs and splunk).
        log_driver = string

        # Driver-specific options -- e.g. awslogs-group / awslogs-stream-prefix
        # to redirect awslogs away from the /aws/batch/job default, or
        # splunk-url for Splunk.
        options = optional(map(string), {})

        # Driver options whose values come from Secrets Manager / SSM (name ->
        # ARN) -- e.g. a splunk-token. Resolved by the agent at job start, never
        # stored in the definition.
        secret_options = optional(map(string), {})
      }))

      # Mounts of the job's named volumes into the container's filesystem.
      mount_points = optional(list(object({
        # The name of a volume declared in container.volumes.
        source_volume = string

        # The path inside the container where the volume mounts.
        # Example: "/mnt/data".
        container_path = string

        # Mount read-only.
        read_only = optional(bool, false)
      })), [])

      # Named volumes containers mount via mount_points: EFS file systems
      # (durable, shared, Fargate-supported) or container-instance host paths
      # (EC2 only).
      volumes = optional(list(object({
        # The volume's name, referenced by container.mount_points.
        name = string

        # Back the volume with an EFS file system.
        efs = optional(object({
          # The EFS file system backing the volume. Reference an
          # AwsElasticFileSystem's file_system_id output or pass a literal file
          # system ID (e.g. "fs-0123456789abcdef0").
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          file_system_id = string

          # The path within the file system to mount as the volume root. Ignored
          # when access_point_id is set (the access point defines the root).
          # Default: "/".
          root_directory = optional(string, "")

          # Mount through this EFS access point -- the recommended pattern: the
          # access point pins the POSIX identity and root path, so jobs cannot
          # wander the file system. Requires transit encryption (which the modules
          # enable whenever an access point or IAM authorization is used).
          # Reference an AwsEfsAccessPoint's access_point_id output or pass a
          # literal ID (e.g. "fsap-0123456789abcdef0").
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          access_point_id = optional(string, "")

          # Authorize the mount with the job's IAM role (job_role must carry
          # elasticfilesystem:ClientMount/ClientWrite). Requires transit
          # encryption, which the modules enable automatically.
          iam_authorization = optional(bool, false)
        }))

        # Back the volume with a path on the container instance (EC2 only).
        # Example: "/mnt/scratch".
        host_path = optional(string, "")
      })), [])

      # Resource limits (ulimits) for the container, e.g. raise "nofile" for
      # connection-heavy jobs. EC2 only -- Fargate rejects ulimit overrides.
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

      # Linux host-level settings: device mappings, tmpfs mounts, shared
      # memory, and swap. EC2 only.
      linux_parameters = optional(object({
        # Run an init process (PID 1) inside the container to reap zombie
        # processes -- maps to Docker's --init. Recommended for images whose
        # entrypoint spawns child processes.
        init_process_enabled = optional(bool, false)

        # Host devices mapped into the container.
        devices = optional(list(object({
          # The device path on the container instance. Example: "/dev/xvdf".
          host_path = string

          # The path the device is exposed at inside the container. Defaults to
          # host_path when omitted.
          container_path = optional(string, "")

          # The cgroup permissions granted: any of "READ", "WRITE", "MKNOD".
          # Defaults to all three when empty.
          permissions = optional(list(string), [])
        })), [])

        # The /dev/shm size in MiB. Raise it for scientific/ML workloads that
        # use shared memory heavily.
        shared_memory_size_mib = optional(number, 0)

        # The container's total swap budget in MiB. 0 disables swap; omit to
        # inherit the instance's configuration. Swap must be enabled on the
        # instance (via the compute environment's launch template) to take
        # effect.
        max_swap_mib = optional(number, 0)

        # Swap aggressiveness, 0-100 (0 = swap only under pressure, 100 = swap
        # aggressively). Only meaningful when max_swap_mib is positive. AWS
        # default: 60.
        swappiness = optional(number, 0)

        # tmpfs (in-memory) mounts inside the container.
        tmpfs = optional(list(object({
          # The mount path inside the container. Example: "/tmp/scratch".
          container_path = string

          # The tmpfs size in MiB.
          size_mib = optional(number, 0)

          # Mount options (e.g. "noexec", "nosuid", "uid=1000").
          mount_options = optional(list(string), [])
        })), [])
      }))

      # Run the container with elevated host permissions (root-equivalent).
      # EC2 only; reserved for host-integration workloads.
      privileged = optional(bool, false)

      # Run the container process as this user ("uid", "uid:gid", or a
      # username present in the image).
      user = optional(string, "")

      # Mount the container's root filesystem read-only -- writable paths must
      # come from volumes. A strong hardening default for jobs that only read
      # inputs and write to mounted storage.
      readonly_root_filesystem = optional(bool, false)

      # Credentials for pulling the image from a private NON-ECR registry: the
      # ARN of an AWS Secrets Manager secret holding {"username","password"}
      # -- a reference resolved at job start, never the credential itself.
      # ECR images need no credentials; grant the execution/instance role pull
      # access instead.
      repository_credentials_secret_arn = optional(string, "")

      # CPU architecture and OS for Fargate jobs. Set cpu_architecture to
      # "ARM64" to run on Graviton (cheaper per vCPU; images must be built for
      # arm64).
      runtime_platform = optional(object({
        # "X86_64" (default) or "ARM64" (Graviton -- cheaper per vCPU; the image
        # must be built for arm64).
        cpu_architecture = optional(string, "")

        # OS family; "LINUX" (the default) for almost everything. Windows
        # containers on Fargate use the WINDOWS_SERVER_* families.
        operating_system_family = optional(string, "")
      }))

      # The Fargate platform version (e.g. "1.4.0" or "LATEST"). Fargate only;
      # omit to let AWS pick LATEST.
      fargate_platform_version = optional(string, "")

      # Give the Fargate job's ENI a public IP -- required for internet access
      # from PUBLIC subnets without a NAT gateway. Fargate only; jobs in
      # private subnets should route through NAT instead.
      assign_public_ip = optional(bool, false)

      # Ephemeral scratch storage for the Fargate job, in GiB (21-200).
      # Fargate includes 20 GiB at no charge; set this only when the job needs
      # more (large intermediate files). Fargate only.
      ephemeral_storage_gib = optional(number, 0)
    }))

    # Where the job may run: "EC2" (default when empty) and/or "FARGATE".
    # A Fargate job definition additionally requires container.execution_role
    # and uses the Fargate-only knobs (platform version, public IP,
    # ephemeral storage, runtime platform); the EC2-only knobs (GPUs,
    # privileged, ulimits, Linux parameters) are rejected for it.
    platform_capabilities = optional(list(string), [])

    # Default placeholder values for the job definition's parameter
    # substitution: a command like ["python", "run.py", "Ref::dataset"]
    # resolves "Ref::dataset" from this map, and SubmitJob can override each
    # key per job -- one definition, many parameterized runs.
    parameters = optional(map(string), {})

    # How failed attempts are retried. When unset, jobs get a single attempt.
    retry_strategy = optional(object({
      # Total attempts, 1-10 (1 = no retries). Attempts re-run the whole job;
      # the workload must be idempotent or checkpoint-aware.
      attempts = optional(number, 0)

      # Ordered conditions evaluated against a FAILED attempt's exit code and
      # status reasons; the FIRST match decides RETRY or EXIT, and a failure
      # matching nothing behaves like EXIT. Up to 5 conditions. The classic
      # use: RETRY on "Host EC2*" status reasons (Spot reclaims) while EXITing
      # on real application failures. CREATE-TIME per revision: changing these
      # registers a new revision.
      evaluate_on_exit = optional(list(object({
        # The decision when this condition matches: "RETRY" (consume another
        # attempt) or "EXIT" (fail the job immediately).
        action = string

        # Glob match on the container's decimal exit code. Only a trailing "*"
        # wildcard is allowed. Example: "137" (SIGKILL / OOM), "1*".
        on_exit_code = optional(string, "")

        # Glob match on the attempt's reason (the container runtime's message,
        # e.g. "DockerTimeoutError*"). Only a trailing "*" wildcard.
        on_reason = optional(string, "")

        # Glob match on the attempt's status reason (Batch's own message, e.g.
        # "Host EC2*" for Spot interruptions). Only a trailing "*" wildcard.
        on_status_reason = optional(string, "")
      })), [])
    }))

    # The hard wall-clock limit per job attempt. Attempts running longer are
    # terminated by Batch (and retried per retry_strategy). SubmitJob can
    # override it per job.
    timeout = optional(object({
      # Seconds an attempt may run before Batch terminates it. Minimum 60.
      attempt_duration_seconds = optional(number, 0)
    }))

    # The job's scheduling priority WITHIN a fair-share queue (0-9999, higher
    # is sooner within the job's share). Only consulted when the queue has a
    # scheduling policy; FIFO queues ignore it.
    scheduling_priority = optional(number, 0)

    # Propagate the job definition's tags to the ECS task (and, from there,
    # to cost reports and IAM tag conditions on the running task).
    propagate_tags = optional(bool, false)

    # Whether registering a new revision deregisters the previous one
    # (marks it INACTIVE). The default (true) keeps exactly one ACTIVE
    # revision -- the one this resource manages. Set false when out-of-band
    # consumers (a manual SubmitJob against a pinned revision) must keep
    # running old revisions.
    deregister_on_new_revision = optional(bool)

    # The Batch-on-EKS pod the job runs: containers, pod networking, and
    # Kubernetes-native volumes. Exactly one of container or eks is set --
    # this arm targets compute environments attached to an EKS cluster
    # (eks_configuration on AwsBatchComputeEnvironment). Jobs are submitted
    # the same way; Batch translates the definition into a pod on the
    # attached cluster.
    eks = optional(object({
      # The pod's main containers (1-10). Batch watches these to decide job
      # success: the job completes when every main container exits.
      containers = list(object({
        # The container image, as a full reference: "<repository>:<tag>" or
        # "<repository>@<digest>". The image architecture must match the
        # cluster's node architecture.
        # Example: "123456789012.dkr.ecr.us-west-2.amazonaws.com/genomics:2.1".
        image = string

        # The container's name -- a Kubernetes DNS-1123 label (lowercase
        # alphanumerics and hyphens, max 63 chars). Required by Kubernetes when
        # the pod has more than one container; Batch names a lone unnamed
        # container "default".
        name = optional(string, "")

        # Entrypoint override (Kubernetes command / Docker ENTRYPOINT).
        # Supports "Ref::<key>" placeholders resolved from spec.parameters.
        command = optional(list(string), [])

        # Arguments to the entrypoint (Kubernetes args / Docker CMD). Supports
        # "Ref::<key>" placeholders resolved from spec.parameters.
        args = optional(list(string), [])

        # Plain-text environment variables (name -> value). Names must not
        # start with "AWS_BATCH" (reserved by the service). For secrets, mount
        # a Kubernetes secret volume instead -- EKS jobs have no ECS-style
        # secrets injection.
        env = optional(map(string), {})

        # When Kubernetes pulls the image: "Always", "IfNotPresent", or
        # "Never". AWS defaults to Always, matching Kubernetes for :latest
        # tags -- pinned tags commonly use IfNotPresent to spare registry
        # traffic.
        image_pull_policy = optional(string, "")

        # The container's compute sizing -- Kubernetes resource requests and
        # limits. Batch schedules the job by these (its EKS counterpart of the
        # container arm's vcpus/memory_mib/gpus).
        resources = optional(object({
          # Hard caps (Kubernetes limits): keys "cpu" (e.g. "1", "500m"),
          # "memory" (e.g. "2Gi", "512Mi"), and "nvidia.com/gpu" for GPU nodes.
          # Batch treats limits as the job's sizing; GPU quantities must be
          # whole numbers.
          limits = optional(map(string), {})

          # Scheduling reservations (Kubernetes requests), same keys as limits.
          # When both are set for a key, request must not exceed limit; Batch
          # fills a missing request from the limit.
          requests = optional(map(string), {})
        }))

        # The container's Kubernetes securityContext -- run-as identity and
        # privilege hardening.
        security_context = optional(object({
          # Run the container process as this numeric UID. 0 (root) is a legal
          # explicit value -- UNSET leaves the image's own USER in effect, which
          # is why presence matters here.
          run_as_user = optional(number)

          # Run the container process as this numeric GID. As with run_as_user,
          # 0 is legal and distinct from unset.
          run_as_group = optional(number)

          # Have Kubernetes REJECT the pod at start if the effective user
          # resolves to root -- an assertion, not an identity setting.
          run_as_non_root = optional(bool, false)

          # Whether the process may gain more privileges than its parent
          # (setuid binaries, file capabilities). UNSET means Kubernetes'
          # default (allowed, unless the container is otherwise restricted);
          # an explicit false is the hardening posture.
          allow_privilege_escalation = optional(bool)

          # Run the container privileged (root-equivalent on the node).
          # Default false, like Kubernetes.
          privileged = optional(bool, false)

          # Mount the container's root filesystem read-only -- writable paths
          # must come from volumes.
          read_only_root_file_system = optional(bool, false)
        }))

        # Mounts of the pod's declared volumes into this container's
        # filesystem.
        volume_mounts = optional(list(object({
          # The name of a volume declared in eks.volumes.
          name = string

          # The path inside the container where the volume mounts.
          # Example: "/mnt/data".
          mount_path = string

          # Mount read-only.
          read_only = optional(bool, false)
        })), [])
      }))

      # Init containers (0-10), run sequentially to completion before the
      # main containers start -- setup steps like fetching data or waiting
      # for a dependency.
      init_containers = optional(list(object({
        # The container image, as a full reference: "<repository>:<tag>" or
        # "<repository>@<digest>". The image architecture must match the
        # cluster's node architecture.
        # Example: "123456789012.dkr.ecr.us-west-2.amazonaws.com/genomics:2.1".
        image = string

        # The container's name -- a Kubernetes DNS-1123 label (lowercase
        # alphanumerics and hyphens, max 63 chars). Required by Kubernetes when
        # the pod has more than one container; Batch names a lone unnamed
        # container "default".
        name = optional(string, "")

        # Entrypoint override (Kubernetes command / Docker ENTRYPOINT).
        # Supports "Ref::<key>" placeholders resolved from spec.parameters.
        command = optional(list(string), [])

        # Arguments to the entrypoint (Kubernetes args / Docker CMD). Supports
        # "Ref::<key>" placeholders resolved from spec.parameters.
        args = optional(list(string), [])

        # Plain-text environment variables (name -> value). Names must not
        # start with "AWS_BATCH" (reserved by the service). For secrets, mount
        # a Kubernetes secret volume instead -- EKS jobs have no ECS-style
        # secrets injection.
        env = optional(map(string), {})

        # When Kubernetes pulls the image: "Always", "IfNotPresent", or
        # "Never". AWS defaults to Always, matching Kubernetes for :latest
        # tags -- pinned tags commonly use IfNotPresent to spare registry
        # traffic.
        image_pull_policy = optional(string, "")

        # The container's compute sizing -- Kubernetes resource requests and
        # limits. Batch schedules the job by these (its EKS counterpart of the
        # container arm's vcpus/memory_mib/gpus).
        resources = optional(object({
          # Hard caps (Kubernetes limits): keys "cpu" (e.g. "1", "500m"),
          # "memory" (e.g. "2Gi", "512Mi"), and "nvidia.com/gpu" for GPU nodes.
          # Batch treats limits as the job's sizing; GPU quantities must be
          # whole numbers.
          limits = optional(map(string), {})

          # Scheduling reservations (Kubernetes requests), same keys as limits.
          # When both are set for a key, request must not exceed limit; Batch
          # fills a missing request from the limit.
          requests = optional(map(string), {})
        }))

        # The container's Kubernetes securityContext -- run-as identity and
        # privilege hardening.
        security_context = optional(object({
          # Run the container process as this numeric UID. 0 (root) is a legal
          # explicit value -- UNSET leaves the image's own USER in effect, which
          # is why presence matters here.
          run_as_user = optional(number)

          # Run the container process as this numeric GID. As with run_as_user,
          # 0 is legal and distinct from unset.
          run_as_group = optional(number)

          # Have Kubernetes REJECT the pod at start if the effective user
          # resolves to root -- an assertion, not an identity setting.
          run_as_non_root = optional(bool, false)

          # Whether the process may gain more privileges than its parent
          # (setuid binaries, file capabilities). UNSET means Kubernetes'
          # default (allowed, unless the container is otherwise restricted);
          # an explicit false is the hardening posture.
          allow_privilege_escalation = optional(bool)

          # Run the container privileged (root-equivalent on the node).
          # Default false, like Kubernetes.
          privileged = optional(bool, false)

          # Mount the container's root filesystem read-only -- writable paths
          # must come from volumes.
          read_only_root_file_system = optional(bool, false)
        }))

        # Mounts of the pod's declared volumes into this container's
        # filesystem.
        volume_mounts = optional(list(object({
          # The name of a volume declared in eks.volumes.
          name = string

          # The path inside the container where the volume mounts.
          # Example: "/mnt/data".
          mount_path = string

          # Mount read-only.
          read_only = optional(bool, false)
        })), [])
      })), [])

      # Whether the pod uses the NODE's network namespace (Kubernetes
      # hostNetwork). UNSET means AWS's default, which is TRUE for Batch
      # pods -- the opposite of the plain-Kubernetes default -- so an
      # explicit false is a real choice: it gives the pod its own namespace
      # (required for VPC-CNI pod networking with security groups per pod).
      host_network = optional(bool)

      # The pod's DNS resolution policy. AWS defaults to "ClusterFirst"
      # (resolve through the cluster's DNS first); "Default" inherits the
      # NODE's resolution; "ClusterFirstWithHostNet" is the cluster-first
      # behavior for pods running with host_network.
      dns_policy = optional(string, "")

      # The Kubernetes service account the pod runs as -- the EKS-native way
      # to grant the JOB's code AWS permissions (IRSA / Pod Identity), the
      # counterpart of the container arm's job_role.
      service_account_name = optional(string, "")

      # Labels applied to the pod's metadata -- Kubernetes selectors,
      # cost-allocation, and policy engines key off these.
      # Example: {"team": "genomics", "workload": "batch"}.
      pod_labels = optional(map(string), {})

      # Names of Kubernetes imagePullSecrets in the job's namespace, for
      # pulling from private non-ECR registries (the EKS counterpart of the
      # container arm's repository_credentials_secret_arn). ECR images need
      # no secret -- the node role's pull access covers them.
      image_pull_secret_names = optional(list(string), [])

      # Share one process namespace across the pod's containers (Kubernetes
      # shareProcessNamespace) -- lets a sidecar signal or observe the main
      # container's processes. Default false, like plain Kubernetes.
      share_process_namespace = optional(bool, false)

      # Kubernetes-native volumes the pod's containers mount by name:
      # emptyDir scratch space, node hostPath directories, or Kubernetes
      # secrets. (EFS rides the cluster's CSI driver and static
      # PersistentVolumes -- outside the job definition's surface.)
      volumes = optional(list(object({
        # The volume's name (a DNS-1123 label), referenced by containers'
        # volume_mounts.
        name = string

        # Scratch space that lives and dies with the job's pod.
        empty_dir = optional(object({
          # Where the scratch lives: unset backs it with node storage; "Memory"
          # backs it with tmpfs (fast, counts against the container's memory
          # sizing).
          medium = optional(string, "")

          # The scratch size cap, as a Kubernetes quantity.
          # Example: "1Gi", "500Mi".
          size_limit = string
        }))

        # A directory on the NODE's filesystem (Kubernetes hostPath.path) --
        # data outlives the pod but is pinned to whichever node ran it.
        # Example: "/mnt/scratch".
        host_path = optional(string, "")

        # Project a Kubernetes secret (from the job's namespace) into the
        # volume.
        secret = optional(object({
          # The name of the Kubernetes secret in the job's namespace (the
          # namespace comes from the compute environment's eks_configuration).
          secret_name = string

          # Mount successfully even when the secret does not exist yet (an empty
          # volume) instead of failing the pod.
          optional = optional(bool, false)
        }))
      })), [])
    }))
  })
}
