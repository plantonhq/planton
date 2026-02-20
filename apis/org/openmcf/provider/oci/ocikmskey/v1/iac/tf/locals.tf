locals {
  resource_id = coalesce(var.metadata.id, var.metadata.name)

  display_name = var.spec.display_name != "" ? var.spec.display_name : var.metadata.name

  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "OciKmsKey"
  }

  org_tag = var.metadata.org != "" ? { "organization" = var.metadata.org } : {}
  env_tag = var.metadata.env != "" ? { "environment" = var.metadata.env } : {}

  freeform_tags = merge(local.base_tags, local.org_tag, local.env_tag, var.metadata.labels)

  protection_mode_map = {
    "hsm"      = "HSM"
    "software" = "SOFTWARE"
    "external" = "EXTERNAL"
  }

  algorithm_map = {
    "aes"   = "AES"
    "rsa"   = "RSA"
    "ecdsa" = "ECDSA"
  }

  curve_id_map = {
    "nist_p256" = "NIST_P256"
    "nist_p384" = "NIST_P384"
    "nist_p521" = "NIST_P521"
  }
}
