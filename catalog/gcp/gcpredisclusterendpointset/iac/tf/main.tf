# The user-created connections registered on a Memorystore for Redis
# Cluster. Google replaces the cluster's whole user-created endpoint list
# with this set on every apply, so the manifest's list IS the set.
#
# Every field of a connection is an output of the block that built it (the
# forwarding rule twice, on two output paths; the reserved address; the
# network; the cluster's service-attachment handle), so the registration
# carries no literals a consumer has to copy by hand. Only the project is
# immutable; the set updates in place.
resource "google_redis_cluster_user_created_connections" "this" {
  name    = local.cluster_name
  project = local.project_id
  region  = local.region

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
  # Sent only when set so the provider default stays in charge otherwise.
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  dynamic "cluster_endpoints" {
    for_each = var.spec.endpoints
    content {
      dynamic "connections" {
        for_each = cluster_endpoints.value.connections
        content {
          psc_connection {
            forwarding_rule    = connections.value.forwarding_rule
            psc_connection_id  = connections.value.psc_connection_id
            address            = connections.value.address
            network            = connections.value.network
            service_attachment = connections.value.service_attachment
            # Optional+Computed: Google records the rule's own project when
            # the spec leaves it unset, so it is sent only when set.
            project_id = connections.value.project_id != "" ? connections.value.project_id : null
          }
        }
      }
    }
  }
}
