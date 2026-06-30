locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Create vault name from metadata.name
  # Azure Key Vault names must be 3-24 characters, alphanumeric and hyphens only
  vault_name_raw = replace(replace(var.metadata.name, ".", "-"), "_", "-")
  vault_name     = substr(local.vault_name_raw, 0, min(24, length(local.vault_name_raw)))

  # Base tags for Azure resources
  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "azure_key_vault"
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

  # Network ACLs with defaults
  network_acls = var.spec.network_acls != null ? var.spec.network_acls : {
    default_action        = "Deny"
    bypass_azure_services = true
    ip_rules              = []
    virtual_network_subnet_ids = []
  }

  # Convert bypass_azure_services boolean to Azure bypass string
  network_bypass = local.network_acls.bypass_azure_services ? "AzureServices" : "None"
}

