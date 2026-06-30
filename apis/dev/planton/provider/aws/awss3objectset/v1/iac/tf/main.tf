resource "aws_s3_object" "objects" {
  for_each = local.objects_map

  bucket = local.bucket_name
  key    = each.value.key

  # Content source: exactly one of content or content_base64
  content        = each.value.content
  content_base64 = each.value.content_base64

  # Optional metadata
  content_type     = each.value.content_type
  cache_control    = each.value.cache_control
  content_encoding = each.value.content_encoding
  acl              = each.value.acl

  # Merge set-level tags with object-level tags (object tags take precedence)
  tags = merge(local.set_tags, try(each.value.tags, {}))
}
