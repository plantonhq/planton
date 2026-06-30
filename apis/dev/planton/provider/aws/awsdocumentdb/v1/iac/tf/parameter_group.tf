# DocumentDB Cluster Parameter Group
# Created when cluster_parameters are provided

resource "aws_docdb_cluster_parameter_group" "main" {
  count = local.need_cluster_parameter_group ? 1 : 0

  name        = local.resource_id
  family      = local.engine_family
  description = "DocumentDB cluster parameter group for ${local.resource_id}"

  dynamic "parameter" {
    for_each = local.parameters
    content {
      name         = parameter.value.name
      value        = parameter.value.value
      apply_method = coalesce(try(parameter.value.apply_method, ""), "immediate")
    }
  }

  tags = local.final_tags
}
