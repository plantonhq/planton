# Create the Azure Machine Learning workspace -- the central home a
# data-science team works in. The workspace is a thin coordination
# object over three REQUIRED companion services (storage account, key
# vault, application insights) and an optional container registry; all
# four attachments are ForceNew.
#
# The spec's CEL contracts already enforce the provider's code-level
# rules (the kind/feature_store pairing, service-side encryption
# requiring the encryption block, the serverless no-public-IP subnet
# rule, cross-type outbound-rule name uniqueness) -- by the time this
# module runs, the shape is legal.
#
# Deletion is a SOFT delete at the service: the workspace becomes a
# purgeable ghost that keeps holding its name. This module's provider
# block (provider.tf) enables
# `machine_learning.purge_soft_deleted_workspace_on_destroy`, so a
# destroy purges the ghost and frees the name.
resource "azurerm_machine_learning_workspace" "main" {
  name                = var.spec.name
  location            = var.spec.region
  resource_group_name = var.spec.resource_group

  # The three required companion services (all ForceNew).
  application_insights_id = var.spec.application_insights_id
  key_vault_id            = var.spec.key_vault_id
  storage_account_id      = var.spec.storage_account_id

  dynamic "identity" {
    for_each = var.spec.identity != null ? [var.spec.identity] : []
    content {
      type         = local.identity_type_wire[identity.value.type]
      identity_ids = length(identity.value.identity_ids) > 0 ? identity.value.identity_ids : null
    }
  }

  # Enum name -> wire value; unspecified ("") omits the property so the
  # provider applies its default, "Default". FEATURE_STORE requires the
  # feature_store block (spec CEL, both directions).
  kind = lookup(local.kind_wire, var.spec.kind, null)

  dynamic "feature_store" {
    for_each = var.spec.feature_store != null && coalesce(var.spec.feature_store.enabled, true) ? [var.spec.feature_store] : []
    content {
      computer_spark_runtime_version = feature_store.value.computer_spark_runtime_version != "" ? feature_store.value.computer_spark_runtime_version : null
      offline_connection_name        = feature_store.value.offline_connection_name != "" ? feature_store.value.offline_connection_name : null
      online_connection_name         = feature_store.value.online_connection_name != "" ? feature_store.value.online_connection_name : null
    }
  }

  primary_user_assigned_identity = var.spec.primary_user_assigned_identity != "" ? var.spec.primary_user_assigned_identity : null

  # ForceNew: attaching or re-pointing a registry replaces the workspace.
  container_registry_id = var.spec.container_registry_id != "" ? var.spec.container_registry_id : null

  # Optional-with-default-true on the provider: emit null when the spec
  # leaves it unset so the provider default applies.
  public_network_access_enabled = var.spec.public_network_access_enabled

  image_build_compute_name = var.spec.image_build_compute_name != "" ? var.spec.image_build_compute_name : null
  description              = var.spec.description != "" ? var.spec.description : null
  friendly_name            = var.spec.friendly_name != "" ? var.spec.friendly_name : null

  # Customer-managed-key encryption; the whole block is ForceNew. The
  # key id is a Key Vault key data-plane URL (versionless follows
  # rotation).
  dynamic "encryption" {
    for_each = var.spec.encryption != null ? [var.spec.encryption] : []
    content {
      key_vault_id              = encryption.value.key_vault_id
      key_id                    = encryption.value.key_id
      user_assigned_identity_id = encryption.value.user_assigned_identity_id != "" ? encryption.value.user_assigned_identity_id : null
    }
  }

  # The managed virtual network. isolation_mode is Optional+Computed on
  # the provider -- unspecified omits it and the value is read back
  # (plans show it known-after-apply when the block is absent).
  dynamic "managed_network" {
    for_each = var.spec.managed_network != null ? [var.spec.managed_network] : []
    content {
      isolation_mode                = lookup(local.isolation_mode_wire, managed_network.value.isolation_mode, null)
      provision_on_creation_enabled = managed_network.value.provision_on_creation_enabled
    }
  }

  high_business_impact = var.spec.high_business_impact

  # "Basic" is the only value the provider accepts at v5; unset applies
  # the provider default (also "Basic").
  sku_name = var.spec.sku_name != "" ? var.spec.sku_name : null

  # SENT ONLY WHEN TRUE: the provider pairs this with the encryption
  # block via RequiredWith, which fires on any SPECIFIED value -- an
  # explicit false without the block fails validation ("all of
  # encryption,service_side_encryption_enabled must be specified" --
  # live-caught). null omits it so the provider default (false)
  # applies; the spec CEL guarantees the encryption block whenever
  # this is true. ForceNew.
  service_side_encryption_enabled = var.spec.service_side_encryption_enabled ? true : null

  v1_legacy_mode_enabled = var.spec.v1_legacy_mode_enabled

  # Enum name -> wire value; unspecified ("") omits the property so the
  # provider applies its default, "AccessKey".
  # PARITY-EXCEPTION: the Pulumi engine's classic SDK cannot express
  # this -- workspaces that set it deploy via this engine only.
  storage_account_access_type = lookup(local.storage_account_access_type_wire, var.spec.storage_account_access_type, null)

  # Serverless compute. NOTE (update behavior, provider-enforced):
  # public_ip_enabled cannot flip true -> false while subnet_id is
  # unset; the static create-time rule is already spec CEL.
  dynamic "serverless_compute" {
    for_each = var.spec.serverless_compute != null ? [var.spec.serverless_compute] : []
    content {
      subnet_id         = serverless_compute.value.subnet_id != "" ? serverless_compute.value.subnet_id : null
      public_ip_enabled = serverless_compute.value.public_ip_enabled
    }
  }

  tags = local.final_tags
}

