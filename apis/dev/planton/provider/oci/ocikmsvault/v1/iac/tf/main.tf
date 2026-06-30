resource "oci_kms_vault" "this" {
  compartment_id = var.spec.compartment_id.value
  display_name   = local.display_name
  vault_type     = lookup(local.vault_type_map, var.spec.vault_type, var.spec.vault_type)
  freeform_tags  = local.freeform_tags

  dynamic "external_key_manager_metadata" {
    for_each = var.spec.external_key_manager_metadata != null ? [var.spec.external_key_manager_metadata] : []
    content {
      external_vault_endpoint_url = external_key_manager_metadata.value.external_vault_endpoint_url
      oauth_metadata {
        client_app_id        = external_key_manager_metadata.value.oauth_metadata.client_app_id
        client_app_secret    = external_key_manager_metadata.value.oauth_metadata.client_app_secret
        idcs_account_name_url = external_key_manager_metadata.value.oauth_metadata.idcs_account_name_url
      }
      private_endpoint_id = external_key_manager_metadata.value.private_endpoint_id
    }
  }
}
