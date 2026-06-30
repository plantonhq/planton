locals {
  # name and tags
  resource_name = coalesce(try(var.metadata.name, null), "awsdynamodb-table")
  tags          = merge({
    "Name" = local.resource_name
  }, try(var.metadata.labels, {}))

  # billing
  is_provisioned_billing = try(var.spec.billing_mode, "BILLING_MODE_PAY_PER_REQUEST") == "BILLING_MODE_PROVISIONED"

  # enum mappings from proto → Terraform acceptable strings
  attribute_type_map = {
    "ATTRIBUTE_TYPE_S" = "S"
    "ATTRIBUTE_TYPE_N" = "N"
    "ATTRIBUTE_TYPE_B" = "B"
  }

  stream_view_type_map = {
    "STREAM_VIEW_TYPE_KEYS_ONLY"         = "KEYS_ONLY"
    "STREAM_VIEW_TYPE_NEW_IMAGE"         = "NEW_IMAGE"
    "STREAM_VIEW_TYPE_OLD_IMAGE"         = "OLD_IMAGE"
    "STREAM_VIEW_TYPE_NEW_AND_OLD_IMAGES" = "NEW_AND_OLD_IMAGES"
  }

  projection_type_map = {
    "PROJECTION_TYPE_ALL"         = "ALL"
    "PROJECTION_TYPE_KEYS_ONLY"   = "KEYS_ONLY"
    "PROJECTION_TYPE_INCLUDE"     = "INCLUDE"
  }

  table_class_map = {
    "TABLE_CLASS_STANDARD"                 = "STANDARD"
    "TABLE_CLASS_STANDARD_INFREQUENT_ACCESS" = "STANDARD_INFREQUENT_ACCESS"
  }

  # streams
  streams_enabled   = coalesce(try(var.spec.stream_enabled, null), false)
  stream_view_type  = try(local.stream_view_type_map[var.spec.stream_view_type], null)

  # ttl
  ttl_enabled        = coalesce(try(var.spec.ttl.enabled, null), false)
  ttl_attribute_name = try(var.spec.ttl.attribute_name, null)

  # sse
  sse_enabled  = coalesce(try(var.spec.server_side_encryption.enabled, null), false)
  sse_kms_arn  = try(var.spec.server_side_encryption.kms_key_arn, null)

  # table class
  table_class = try(local.table_class_map[var.spec.table_class], null)

  # contributor insights
  contributor_insights_enabled = coalesce(try(var.spec.contributor_insights_enabled, null), false)

  # deletion protection
  deletion_protection_enabled = coalesce(try(var.spec.deletion_protection_enabled, null), false)

  # derive hash and range keys from key_schema
  table_hash_key  = [for k in try(var.spec.key_schema, []) : k.attribute_name if k.key_type == "KEY_TYPE_HASH"][0]
  table_range_key = try([for k in try(var.spec.key_schema, []) : k.attribute_name if k.key_type == "KEY_TYPE_RANGE"][0], null)

  # autoscaling toggle and labels used by autoscaling resources (if present)
  auto_scaling_is_enabled = coalesce(try(var.spec.auto_scale.is_enabled, null), false)
  final_labels            = local.tags
}


