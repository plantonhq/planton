locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The cluster's GCP name: spec.cluster_name when set, otherwise
  # metadata.name -- the same fallback the Pulumi module applies.
  cluster_name = var.spec.cluster_name != "" ? var.spec.cluster_name : var.metadata.name
  region       = var.spec.region
  shard_count  = var.spec.shard_count

  # 0 means "no replicas" -- but the provider treats null as "server
  # default" and 0 as an explicit choice; send 0 through so the manifest
  # value is authoritative either way.
  replica_count = var.spec.replica_count

  # The two immutable security modes are sent explicitly with Google's
  # defaults when unset, so the cluster's posture never depends on a
  # provider default (identical to the Pulumi module).
  authorization_mode      = var.spec.authorization_mode != "" ? var.spec.authorization_mode : "AUTH_MODE_DISABLED"
  transit_encryption_mode = var.spec.transit_encryption_mode != "" ? var.spec.transit_encryption_mode : "TRANSIT_ENCRYPTION_MODE_DISABLED"

  # Optional+Computed levers: empty translates to null so Google's own
  # default applies instead of an empty value being sent verbatim.
  node_type           = var.spec.node_type != "" ? var.spec.node_type : null
  server_ca_mode      = var.spec.server_ca_mode != "" ? var.spec.server_ca_mode : null
  server_ca_pool      = var.spec.server_ca_pool != "" ? var.spec.server_ca_pool : null
  kms_key             = var.spec.kms_key != "" ? var.spec.kms_key : null
  maintenance_version = var.spec.maintenance_version != "" ? var.spec.maintenance_version : null
  acl_policy          = var.spec.acl_policy != "" ? var.spec.acl_policy : null

  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = local.cluster_name
    "planton-ai_kind"     = "gcprediscluster"
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

  # User labels first so platform attribution labels win on key conflicts --
  # identical merge order to the Pulumi module.
  final_labels = merge(
    var.spec.labels,
    local.base_labels,
    local.org_label,
    local.env_label,
    local.id_label,
  )

  # One scalar handle per connection type, so a consumer's forwarding rule
  # (and GcpRedisClusterEndpointSet) can reference the attachment it needs
  # -- a reference cannot index a repeated output. Empty when Google
  # publishes no attachment of that type (no reader without replicas).
  service_attachments_by_type = {
    for a in google_redis_cluster.this.psc_service_attachments : a.connection_type => a.service_attachment...
  }
}
