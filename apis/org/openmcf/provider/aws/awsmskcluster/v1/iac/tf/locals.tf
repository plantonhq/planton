locals {
  tags = {
    "planton.org/resource"      = "true"
    "planton.org/organization"  = var.metadata.org
    "planton.org/environment"   = var.metadata.env
    "planton.org/resource-kind" = "AwsMskCluster"
    "planton.org/resource-id"   = var.metadata.id
  }

  has_ingress_refs = length(var.spec.security_group_ids) > 0 || length(var.spec.allowed_cidr_blocks) > 0

  # Serialize server_properties map to .properties format
  server_properties = length(var.spec.server_properties) > 0 ? join("\n", [
    for k, v in var.spec.server_properties : "${k} = ${v}"
  ]) : ""
}
