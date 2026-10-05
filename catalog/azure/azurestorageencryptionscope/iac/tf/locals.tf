locals {
  # The spec's source enum arrives as the FULL proto value name; the map
  # carries the complete vocabulary, translated to ARM's dotted wire
  # values.
  source_map = {
    "MICROSOFT_STORAGE"   = "Microsoft.Storage"
    "MICROSOFT_KEY_VAULT" = "Microsoft.KeyVault"
  }

  source = local.source_map[var.spec.source]

  # Sent only when set -- ARM pairs the key with the Microsoft.KeyVault
  # source (the spec enforces required-when-KeyVault).
  key_vault_key_id = (
    var.spec.key_vault_key_id == null || var.spec.key_vault_key_id == "" ? null :
    var.spec.key_vault_key_id
  )

  # The account name, parsed from the resolved account ARM ID for the
  # output -- consumers frequently need the account/scope name
  # pair, and this saves them a second reference. The named-group regex
  # fails the plan loudly if the ID is not a storage-account ARM ID.
  storage_account_name = regex("/storageAccounts/(?P<name>[^/]+)$", var.spec.storage_account_id)["name"]
}
