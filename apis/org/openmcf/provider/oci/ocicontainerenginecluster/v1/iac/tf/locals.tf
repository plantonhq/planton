locals {
  resource_id = coalesce(var.metadata.id, var.metadata.name)

  display_name = coalesce(var.spec.name, var.metadata.name)

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "OciContainerEngineCluster"
  }

  org_tag = var.metadata.org != "" ? { "organization" = var.metadata.org } : {}
  env_tag = var.metadata.env != "" ? { "environment" = var.metadata.env } : {}

  freeform_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.metadata.labels)

  cluster_type_map = {
    "basic_cluster"    = "BASIC_CLUSTER"
    "enhanced_cluster" = "ENHANCED_CLUSTER"
  }

  cni_type_map = {
    "flannel_overlay"   = "FLANNEL_OVERLAY"
    "oci_vcn_ip_native" = "OCI_VCN_IP_NATIVE"
  }

  ip_family_map = {
    "ipv4" = "IPv4"
    "ipv6" = "IPv6"
  }

  endpoint_nsg_ids = var.spec.endpoint_config != null ? [
    for nsg in var.spec.endpoint_config.nsg_ids : nsg.value
  ] : []

  service_lb_subnet_ids = var.spec.options != null ? [
    for s in var.spec.options.service_lb_subnet_ids : s.value
  ] : []

  service_lb_backend_nsg_ids = var.spec.options != null && var.spec.options.service_lb_config != null ? [
    for nsg in var.spec.options.service_lb_config.backend_nsg_ids : nsg.value
  ] : []

  ip_families = var.spec.options != null ? [
    for f in var.spec.options.ip_families : lookup(local.ip_family_map, f, f)
  ] : []
}
