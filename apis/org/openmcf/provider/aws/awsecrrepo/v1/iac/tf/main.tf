resource "aws_ecr_repository" "this" {
  name                 = var.spec.repository_name
  image_tag_mutability = local.image_tag_mutability
  force_delete         = try(var.spec.force_delete, false)

  encryption_configuration {
    encryption_type = local.encryption_type
    kms_key         = local.kms_key_id
  }

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = local.tags
}

# Lifecycle policy for cost control. Created only when spec.lifecycle_policy is set and yields at
# least one rule; the rules themselves (a single untagged-by-age rule and/or an any-keep-last-N
# rule) are assembled in locals so the AWS "only one untagged rule" constraint always holds.
resource "aws_ecr_lifecycle_policy" "this" {
  count      = local.lifecycle_policy != null && length(local.lifecycle_rules) > 0 ? 1 : 0
  repository = aws_ecr_repository.this.name

  policy = jsonencode({
    rules = local.lifecycle_rules
  })
}


