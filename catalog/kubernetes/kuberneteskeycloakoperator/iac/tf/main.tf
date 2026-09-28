# KubernetesKeycloakOperator Terraform module.
#
# Installs the official Keycloak Operator from the keycloak-k8s-resources
# release manifests — the operator's first-party Kubernetes distribution
# (Keycloak ships NO official Helm chart). The bundle applies per
# document: ServiceAccount, ClusterRoles, bindings, Role, the
# metrics/health Service and the operator Deployment (with the spec's
# typed overrides patched on — see locals.tf), plus the four
# k8s.keycloak.org CRDs published beside it. spec.cluster_wide selects
# the watch-scope variant.
#
# NAMESPACE STAMPING: the bundle ships every document WITHOUT a
# namespace field (upstream expects kustomize to set it). The module
# owns that kustomize step: metadata.namespace = <spec namespace> on
# every namespaced document, and every binding's ServiceAccount subject
# namespace rewritten to match (locals.stamped_documents). Every
# resource is FIXED-NAME (`keycloak-operator` etc. — upstream's own
# names), so exactly ONE operator install fits per namespace.
#
# DESTROY SEMANTICS: every document deletes with the resource, INCLUDING
# the CRDs — which cascade-deletes every Keycloak / KeycloakRealmImport /
# KeycloakOidcClient / KeycloakSamlClient CR on the cluster. Always
# destroy KubernetesKeycloak resources FIRST while the operator still
# runs.

# Fetch the operator bundle (the watch-scope variant chosen by
# spec.cluster_wide) from the pinned release tag.
data "http" "keycloak_operator_bundle" {
  url = local.bundle_url

  request_headers = {
    Accept = "application/yaml"
  }
}

# Fetch the four k8s.keycloak.org CRD files (shared by both variants)
# from the same pinned tag.
data "http" "keycloak_operator_crds" {
  for_each = toset(local.crd_files)

  url = "${local.bundle_base_url}/${each.value}"

  request_headers = {
    Accept = "application/yaml"
  }
}

# Apply every document. Keyed by each document's COMPOSED IDENTITY
# (apiVersion//kind//name[//namespace]) — stable across manifest
# reorderings, and exactly the ID form the kubectl importer takes so the
# import map derives IDs blind from the address keys (names repeat
# across kinds in this bundle: ServiceAccount, Service and Deployment
# are all named `keycloak-operator`).
#
# server_side_apply is REQUIRED, not just preferred: the keycloaks CRD
# document (~9,900 lines) blows past the client-side
# last-applied-configuration annotation cap, and SSA keeps re-installs
# tolerant of the operator's own field management. Matches the Pulumi
# twin's apply mode.
#
# The module-authored namespace applies FIRST and deletes LAST — and its
# DELETE blocks until the namespace is fully gone (the provider's wait
# attribute). Namespace deletion is asynchronous and this resource's
# namespace name is FIXED per install, so without the ordering a
# destroy-then-apply sequence races the terminating namespace: every
# namespaced document fails with "unable to create new content in
# namespace ... because it is being terminated". Empty when
# create_namespace is false (the namespace must then already exist).
resource "kubectl_manifest" "namespace" {
  for_each = local.namespace_documents

  yaml_body = yamlencode(each.value)

  server_side_apply = true
  force_conflicts   = true
  wait              = true
}

# The bundle's 16 documents: ServiceAccount, RBAC, the metrics/health
# Service and the operator Deployment (with the spec's typed overrides
# patched on). Rollout waiting is deliberately OFF — the group applies
# BEFORE the CRDs (see the ordering rationale in locals.tf), and the
# JOSDK operator crash-loops until its CRDs exist; blocking here would
# deadlock the create ordering. No readiness gate is needed after the
# CRDs either: the operator registers no admission webhook, so a
# Keycloak resource applied before it runs is stored as written and
# reconciled once it starts.
resource "kubectl_manifest" "keycloak_operator" {
  for_each = local.workload_documents

  yaml_body = yamlencode(each.value)

  server_side_apply = true
  force_conflicts   = true
  wait_for_rollout  = false

  depends_on = [kubectl_manifest.namespace]
}

# The four k8s.keycloak.org CRDs apply LAST and — critically — delete
# FIRST, while the operator Deployment still runs: CRD deletion
# cascade-deletes every Keycloak CR, and any operator-processed
# finalizers on those CRs need the LIVE operator to drain. The delete
# blocks until each CRD is fully gone so the next group's teardown never
# overtakes the drain.
resource "kubectl_manifest" "crds" {
  for_each = local.crd_documents

  yaml_body = yamlencode(each.value)

  server_side_apply = true
  force_conflicts   = true
  wait              = true

  depends_on = [kubectl_manifest.keycloak_operator]
}
