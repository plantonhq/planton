locals {
  cluster_parameters = [for p in local.parameters : {
    apply_method = try(p.apply_method, null)
    name         = try(p.name, null)
    value        = try(p.value, null)
  }]
}

resource "aws_rds_cluster_parameter_group" "this" {
  count  = local.need_cluster_parameter_group ? 1 : 0
  name   = "${local.resource_id}-cluster"
  family = local.engine_family
  tags   = local.final_labels

  dynamic "parameter" {
    for_each = local.cluster_parameters
    content {
      apply_method = parameter.value.apply_method
      name         = parameter.value.name
      value        = parameter.value.value
    }
  }
}


