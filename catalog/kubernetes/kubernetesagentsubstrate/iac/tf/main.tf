# KubernetesAgentSubstrate Terraform module.
#
# Installs the Agent Substrate control plane of the release spec.version names
# (local.release when empty), straight from what upstream publishes: the
# tag's own manifests (documents.tf), every image pinned by the digest
# upstream's registry serves (images.tf), composed the way upstream's
# installer composes them, with the typed settings applied onto upstream's
# own objects.
#
# ORDER: every refusal fires at plan time, before anything is written
# (preflight.tf: a file or image the release does not publish, a placeholder
# this module does not know, a release move upstream cannot make; and
# helm_crds.tf for a definition someone else owns). Then the namespaces; the kept definitions; the generated and
# derived Secrets (secrets.tf); and the four document phases, each awaited
# before the next: prerequisites (the pod-certificate controller every
# certificate comes from), database, system, runtime.
#
# DESTROY: in reverse. The default SandboxConfig goes first, while the
# controller that clears its finalizer still runs; the namespaces last. The
# definitions are kept (apply_only) unless crds.keep_on_uninstall is false,
# so worker pools and their actors are not deleted with the control plane;
# the bundled database's volume is kept by its StatefulSet, and the upgrade
# refusal reads its stamp.

locals {
  # The kept-definition contract (the generated helm_crds.tf): the bundle
  # branch, upstream's own definitions, one file per kind, at the tag
  # spec.version names, so a release that changes them installs its changes.
  # The upgrade rule (preflight.tf) judges every move between releases, a
  # rollback included, and upstream's rollback restores the older definitions,
  # so the never-downgrade refusal defers to it.
  helm_crds_args = {
    install                  = local.crds_install
    keep_on_uninstall        = local.crds_keep
    expect_crds              = true
    version                  = local.version
    version_field            = "spec.version"
    downgrade_judged_by_kind = true
    chart                    = null
    bundle = {
      name = local.bundle_name
      urls = [for path in local.crd_files : "${local.upstream_files_base}{{version}}/${path}"]
    }
  }
}

# The namespaces: ate-system when this resource owns it, and
# podcertificate-controller-system, which it always owns. Applied first and
# deleted last; the delete blocks until a namespace is fully gone, because
# the names are fixed and a destroy followed by an apply would otherwise race
# the terminating namespace.
resource "kubectl_manifest" "namespace" {
  for_each = local.namespace_docs

  yaml_body = yamlencode(each.value)

  server_side_apply = true
  force_conflicts   = true
  wait              = true

  depends_on = [terraform_data.preflight]
}

# The controller's RBAC, the spec-derived ConfigMaps, the SandboxConfig
# admission policy, and the pod-certificate controller.
resource "kubectl_manifest" "prerequisites" {
  for_each = local.prerequisite_docs

  yaml_body = yamlencode(each.value)

  server_side_apply = true
  force_conflicts   = true
  wait_for_rollout  = true

  depends_on = [
    terraform_data.preflight,
    kubectl_manifest.namespace,
    kubectl_manifest.helm_crds,
    kubernetes_secret_v1.ca_pool,
    kubectl_manifest.trust_bundle,
    kubernetes_secret_v1.actor_ca_certs,
    kubernetes_secret_v1.actor_jwt_pool,
    kubernetes_secret_v1.api_env,
    kubernetes_secret_v1.postgres_server_ca,
  ]
}

# The OpenTelemetry ConfigMap every workload reads, and the bundled database.
resource "kubectl_manifest" "database" {
  for_each = local.database_docs

  yaml_body = yamlencode(each.value)

  server_side_apply = true
  force_conflicts   = true
  wait_for_rollout  = true

  depends_on = [kubectl_manifest.prerequisites]
}

# The API server, controller, node agent, router, egress gateway,
# credential provider and network policies.
resource "kubectl_manifest" "system" {
  for_each = local.system_docs

  yaml_body = yamlencode(each.value)

  server_side_apply = true
  force_conflicts   = true
  wait_for_rollout  = true

  depends_on = [kubectl_manifest.database]
}

# The default gVisor SandboxConfig, last in and first out.
resource "kubectl_manifest" "runtime" {
  for_each = local.runtime_docs

  yaml_body = yamlencode(each.value)

  server_side_apply = true
  force_conflicts   = true

  depends_on = [kubectl_manifest.system]
}
