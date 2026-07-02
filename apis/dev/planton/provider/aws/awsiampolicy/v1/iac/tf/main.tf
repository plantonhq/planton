# A customer-managed IAM policy is the reusable unit of AWS permissions: one
# document, one ARN, attachable to many roles/users at once. Name, path, and
# description are create-only in AWS (the provider replaces the policy when
# they change); only the document itself is updatable in place.
#
# Document updates create a new policy version and mark it default. AWS keeps
# at most 5 versions per policy; the provider prunes the oldest non-default
# version before saving a new one, so repeated updates keep working without
# manual version cleanup.
resource "aws_iam_policy" "this" {
  name        = var.metadata.name
  path        = var.spec.path != "" ? var.spec.path : null
  description = var.spec.description != "" ? var.spec.description : null
  policy      = local.policy_document_json

  tags = local.aws_tags
}
