# The refusals (module/preflight.go), as preconditions on a terraform_data
# resource that everything else depends on. Every value they read is known at
# plan time, so a failing condition stops the plan before anything is
# written, and conditions are not evaluated on destroy, so no refusal stands
# between a user and deleting their stack. The words are the Go module's,
# three parts each.
#
# NO depends_on ON THE DATA SOURCES: a data source that depends on a resource
# is read at apply time, and the refusals must fire at plan time. Stigmers are
# read only when their definition exists (count): a read of a kind the
# cluster does not serve fails the plan instead of returning nothing. A
# namespaced read lists one namespace (the data source has no cluster-wide
# list), so Stigmers are read in the Stigmer's own namespace, as the Go
# module reads them.

# The Stigmer's namespace and the operator's, read by name.
data "kubernetes_resources" "stigmer_namespace" {
  api_version    = "v1"
  kind           = "Namespace"
  field_selector = "metadata.name=${local.namespace}"
}

data "kubernetes_resources" "operator_namespace" {
  api_version    = "v1"
  kind           = "Namespace"
  field_selector = "metadata.name=${local.operator_namespace}"
}

# The Stigmer definition: whose Helm release owns it, and whether Stigmers
# can be read at all.
data "kubernetes_resources" "stigmer_definition" {
  api_version    = "apiextensions.k8s.io/v1"
  kind           = "CustomResourceDefinition"
  field_selector = "metadata.name=${local.stigmer_definition}"
}

# The Stigmers in the Stigmer's namespace, when the definition exists.
data "kubernetes_resources" "stigmers" {
  count = length(data.kubernetes_resources.stigmer_definition.objects) > 0 ? 1 : 0

  api_version = local.stigmer_api_version
  kind        = local.stigmer_kind
  namespace   = local.namespace
}

locals {
  stigmer_namespace_found  = length(data.kubernetes_resources.stigmer_namespace.objects) > 0
  operator_namespace_found = length(data.kubernetes_resources.operator_namespace.objects) > 0
  operator_namespace_ours = local.operator_namespace_found && try(
    data.kubernetes_resources.operator_namespace.objects[0].metadata.labels[local.owner_label], ""
  ) == local.kind_name
  separate_namespaces = local.operator_namespace != local.namespace

  definition                 = try(data.kubernetes_resources.stigmer_definition.objects[0], null)
  definition_release         = try(local.definition.metadata.annotations["meta.helm.sh/release-name"], "")
  definition_release_ns      = try(local.definition.metadata.annotations["meta.helm.sh/release-namespace"], "")
  definition_foreign         = local.definition != null && !(local.definition_release == local.operator_release && local.definition_release_ns == local.operator_namespace)
  definition_owner           = local.definition_release == "" ? "no Helm release (it was applied another way)" : "the Helm release ${local.definition_release} in namespace ${local.definition_release_ns}"
  definition_next_step_moved = "the operator cannot move to another namespace in place: destroy this resource, delete the kept definition once no Stigmer is left (kubectl delete crd ${local.stigmer_definition}), then deploy again with spec.operator.namespace ${local.operator_namespace}; or set spec.operator.namespace back to ${local.definition_release_ns}"
  definition_next_step       = local.definition_release == local.operator_release ? local.definition_next_step_moved : "remove the other Stigmer operator install first, or keep it and do not install KubernetesStigmer on this cluster"

  stigmers = try(data.kubernetes_resources.stigmers[0].objects, [])
  foreign_stigmers = [
    for s in local.stigmers : s
    if !(s.metadata.namespace == local.namespace && s.metadata.name == local.stigmer_name && try(s.metadata.labels[local.owner_label], "") == local.kind_name)
  ]
  our_stigmer     = try([for s in local.stigmers : s if s.metadata.namespace == local.namespace && s.metadata.name == local.stigmer_name][0], null)
  running_version = trimprefix(try(local.our_stigmer.status.version, ""), "v")
  running_parts   = can(regex("^[0-9]+\\.[0-9]+\\.[0-9]+$", local.running_version)) ? [for p in split(".", local.running_version) : tonumber(p)] : []
  requested_parts = [for p in split(".", var.spec.version) : tonumber(p)]
  downgrade = length(local.running_parts) == 3 && (
    local.requested_parts[0] < local.running_parts[0] ||
    (local.requested_parts[0] == local.running_parts[0] && local.requested_parts[1] < local.running_parts[1]) ||
    (local.requested_parts[0] == local.running_parts[0] && local.requested_parts[1] == local.running_parts[1] && local.requested_parts[2] < local.running_parts[2])
  )
}

