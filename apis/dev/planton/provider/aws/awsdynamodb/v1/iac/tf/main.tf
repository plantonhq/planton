resource "aws_dynamodb_table" "this" {
  name         = local.resource_name
  billing_mode = local.is_provisioned_billing ? "PROVISIONED" : "PAY_PER_REQUEST"

  # When PROVISIONED, the aws provider expects top-level read/write capacity
  read_capacity  = local.is_provisioned_billing ? try(var.spec.provisioned_throughput.read_capacity_units, 0) : null
  write_capacity = local.is_provisioned_billing ? try(var.spec.provisioned_throughput.write_capacity_units, 0) : null

  # attribute definitions
  dynamic "attribute" {
    for_each = try(var.spec.attribute_definitions, [])
    content {
      name = attribute.value.name
      type = local.attribute_type_map[attribute.value.type]
    }
  }

  # key schema
  hash_key  = local.table_hash_key
  range_key = local.table_range_key

  # local secondary indexes
  dynamic "local_secondary_index" {
    for_each = coalesce(try(var.spec.local_secondary_indexes, null), [])
    content {
      name               = local_secondary_index.value.name
      projection_type    = try(local.projection_type_map[local_secondary_index.value.projection.type], upper(local_secondary_index.value.projection.type))
      non_key_attributes = try(local_secondary_index.value.projection.non_key_attributes, null)
      range_key          = [for k in local_secondary_index.value.key_schema : k.attribute_name if k.key_type == "KEY_TYPE_RANGE"][0]
    }
  }

  # global secondary indexes
  dynamic "global_secondary_index" {
    for_each = coalesce(try(var.spec.global_secondary_indexes, null), [])
    content {
      name               = global_secondary_index.value.name
      hash_key           = [for k in global_secondary_index.value.key_schema : k.attribute_name if k.key_type == "KEY_TYPE_HASH"][0]
      range_key          = try([for k in global_secondary_index.value.key_schema : k.attribute_name if k.key_type == "KEY_TYPE_RANGE"][0], null)
      projection_type    = try(local.projection_type_map[global_secondary_index.value.projection.type], upper(global_secondary_index.value.projection.type))
      non_key_attributes = try(global_secondary_index.value.projection.non_key_attributes, null)

      read_capacity  = local.is_provisioned_billing ? try(global_secondary_index.value.provisioned_throughput.read_capacity_units, 0) : null
      write_capacity = local.is_provisioned_billing ? try(global_secondary_index.value.provisioned_throughput.write_capacity_units, 0) : null
    }
  }

  # TTL
  dynamic "ttl" {
    for_each = local.ttl_enabled ? [1] : []
    content {
      enabled        = true
      attribute_name = local.ttl_attribute_name
    }
  }

  # Streams
  stream_enabled   = local.streams_enabled
  stream_view_type = local.streams_enabled ? local.stream_view_type : null

  # Point-in-time recovery (only when enabled)
  dynamic "point_in_time_recovery" {
    for_each = coalesce(try(var.spec.point_in_time_recovery_enabled, null), false) ? [1] : []
    content {
      enabled = true
    }
  }

  # Server-side encryption
  server_side_encryption {
    enabled     = local.sse_enabled
    kms_key_arn = local.sse_kms_arn
  }

  # Table class
  table_class = local.table_class

  # Deletion protection
  deletion_protection_enabled = local.deletion_protection_enabled

  # Contributor Insights (separate resource in AWS provider v5)

  tags = local.tags
}

resource "aws_dynamodb_contributor_insights" "this" {
  count      = local.contributor_insights_enabled ? 1 : 0
  table_name = aws_dynamodb_table.this.name
}


