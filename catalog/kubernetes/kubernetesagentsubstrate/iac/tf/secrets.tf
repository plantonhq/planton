# Every Secret the install reads (secrets.go), as typed Secret resources
# rather than manifest documents, so key material never sits in the
# documents the engines compare: the CA pools and the actor token signing
# key (pools.tf), the actor CA certificate, the API server's environment
# Secret, and an external database's server CA. They follow the namespaces
# and precede everything that mounts them.

resource "kubernetes_secret_v1" "ca_pool" {
  for_each = local.ca_pools

  metadata {
    name      = each.key
    namespace = each.value.namespace
    labels    = local.identity_labels
  }
  type = "kubernetes.io/tls"
  data = local.ca_pool_data[each.key]

  depends_on = [kubectl_manifest.namespace]
}

resource "kubernetes_secret_v1" "actor_ca_certs" {
  metadata {
    name      = "actor-id-ca-certs"
    namespace = local.namespace
    labels    = local.identity_labels
  }
  type = "Opaque"
  data = { "ca.crt" = tls_self_signed_cert.ca["actor-id-ca-pool"].cert_pem }

  depends_on = [kubectl_manifest.namespace]
}

resource "kubernetes_secret_v1" "actor_jwt_pool" {
  metadata {
    name      = "actor-id-jwt-pool"
    namespace = local.namespace
    labels    = local.identity_labels
  }
  type = "Opaque"
  data = {
    pool = local.actor_jwt_pool
    jwks = local.actor_jwt_jwks
  }

  depends_on = [kubectl_manifest.namespace]
}

resource "kubernetes_secret_v1" "api_env" {
  metadata {
    name      = "ate-api-server-secret-envvars"
    namespace = local.namespace
    labels    = local.identity_labels
  }
  type = "Opaque"
  data = local.api_env_secret

  depends_on = [kubectl_manifest.namespace]
}

resource "kubernetes_secret_v1" "postgres_server_ca" {
  count = local.db.server_ca != "" ? 1 : 0

  metadata {
    name      = "postgres-server-ca"
    namespace = local.namespace
    labels    = local.identity_labels
  }
  type = "Opaque"
  data = { "server-ca.pem" = local.db.server_ca }

  depends_on = [kubectl_manifest.namespace]
}
