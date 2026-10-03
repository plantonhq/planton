variable "metadata" {
  description = "Catalog object metadata"
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
  description = "GcpDeployTarget specification"
  type = object({
    # The project the target lives in: a literal project ID or a GcpProject
    # reference. Empty means the provider's default project. The delivery
    # pipelines that deploy to it must be in the same project. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The region the target lives in, e.g. "us-central1" -- the region of
    # the delivery pipelines that deploy to it. It need not be the region
    # the workload runs in (a run target names its own region). Required.
    # Immutable.
    location = string

    # The target's ID, unique in the project and region: a lowercase letter,
    # then up to 62 lowercase letters, digits, or hyphens, not ending with a
    # hyphen (Google's rule). Pipeline stages name the target by this ID.
    # Defaults to metadata.name. Immutable.
    target_id = optional(string, "")

    # A description shown in the console, up to 255 characters.
    description = optional(string, "")

    # Labels on the target. Cloud Deploy also reads them: a deploy policy
    # or an automation can select targets by label. Keys and values are
    # lowercase letters, digits, underscores, and dashes, at most 64 labels.
    # The platform attribution labels are added on top and win on key
    # conflicts.
    labels = optional(map(string), {})

    # Annotations on the target (AIP-128 key/value metadata Cloud Deploy
    # never reads). Only the keys declared here are managed.
    annotations = optional(map(string), {})

    # Require an approval before any rollout to this target proceeds: a
    # principal with clouddeploy.rollouts.approve (roles/clouddeploy.approver)
    # approves or rejects each one. The usual gate in front of production.
    require_approval = optional(bool, false)

    # Deploy parameters for every rollout to this target: key/value pairs
    # substituted into the rendered manifests wherever a
    # "# from-param: ${key}" marker names the key. Pipeline stages and
    # releases can declare parameters too.
    deploy_parameters = optional(map(string), {})

    # Deploy to a GKE cluster. Exactly one target type.
    gke = optional(object({
      # The cluster, by full name
      # (projects/{project}/locations/{location}/clusters/{name}): a
      # GcpGkeCluster reference (its cluster_id output) or the literal name.
      # The cluster may be in another project; the execution service account
      # then needs access there.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      cluster = optional(string, "")

      # Reach the control plane on its private IP address. Only for clusters
      # with a private endpoint; the default address is already private for a
      # cluster whose endpoint is private-only. Cloud Build must then run on a
      # private pool with a route to that network (execution_configs).
      # Cannot be combined with dns_endpoint.
      internal_ip = optional(bool, false)

      # Reach the control plane through its DNS-based endpoint. Cannot be
      # combined with internal_ip.
      dns_endpoint = optional(bool, false)

      # An HTTP proxy Cloud Deploy reaches the Kubernetes API server through
      # (the kubeconfig proxy-url), e.g. "http://10.0.0.5:3128".
      proxy_url = optional(string, "")
    }))

    # Deploy to a cluster registered to a fleet, through its membership
    # (GKE Enterprise, including attached and on-premises clusters). Exactly
    # one target type.
    anthos_cluster = optional(object({
      # The cluster's fleet membership, by full name
      # (projects/{project}/locations/{location}/memberships/{id}): a
      # GcpGkeFleetMembership reference (its name output), a GcpGkeCluster
      # reference (its fleet_membership output, for a cluster that joined
      # through fleet_project), or the literal name.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      membership = optional(string, "")
    }))

    # Deploy Cloud Run services or jobs into a region. Exactly one target
    # type.
    run = optional(object({
      # Where the Cloud Run services or jobs are deployed, as
      # projects/{project}/locations/{region}, e.g.
      # "projects/my-app-prod/locations/us-central1". The project may differ
      # from the target's own (a pipeline in a CI project deploying into
      # per-environment projects). Required.
      location = string
    }))

    # Deploy to several other targets at once, in parallel. Exactly one
    # target type.
    multi_target = optional(object({
      # The child targets, by ID: GcpDeployTarget references (their target_id
      # output) or literal IDs. Each must be in this target's project and
      # region. A rollout to the multi-target deploys to every child in
      # parallel. At least one.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      target_ids = list(string)
    }))

    # Deploy with a custom target type: your own render and deploy
    # containers or Skaffold modules, for platforms Cloud Deploy does not
    # deploy natively. Exactly one target type.
    custom_target = optional(object({
      # The custom target type, by full name
      # (projects/{project}/locations/{location}/customTargetTypes/{id}): a
      # GcpDeployCustomTargetType reference (its name output) or the literal
      # name. Required.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      custom_target_type = string
    }))

    # Clusters other than the deployment target that some features deploy
    # to, each under an entity ID. The Gateway API canary, for example, can
    # deploy its HTTPRoute to a separate config cluster named here. Keyed by
    # entity_id.
    associated_entities = optional(list(object({
      # The entity's ID, the key features use to find these clusters: a
      # lowercase letter, then up to 62 lowercase letters, digits, or hyphens,
      # not ending with a hyphen (Google's rule). Required.
      entity_id = string

      # GKE clusters for this entity.
      gke_clusters = optional(list(object({
        # The cluster, by full name
        # (projects/{project}/locations/{location}/clusters/{name}): a
        # GcpGkeCluster reference (its cluster_id output) or the literal name.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        cluster = optional(string, "")

        # Reach the control plane on its private IP address. Only for clusters
        # with a private endpoint.
        internal_ip = optional(bool, false)

        # An HTTP proxy Cloud Deploy reaches the Kubernetes API server through.
        proxy_url = optional(string, "")
      })), [])

      # Fleet-registered clusters for this entity.
      anthos_clusters = optional(list(object({
        # The cluster's fleet membership, by full name
        # (projects/{project}/locations/{location}/memberships/{id}): a
        # GcpGkeFleetMembership reference (its name output), a GcpGkeCluster
        # reference (its fleet_membership output), or the literal name.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        membership = optional(string, "")
      })), [])
    })), [])

    # How Cloud Deploy runs this target's jobs, one configuration per group
    # of usages. Each usage (RENDER, DEPLOY, VERIFY, PREDEPLOY, POSTDEPLOY)
    # may appear in only one configuration, and when any are declared,
    # RENDER and DEPLOY must both be covered. With none, every job runs on
    # Cloud Build's default pool with the project's default compute service
    # account.
    execution_configs = optional(list(object({
      # The jobs this configuration applies to, each at most once across all
      # of the target's configurations:
      #   "RENDER"     -- rendering manifests for a release
      #   "DEPLOY"     -- deploying, and the deployment hooks
      #   "VERIFY"     -- deployment verification
      #   "PREDEPLOY"  -- predeploy jobs
      #   "POSTDEPLOY" -- postdeploy jobs
      # Required.
      usages = list(string)

      # The Cloud Build private pool the jobs run on: a GcpCloudBuildWorkerPool
      # reference (its name output) or a literal
      # projects/{project}/locations/{location}/workerPools/{id}. Empty means
      # Cloud Build's default pool. A private pool is how jobs reach a GKE
      # cluster's private endpoint. The top-level worker_pool,
      # service_account, and artifact_storage and the default_pool and
      # private_pool blocks express the same choice; use one form.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      worker_pool = optional(string, "")

      # The service account the jobs run as, by email: a GcpServiceAccount
      # reference or a literal. Empty means the project's default compute
      # service account. It needs roles/clouddeploy.jobRunner, and the caller
      # creating releases needs roles/iam.serviceAccountUser on it.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      service_account = optional(string, "")

      # Where the jobs store their outputs (rendered manifests, logs): a
      # bucket ("gs://my-bucket") or a path in one ("gs://my-bucket/deploy"),
      # a GcpGcsBucket reference (its gs:// url output) or a literal. Empty
      # means a default bucket Cloud Deploy creates in the target's region.
      # The service account needs write access to it.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      artifact_storage = optional(string, "")

      # How long one job may run, in seconds with an "s" suffix, between 10
      # minutes and 24 hours ("600s" to "86400s"). Empty means one hour.
      execution_timeout = optional(string, "")

      # Turn on verbose logging in the jobs' builds, for debugging.
      verbose = optional(bool, false)

      # Use Cloud Build's default pool, with its own service account and
      # artifact storage (the block form of the top-level fields). Mutually
      # exclusive with private_pool.
      default_pool = optional(object({
        # The service account the jobs run as, by email: a GcpServiceAccount
        # reference or a literal. Empty means the project's default compute
        # service account.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        service_account = optional(string, "")

        # Where the jobs store their outputs: a gs:// bucket or path, a
        # GcpGcsBucket reference (its url output) or a literal. Empty means a
        # default bucket in the target's region.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        artifact_storage = optional(string, "")
      }))

      # Use a Cloud Build private pool, with its own service account and
      # artifact storage (the block form of the top-level fields). Mutually
      # exclusive with default_pool.
      private_pool = optional(object({
        # The private pool: a GcpCloudBuildWorkerPool reference (its name
        # output) or a literal
        # projects/{project}/locations/{location}/workerPools/{id}. Required.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        worker_pool = string

        # The service account the jobs run as, by email: a GcpServiceAccount
        # reference or a literal. Empty means the project's default compute
        # service account.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        service_account = optional(string, "")

        # Where the jobs store their outputs: a gs:// bucket or path, a
        # GcpGcsBucket reference (its url output) or a literal. Empty means a
        # default bucket in the target's region.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        artifact_storage = optional(string, "")
      }))
    })), [])

    # What destroy does:
    #   "" / "DELETE" -- the target is deleted
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the target leaves management and stays in Google
    deletion_policy = optional(string, "")
  })
}