resource "terraform_data" "preflight" {
  lifecycle {
    precondition {
      condition     = !(local.create_namespace && !local.separate_namespaces)
      error_message = "observed: spec.operator.namespace and spec.namespace both resolve to ${local.namespace}, and spec.operator.create_namespace is true\nmeaning: this resource would create the namespace and delete it on destroy, and the Stigmer's encryption keys live there: deleting it would make every backup unreadable\nnext step: install the operator in a namespace of its own (stigmer-system by convention), or turn spec.operator.create_namespace off"
    }
    precondition {
      condition     = local.stigmer_namespace_found
      error_message = "observed: namespace ${local.namespace} (spec.namespace) does not exist on the cluster\nmeaning: this resource never creates the Stigmer's namespace, because deleting it would delete the encryption keys and make every backup unreadable\nnext step: deploy a KubernetesNamespace named ${local.namespace} first and reference it from spec.namespace, or create it yourself"
    }
    precondition {
      condition     = !local.separate_namespaces || local.create_namespace || local.operator_namespace_found
      error_message = "observed: namespace ${local.operator_namespace} (spec.operator.namespace) does not exist on the cluster, and spec.operator.create_namespace is false\nmeaning: the operator installs into that namespace, so its owner must create it first\nnext step: deploy a KubernetesNamespace named ${local.operator_namespace} first and reference it from spec.operator.namespace, or set spec.operator.create_namespace to true"
    }
    precondition {
      condition     = !local.separate_namespaces || !local.create_namespace || !local.operator_namespace_found || local.operator_namespace_ours
      error_message = "observed: namespace ${local.operator_namespace} already exists and was not created by this KubernetesStigmer, and spec.operator.create_namespace is true\nmeaning: this resource would take the namespace over, and destroying it later would delete the namespace and everything other components put in it\nnext step: set spec.operator.create_namespace to false to install into the existing namespace; or, if you created it for this install, mark it as this kind's (`kubectl label namespace ${local.operator_namespace} ${local.owner_label}=${local.kind_name}`) and deploy again"
    }
    precondition {
      condition     = !local.separate_namespaces || local.create_namespace || !local.operator_namespace_ours
      error_message = "observed: namespace ${local.operator_namespace} was created by this KubernetesStigmer, and spec.operator.create_namespace is now false\nmeaning: turning it off would delete the namespace, and the operator with it\nnext step: keep spec.operator.create_namespace true; to hand the namespace to a KubernetesNamespace, destroy this resource first, then deploy it again with create_namespace false"
    }
    precondition {
      condition     = !local.definition_foreign
      error_message = "observed: the Stigmer definition ${local.stigmer_definition} on the cluster belongs to ${local.definition_owner}, and this resource installs the operator as release ${local.operator_release} in namespace ${local.operator_namespace}\nmeaning: Helm refuses to take over a definition another release owns, and Stigmer runs one operator per cluster\nnext step: ${local.definition_next_step}"
    }
    precondition {
      condition     = length(local.foreign_stigmers) == 0
      error_message = length(local.foreign_stigmers) == 0 ? "" : "observed: Stigmer ${local.foreign_stigmers[0].metadata.namespace}/${local.foreign_stigmers[0].metadata.name} is declared on the cluster, and this resource did not declare it\nmeaning: Stigmer runs one install per cluster, and this resource declares the namespace's Stigmer as stigmer: the operator serves only the oldest Stigmer and refuses every other one\nnext step: delete the other Stigmer first (kubectl -n ${local.foreign_stigmers[0].metadata.namespace} delete stigmer ${local.foreign_stigmers[0].metadata.name}; save its keys Secret before you delete its namespace), or keep it and do not install KubernetesStigmer on this cluster"
    }
    precondition {
      condition     = !local.downgrade
      error_message = "observed: Stigmer ${local.namespace}/${local.stigmer_name} records Stigmer v${local.running_version} on this cluster, and this ${local.kind_name} installs ${var.spec.version}\nmeaning: Stigmer does not move from v${local.running_version} to ${var.spec.version} in place: its server refuses to start on a database a newer release migrated\nnext step: set spec.version back to ${local.running_version} or later; to go back to ${var.spec.version}, restore a backup taken under it into a new install instead"
    }
  }
}
