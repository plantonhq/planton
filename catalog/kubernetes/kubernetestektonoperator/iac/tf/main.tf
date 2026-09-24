# KubernetesTektonOperator Terraform module.
#
# Installs the Tekton Operator from its released single-file manifest —
# the operator's OFFICIAL distribution (the in-repo Helm chart is
# unpublished, version "devel"). The manifest applies per document: the
# namespace `tekton-operator`, the 14 operator.tekton.dev CRDs, the
# operator and webhook Deployments (with the spec's typed overrides
# patched on — see locals.tf), ConfigMaps, RBAC, Services and the webhook
# cert Secret.
#
# AUTO-INSTALL IS ALWAYS DISABLED: the tekton-config-defaults ConfigMap's
# AUTOINSTALL_COMPONENTS key is patched from the release's "true" to
# "false" so the operator never creates its own TektonConfig — the
# KubernetesTekton declaration kind is the single owner of the cluster's
# Tekton configuration. Installing the operator alone deploys no Tekton
# components.
#
# DESTROY SEMANTICS: every document deletes with the resource, INCLUDING
# the CRDs — which cascade-deletes any TektonConfig on the cluster.
# Always destroy the KubernetesTekton resource FIRST while the operator
# still runs (TektonInstallerSet finalizers are operator-processed; see
# the spec's destroy note).

# Fetch the released manifest from the pinned release tag.
data "http" "tekton_operator_manifest" {
  url = local.manifest_url

  request_headers = {
    Accept = "application/yaml"
  }
}

# Apply every document. Keyed by each document's COMPOSED IDENTITY
# (apiVersion//kind//name[//namespace]) — stable across manifest
# reorderings, and exactly the ID form the kubectl importer takes so the
# import map derives IDs blind from the address keys.
#
# server_side_apply keeps re-installs tolerant of the operator's own
# field management on its ConfigMaps and matches the Pulumi twin's apply
# mode.
# The fixed namespace applies FIRST and deletes LAST — and its DELETE
# blocks until the namespace is fully gone (the provider's wait
# attribute). Namespace deletion is asynchronous and this kind's
# namespace name is FIXED, so without the ordering a destroy-then-apply
# sequence races the terminating namespace: every namespaced document
# fails with "unable to create new content in namespace ... because it
# is being terminated" (verified live).
resource "kubectl_manifest" "namespace" {
  for_each = local.namespace_documents

  yaml_body = yamlencode(each.value)

  server_side_apply = true
  force_conflicts   = true
  wait              = true
}

# Everything that is not a Namespace or CRD: the two Deployments (with
# the spec's typed overrides patched on), RBAC, ConfigMaps, Services,
# the webhook Secret. Rollout waiting is deliberately OFF — the group
# applies BEFORE the CRDs (see the ordering rationale in locals.tf),
# and the operator only becomes ready once its CRDs exist; blocking
# here would deadlock the create ordering. The E2E verifier (and any
# health check) owns rollout readiness.
resource "kubectl_manifest" "tekton_operator" {
  for_each = local.workload_documents

  yaml_body = yamlencode(each.value)

  server_side_apply = true
  force_conflicts   = true
  wait_for_rollout  = false

  lifecycle {
    precondition {
      condition     = local.image_registry == "" || local.image_table_matches_release
      error_message = local.image_table_mismatch_message
    }
  }

  depends_on = [kubectl_manifest.namespace]
}

# The 14 operator.tekton.dev CRDs apply LAST and — critically — delete
# FIRST, while the operator Deployments still run: CRD deletion drains
# every CR, and the operator's runtime InstallerSets carry a finalizer
# only the LIVE operator can process (verified live: deleting CRDs and
# operator in one pass wedges tektoninstallersets in Terminating). The
# delete blocks until each CRD is fully gone so the next group's
# teardown never overtakes the drain.
resource "kubectl_manifest" "crds" {
  for_each = local.crd_documents

  yaml_body = yamlencode(each.value)

  server_side_apply = true
  force_conflicts   = true
  wait              = true

  depends_on = [kubectl_manifest.tekton_operator]
}
