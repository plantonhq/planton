locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  instance_name = var.spec.instance_name
  location      = var.spec.location
  shard_count   = var.spec.shard_count

  # Empty-string enums translate to null so the provider applies its own
  # defaults instead of receiving an empty value verbatim.
  mode           = var.spec.mode != "" ? var.spec.mode : null
  node_type      = var.spec.node_type != "" ? var.spec.node_type : null
  engine_version = var.spec.engine_version != "" ? var.spec.engine_version : null

  # 0 means "no replicas" — but the provider treats null as "server
  # default" and 0 as an explicit choice; send 0 through so the manifest
  # value is authoritative either way.
  replica_count = var.spec.replica_count

  authorization_mode      = var.spec.authorization_mode != "" ? var.spec.authorization_mode : null
  transit_encryption_mode = var.spec.transit_encryption_mode != "" ? var.spec.transit_encryption_mode : null
  kms_key                 = var.spec.kms_key != "" ? var.spec.kms_key : null
  server_ca_mode          = var.spec.server_ca_mode != "" ? var.spec.server_ca_mode : null
  server_ca_pool          = var.spec.server_ca_pool != "" ? var.spec.server_ca_pool : null
  maintenance_version     = var.spec.maintenance_version != "" ? var.spec.maintenance_version : null
  acl_policy              = var.spec.acl_policy != "" ? var.spec.acl_policy : null

  # The ambient-project data source is only instantiated when some PSC
  # endpoint entry omits its consumer project AND the spec carries no
  # explicit project to inherit — mirroring the Pulumi module's
  # only-when-needed GetClientConfig lookup and keeping offline plans
  # free of live API calls.
  needs_ambient_endpoint_project = (
    var.spec.project_id == "" &&
    anytrue([for c in var.spec.psc_auto_connections : c.project_id == ""])
  )

  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = local.instance_name
    "planton-ai_kind"     = "gcpmemorystoreinstance"
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

  # User labels first so platform attribution labels win on key conflicts —
  # identical merge order to the Pulumi module.
  final_labels = merge(
    var.spec.labels,
    local.base_labels,
    local.org_label,
    local.env_label,
    local.id_label,
  )
}
