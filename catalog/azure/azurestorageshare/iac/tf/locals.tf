locals {
  # The spec's protocol and tier enums arrive as the FULL proto value
  # names; the maps carry the complete vocabularies, translated to
  # azurerm's wire values. Protocol unset means SMB -- azurerm's own
  # default -- and the tier is sent only when the spec chooses one, so
  # Azure's per-account-kind default (TransactionOptimized on standard,
  # Premium on FileStorage) applies when unset.
  enabled_protocol_map = {
    "SMB" = "SMB"
    "NFS" = "NFS"
  }

  access_tier_map = {
    "TRANSACTION_OPTIMIZED" = "TransactionOptimized"
    "HOT"                   = "Hot"
    "COOL"                  = "Cool"
    "PREMIUM"               = "Premium"
  }

  enabled_protocol = (
    var.spec.enabled_protocol == null || var.spec.enabled_protocol == "" ? "SMB" :
    local.enabled_protocol_map[var.spec.enabled_protocol]
  )

  access_tier = (
    var.spec.access_tier == null || var.spec.access_tier == "" ? null :
    local.access_tier_map[var.spec.access_tier]
  )

  # The account name, parsed from the resolved account ARM ID for the
  # output -- consumers frequently need the account/share name
  # pair, and this saves them a second reference. The named-group regex
  # fails the plan loudly if the ID is not a storage-account ARM ID.
  storage_account_name = regex("/storageAccounts/(?P<name>[^/]+)$", var.spec.storage_account_id)["name"]
}
