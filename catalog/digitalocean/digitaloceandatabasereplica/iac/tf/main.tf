# DigitalOcean Database Replica
#
# Provisions a single-node read-only replica of a DigitalOcean managed
# database cluster -- the complete digitalocean_database_replica resource
# surface. PostgreSQL and MySQL primaries support replicas, in the
# primary's region or a different one (cross-region read replica).
#
# Semantics worth knowing before editing:
#   - region and size are REQUIRED by the spec even though the provider
#     marks them optional: the provider reads both back from the API but
#     never computes them, so omitted values drift on the next apply --
#     and region's drift schedules a full replica REPLACEMENT. The spec
#     makes that class unrepresentable by requiring explicit values.
#   - Only size and storage_size_mib change in place (a resize, waited to
#     "online"); every other argument change REPLACES the replica --
#     including tags, which are create-only upstream.
#   - DigitalOcean retries creation through 412 responses while the
#     primary's first backup completes; expect create to take about as
#     long as the primary's own provisioning.

resource "digitalocean_database_replica" "replica" {
  cluster_id = var.spec.cluster
  name       = var.spec.replica_name
  region     = local.region_slug
  size       = var.spec.size

  # Optional VPC placement for the replica's region (create-only).
  private_network_uuid = local.vpc_uuid

  # Custom disk size in MiB (string in the provider schema). Resizes in
  # place together with size; must stay >= the primary's storage.
  storage_size_mib = local.storage_size_mib

  # User tags plus the standard Planton labels (identical set in both
  # provisioners). CREATE-ONLY upstream: a tag edit replaces the replica.
  tags = local.tags

  # Fail loud on DigitalOcean's combined-tags budget before anything
  # renders (see local.tags_combined_budget; twin of the Pulumi module's
  # error).
  lifecycle {
    precondition {
      condition     = length(local.tags_combined) <= local.tags_combined_budget
      error_message = "DigitalOcean caps a database replica's combined tags (joined by commas) at ${local.tags_combined_budget} characters; this replica's ${length(local.tags)} tags join to ${length(local.tags_combined)} characters. Shorten metadata.name or metadata.id, or remove entries from spec.tags -- the Planton label tags alone use ${local.planton_tags_length} characters here."
    }
  }
}
