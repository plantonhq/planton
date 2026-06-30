locals {
  bucket_name = var.metadata.name
  is_public   = try(var.spec.is_public, false)
  
  # Merge Planton labels with user-provided tags
  merged_tags = merge(
    {
      "planton.org/resource"      = "true"
      "planton.org/organization"  = var.metadata.org
      "planton.org/environment"   = var.metadata.env
      "planton.org/resource-kind" = "AwsS3Bucket"
      "planton.org/resource-id"   = var.metadata.id
    },
    try(var.spec.tags, {})
  )

  # Map storage class enum values to AWS storage class strings
  storage_class_map = {
    "STANDARD"                        = "STANDARD"
    "STANDARD_IA"                     = "STANDARD_IA"
    "ONE_ZONE_IA"                     = "ONEZONE_IA"
    "INTELLIGENT_TIERING"             = "INTELLIGENT_TIERING"
    "GLACIER_INSTANT_RETRIEVAL"       = "GLACIER_IR"
    "GLACIER_FLEXIBLE_RETRIEVAL"      = "GLACIER"
    "GLACIER_DEEP_ARCHIVE"            = "DEEP_ARCHIVE"
  }
}
