locals {
  # Convert access_control enum to ACL string
  # 0 = PRIVATE, 1 = PUBLIC_READ
  acl = var.spec.access_control == 1 ? "public-read" : "private"

  # Combine metadata tags and spec tags
  all_tags = distinct(concat(
    var.metadata.tags != null ? var.metadata.tags : [],
    var.spec.tags
  ))

  # Resource labels
  resource_labels = merge(
    var.metadata.labels != null ? var.metadata.labels : {},
    {
      "managed-by" = "planton"
      "component"  = "digitalocean-bucket"
    }
  )
}

