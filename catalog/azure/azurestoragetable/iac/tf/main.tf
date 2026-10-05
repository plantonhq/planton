# The Storage table, addressed by the parent account's ARM ID (the
# control-plane path -- the account-name form is the provider's legacy
# data-plane path, removed in azurerm v5). Tables carry no Azure tags:
# ARM does not support tags on tableServices/tables, so the platform's
# identity tags live on the account.
#
# PARITY-EXCEPTION: this module addresses the table by storage_account_id
# (the resource-manager path) while the Pulumi module passes the account
# NAME parsed from the same ARM id -- pulumi-azure v6 has not yet bridged
# the table's storage_account_id input (verified at v6.38, the latest v6).
# The created table is identical and all outputs match byte-for-byte
# (both engines export the same resource_manager_id); only the provider's
# internal addressing differs. Re-align the Pulumi module when a bridge
# release carries storageAccountId on storage.Table.
#
# Operational contract: the provider drives table creation and ACLs
# through the table DATA PLANE with shared-key authorization, so the
# parent account must keep shared_access_key_enabled true (Azure's
# default) for deploys to work.
resource "azurerm_storage_table" "main" {
  name               = var.spec.table_name
  storage_account_id = var.spec.storage_account_id

  # Stored access policies (signed identifiers): revoking or shortening
  # a policy immediately revokes every SAS token anchored to it. Table
  # policies require the full validity window (start + expiry). At most
  # five per table (Azure's limit, enforced in the spec).
  dynamic "acl" {
    for_each = var.spec.acls
    content {
      id = acl.value.id

      dynamic "access_policy" {
        for_each = acl.value.access_policies
        content {
          permissions = access_policy.value.permissions
          start       = access_policy.value.start
          expiry      = access_policy.value.expiry
        }
      }
    }
  }
}
