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
  description = "AwsAppRunnerService specification"
  type = object({
    # The AWS region where the App Runner service will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Image-based deployment source. Use this to deploy a container image
    # stored in Amazon ECR (private) or ECR Public Gallery.
    image_source = optional(object({
      # Full container image identifier including tag or digest.
      # ECR format: "ACCOUNT_ID.dkr.ecr.REGION.amazonaws.com/REPO:TAG"
      # ECR Public format: "public.ecr.aws/ALIAS/REPO:TAG"
      # This stays a literal string (not a reference) because it carries a
      # repository-plus-tag coordinate no single upstream output represents.
      # The format rule below is AWS's own ImageIdentifier pattern (mirrored
      # by the provider): a 12-digit-account private ECR host path, or a
      # public.ecr.aws gallery path.
      image_identifier = string

      # Type of image repository.
      # "ECR": private Amazon ECR (requires access_role_arn for pull access).
      # "ECR_PUBLIC": public ECR Gallery (no pull authentication; automatic
      # deployments are not supported for public images).
      image_repository_type = string

      # IAM role ARN that grants App Runner permission to pull images from
      # private ECR. Required when image_repository_type is "ECR"; not used for
      # "ECR_PUBLIC". The role must be assumable by build.apprunner.amazonaws.com
      # and carry ECR read permissions (the AWS-managed
      # AWSAppRunnerServicePolicyForECRAccess policy covers it).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      access_role_arn = optional(string, "")
    }))

    # Code-based deployment source. Use this to deploy from a source
    # repository through an App Runner connection; App Runner clones the
    # repository, builds the application with the selected managed runtime,
    # and deploys the resulting container.
    code_source = optional(object({
      # Repository URL (e.g. "https://github.com/owner/repo").
      repository_url = string

      # Branch to deploy from (e.g. "main", "production"). App Runner tracks
      # this branch; with auto_deployments_enabled, every push deploys.
      branch = string

      # Subdirectory within the repository containing the application source.
      # Defaults to the repository root. Useful for monorepos where the service
      # lives in a subfolder; with configuration_source="REPOSITORY", the
      # apprunner.yaml is read from this directory. One-way on a live service:
      # clearing this does not return the build to the repository root (the
      # provider keeps the last-applied value) -- set "/" explicitly instead.
      source_directory = optional(string, "")

      # ARN of an App Runner connection that authorizes repository access
      # (GitHub or Bitbucket). Connections require a one-time OAuth handshake
      # completed in the AWS console, so they are created out-of-band and
      # referenced here by ARN; one connection is shared across services.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      connection_arn = string

      # Where App Runner reads the build/runtime configuration from.
      # "API": configuration comes from this spec (runtime, build_command,
      #   port, start_command).
      # "REPOSITORY": configuration is read from an apprunner.yaml at the
      #   repository root (or source_directory); runtime and build_command in
      #   this spec are ignored.
      configuration_source = string

      # Managed runtime that builds and runs the application. Required when
      # configuration_source is "API", ignored for "REPOSITORY".
      runtime = optional(string, "")

      # Shell command that builds the application (e.g. "npm ci && npm run
      # build", "pip install -r requirements.txt"). Only used when
      # configuration_source is "API".
      build_command = optional(string, "")
    }))

    # Port the application listens on inside the container. App Runner
    # terminates TLS on 443 and forwards requests to this port.
    port = optional(string)

    # Override the container start command. For image_source this overrides
    # the image ENTRYPOINT/CMD; for code_source with
    # configuration_source="API" it is the command that starts the built
    # application. One-way on a live service: the provider drops empty
    # values on update, so clearing this does not restore the image's own
    # ENTRYPOINT/CMD -- set the desired command explicitly instead.
    start_command = optional(string, "")

    # Environment variables injected into every instance at runtime. Keys are
    # variable names, values are plaintext strings. Never put secret values
    # here -- use environment_secrets for anything sensitive. Keys prefixed
    # with "AWSAPPRUNNER" are reserved by the service.
    environment_variables = optional(map(string), {})

    # Environment secrets injected at runtime. Keys are variable names;
    # values are full ARNs of AWS Secrets Manager secrets or SSM Parameter
    # Store parameters -- App Runner resolves each ARN at deploy time and
    # injects the resolved value as an environment variable. The
    # instance_role_arn role must be allowed to read the referenced secrets
    # (secretsmanager:GetSecretValue / ssm:GetParameters).
    environment_secrets = optional(map(string), {})

    # CPU allocation per instance. Accepts millicore strings or
    # human-readable vCPU format.
    # Numeric: "256", "512", "1024", "2048", "4096"
    # Human-readable: "0.25 vCPU", "0.5 vCPU", "1 vCPU", "2 vCPU", "4 vCPU"
    cpu = optional(string)

    # Memory allocation per instance. Accepts megabyte strings or
    # human-readable GB format.
    # Numeric: "512", "1024", "2048", "3072", "4096", "6144", "8192",
    # "10240", "12288"
    # Human-readable: "0.5 GB", "1 GB", "2 GB", "3 GB", "4 GB", "6 GB",
    # "8 GB", "10 GB", "12 GB"
    # Not every CPU/memory pairing is valid -- App Runner accepts specific
    # combinations (e.g. 4 vCPU requires 8-12 GB); the create call rejects
    # invalid pairs.
    memory = optional(string)

    # IAM role that service instances assume at runtime to call AWS APIs --
    # the role your application code uses (reading S3, writing DynamoDB,
    # resolving environment_secrets). This is NOT the image-pull role (that
    # is image_source.access_role_arn). One-way on a live service: the
    # provider drops empty values on update, so REMOVING the role does not
    # detach it from a running service -- attach a replacement role instead.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    instance_role_arn = optional(string, "")

    # Health check configuration App Runner uses to monitor instance
    # readiness; unhealthy instances are replaced automatically. When
    # omitted, App Runner performs TCP checks on the configured port with
    # AWS defaults. Removing the block from a live service stops managing
    # the check but does not reset it -- to return to defaults, set the
    # default values explicitly (protocol TCP, interval 5, timeout 2,
    # thresholds 1/5).
    health_check = optional(object({
      # Health check protocol.
      # "TCP": checks the port accepts connections (default).
      # "HTTP": sends GET requests to path and expects a 2xx response.
      protocol = optional(string)

      # URL path for HTTP health checks (e.g. "/health", "/readyz"). Ignored
      # when protocol is "TCP".
      path = optional(string)

      # Seconds between consecutive health checks.
      interval = optional(number)

      # Seconds to wait for a health check response before counting the check
      # as failed.
      timeout = optional(number)

      # Consecutive successful checks required to mark an instance healthy.
      healthy_threshold = optional(number)

      # Consecutive failed checks before an instance is marked unhealthy and
      # replaced.
      unhealthy_threshold = optional(number)
    }))

    # ARN of an AwsAppRunnerAutoScalingConfiguration revision that governs
    # concurrency-based scaling. When omitted, AWS applies the account's
    # default auto scaling configuration (1 min / 25 max / 100 concurrency
    # unless the account default was changed). The referenced ARN carries a
    # revision, so registering a new revision rolls this service on its next
    # deployment. One-way on a live service: once set, REMOVING this
    # reference does not return the service to the account default -- the
    # provider keeps the last-applied configuration (its attribute is
    # Optional+Computed, so removal produces no change). To move back to the
    # default, reference the default configuration's ARN explicitly.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    auto_scaling_configuration_arn = optional(string, "")

    # ARN of an AwsAppRunnerVpcConnector for outbound VPC access -- lets the
    # service reach private resources (databases, caches, internal APIs)
    # inside a VPC. When omitted, egress uses App Runner's default public
    # path and the service can only reach public endpoints. One connector is
    # shared by any number of services.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    vpc_connector_arn = optional(string, "")

    # ARN of an AwsAppRunnerObservabilityConfiguration revision. When set,
    # the service sends request traces to the configured vendor (AWS X-Ray);
    # when omitted at creation, tracing is off. Presence of the reference IS
    # the enable switch -- there is no separate toggle to keep in sync.
    # One-way on a live service (upstream provider gap): once tracing is
    # enabled, REMOVING this reference does not disable it -- the provider
    # sends nothing for an absent block, so AWS keeps tracing on and the
    # plan never converges. To genuinely disable tracing today, replace the
    # service (or disable it in the AWS console) until the provider gains a
    # disable path.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    observability_configuration_arn = optional(string, "")

    # Whether the service endpoint is publicly reachable from the internet.
    # When true (default), the service gets a public HTTPS URL. When false,
    # the endpoint is reachable only from within VPCs that attach a VPC
    # Ingress Connection to this service -- declare those below in
    # vpc_ingress_connections.
    is_publicly_accessible = optional(bool)

    # VPC Ingress Connections for this service -- each entry publishes the
    # service into one VPC through an interface VPC endpoint
    # (AWS PrivateLink), so clients inside that VPC reach the service
    # privately at the exported per-connection domain name. Pair with
    # is_publicly_accessible: false for a service reachable ONLY from inside
    # the named VPCs. Keyed by name: adding or removing entries updates in
    # place; changing an entry's VPC or endpoint replaces that one
    # connection.
    vpc_ingress_connections = optional(list(object({
      # Name for this VPC Ingress Connection. Must be unique across all active
      # VPC Ingress Connections in the AWS account and region (it is the AWS
      # resource name, not a label). App Runner family names are 4-40
      # characters.
      name = string

      # The VPC to publish the service into.
      #
      # Containment-exempt: the connection lets clients INSIDE this VPC
      # reach the service over PrivateLink; the service itself runs on
      # App Runner's managed infrastructure and is never deployed into the
      # VPC (a VPC connector, not an ingress connection, is what places a
      # service's egress in a VPC). On a diagram the service stands outside
      # the VPC with a line to the network it is published into.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      vpc_id = string

      # The interface VPC endpoint in that VPC that carries the traffic. AWS
      # requires both members -- an ingress connection cannot exist without its
      # endpoint (the provider leaves both optional and lets the create call
      # fail server-side; this spec enforces AWS's contract up front).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      vpc_endpoint_id = string
    })), [])

    # IP address type for the service endpoint.
    # "IPV4": IPv4-only (default). "DUAL_STACK": IPv4 + IPv6.
    ip_address_type = optional(string)

    # Customer-managed KMS key ARN App Runner uses to encrypt its stored copy
    # of the deployment source (image or repository archive) and build logs.
    # When omitted, App Runner uses an AWS-managed key. ForceNew: changing
    # this replaces the service.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_arn = optional(string, "")

    # Whether App Runner automatically starts a deployment when the source
    # changes (a new image pushed to the tracked tag, or a new commit on the
    # tracked branch). Disabled by default: deployments then happen only when
    # this resource is applied or a deployment is started explicitly, which
    # keeps rollouts deterministic and graph-driven. AWS supports automatic
    # deployments only for private same-account ECR and code repositories --
    # ECR Public images cannot enable this (AWS rejects the create call).
    # The modules always send this value explicitly: AWS's own default is
    # conditional on the source type (on for code repos and same-account
    # ECR, off otherwise), and the provider substitutes an unconditional
    # "on" for any omitted value -- explicit send is the only deterministic
    # path through both layers.
    auto_deployments_enabled = optional(bool, false)

    # Custom domains associated with this service, keyed by domain name. App
    # Runner provisions and renews the TLS certificate for each domain; you
    # prove ownership by creating the CNAME records exported per domain in
    # status.outputs.custom_domains (compose them into AwsRoute53DnsRecord
    # resources), plus a CNAME/alias from the domain to the exported
    # dns_target. Adding or removing entries updates in place; changing an
    # entry replaces that one association.
    custom_domains = optional(list(object({
      # The domain to associate (e.g. "app.example.com" or "example.com").
      domain_name = string

      # Whether to also associate the "www." subdomain of domain_name.
      # AWS defaults this to true; meaningful mainly for apex domains.
      enable_www_subdomain = optional(bool)
    })), [])

    # ARN of a REGIONAL AwsWafWebAcl to associate with this service. All
    # requests pass WAF inspection before reaching the application. When
    # omitted, no WAF is attached.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    web_acl_arn = optional(string, "")
  })
}
