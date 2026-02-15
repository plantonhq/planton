resource "aws_cognito_identity_provider" "this" {
  user_pool_id  = try(var.spec.userPoolId.value, var.spec.userPoolId, "")
  provider_name = try(var.spec.providerName, var.spec.provider_name, "")
  provider_type = local.provider_type

  provider_details = local.provider_details

  attribute_mapping = local.attribute_mapping

  idp_identifiers = local.idp_identifiers

  lifecycle {
    # AWS auto-populates ActiveEncryptionCertificate for SAML providers.
    # Ignore it to prevent perpetual drift.
    ignore_changes = [
      provider_details["ActiveEncryptionCertificate"],
    ]
  }
}
