# DigitalOcean Database User
#
# Provisions an additional user on a DigitalOcean managed database cluster,
# modeling the complete digitalocean_database_user resource surface: the
# MySQL authentication plugin choice and the Kafka / OpenSearch access
# control lists. DigitalOcean generates the password (and, on Kafka, the
# mTLS certificate pair) server-side; they surface only as outputs.
#
# The DigitalOcean API serializes user creation/deletion per cluster, so
# composing many users on one cluster deploys sequentially by design.

resource "digitalocean_database_user" "user" {
  cluster_id = var.spec.cluster
  name       = var.spec.user_name

  # MySQL clusters only (API-enforced). Unset defers to DigitalOcean's
  # caching_sha2_password default; updates apply through a
  # password-preserving auth reset.
  mysql_auth_plugin = local.mysql_auth_plugin

  # Engine-specific ACLs. The provider records `settings` only from the
  # CREATE response and never refreshes it from the API, so this
  # configuration is the source of truth and imports can never recover it
  # (recorded as a config-only import tolerance). Each ACL row also carries
  # a computed server-side id, which is provisioning noise and not modeled.
  #
  # The block is sent exactly when the spec sets `settings` -- an EMPTY
  # message counts -- because what DigitalOcean stores is engine-specific
  # (measured 2026-09-17): every PostgreSQL user comes back with a settings
  # object (`pg_allow_replication: false`), which the provider keeps as one
  # empty block, so a PostgreSQL manifest declares `settings: {}` to match
  # it; a MySQL user never carries one and the API REFUSES a settings update
  # (`422 operation is not supported for this cluster type`), so a MySQL
  # manifest leaves it out. The module cannot know the engine from a cluster
  # UUID, so the manifest carries that knowledge. Same shape as the Pulumi
  # module.
  dynamic "settings" {
    for_each = var.spec.settings != null ? [var.spec.settings] : []
    content {
      dynamic "acl" {
        for_each = coalesce(settings.value.kafka_acls, [])
        content {
          topic      = acl.value.topic
          permission = acl.value.permission
        }
      }
      dynamic "opensearch_acl" {
        for_each = coalesce(settings.value.opensearch_acls, [])
        content {
          index      = opensearch_acl.value.index
          permission = opensearch_acl.value.permission
        }
      }
    }
  }
}
