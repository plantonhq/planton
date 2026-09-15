# Create the Data Protection backup instance. Exactly one variant
# block is set in the spec (validated at admission); each variant
# creates its own provider resource -- ONE resource exists per
# deployment.
#
# The vault's managed identity must already hold the datasource roles
# Azure Backup requires (disk: "Disk Backup Reader" on the disk +
# "Disk Snapshot Contributor" on the snapshot resource group; blob and
# Data Lake: "Storage Account Backup Contributor" on the storage
# account; Kubernetes: the AKS Backup extension + trusted access;
# MySQL/PostgreSQL: the vault identity's backup roles on the server).
# Azure validates the grants at create time -- an authorization-class
# failure here means missing role assignments, not a module defect.
#
# Nearly everything is ForceNew; only backup_policy_id updates in
# place (and on the kubernetes_cluster variant even that replaces the
# instance -- the provider ships no update path for it). Backup
# instances carry NO tags argument (deliberately absent here -- the
# provider has none).

# --- Blob storage ------------------------------------------------------------
resource "azurerm_data_protection_backup_instance_blob_storage" "main" {
  count = var.spec.blob_storage != null ? 1 : 0

  name     = var.spec.name
  location = var.spec.region
  vault_id = var.spec.vault_id

  backup_policy_id   = var.spec.backup_policy_id
  storage_account_id = var.spec.blob_storage.storage_account_id

  # Sent only when non-empty: the provider omits the containers list
  # from the ARM body when absent (operational-only protection covers
  # the whole account). ONE-WAY once set -- the provider ForceNews on
  # clearing the list, never on changing it.
  storage_account_container_names = length(var.spec.blob_storage.storage_account_container_names) > 0 ? var.spec.blob_storage.storage_account_container_names : null
}

# --- Managed disks -----------------------------------------------------------
resource "azurerm_data_protection_backup_instance_disk" "main" {
  count = var.spec.disk != null ? 1 : 0

  name     = var.spec.name
  location = var.spec.region
  vault_id = var.spec.vault_id

  backup_policy_id = var.spec.backup_policy_id
  disk_id          = var.spec.disk.disk_id

  # Where the incremental snapshots land. Null subscription means the
  # vault's own subscription -- the provider's default.
  snapshot_resource_group_name = var.spec.disk.snapshot_resource_group_name
  snapshot_subscription_id     = var.spec.disk.snapshot_subscription_id
}

# --- Kubernetes (AKS) clusters -----------------------------------------------
# The one variant with NO update path at all -- every field including
# backup_policy_id replaces the instance when changed.
resource "azurerm_data_protection_backup_instance_kubernetes_cluster" "main" {
  count = var.spec.kubernetes_cluster != null ? 1 : 0

  name     = var.spec.name
  location = var.spec.region
  vault_id = var.spec.vault_id

  backup_policy_id      = var.spec.backup_policy_id
  kubernetes_cluster_id = var.spec.kubernetes_cluster.kubernetes_cluster_id

  snapshot_resource_group_name = var.spec.kubernetes_cluster.snapshot_resource_group_name

  dynamic "backup_datasource_parameters" {
    for_each = var.spec.kubernetes_cluster.backup_datasource_parameters != null ? [var.spec.kubernetes_cluster.backup_datasource_parameters] : []
    content {
      included_namespaces              = length(backup_datasource_parameters.value.included_namespaces) > 0 ? backup_datasource_parameters.value.included_namespaces : null
      excluded_namespaces              = length(backup_datasource_parameters.value.excluded_namespaces) > 0 ? backup_datasource_parameters.value.excluded_namespaces : null
      included_resource_types          = length(backup_datasource_parameters.value.included_resource_types) > 0 ? backup_datasource_parameters.value.included_resource_types : null
      excluded_resource_types          = length(backup_datasource_parameters.value.excluded_resource_types) > 0 ? backup_datasource_parameters.value.excluded_resource_types : null
      label_selectors                  = length(backup_datasource_parameters.value.label_selectors) > 0 ? backup_datasource_parameters.value.label_selectors : null
      cluster_scoped_resources_enabled = coalesce(backup_datasource_parameters.value.cluster_scoped_resources_enabled, false)
      volume_snapshot_enabled          = coalesce(backup_datasource_parameters.value.volume_snapshot_enabled, false)
    }
  }
}

# --- MySQL flexible servers ---------------------------------------------------
resource "azurerm_data_protection_backup_instance_mysql_flexible_server" "main" {
  count = var.spec.mysql_flexible_server != null ? 1 : 0

  name     = var.spec.name
  location = var.spec.region
  vault_id = var.spec.vault_id

  backup_policy_id = var.spec.backup_policy_id
  server_id        = var.spec.mysql_flexible_server.server_id
}

# --- PostgreSQL flexible servers -----------------------------------------------
resource "azurerm_data_protection_backup_instance_postgresql_flexible_server" "main" {
  count = var.spec.postgresql_flexible_server != null ? 1 : 0

  name     = var.spec.name
  location = var.spec.region
  vault_id = var.spec.vault_id

  backup_policy_id = var.spec.backup_policy_id
  server_id        = var.spec.postgresql_flexible_server.server_id
}

# --- Data Lake storage ---------------------------------------------------------
# The one variant that names its vault and policy arguments
# differently (data_protection_backup_vault_id /
# backup_policy_data_lake_storage_id) -- same values, provider-side
# renames recorded in the parity manifest.
resource "azurerm_data_protection_backup_instance_data_lake_storage" "main" {
  count = var.spec.data_lake_storage != null ? 1 : 0

  name                            = var.spec.name
  location                        = var.spec.region
  data_protection_backup_vault_id = var.spec.vault_id

  backup_policy_data_lake_storage_id = var.spec.backup_policy_id
  storage_account_id                 = var.spec.data_lake_storage.storage_account_id

  storage_container_names = var.spec.data_lake_storage.storage_container_names
}
