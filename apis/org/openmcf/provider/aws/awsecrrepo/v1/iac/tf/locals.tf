locals {
  # resource name and tags
  resource_name = coalesce(try(var.metadata.name, null), "aws-ecr-repo")
  tags = merge({
    "Name" = local.resource_name
  }, try(var.metadata.labels, {}))

  # image tag mutability
  image_tag_mutability = try(var.spec.image_immutable, false) ? "IMMUTABLE" : "MUTABLE"

  # encryption settings
  encryption_type   = upper(try(var.spec.encryption_type, "AES256"))
  is_kms_encryption = local.encryption_type == "KMS"
  kms_key_id        = local.is_kms_encryption ? try(var.spec.kms_key_id.value, null) : null

  # Lifecycle policy (cost control), driven entirely by spec.lifecycle_policy. When the block is
  # absent, no policy is created. AWS rejects a lifecycle policy that has more than one rule
  # selecting untagged images, so we emit at most ONE untagged rule (expire by age) plus an
  # "any" keep-last-N rule. The "any" rule must carry the highest rulePriority, which it does.
  # Each rule is included only when its knob is a positive value, so a partially-specified block
  # (or a pruned zero) never produces an invalid rule.
  lifecycle_policy     = try(var.spec.lifecycle_policy, null)
  expire_untagged_days = try(local.lifecycle_policy.expire_untagged_after_days, 0)
  max_image_count      = try(local.lifecycle_policy.max_image_count, 0)
  lifecycle_rules = concat(
    local.expire_untagged_days > 0 ? [{
      rulePriority = 1
      description  = "Expire untagged images older than ${local.expire_untagged_days} day(s)"
      selection = {
        tagStatus   = "untagged"
        countType   = "sinceImagePushed"
        countUnit   = "days"
        countNumber = local.expire_untagged_days
      }
      action = {
        type = "expire"
      }
    }] : [],
    local.max_image_count > 0 ? [{
      rulePriority = 2
      description  = "Keep only the most recent ${local.max_image_count} images"
      selection = {
        tagStatus   = "any"
        countType   = "imageCountMoreThan"
        countNumber = local.max_image_count
      }
      action = {
        type = "expire"
      }
    }] : [],
  )
}


