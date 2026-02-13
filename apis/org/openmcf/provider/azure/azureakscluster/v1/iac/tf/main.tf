# Create Azure Kubernetes Service (AKS) Cluster
resource "azurerm_kubernetes_cluster" "aks" {
  name                = local.cluster_name
  location            = var.spec.region
  resource_group_name = local.resource_group_name
  dns_prefix          = local.dns_prefix
  kubernetes_version  = var.spec.kubernetes_version

  # Control Plane SKU
  sku_tier = local.sku_tier

  # System node pool (default node pool)
  default_node_pool {
    name                = "system"
    node_count          = var.spec.system_node_pool.autoscaling.min_count
    vm_size             = var.spec.system_node_pool.vm_size
    vnet_subnet_id      = var.spec.vnet_subnet_id
    enable_auto_scaling = true
    min_count           = var.spec.system_node_pool.autoscaling.min_count
    max_count           = var.spec.system_node_pool.autoscaling.max_count
    zones               = var.spec.system_node_pool.availability_zones
    
    # System node pool should be labeled and tainted for system workloads only
    node_labels = {
      "node-role" = "system"
    }

    only_critical_addons_enabled = true
  }

  # System-Assigned Managed Identity
  identity {
    type = "SystemAssigned"
  }

  # Network Profile
  network_profile {
    network_plugin      = local.network_plugin
    network_plugin_mode = local.network_plugin_mode
    load_balancer_sku   = "standard"
    service_cidr        = local.service_cidr
    dns_service_ip      = local.dns_service_ip
    pod_cidr            = local.pod_cidr
  }

  # API Server Access Profile
  dynamic "api_server_access_profile" {
    for_each = length(var.spec.authorized_ip_ranges) > 0 ? [1] : []
    content {
      authorized_ip_ranges = var.spec.authorized_ip_ranges
    }
  }

  # Private cluster configuration
  private_cluster_enabled = var.spec.private_cluster_enabled

  # Azure AD RBAC Integration
  dynamic "azure_active_directory_role_based_access_control" {
    for_each = local.azure_ad_rbac_enabled ? [1] : []
    content {
      managed            = true
      azure_rbac_enabled = true
    }
  }

  # OMS Agent (Container Insights) addon
  dynamic "oms_agent" {
    for_each = local.enable_container_insights ? [1] : []
    content {
      log_analytics_workspace_id = var.spec.addons.log_analytics_workspace_id
    }
  }

  # Key Vault Secrets Provider addon
  dynamic "key_vault_secrets_provider" {
    for_each = var.spec.addons.enable_key_vault_csi_driver ? [1] : []
    content {
      secret_rotation_enabled = true
    }
  }

  # Azure Policy addon
  dynamic "azure_policy_enabled" {
    for_each = var.spec.addons.enable_azure_policy ? [1] : []
    content {}
  }

  # Workload Identity (OIDC Issuer)
  oidc_issuer_enabled       = var.spec.addons.enable_workload_identity
  workload_identity_enabled = var.spec.addons.enable_workload_identity

  tags = local.tags

  lifecycle {
    ignore_changes = [
      # Ignore changes to tags added by Azure or other processes
      tags["LastModified"],
    ]
  }
}

# Create User Node Pools
resource "azurerm_kubernetes_cluster_node_pool" "user_pools" {
  for_each = { for pool in var.spec.user_node_pools : pool.name => pool }

  name                  = each.value.name
  kubernetes_cluster_id = azurerm_kubernetes_cluster.aks.id
  vm_size               = each.value.vm_size
  vnet_subnet_id        = var.spec.vnet_subnet_id
  
  enable_auto_scaling = true
  min_count           = each.value.autoscaling.min_count
  max_count           = each.value.autoscaling.max_count
  zones               = each.value.availability_zones

  # Spot instance configuration
  priority        = each.value.spot_enabled ? "Spot" : "Regular"
  eviction_policy = each.value.spot_enabled ? "Delete" : null
  spot_max_price  = each.value.spot_enabled ? -1 : null  # -1 means pay up to regular price

  node_labels = {
    "node-role" = "user"
    "pool-name" = each.value.name
  }

  tags = local.tags

  lifecycle {
    ignore_changes = [
      tags["LastModified"],
    ]
  }
}

