locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Create storage account name from metadata.name
  # Azure Storage Account names must be 3-24 characters, lowercase letters and numbers only
  storage_account_name_raw = lower(replace(replace(replace(var.metadata.name, ".", ""), "-", ""), "_", ""))
  storage_account_name     = substr(local.storage_account_name_raw, 0, min(24, length(local.storage_account_name_raw)))

  # Base tags for Azure resources
  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "azure_storage_account"
    "resource_name" = var.metadata.name
  }

  # Organization tag only if var.metadata.org is non-empty
  org_tag = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  # Environment tag only if var.metadata.env is non-empty
  env_tag = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "environment" = var.metadata.env } : {}

  # Merge base, org, and environment tags
  final_tags = merge(local.base_tags, local.org_tag, local.env_tag)

  # Network rules with defaults
  network_rules = var.spec.network_rules != null ? var.spec.network_rules : {
    default_action             = "Deny"
    bypass_azure_services      = true
    ip_rules                   = []
    virtual_network_subnet_ids = []
  }

  # Convert bypass_azure_services boolean to Azure bypass list
  network_bypass = local.network_rules.bypass_azure_services ? ["AzureServices"] : []

  # Blob properties with defaults
  blob_properties = var.spec.blob_properties != null ? var.spec.blob_properties : {
    enable_versioning                    = false
    soft_delete_retention_days           = 7
    container_soft_delete_retention_days = 7
  }

  # Create a map of containers for for_each
  containers_map = {
    for container in var.spec.containers :
    container.name => container
  }
}
