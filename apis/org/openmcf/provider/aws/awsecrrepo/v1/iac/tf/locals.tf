locals {
  # resource name and tags
  resource_name = coalesce(try(var.metadata.name, null), "aws-ecr-repo")
  tags          = merge({
    "Name" = local.resource_name
  }, try(var.metadata.labels, {}))

  # image tag mutability
  image_tag_mutability = try(var.spec.image_immutable, false) ? "IMMUTABLE" : "MUTABLE"

  # encryption settings
  encryption_type     = upper(try(var.spec.encryption_type, "AES256"))
  is_kms_encryption   = local.encryption_type == "KMS"
  kms_key_id          = local.is_kms_encryption ? try(var.spec.kms_key_id.value, null) : null
}


