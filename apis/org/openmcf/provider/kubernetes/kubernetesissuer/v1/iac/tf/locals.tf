locals {
  namespace   = var.spec.namespace
  issuer_name = var.metadata.name

  is_ca          = var.spec.ca != null
  is_self_signed = var.spec.self_signed != null

  # Build the cert-manager Issuer spec based on issuer type.
  # CA requires a secretName reference; SelfSigned is an empty object.
  issuer_spec = local.is_ca ? {
    ca = {
      secretName = var.spec.ca.ca_secret_name
    }
  } : {
    selfSigned = {}
  }
}
