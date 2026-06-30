locals {
  # Resource group name (provided externally)
  resource_group_name = var.spec.resource_group

  # Cluster name
  cluster_name = var.metadata.name

  # Control plane SKU tier
  sku_tier = var.spec.control_plane_sku == "FREE" ? "Free" : "Standard"

  # Determine network plugin
  network_plugin = var.spec.network_plugin == "KUBENET" ? "kubenet" : "azure"

  # Determine network plugin mode
  network_plugin_mode = var.spec.network_plugin == "AZURE_CNI" ? (
    var.spec.network_plugin_mode == "DYNAMIC" ? "dynamic" : "overlay"
  ) : null

  # Azure AD RBAC enabled
  azure_ad_rbac_enabled = !var.spec.disable_azure_ad_rbac

  # Container Insights enabled
  enable_container_insights = var.spec.addons.enable_container_insights && var.spec.addons.log_analytics_workspace_id != ""

  # Network CIDRs with defaults
  pod_cidr = (
    var.spec.advanced_networking.pod_cidr != "" 
    ? var.spec.advanced_networking.pod_cidr 
    : (var.spec.network_plugin_mode == "OVERLAY" ? "10.244.0.0/16" : null)
  )
  
  service_cidr = (
    var.spec.advanced_networking.service_cidr != "" 
    ? var.spec.advanced_networking.service_cidr 
    : "10.0.0.0/16"
  )
  
  dns_service_ip = (
    var.spec.advanced_networking.dns_service_ip != "" 
    ? var.spec.advanced_networking.dns_service_ip 
    : "10.0.0.10"
  )

  # Tags from metadata
  tags = merge(
    var.metadata.labels != null ? var.metadata.labels : {},
    {
      Name        = var.metadata.name
      Environment = var.metadata.env != null ? var.metadata.env : "default"
      ManagedBy   = "terraform"
    }
  )

  # DNS prefix
  dns_prefix = "${var.metadata.name}-dns"
}

