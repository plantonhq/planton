resource "alicloud_ram_role" "main" {
  role_name                    = var.spec.role_name
  assume_role_policy_document = var.spec.assume_role_policy_document
  description                  = var.spec.description != "" ? var.spec.description : null
  max_session_duration         = var.spec.max_session_duration
  force                        = var.spec.force
  tags                         = local.final_tags
}

resource "alicloud_ram_role_policy_attachment" "attachments" {
  for_each = local.policy_attachments_map

  role_name   = alicloud_ram_role.main.role_name
  policy_name = each.value.policy_name
  policy_type = each.value.policy_type
}
