locals {
  bucket_name = var.spec.bucket

  # Standard Planton labels
  base_tags = {
    "planton.org/resource"      = "true"
    "planton.org/organization"  = var.metadata.org
    "planton.org/environment"   = var.metadata.env
    "planton.org/resource-kind" = "AwsS3ObjectSet"
    "planton.org/resource-id"   = var.metadata.id
  }

  # Merge base tags with set-level tags
  set_tags = merge(local.base_tags, try(var.spec.tags, {}))

  # Create a map of objects keyed by their S3 key for stable for_each iteration
  objects_map = {
    for obj in var.spec.objects : obj.key => obj
  }
}
