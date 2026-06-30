resource "oci_streaming_stream" "this" {
  for_each = { for s in var.spec.streams : s.name => s }

  name             = each.value.name
  partitions       = each.value.partitions
  stream_pool_id   = oci_streaming_stream_pool.this.id
  retention_in_hours = each.value.retention_in_hours
  freeform_tags    = local.freeform_tags
}
