resource "aws_db_subnet_group" "this" {
  count      = local.need_subnet_group ? 1 : 0
  name       = local.resource_id
  subnet_ids = local.safe_subnet_ids
  tags       = local.final_labels
}


