resource "alicloud_ram_policy" "main" {
  policy_name     = var.spec.policy_name
  policy_document = var.spec.policy_document
  description     = var.spec.description != "" ? var.spec.description : null
  rotate_strategy = var.spec.rotate_strategy
  force           = var.spec.force
  tags            = local.final_tags
}
