resource "oci_objectstorage_replication_policy" "this" {
  for_each = { for p in var.spec.replication_policies : p.name => p }

  bucket                = oci_objectstorage_bucket.this.name
  namespace             = var.spec.namespace
  name                  = each.value.name
  destination_bucket_name = each.value.destination_bucket_name
  destination_region_name = each.value.destination_region_name
}
