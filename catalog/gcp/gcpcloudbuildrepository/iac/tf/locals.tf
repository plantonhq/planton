locals {
  # The repository lives in its connection's project and region, both
  # parsed from the connection's full name
  # ("projects/{p}/locations/{l}/connections/{c}") -- the same rule as the
  # Pulumi module.
  connection_segments = split("/", var.spec.parent_connection)
  project_id          = local.connection_segments[1]
  location            = local.connection_segments[3]

  # The repository ID defaults to metadata.name -- identical to the Pulumi
  # module.
  repository_id   = var.spec.repository_id != "" ? var.spec.repository_id : var.metadata.name
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