# The composed FQDN outbound rules: standalone ARM children of the
# workspace's managed network, one per spec entry, keyed by name (the
# three rule types share ONE ARM collection -- cross-type name
# uniqueness is spec CEL). The fqdn_outbound_rule_ids output publishes
# each rule's ARM id. Only effective under AllowOnlyApprovedOutbound
# isolation.
resource "azurerm_machine_learning_workspace_network_outbound_rule_fqdn" "fqdn_rules" {
  for_each = { for rule in var.spec.fqdn_outbound_rules : rule.name => rule }

  name             = each.value.name
  workspace_id     = azurerm_machine_learning_workspace.main.id
  destination_fqdn = each.value.destination_fqdn
}

# The composed private-endpoint outbound rules: the managed VNet
# creates a private endpoint to the named Azure resource. Every field
# is ForceNew (the provider ships no update for this rule type); the
# target/sub-resource pairing is spec CEL for literal ids and
# provider-checked for references.
resource "azurerm_machine_learning_workspace_network_outbound_rule_private_endpoint" "private_endpoint_rules" {
  for_each = { for rule in var.spec.private_endpoint_outbound_rules : rule.name => rule }

  name                = each.value.name
  workspace_id        = azurerm_machine_learning_workspace.main.id
  service_resource_id = each.value.service_resource_id
  sub_resource_target = each.value.sub_resource_target
  spark_enabled       = each.value.spark_enabled
}

# The composed service-tag outbound rules: allow outbound traffic to
# an Azure service tag on the given protocol and ports. Only effective
# under AllowOnlyApprovedOutbound isolation.
resource "azurerm_machine_learning_workspace_network_outbound_rule_service_tag" "service_tag_rules" {
  for_each = { for rule in var.spec.service_tag_outbound_rules : rule.name => rule }

  name         = each.value.name
  workspace_id = azurerm_machine_learning_workspace.main.id
  service_tag  = each.value.service_tag
  protocol     = each.value.protocol
  port_ranges  = each.value.port_ranges
}
