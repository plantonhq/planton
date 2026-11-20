locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain; an empty string would be sent
  # verbatim and rejected by the API.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # Cluster name defaults to metadata.name (the spec-level contract), so a
  # manifest only sets cluster_name when the cloud name must differ from the
  # Planton object name.
  cluster_name = var.spec.cluster_name != "" ? var.spec.cluster_name : var.metadata.name

  # The spec's NONE channel (opt out of channel-based upgrades) is spelled
  # UNSPECIFIED on the provider — the API has no literal NONE value.
  release_channel = var.spec.release_channel == "NONE" ? "UNSPECIFIED" : var.spec.release_channel

  # Empty optional strings become null so the provider omits them from the
  # API payload instead of sending empty values it would reject or diff on.
  description                = var.spec.description != "" ? var.spec.description : null
  datapath_provider          = var.spec.datapath_provider != "" ? var.spec.datapath_provider : null
  private_ipv6_google_access = var.spec.private_ipv6_google_access != "" ? var.spec.private_ipv6_google_access : null
  in_transit_encryption      = var.spec.in_transit_encryption != "" ? var.spec.in_transit_encryption : null
  min_master_version         = var.spec.min_master_version != "" ? var.spec.min_master_version : null

  # Workload Identity is a project-scoped pool with exactly one valid value
  # for standard clusters. Autopilot has it always on — sending the block
  # would conflict, so it is suppressed there.
  workload_pool = "${coalesce(local.project_id, try(data.google_client_config.current[0].project, null))}.svc.id.goog"

  # Autopilot clusters manage their own nodes: the default-node-pool
  # removal fields, shielded-nodes flag, and Calico network policy must be
  # omitted entirely (the API rejects them). Standard clusters always drop
  # the default node pool — node pools are separate GcpGkeNodePool
  # resources.
  is_autopilot = var.spec.enable_autopilot

  # Logging is on unless the spec declares it off (enabled: false); an
  # omitted logging block keeps GKE's default, which ships logs.
  logging_enabled = try(var.spec.logging.enabled, null) != false

  # The same planton-ai_* label set the Pulumi module applies, so a cluster
  # is attributable to its Planton object regardless of the engine that
  # created it. The cluster name (not metadata.name) keys the name label so
  # the label matches what is visible in the GCP console.
  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = local.cluster_name
    "planton-ai_kind"     = "gcpgkecluster"
  }

  org_label = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "planton-ai_organization" = var.metadata.org } : {}

  env_label = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "planton-ai_environment" = var.metadata.env } : {}

  id_label = (
    var.metadata.id != null && var.metadata.id != ""
  ) ? { "planton-ai_id" = var.metadata.id } : {}

  # User labels merge in first so the platform attribution labels can never
  # be clobbered by a spec label with the same key.
  final_labels = merge(
    var.spec.resource_labels,
    local.base_labels,
    local.org_label,
    local.env_label,
    local.id_label,
  )
}

# Resolves the provider's ambient project for the Workload Identity pool
# name when spec.project_id is empty (the pool is always
# PROJECT_ID.svc.id.goog and the API does not accept a shorthand). Read
# ONLY in that case: an explicit project_id must not force a live
# credential lookup at plan time.
data "google_client_config" "current" {
  count = var.spec.project_id != "" ? 0 : 1
}
