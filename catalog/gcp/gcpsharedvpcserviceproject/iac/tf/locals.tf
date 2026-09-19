locals {
  # Both projects arrive as resolved references (the host from a
  # GcpSharedVpcHost's host_project_id, the service project from a
  # GcpProject's project_id) and the attachment carries no name or labels of
  # its own, so nothing is derived here. The one projection: an empty
  # deletion_policy defers to the provider (detach on destroy).
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
