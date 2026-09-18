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
  description = "AwsCodeBuildProject specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # source defines where the primary build input comes from and how to
    # fetch it. Its source_identifier must stay empty — identifiers exist only
    # to name secondary sources.
    source = object({
      # type is the source provider.
      #   GITHUB:                GitHub.com repository
      #   BITBUCKET:             Bitbucket Cloud repository
      #   CODECOMMIT:            AWS CodeCommit repository
      #   CODEPIPELINE:          Source provided by CodePipeline (no location needed)
      #   GITHUB_ENTERPRISE:     GitHub Enterprise Server
      #   GITLAB:                GitLab.com repository
      #   GITLAB_SELF_MANAGED:   Self-hosted GitLab instance
      #   NO_SOURCE:             No source — buildspec must be provided inline
      #   S3:                    S3 bucket containing source archive
      type = string

      # location is the source code repository URL or S3 path.
      # Required for all types except CODEPIPELINE and NO_SOURCE.
      # Format depends on type:
      #   GITHUB/BITBUCKET/GITLAB: https URL (e.g., https://github.com/owner/repo.git)
      #   CODECOMMIT: HTTPS clone URL
      #   S3: bucket/path (e.g., my-bucket/source.zip)
      #   GITHUB_ENTERPRISE: HTTPS URL of the repo
      location = optional(string, "")

      # buildspec is the build specification, either as an inline YAML string or
      # a path relative to the source root (e.g., "buildspec.yml").
      # If omitted, CodeBuild looks for buildspec.yml at the source root.
      # Required when source type is NO_SOURCE.
      buildspec = optional(string, "")

      # git_clone_depth limits the Git clone depth. 0 means full clone.
      # Only applicable for Git-based source types.
      git_clone_depth = optional(number, 0)

      # git_submodules_config controls Git submodule fetching during the source
      # download phase. Presence of the block opts into explicit submodule
      # handling. Only supported for BITBUCKET, CODECOMMIT, GITHUB, and
      # GITHUB_ENTERPRISE sources.
      git_submodules_config = optional(object({
        # fetch_submodules fetches all Git submodules during the source download
        # phase when true.
        fetch_submodules = optional(bool, false)
      }))

      # insecure_ssl skips TLS certificate verification when fetching source —
      # only for GitHub Enterprise / self-managed GitLab instances with
      # self-signed certificates. Never enable against public providers.
      insecure_ssl = optional(bool, false)

      # report_build_status reports build start and finish status back to the
      # source provider (e.g., GitHub commit status checks).
      # Only applicable for GITHUB, BITBUCKET, GITHUB_ENTERPRISE, GITLAB,
      # GITLAB_SELF_MANAGED.
      report_build_status = optional(bool, false)

      # build_status_config customizes how the commit status is reported —
      # the status-check context label and the URL it links to. Only applicable
      # to the same source types as report_build_status.
      build_status_config = optional(object({
        # context is the status-check label shown by the provider (e.g., the
        # GitHub status context). Defaults to the CodeBuild project identity when
        # omitted.
        context = optional(string, "")

        # target_url is the URL the reported status links to. Defaults to the
        # CodeBuild console page for the build when omitted.
        target_url = optional(string, "")
      }))

      # auth pins how CodeBuild authenticates to this source, overriding the
      # account-level source credential. The modern path is CODECONNECTIONS: a
      # CodeConnections (formerly CodeStar Connections) connection ARN that
      # grants repository access without long-lived tokens. SECRETS_MANAGER
      # points at a secret holding a provider token; OAUTH uses the legacy
      # account-level OAuth grant.
      auth = optional(object({
        # type is the authorization mechanism.
        #   CODECONNECTIONS:  A CodeConnections connection (recommended; no stored
        #                     tokens, resource is the connection ARN)
        #   SECRETS_MANAGER:  A Secrets Manager secret holding the provider token
        #                     (resource is the secret ARN)
        #   OAUTH:            The account-level OAuth authorization (legacy)
        type = string

        # resource is the ARN of the authorization object — the CodeConnections
        # connection ARN or the Secrets Manager secret ARN, depending on type.
        # Always a reference, never the credential value itself.
        resource = string
      }))

      # source_identifier names this source inside the project. REQUIRED for
      # secondary sources (the buildspec addresses the checkout as
      # $CODEBUILD_SRC_DIR_<source_identifier>); must stay EMPTY on the primary
      # source. Alphanumeric and underscore.
      source_identifier = optional(string, "")
    })

    # secondary_sources add up to 12 extra inputs alongside the primary
    # source (e.g., a shared build-tooling repository next to the application
    # repository). Each entry MUST set source_identifier; the buildspec
    # addresses the checkout via $CODEBUILD_SRC_DIR_<source_identifier>.
    secondary_sources = optional(list(object({
      # type is the source provider.
      #   GITHUB:                GitHub.com repository
      #   BITBUCKET:             Bitbucket Cloud repository
      #   CODECOMMIT:            AWS CodeCommit repository
      #   CODEPIPELINE:          Source provided by CodePipeline (no location needed)
      #   GITHUB_ENTERPRISE:     GitHub Enterprise Server
      #   GITLAB:                GitLab.com repository
      #   GITLAB_SELF_MANAGED:   Self-hosted GitLab instance
      #   NO_SOURCE:             No source — buildspec must be provided inline
      #   S3:                    S3 bucket containing source archive
      type = string

      # location is the source code repository URL or S3 path.
      # Required for all types except CODEPIPELINE and NO_SOURCE.
      # Format depends on type:
      #   GITHUB/BITBUCKET/GITLAB: https URL (e.g., https://github.com/owner/repo.git)
      #   CODECOMMIT: HTTPS clone URL
      #   S3: bucket/path (e.g., my-bucket/source.zip)
      #   GITHUB_ENTERPRISE: HTTPS URL of the repo
      location = optional(string, "")

      # buildspec is the build specification, either as an inline YAML string or
      # a path relative to the source root (e.g., "buildspec.yml").
      # If omitted, CodeBuild looks for buildspec.yml at the source root.
      # Required when source type is NO_SOURCE.
      buildspec = optional(string, "")

      # git_clone_depth limits the Git clone depth. 0 means full clone.
      # Only applicable for Git-based source types.
      git_clone_depth = optional(number, 0)

      # git_submodules_config controls Git submodule fetching during the source
      # download phase. Presence of the block opts into explicit submodule
      # handling. Only supported for BITBUCKET, CODECOMMIT, GITHUB, and
      # GITHUB_ENTERPRISE sources.
      git_submodules_config = optional(object({
        # fetch_submodules fetches all Git submodules during the source download
        # phase when true.
        fetch_submodules = optional(bool, false)
      }))

      # insecure_ssl skips TLS certificate verification when fetching source —
      # only for GitHub Enterprise / self-managed GitLab instances with
      # self-signed certificates. Never enable against public providers.
      insecure_ssl = optional(bool, false)

      # report_build_status reports build start and finish status back to the
      # source provider (e.g., GitHub commit status checks).
      # Only applicable for GITHUB, BITBUCKET, GITHUB_ENTERPRISE, GITLAB,
      # GITLAB_SELF_MANAGED.
      report_build_status = optional(bool, false)

      # build_status_config customizes how the commit status is reported —
      # the status-check context label and the URL it links to. Only applicable
      # to the same source types as report_build_status.
      build_status_config = optional(object({
        # context is the status-check label shown by the provider (e.g., the
        # GitHub status context). Defaults to the CodeBuild project identity when
        # omitted.
        context = optional(string, "")

        # target_url is the URL the reported status links to. Defaults to the
        # CodeBuild console page for the build when omitted.
        target_url = optional(string, "")
      }))

      # auth pins how CodeBuild authenticates to this source, overriding the
      # account-level source credential. The modern path is CODECONNECTIONS: a
      # CodeConnections (formerly CodeStar Connections) connection ARN that
      # grants repository access without long-lived tokens. SECRETS_MANAGER
      # points at a secret holding a provider token; OAUTH uses the legacy
      # account-level OAuth grant.
      auth = optional(object({
        # type is the authorization mechanism.
        #   CODECONNECTIONS:  A CodeConnections connection (recommended; no stored
        #                     tokens, resource is the connection ARN)
        #   SECRETS_MANAGER:  A Secrets Manager secret holding the provider token
        #                     (resource is the secret ARN)
        #   OAUTH:            The account-level OAuth authorization (legacy)
        type = string

        # resource is the ARN of the authorization object — the CodeConnections
        # connection ARN or the Secrets Manager secret ARN, depending on type.
        # Always a reference, never the credential value itself.
        resource = string
      }))

      # source_identifier names this source inside the project. REQUIRED for
      # secondary sources (the buildspec addresses the checkout as
      # $CODEBUILD_SRC_DIR_<source_identifier>); must stay EMPTY on the primary
      # source. Alphanumeric and underscore.
      source_identifier = optional(string, "")
    })), [])

    # secondary_source_versions pin a branch, tag, or commit per secondary
    # source (by its source_identifier). Secondary sources without an entry
    # build from their default branch.
    secondary_source_versions = optional(list(object({
      # source_identifier matches the source_identifier of an entry in
      # secondary_sources.
      source_identifier = string

      # source_version is the branch name, tag, commit SHA, or S3 object
      # version for that source.
      source_version = string
    })), [])

    # environment defines the build container: image, compute, and variables.
    environment = object({
      # type is the build environment type.
      #   LINUX_CONTAINER:               Standard Linux (x86_64)
      #   LINUX_GPU_CONTAINER:           Linux with GPU support
      #   ARM_CONTAINER:                 Linux (ARM64)
      #   WINDOWS_CONTAINER:             Windows (legacy alias; prefer the
      #                                  versioned Windows Server types)
      #   WINDOWS_SERVER_2019_CONTAINER: Windows Server 2019
      #   WINDOWS_SERVER_2022_CONTAINER: Windows Server 2022
      #   LINUX_LAMBDA_CONTAINER:        Lambda-based Linux (x86_64) — fastest
      #                                  start, no privileged mode or timeouts
      #   ARM_LAMBDA_CONTAINER:          Lambda-based Linux (ARM64)
      #   LINUX_EC2:                     EC2-based Linux (reserved fleets)
      #   ARM_EC2:                       EC2-based Linux ARM (reserved fleets)
      #   WINDOWS_EC2:                   EC2-based Windows (reserved fleets)
      #   MAC_ARM:                       macOS on Apple Silicon (reserved
      #                                  fleets only — iOS/macOS builds)
      type = string

      # compute_type is the compute capacity for the build container.
      #   Standard:  BUILD_GENERAL1_SMALL, BUILD_GENERAL1_MEDIUM,
      #              BUILD_GENERAL1_LARGE, BUILD_GENERAL1_XLARGE,
      #              BUILD_GENERAL1_2XLARGE
      #   Lambda:    BUILD_LAMBDA_1GB through BUILD_LAMBDA_10GB
      #   Fleets:    ATTRIBUTE_BASED_COMPUTE (the fleet picks machines by
      #              attributes), CUSTOM_INSTANCE_TYPE (the fleet pins an EC2
      #              instance type) — both only meaningful when the project or
      #              its fleet uses reserved capacity
      compute_type = string

      # image is the Docker image identifier for the build environment.
      # Use AWS managed images (e.g., "aws/codebuild/amazonlinux2-x86_64-standard:5.0")
      # or a custom image URI from ECR or Docker Hub.
      image = string

      # certificate is an S3 path (bucket/key ending in .pem or .zip) to a
      # certificate bundle the build trusts — for reaching source providers or
      # registries behind private CAs.
      certificate = optional(string, "")

      # privileged_mode enables Docker daemon access inside the build container.
      # Required for building Docker images. Not supported for Lambda types.
      privileged_mode = optional(bool, false)

      # image_pull_credentials_type controls how the build image is pulled.
      #   CODEBUILD:    CodeBuild uses its own credentials (default, for AWS images)
      #   SERVICE_ROLE: CodeBuild uses the project's service role (for ECR private images)
      image_pull_credentials_type = optional(string)

      # environment_variables define key-value pairs available during the build.
      # Supports plaintext values, SSM Parameter Store references, and
      # Secrets Manager references.
      environment_variables = optional(list(object({
        # name is the environment variable name.
        name = string

        # value is the environment variable value. For PARAMETER_STORE type,
        # this is the SSM parameter name. For SECRETS_MANAGER, this is the
        # secret ARN or name. Never put secret material in a PLAINTEXT value —
        # it is visible to anyone who can describe the project.
        value = string

        # type controls how the value is interpreted.
        #   PLAINTEXT:       Value is used as-is (default)
        #   PARAMETER_STORE: Value is an SSM Parameter Store parameter name
        #   SECRETS_MANAGER: Value is a Secrets Manager secret ARN or name
        type = optional(string)
      })), [])

      # registry_credential provides credentials for pulling the build image
      # from a private Docker registry. Only needed when image_pull_credentials_type
      # is SERVICE_ROLE and the image is in a non-ECR private registry.
      registry_credential = optional(object({
        # credential is the ARN or name of the Secrets Manager secret containing
        # the Docker registry credentials (username + password).
        credential = string

        # credential_provider is the credential provider type.
        # Currently only SECRETS_MANAGER is supported by AWS.
        credential_provider = string
      }))

      # docker_server provisions a persistent, dedicated Docker server for the
      # project's builds — Docker layer state survives across builds, giving
      # dramatically faster image builds than a per-build daemon. Choose the
      # server's compute size independently of the build compute.
      docker_server = optional(object({
        # compute_type sizes the Docker server (same scale as the build compute
        # types, e.g. BUILD_GENERAL1_MEDIUM).
        compute_type = string

        # security_group_ids attach VPC security groups to the Docker server when
        # the project runs inside a VPC. Maximum 5.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        security_group_ids = optional(list(string), [])
      }))

      # fleet_arn joins this project to a reserved-capacity fleet — a pool of
      # pre-provisioned, always-warm build machines (zero queue/provisioning
      # time; required for MAC_ARM and the EC2 environment types). The fleet is
      # a shared, account-level resource managed outside this project; reference
      # it by ARN.
      fleet_arn = optional(string, "")

      # host_kernel selects the Linux kernel version for the build hosts.
      #   LINUX_KERNEL_4:      Kernel 4.x (the long-standing default)
      #   LINUX_KERNEL_6:      Kernel 6.x (newer eBPF/io_uring features)
      #   LINUX_KERNEL_LATEST: Always the newest kernel CodeBuild offers
      # Applies only to the LINUX_CONTAINER, ARM_CONTAINER, LINUX_EC2, and
      # ARM_EC2 environment types — AWS's contract; Windows, Lambda, and Mac
      # environments have no kernel selection. Omit to let AWS choose.
      host_kernel = optional(string, "")
    })

    # artifacts defines where the primary build output goes. Its
    # artifact_identifier is optional on the primary output.
    artifacts = object({
      # type is the artifact output type.
      #   NO_ARTIFACTS:  No output artifacts (CI-only builds, or push to ECR in buildspec)
      #   S3:            Write artifacts to an S3 bucket
      #   CODEPIPELINE:  Artifacts managed by CodePipeline
      type = string

      # location is the S3 bucket name for artifact output.
      # Required when type is S3.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      location = optional(string, "")

      # name is the artifact output name (object key in S3).
      name = optional(string, "")

      # path is the S3 prefix for the artifact output.
      path = optional(string, "")

      # packaging controls how artifacts are packaged.
      #   NONE: No packaging (files uploaded as-is)
      #   ZIP:  Files packaged into a ZIP archive
      packaging = optional(string, "")

      # namespace_type controls whether artifact paths include the build ID.
      #   NONE:     No namespace (artifacts at path/name)
      #   BUILD_ID: Artifacts at path/<build-id>/name
      namespace_type = optional(string, "")

      # encryption_disabled disables server-side encryption for artifacts.
      # By default, CodeBuild encrypts artifacts using the project's encryption key.
      encryption_disabled = optional(bool, false)

      # override_artifact_name lets the buildspec's artifacts.name override the
      # name configured here — used for per-build artifact names (e.g., embedding
      # the version or commit).
      override_artifact_name = optional(bool, false)

      # bucket_owner_access grants the owning account of the destination bucket
      # access to the uploaded artifacts (for cross-account artifact buckets).
      #   NONE:      Bucket owner gets no access (default)
      #   READ_ONLY: Bucket owner can read the artifacts
      #   FULL:      Bucket owner has full control
      bucket_owner_access = optional(string, "")

      # artifact_identifier names this output. REQUIRED for secondary artifacts
      # (matching the buildspec's secondary-artifacts key); optional on the
      # primary output. Alphanumeric and underscore.
      artifact_identifier = optional(string, "")
    })

    # secondary_artifacts add up to 12 extra output locations (e.g., publish
    # a container context to one bucket and test reports to another). Each
    # entry MUST set artifact_identifier, matching the identifier used in the
    # buildspec's secondary-artifacts section.
    secondary_artifacts = optional(list(object({
      # type is the artifact output type.
      #   NO_ARTIFACTS:  No output artifacts (CI-only builds, or push to ECR in buildspec)
      #   S3:            Write artifacts to an S3 bucket
      #   CODEPIPELINE:  Artifacts managed by CodePipeline
      type = string

      # location is the S3 bucket name for artifact output.
      # Required when type is S3.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      location = optional(string, "")

      # name is the artifact output name (object key in S3).
      name = optional(string, "")

      # path is the S3 prefix for the artifact output.
      path = optional(string, "")

      # packaging controls how artifacts are packaged.
      #   NONE: No packaging (files uploaded as-is)
      #   ZIP:  Files packaged into a ZIP archive
      packaging = optional(string, "")

      # namespace_type controls whether artifact paths include the build ID.
      #   NONE:     No namespace (artifacts at path/name)
      #   BUILD_ID: Artifacts at path/<build-id>/name
      namespace_type = optional(string, "")

      # encryption_disabled disables server-side encryption for artifacts.
      # By default, CodeBuild encrypts artifacts using the project's encryption key.
      encryption_disabled = optional(bool, false)

      # override_artifact_name lets the buildspec's artifacts.name override the
      # name configured here — used for per-build artifact names (e.g., embedding
      # the version or commit).
      override_artifact_name = optional(bool, false)

      # bucket_owner_access grants the owning account of the destination bucket
      # access to the uploaded artifacts (for cross-account artifact buckets).
      #   NONE:      Bucket owner gets no access (default)
      #   READ_ONLY: Bucket owner can read the artifacts
      #   FULL:      Bucket owner has full control
      bucket_owner_access = optional(string, "")

      # artifact_identifier names this output. REQUIRED for secondary artifacts
      # (matching the buildspec's secondary-artifacts key); optional on the
      # primary output. Alphanumeric and underscore.
      artifact_identifier = optional(string, "")
    })), [])

    # service_role is the IAM role ARN that grants CodeBuild permission to
    # access source code, write artifacts, publish logs, and interact with
    # other AWS services during the build.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    service_role = string

    # description is a human-readable description of the project (max 255 chars).
    description = optional(string, "")

    # encryption_key is the ARN of a KMS key used to encrypt build artifacts.
    # If omitted, CodeBuild uses the AWS-managed key for S3.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    encryption_key = optional(string, "")

    # build_timeout is the maximum duration of a single build, in minutes.
    # Range: 5-2160 (36 hours). Default: 60. Not supported for Lambda compute
    # environment types (LINUX_LAMBDA_CONTAINER, ARM_LAMBDA_CONTAINER), where
    # AWS caps every build at the Lambda maximum instead.
    build_timeout = optional(number)

    # queued_timeout is the maximum time a build can wait in the queue before
    # timing out, in minutes. Range: 5-480 (8 hours). Default: 480. Not
    # supported for Lambda compute environment types.
    queued_timeout = optional(number)

    # concurrent_build_limit caps the number of concurrent builds. Useful for
    # cost control. Minimum 1. Omit to allow unlimited concurrency.
    concurrent_build_limit = optional(number, 0)

    # auto_retry_limit is the number of ADDITIONAL automatic retries after a
    # failed build (e.g., 2 means up to three attempts total). AWS allows up
    # to 10. Omit (or 0) to disable automatic retry.
    auto_retry_limit = optional(number, 0)

    # badge_enabled publishes a dynamic build badge for the project. The badge
    # URL is exported as the badge_url stack output and can be embedded in a
    # repository README. Not supported for CODEPIPELINE or S3 sources.
    badge_enabled = optional(bool, false)

    # source_version is the default branch, tag, or commit ID to build.
    # For GitHub: branch name, tag, or full commit SHA.
    # For S3: object version ID.
    source_version = optional(string, "")

    # cache configures build caching to speed up subsequent builds.
    cache = optional(object({
      # type is the cache type.
      #   NO_CACHE: Caching disabled (default)
      #   S3:       Cache stored in an S3 bucket
      #   LOCAL:    Cache stored on the build host (ephemeral, useful for Docker layers)
      type = optional(string)

      # location is the S3 bucket and optional prefix for cache storage.
      # Required when type is S3. Format: "bucket-name" or "bucket-name/prefix".
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      location = optional(string, "")

      # modes specifies what to cache when type is LOCAL.
      #   LOCAL_SOURCE_CACHE:        Cache Git metadata
      #   LOCAL_DOCKER_LAYER_CACHE:  Cache Docker layers
      #   LOCAL_CUSTOM_CACHE:        Cache paths specified in buildspec
      modes = optional(list(string), [])

      # cache_namespace scopes S3 cache keys so multiple projects (or branches)
      # can share one cache bucket without collisions.
      cache_namespace = optional(string, "")
    }))

    # logs_config controls where build logs are sent.
    logs_config = optional(object({
      # cloudwatch_logs configures CloudWatch Logs for build output.
      cloudwatch_logs = optional(object({
        # status controls whether CloudWatch logging is enabled.
        # Default: ENABLED.
        status = optional(string)

        # group_name is the CloudWatch Logs log group name.
        # If omitted, CodeBuild creates a default log group.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        group_name = optional(string, "")

        # stream_name is the CloudWatch Logs log stream name prefix.
        # If omitted, CodeBuild generates a default stream name.
        stream_name = optional(string, "")
      }))

      # s3_logs configures S3 logging for build output.
      s3_logs = optional(object({
        # status controls whether S3 logging is enabled.
        # Default: DISABLED.
        status = optional(string)

        # bucket is the S3 bucket receiving build logs — a reference to an
        # AwsS3Bucket's bucket_id output, a literal bucket name, or the bucket
        # ARN. AWS stores S3 build logs only under a prefix, so the bucket and
        # `prefix` are separate fields here and both IaC modules compose the
        # provider's single location argument as "{bucket}/{prefix}".
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        bucket = optional(string, "")

        # encryption_disabled disables server-side encryption for log files.
        encryption_disabled = optional(bool, false)

        # bucket_owner_access grants the owning account of the log bucket access
        # to the uploaded log files (for centralized cross-account log buckets).
        bucket_owner_access = optional(string, "")

        # prefix is the path under the bucket where log objects land. AWS
        # REQUIRES it for S3 build logs — a bare bucket with no prefix is
        # rejected at CreateProject, which is why it is a separate required
        # companion to `bucket` rather than a "bucket/prefix" literal format.
        prefix = optional(string, "")
      }))
    }))

    # vpc_config places the build in a VPC, giving it access to private
    # resources such as RDS databases, ElastiCache clusters, or internal APIs.
    vpc_config = optional(object({
      # vpc_id is the VPC where build containers are launched.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      vpc_id = string

      # subnet_ids are the VPC subnets where build containers are placed.
      # Use private subnets for security. Maximum 16 subnets.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      subnet_ids = list(string)

      # security_group_ids are the VPC security groups applied to build containers.
      # Maximum 5 security groups.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      security_group_ids = list(string)
    }))

    # file_system_locations mount Amazon EFS file systems into the build
    # container — useful for large shared caches (e.g., a Gradle or Bazel
    # cache) that outlive individual builds. Requires vpc_config placing the
    # build in subnets that can reach the file system's mount targets.
    file_system_locations = optional(list(object({
      # identifier names this mount; CodeBuild exposes it to the build as the
      # environment variable CODEBUILD_<identifier> (uppercased).
      identifier = string

      # location is the EFS mount source in the form
      # <file-system-dns>:/<path>, e.g.
      # "fs-0abc123.efs.us-east-1.amazonaws.com:/build-cache". The DNS name
      # composes from an AwsElasticFileSystem's file_system_id as
      # <file_system_id>.efs.<region>.amazonaws.com.
      location = string

      # mount_point is the absolute path inside the build container where the
      # file system is mounted (e.g., "/mnt/build-cache").
      mount_point = string

      # mount_options are NFS mount options. Omit for the EFS recommended
      # defaults (nfsvers=4.1,rsize=1048576,wsize=1048576,hard,timeo=600,retrans=2).
      mount_options = optional(string, "")

      # type is the file system protocol. AWS currently supports only EFS.
      type = optional(string)
    })), [])

    # build_batch_config enables batch builds — a single StartBuildBatch call
    # fans out into multiple coordinated builds defined by the buildspec's
    # batch section (build graphs, build lists, matrix builds).
    build_batch_config = optional(object({
      # service_role is the IAM role CodeBuild assumes to launch the child
      # builds of a batch. May be the project's service role or a dedicated one.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      service_role = string

      # combine_artifacts merges the child builds' artifacts into a single
      # artifact location for the whole batch.
      combine_artifacts = optional(bool, false)

      # timeout_in_mins is the maximum duration of the entire batch, in minutes
      # (5-2160).
      timeout_in_mins = optional(number, 0)

      # restrictions bound what the batch's child builds may consume.
      restrictions = optional(object({
        # compute_types_allowed restricts which compute types child builds may
        # request (values from the environment compute_type set). Empty allows all.
        compute_types_allowed = optional(list(string), [])

        # maximum_builds_allowed caps how many child builds one batch may spawn
        # (1-100).
        maximum_builds_allowed = optional(number, 0)
      }))
    }))

    # project_visibility controls whether the project's builds are publicly
    # readable.
    #   PRIVATE:     Only principals in the account can view builds (default)
    #   PUBLIC_READ: Build results, logs, and artifacts are world-readable —
    #                used by open-source projects publishing CI results
    project_visibility = optional(string)

    # resource_access_role is the IAM role CodeBuild uses to read the CloudWatch
    # logs and S3 artifacts it exposes for public builds. Required when
    # project_visibility is PUBLIC_READ.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    resource_access_role = optional(string, "")

    # resource_policy is a resource-based IAM policy document attached to the
    # project — the mechanism for cross-account access (e.g., letting a
    # central CI account start builds in this account). Provide the policy as
    # a structured JSON document.
    resource_policy = optional(any)

    # webhook configures automatic build triggers from the source provider.
    # Only valid when source type supports webhooks: GITHUB, BITBUCKET,
    # GITHUB_ENTERPRISE, GITLAB, GITLAB_SELF_MANAGED, CODECOMMIT.
    # Omit for CodePipeline-triggered or manual-only projects.
    webhook = optional(object({
      # build_type controls the build type triggered by the webhook.
      #   BUILD:                  Standard single build (default)
      #   BUILD_BATCH:            Batch build (requires build_batch_config)
      #   RUNNER_BUILDKITE_BUILD: The project acts as a Buildkite runner —
      #                           webhook events dispatch Buildkite jobs
      build_type = optional(string, "")

      # manual_creation makes CodeBuild return the payload URL and HMAC secret
      # WITHOUT registering the webhook with the provider — you configure the
      # repository webhook by hand from the webhook_payload_url and
      # webhook_secret stack outputs. Required for GitHub Enterprise; useful
      # when the connection lacks admin rights on the repository.
      manual_creation = optional(bool, false)

      # filter_groups define which repository events trigger a build.
      # Multiple groups are OR'd together — a build triggers if ANY group matches.
      # Within a group, filters are AND'd — ALL filters in the group must match.
      filter_groups = optional(list(object({
        # filters are the individual conditions in this group.
        filters = list(object({
          # type is the event field to match against.
          #   EVENT:             Event type (PUSH, PULL_REQUEST_CREATED, PULL_REQUEST_UPDATED, ...)
          #   BASE_REF:          Base branch for PRs (regex)
          #   HEAD_REF:          Branch or tag name (regex)
          #   ACTOR_ACCOUNT_ID:  Source provider account ID
          #   FILE_PATH:         Changed file paths (regex)
          #   COMMIT_MESSAGE:    Commit message (regex)
          #   WORKFLOW_NAME:     GitHub Actions workflow name (runner projects)
          #   TAG_NAME:          Release tag name (regex)
          #   RELEASE_NAME:      Release name (regex)
          #   REPOSITORY_NAME:   Repository name (org-scoped webhooks; regex)
          #   ORGANIZATION_NAME: Organization name (global-scoped webhooks; regex)
          type = string

          # pattern is the regex pattern or comma-separated event list to match.
          # For EVENT type: comma-separated values like "PUSH, PULL_REQUEST_CREATED".
          # For other types: a regex pattern (e.g., "^refs/heads/main$").
          pattern = string

          # exclude_matched_pattern inverts the match — the filter passes when the
          # pattern does NOT match. Useful for excluding branches or paths.
          exclude_matched_pattern = optional(bool, false)
        }))
      })), [])

      # scope_configuration widens the webhook beyond one repository — an
      # organization- or group-level webhook (used with runner projects and
      # org-wide CI). The webhook then fires for every repository in scope.
      scope_configuration = optional(object({
        # name is the organization (GitHub) or group (GitLab) the webhook covers.
        name = string

        # scope is the webhook's coverage.
        #   GITHUB_ORGANIZATION: All repositories in a GitHub organization
        #   GITHUB_GLOBAL:       All repositories the connection can reach
        #   GITLAB_GROUP:        All projects in a GitLab group
        scope = string

        # domain is the self-hosted provider domain (GitHub Enterprise Server /
        # self-managed GitLab). Omit for the cloud-hosted providers.
        domain = optional(string, "")
      }))

      # pull_request_build_policy gates pull-request-triggered builds behind a
      # comment approval — protection against untrusted code running in CI from
      # fork PRs.
      pull_request_build_policy = optional(object({
        # requires_comment_approval controls which pull requests need an approval
        # comment before building.
        #   DISABLED:           All PRs build immediately (default provider behavior)
        #   FORK_PULL_REQUESTS: Only PRs from forks wait for approval — the common
        #                       open-source posture (protects secrets from
        #                       untrusted fork code)
        #   ALL_PULL_REQUESTS:  Every PR waits for an approval comment
        requires_comment_approval = string

        # approver_roles are the repository roles whose comments count as
        # approval. Defaults are provider-managed when empty.
        approver_roles = optional(list(string), [])
      }))
    }))
  })
}
