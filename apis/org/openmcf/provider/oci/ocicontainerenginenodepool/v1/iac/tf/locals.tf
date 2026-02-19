locals {
  resource_id = coalesce(var.metadata.id, var.metadata.name)

  display_name = coalesce(var.spec.name, var.metadata.name)

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "OciContainerEngineNodePool"
  }

  org_tag = var.metadata.org != "" ? { "organization" = var.metadata.org } : {}
  env_tag = var.metadata.env != "" ? { "environment" = var.metadata.env } : {}

  freeform_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.metadata.labels)

  cni_type_map = {
    "flannel_overlay"   = "FLANNEL_OVERLAY"
    "oci_vcn_ip_native" = "OCI_VCN_IP_NATIVE"
  }

  node_nsg_ids = [
    for nsg in var.spec.node_config_details.nsg_ids : nsg.value
  ]

  pod_nsg_ids = (
    var.spec.node_config_details.pod_network_option_details != null
    ? [for nsg in var.spec.node_config_details.pod_network_option_details.pod_nsg_ids : nsg.value]
    : []
  )

  pod_subnet_ids = (
    var.spec.node_config_details.pod_network_option_details != null
    ? [for s in var.spec.node_config_details.pod_network_option_details.pod_subnet_ids : s.value]
    : []
  )
}
