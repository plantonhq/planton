# Neptune Subnet Group
# Created when subnet_ids are provided and neptune_subnet_group_name is not set

resource "aws_neptune_subnet_group" "main" {
  count = local.need_subnet_group ? 1 : 0

  name        = local.resource_id
  subnet_ids  = local.safe_subnet_ids
  description = "Neptune subnet group for ${local.resource_id}"

  tags = local.final_tags
}
