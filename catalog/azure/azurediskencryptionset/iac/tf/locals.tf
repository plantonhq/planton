locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CatalogKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azurediskencryptionset"
    "resource_name" = var.metadata.name
  }

  org_tag = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  env_tag = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "environment" = var.metadata.env } : {}

  id_tag = (
    var.metadata.id != null && var.metadata.id != ""
  ) ? { "resource_id" = var.metadata.id } : {}

  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, local.id_tag, var.spec.tags)

  # Map the spec enums' name strings to ARM's values. null lets Azure apply
  # its default (EncryptionAtRestWithCustomerKey); only an explicit choice is
  # sent so an unspecified spec and Azure's default deploy identically on
  # both engines.
  encryption_type = (
    var.spec.encryption_type == "ENCRYPTION_AT_REST_WITH_CUSTOMER_KEY" ? "EncryptionAtRestWithCustomerKey" :
    var.spec.encryption_type == "ENCRYPTION_AT_REST_WITH_PLATFORM_AND_CUSTOMER_KEYS" ? "EncryptionAtRestWithPlatformAndCustomerKeys" :
    var.spec.encryption_type == "CONFIDENTIAL_VM_ENCRYPTED_WITH_CUSTOMER_KEY" ? "ConfidentialVmEncryptedWithCustomerKey" : null
  )

  identity_type = (
    var.spec.identity.type == "SYSTEM_ASSIGNED" ? "SystemAssigned" :
    var.spec.identity.type == "USER_ASSIGNED" ? "UserAssigned" :
    var.spec.identity.type == "SYSTEM_AND_USER_ASSIGNED" ? "SystemAssigned, UserAssigned" : null
  )
}
