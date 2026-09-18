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
  description = "GcpVertexAiDeployedIndex specification"
  type = object({
    # Region of the index endpoint (e.g., "us-central1"). Must match the
    # endpoint's own region — the Vertex AI API host is regional
    # (https://{region}-aiplatform.googleapis.com) and a deployment
    # cannot cross regions. Immutable after creation.
    location = string

    # User-chosen ID of this deployment, unique within the project: up to
    # 128 characters, starting with a letter, containing only letters,
    # numbers, and underscores. This is the handle queries and undeploy
    # operations address the deployment by. Immutable after creation.
    deployed_index_id = string

    # The index being deployed — the fully qualified index resource path
    # (projects/{project}/locations/{location}/indexes/{indexId}); a
    # GcpVertexAiIndex reference resolves to it. The index must live in
    # the same region as the endpoint. Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    index = string

    # The index endpoint being deployed onto — the fully qualified
    # endpoint resource path
    # (projects/{project}/locations/{location}/indexEndpoints/{id}); a
    # GcpVertexAiIndexEndpoint reference resolves to it. Immutable after
    # creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    index_endpoint = string

    # Display name of the deployment (up to 128 UTF-8 characters).
    # Unusually for a display name, the API treats it as IMMUTABLE on a
    # deployed index — changing it replaces the deployment.
    display_name = optional(string, "")

    # Vertex-managed serving compute (machine types chosen by GCP,
    # replicas scale between bounds). Mutually exclusive with
    # dedicated_resources; omitting both lets GCP deploy with automatic
    # defaults.
    automatic_resources = optional(object({
      # Minimum replicas always running. GCP's default is 2 (no SLA is
      # provided at 1). Mutable in place — replica bounds are the only
      # post-deploy tuning knobs.
      min_replica_count = optional(number)

      # Maximum replicas under load, up to 1000. Defaults to
      # min_replica_count when unset. Mutable in place.
      max_replica_count = optional(number, 0)
    }))

    # Explicitly pinned serving compute (machine type + replica bounds).
    # Mutually exclusive with automatic_resources.
    dedicated_resources = optional(object({
      # Machine type serving the index (e.g. "e2-standard-16"). Must be
      # compatible with the index's shard_size. If omitted, the API applies
      # its own default. Immutable.
      machine_type = optional(string, "")

      # Minimum replicas always running, at least 1 (no SLA at 1). Required
      # by the API for dedicated sizing. Mutable in place — replica bounds
      # are the only post-deploy tuning knobs.
      min_replica_count = number

      # Maximum replicas under load, up to 1000. Defaults to
      # min_replica_count when unset. Mutable in place.
      max_replica_count = optional(number, 0)
    }))

    # Deployment group for IP-space partitioning, up to 64 characters
    # (e.g. "test", "prod"); GCP's default group is "default". Pairing
    # groups with reserved_ip_ranges gives each group a predictable IP
    # space when the peered network has multiple peering ranges — and the
    # API HOLDS the pairing: a non-default group, once used with a set of
    # reserved ranges, can only ever be used with exactly that set again.
    # At most 5 groups besides "default". Immutable after creation.
    deployment_group = optional(string)

    # If true, private-endpoint access logs are sent to Cloud Logging.
    # Immutable after creation.
    enable_access_logging = optional(bool, false)

    # Names of reserved compute address ranges under the endpoint's
    # peered VPC network to deploy into (e.g. ["vertex-ai-ip-range"]) —
    # Vertex peering ranges are global INTERNAL VPC_PEERING addresses, so
    # a GcpGlobalAddress reference resolves to its name. If omitted, the
    # index may deploy to any range under the network. Only meaningful on
    # a VPC-peered endpoint. Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    reserved_ip_ranges = optional(list(string), [])

    # JWT authentication for the private query endpoint. If omitted, the
    # endpoint relies on network reachability alone. Immutable after
    # creation.
    auth_config = optional(object({
      # Service accounts whose signed JWTs are accepted, each in the form
      # service-account-name@project-id.iam.gserviceaccount.com. Immutable.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      allowed_issuers = optional(list(string), [])

      # JWT audiences accepted on the query endpoint; a JWT carrying any of
      # them is accepted. Immutable.
      audiences = optional(list(string), [])
    }))

    # Deletion policy for the deployment — what happens when this
    # resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the index is undeployed from the endpoint; queries
    #                against this deployment stop being served
    #   "PREVENT" -- destroy FAILS; a guard for a serving path whose
    #                disappearance would break live query traffic
    #   "ABANDON" -- the deployment is removed from management but keeps
    #                serving (and billing for its replicas) in GCP
    deletion_policy = optional(string, "")
  })
}
