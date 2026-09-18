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
  description = "GcpCloudFunction specification"
  type = object({
    # The GCP project the function is created in. Accepts a literal project
    # ID or a reference to a GcpProject resource. If omitted, the provider's
    # default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Region the function is deployed in, e.g. "us-central1". Immutable.
    region = string

    # Name of the function in GCP. Immutable. If not specified, defaults to
    # metadata.name. Must be 1-63 characters: lowercase letters, digits, and
    # hyphens; starting with a letter.
    function_name = optional(string, "")

    # Human-readable description of what the function does.
    description = optional(string, "")

    # Labels applied to the function object. User labels are merged beneath
    # Planton's attribution labels and shared with Google's billing system.
    labels = optional(map(string), {})

    # Cloud KMS key encrypting the function's resources (CMEK) — the
    # container image and source artifacts. The Cloud Functions and Artifact
    # Registry service agents must hold cryptoKeyEncrypterDecrypter on it,
    # and CMEK deployments require a customer-managed docker_repository.
    # Accepts a full crypto-key path or a reference to a GcpKmsKey resource.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_name = optional(string, "")

    # How the source becomes a runnable container: runtime, entry point,
    # source location, and build identity. Required.
    build_config = object({
      # Runtime the function executes in, e.g. "python312", "nodejs22",
      # "go123", "java21". Any current Gen 2 runtime GCP publishes is valid —
      # run `gcloud functions runtimes list` for the live set; deprecated
      # runtimes are rejected by the API at deploy time.
      runtime = string

      # Name of the function in source code that will be executed (the entry
      # point). For example: "hello_http" in Python, "helloHttp" in Node.js.
      entry_point = string

      # Where the source code lives. Required.
      source = object({
        # Source archive in Google Cloud Storage — a .zip of the function code
        # and dependency manifest. The standard path for CI/CD-shipped source.
        storage_source = optional(object({
          # GCS bucket holding the source archive. The build service account needs
          # read access. Accepts a literal bucket name or a reference to a
          # GcpGcsBucket resource.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          bucket = string

          # Object name (path) of the source archive in the bucket, e.g.
          # "functions/my-function-v1.2.3.zip". Version the object name per
          # release — a changed object name is what makes the deploy roll.
          object = string

          # Generation number of the object to pin an exact object version even
          # if the path is overwritten. If unset, the current generation is used.
          generation = optional(number)
        }))

        # Source in Cloud Source Repositories. Note GCP deprecated CSR for new
        # customers in June 2024 — existing repositories keep working, but new
        # integrations should ship archives to GCS instead.
        repo_source = optional(object({
          # Name of the Cloud Source Repository, e.g. "my-repo".
          repo_name = string

          # Branch to build from (the branch's current HEAD at deploy time).
          branch_name = optional(string, "")

          # Tag to build from.
          tag_name = optional(string, "")

          # Exact commit to build from — the only fully reproducible pin.
          commit_sha = optional(string, "")

          # Directory within the repository containing the function source. If
          # unset, the repository root is used.
          dir = optional(string, "")

          # Invert the revision match: build from revisions that do NOT match
          # the configured branch/tag regex.
          invert_regex = optional(bool, false)

          # Project that owns the repository, when it lives outside the
          # function's project. Immutable.
          project_id = optional(string, "")
        }))
      })

      # Environment variables available at build time (e.g. buildpack knobs
      # like GOOGLE_ENTRYPOINT). Not injected into the runtime — use
      # service_config.environment_variables for that.
      build_environment_variables = optional(map(string), {})

      # Service account Cloud Build runs the build as — the identity that
      # reads the source and pushes the image. FULLY-QUALIFIED resource name
      # (projects/{project}/serviceAccounts/{email}), not a bare email.
      # Accepts a literal or a reference to a GcpServiceAccount resource. If
      # omitted, GCP uses its default build identity.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      service_account = optional(string, "")

      # Cloud Build Custom Worker Pool that builds the function — for builds
      # that must run inside a private network perimeter. Format:
      # projects/{project}/locations/{region}/workerPools/{name}.
      worker_pool = optional(string, "")

      # User-managed Artifact Registry repository the built container is
      # stored in, optionally CMEK-protected (required when kms_key_name is
      # set). FULLY-QUALIFIED path
      # (projects/{project}/locations/{location}/repositories/{name}).
      # Accepts a literal path or a reference to a GcpArtifactRegistryRepo
      # resource (its repository_path output is exactly this value). If
      # omitted, GCP manages a default repository.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      docker_repository = optional(string, "")

      # How the runtime base image is patched. AUTOMATIC (the API default when
      # unset) applies security updates continuously; ON_DEPLOY pins the
      # runtime version at deploy time so instances never change under you
      # between deploys.
      update_policy = optional(string, "")
    })

    # How the function runs: compute resources, environment, secrets,
    # networking, scaling, and invocation policy.
    service_config = optional(object({
      # Email of the IAM service account the function runs as — the identity
      # whose permissions the code exercises when calling other GCP APIs.
      # Accepts a literal email or a reference to a GcpServiceAccount
      # resource. If omitted, the project's Compute Engine default service
      # account is used — fine for experiments, too broad for production.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      service_account_email = optional(string, "")

      # Memory available to each instance, as a quantity string: "256M",
      # "512M", "1Gi", "16Gi". CPU scales with memory unless available_cpu is
      # set explicitly. If unset, GCP defaults to 256M.
      available_memory = optional(string, "")

      # CPUs available to each instance, e.g. "1", "2", "0.5". If unset, GCP
      # derives CPU from memory. Concurrency above 1 requires at least 1 CPU.
      available_cpu = optional(string, "")

      # Per-request timeout in seconds. HTTP functions support up to 3600
      # (60 minutes); event-driven functions are capped at 540 by Eventarc's
      # delivery timeout. If unset, GCP defaults to 60.
      timeout_seconds = optional(number, 0)

      # Concurrent requests each instance handles (1-1000). GCP defaults to 1
      # — every request gets its own instance, safe for any runtime. Raising
      # it cuts instance count and cold starts for I/O-bound code, but needs
      # at least 1 CPU and thread-safe code.
      max_instance_request_concurrency = optional(number, 0)

      # Environment variables injected into the runtime as plain-text
      # KEY=VALUE pairs. Configuration only — never place credentials here;
      # use secret_environment_variables so material stays in Secret Manager.
      environment_variables = optional(map(string), {})

      # Secret Manager references injected as environment variables. The
      # material never appears in the spec — each entry names a secret and
      # version resolved at instance start. The runtime service account needs
      # roles/secretmanager.secretAccessor on each secret.
      secret_environment_variables = optional(list(object({
        # Environment variable name, e.g. "DATABASE_PASSWORD".
        key = string

        # The secret: a short name for a secret in the function's project
        # ("my-secret"). Cross-project secrets set project_id.
        secret = string

        # Secret version to resolve: a version number or "latest" — the common
        # choice, at the cost of new instances silently picking up rotations.
        version = optional(string, "")

        # Project the secret lives in, when it is not the function's project.
        project_id = optional(string, "")
      })), [])

      # Secret Manager secret versions projected as files under a mount path
      # — for consumers that read credentials from disk (certificates, config
      # files). Same accessor-role requirement as secret env vars.
      secret_volumes = optional(list(object({
        # Absolute path the volume is mounted at, e.g. "/etc/secrets". Each
        # configured version appears as a file under it.
        mount_path = string

        # The secret: a short name for a secret in the function's project.
        # Cross-project secrets set project_id.
        secret = string

        # Project the secret lives in, when it is not the function's project.
        project_id = optional(string, "")

        # Which versions land at which relative paths. If empty, the "latest"
        # version is projected at a file named after the secret.
        versions = optional(list(object({
          # Secret version to project: a version number or "latest".
          version = string

          # Relative path of the file under the volume's mount path.
          path = string
        })), [])
      })), [])

      # Serverless VPC Access connector routing the function's egress into a
      # VPC — how the function reaches private IPs (Cloud SQL private IP,
      # Memorystore, internal load balancers). Accepts the connector's full
      # resource name (projects/*/locations/*/connectors/*) or a reference to
      # a GcpServerlessVpcConnector resource.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      vpc_connector = optional(string, "")

      # Which egress traffic uses the connector: only RFC1918/private
      # destinations (the default; public egress keeps the normal path), or
      # everything (enables static egress IPs via Cloud NAT).
      vpc_connector_egress_settings = optional(string, "")

      # Who can reach the function's endpoint at the network level. Use
      # ALLOW_INTERNAL_ONLY for private functions, ALLOW_INTERNAL_AND_GCLB
      # when fronting with an external Application Load Balancer.
      ingress_settings = optional(string, "")

      # Instance scaling bounds.
      scaling = optional(object({
        # Minimum instances kept warm. Above 0 eliminates cold starts at idle
        # compute cost — for latency-sensitive production endpoints.
        min_instance_count = optional(number, 0)

        # Maximum instances the function scales to — the cost and downstream-
        # pressure ceiling (a runaway event storm stops here).
        max_instance_count = optional(number, 0)
      }))

      # Whether 100% of traffic goes to the latest revision as soon as it is
      # ready (the API default, true). Set false to hold traffic on the
      # previous revision — the lever for manual canary/rollback via the
      # underlying Cloud Run service.
      all_traffic_on_latest_revision = optional(bool)

      # Binary Authorization policy checked before instances start, e.g.
      # "default" or "projects/{project}/platforms/gae/policies/{policy}".
      binary_authorization_policy = optional(string, "")

      # Makes the function publicly invokable by unauthenticated callers by
      # granting run.invoker to allUsers on the underlying Cloud Run service.
      # Leave false for private functions and grant invoker to specific
      # identities instead.
      allow_unauthenticated = optional(bool, false)

      # Direct VPC egress: attach the function straight to a VPC network or
      # subnet — no Serverless VPC Access connector to size, pay for, or
      # saturate. The modern alternative to vpc_connector (mutually
      # exclusive with it); instances get IPs from the subnet, so size its
      # range for the scaling ceiling.
      direct_vpc_network_interface = optional(object({
        # The VPC network to attach to. Accepts a literal network name or a
        # reference to a GcpVpcNetwork resource.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        network = optional(string, "")

        # The subnetwork instances draw their IPs from. Accepts a literal
        # subnetwork name or a reference to a GcpSubnetwork resource. The
        # subnet's free range caps how far the function can scale.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        subnetwork = optional(string, "")

        # Network tags applied to the function's instances — how VPC firewall
        # rules select their egress.
        tags = optional(list(string), [])
      }))

      # Which egress traffic takes the direct-VPC path: only RFC1918/private
      # destinations (the API default; public egress keeps the normal path),
      # or everything (enables static egress IPs via Cloud NAT). Only
      # meaningful with direct_vpc_network_interface — the connector path is
      # steered by vpc_connector_egress_settings instead.
      direct_vpc_egress = optional(string, "")
    }))

    # What invokes the function. If not specified, defaults to HTTP.
    trigger = optional(object({
      # Type of trigger. Defaults to HTTP if not specified.
      trigger_type = optional(string, "")

      # Event trigger configuration. Required when trigger_type is
      # EVENT_TRIGGER.
      event_trigger = optional(object({
        # Event type that triggers the function, in CloudEvents format:
        # - "google.cloud.pubsub.topic.v1.messagePublished" (Pub/Sub)
        # - "google.cloud.storage.object.v1.finalized" (object created)
        # - "google.cloud.storage.object.v1.deleted" (object deleted)
        # - "google.cloud.firestore.document.v1.written" (document write)
        event_type = string

        # Pub/Sub topic for messagePublished triggers. Accepts the full
        # resource name (projects/{project}/topics/{name}) or a reference to a
        # GcpPubSubTopic resource.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        pubsub_topic = optional(string, "")

        # Event filters narrowing which events invoke the function. For Storage
        # triggers, filter by bucket: attribute="bucket" value="my-bucket". For
        # Firestore, filter by document path pattern.
        event_filters = optional(list(object({
          # Attribute to filter on (e.g. "bucket" for Storage events).
          attribute = string

          # Value to match.
          value = string

          # Matching operator. Unset means exact match; "match-path-pattern"
          # (the only other value GCP accepts) enables path-pattern wildcards for
          # Firestore/audit-log filters.
          operator = optional(string)
        })), [])

        # Region the trigger listens in. Storage/audit-log sources fire in the
        # bucket's region (multi-region sources use "us"/"eu"); if unset, GCP
        # uses the function's region.
        trigger_region = optional(string, "")

        # Retry policy for failed deliveries. RETRY_POLICY_RETRY redelivers
        # with exponential backoff (at-least-once — handlers must be
        # idempotent); RETRY_POLICY_DO_NOT_RETRY delivers at most once.
        retry_policy = optional(string, "")

        # Email of the service account Eventarc uses to invoke the function —
        # it needs run.invoker on the underlying service. Accepts a literal
        # email or a reference to a GcpServiceAccount resource. If omitted, the
        # default compute service account is used.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        service_account_email = optional(string, "")
      }))
    }))

    # What destroying this resource does to the function:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the function (and the Cloud Run service serving it) is
    #                deleted; event triggers stop firing
    #   "PREVENT" -- destroy FAILS; protects a function other systems invoke
    #   "ABANDON" -- the function is removed from management but keeps
    #                serving and consuming events in GCP
    deletion_policy = optional(string, "")
  })
}
