locals {
  # The account name, parsed from the resolved account ARM ID for the
  # output -- consumers frequently need the account/queue name
  # pair, and this saves them a second reference. The named-group regex
  # fails the plan loudly if the ID is not a storage-account ARM ID.
  storage_account_name = regex("/storageAccounts/(?P<name>[^/]+)$", var.spec.storage_account_id)["name"]
}
