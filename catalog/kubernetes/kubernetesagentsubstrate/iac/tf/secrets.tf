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

# The ClusterTrustBundle each pod-certificate pool's root is published under
# (secrets.go): applied before the controller and every workload, so a
# reinstall never starts them against the previous install's roots (a bundle
# is cluster-scoped and outlives the namespaces a destroy removes), and
# removed by a destroy, as upstream's installer removes them. The object is
# the controller's own (its name, signer, its one label and the root in its
# PEM form), so the controller finds it already in the state it wants. The
# API is v1 from Kubernetes 1.37 and v1beta1 before it (GKE 1.36).
data "kubernetes_server_version" "cluster" {}

locals {
  trust_bundle_api_version = (
    tonumber(trimsuffix(data.kubernetes_server_version.cluster.major, "+")) > 1 ||
    tonumber(trimsuffix(data.kubernetes_server_version.cluster.minor, "+")) >= 37
  ) ? "certificates.k8s.io/v1" : "certificates.k8s.io/v1beta1"
}

resource "kubectl_manifest" "trust_bundle" {
  for_each = { for name, pool in local.ca_pools : name => pool if pool.trust_bundle != "" }

  yaml_body = yamlencode({
    apiVersion = local.trust_bundle_api_version
    kind       = "ClusterTrustBundle"
    metadata = {
      name   = each.value.trust_bundle
      labels = { "podcert.ate.dev/canarying" = "live" }
    }
    spec = {
      signerName  = each.value.signer
      trustBundle = tls_self_signed_cert.ca[each.key].cert_pem
    }
  })

  server_side_apply = true
  force_conflicts   = true

  depends_on = [terraform_data.preflight]
}
