resource "aws_kms_key" "this" {
  description             = local.description
  deletion_window_in_days = local.deletion_window_days
  enable_key_rotation     = local.rotation_enabled
  customer_master_key_spec = local.customer_master_key_spec

  tags = local.tags
}

resource "aws_kms_alias" "this" {
  count         = local.alias_name != null && local.alias_name != "" ? 1 : 0
  name          = local.alias_name
  target_key_id = aws_kms_key.this.key_id
}



