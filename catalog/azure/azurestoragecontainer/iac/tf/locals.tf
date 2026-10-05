locals {
  # The spec's access-type enum arrives as the FULL proto value name;
  # the map carries the complete vocabulary, translated to azurerm's
  # lowercase wire values. Unset means private -- the container is born
  # locked down unless the spec says otherwise.
  container_access_type_map = {
    "PRIVATE"   = "private"
    "BLOB"      = "blob"
    "CONTAINER" = "container"
  }

  container_access_type = (
    var.spec.container_access_type == null || var.spec.container_access_type == "" ? "private" :
    local.container_access_type_map[var.spec.container_access_type]
  )

  # The encryption-scope pair: the scope is sent only when non-empty
  # (azurerm treats empty and omitted differently on ForceNew fields),
  # and the override flag rides with it -- Azure's default is true when
  # a scope is set.
  default_encryption_scope = (
    var.spec.default_encryption_scope == null || var.spec.default_encryption_scope == "" ? null :
    var.spec.default_encryption_scope
  )

  # The account name, parsed from the resolved account ARM ID for the
  # output -- consumers frequently need the account/container name
  # pair, and this saves them a second reference. The named-group regex
  # fails the plan loudly if the ID is not a storage-account ARM ID.
  storage_account_name = regex("/storageAccounts/(?P<name>[^/]+)$", var.spec.storage_account_id)["name"]
}
