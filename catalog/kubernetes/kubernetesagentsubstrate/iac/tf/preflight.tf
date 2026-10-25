# The refusals (preflight.go, upgrade.go, workloadowner), as preconditions
# on a terraform_data resource that everything else depends on. Every value
# they read is known at plan time, so a failing condition stops the plan
# before anything is written, and conditions are not evaluated on destroy,
# so no refusal stands between a user and deleting their stack. The words
# are the Go module's, three parts each.
#
# NO depends_on ON THE DATA SOURCES: a data source that depends on a resource
# is read at apply time, and the refusals must fire at plan time.

locals {
  owned_workloads = merge(
    {
      "Deployment/${local.pod_certificates_namespace}/${local.pod_certificates_name}" = { api_version = "apps/v1", kind = "Deployment", namespace = local.pod_certificates_namespace, name = local.pod_certificates_name }
      "Deployment/${local.namespace}/${local.api_server_name}"                        = { api_version = "apps/v1", kind = "Deployment", namespace = local.namespace, name = local.api_server_name }
      "Deployment/${local.namespace}/${local.identity_server_name}"                   = { api_version = "apps/v1", kind = "Deployment", namespace = local.namespace, name = local.identity_server_name }
      "Deployment/${local.namespace}/${local.controller_name}"                        = { api_version = "apps/v1", kind = "Deployment", namespace = local.namespace, name = local.controller_name }
      "DaemonSet/${local.namespace}/${local.node_agent_name}"                         = { api_version = "apps/v1", kind = "DaemonSet", namespace = local.namespace, name = local.node_agent_name }
      "Deployment/${local.namespace}/${local.router_name}"                            = { api_version = "apps/v1", kind = "Deployment", namespace = local.namespace, name = local.router_name }
      "Deployment/${local.namespace}/${local.egress_name}"                            = { api_version = "apps/v1", kind = "Deployment", namespace = local.namespace, name = local.egress_name }
    },
    local.egress.provider == "kubernetes" ? { "Deployment/${local.namespace}/${local.credential_provider_name}" = { api_version = "apps/v1", kind = "Deployment", namespace = local.namespace, name = local.credential_provider_name } } : {},
    local.db_kind == "bundled" ? { "StatefulSet/${local.namespace}/${local.postgres_name}" = { api_version = "apps/v1", kind = "StatefulSet", namespace = local.namespace, name = local.postgres_name } } : {},
  )
  checked_namespaces = {
    (local.namespace)                  = local.create_namespace
    (local.pod_certificates_namespace) = true
  }
}

# Each workload the module applies under upstream's fixed names, read by
# name: one that exists without this kind's ownership label belongs to
# another installation and must not be taken over.
data "kubernetes_resources" "owned_workload" {
  for_each = local.owned_workloads

  api_version    = each.value.api_version
  kind           = each.value.kind
  namespace      = each.value.namespace
  field_selector = "metadata.name=${each.value.name}"
}

# Both namespaces, read by name: whether each exists, and who created it.
data "kubernetes_resources" "namespace" {
  for_each = local.checked_namespaces

  api_version    = "v1"
  kind           = "Namespace"
  field_selector = "metadata.name=${each.key}"
}

# The bundled database's volume, which outlives a destroy while another
# resource owns the namespace and carries the minor that wrote it.
data "kubernetes_resources" "database_volume" {
  api_version    = "v1"
  kind           = "PersistentVolumeClaim"
  namespace      = local.namespace
  field_selector = "metadata.name=${local.postgres_volume_claim}"
}

locals {
  # The first workload another installation holds, workloads in the Go
  # module's order (pod certificates first).
  owned_workload_order = [
    "Deployment/${local.pod_certificates_namespace}/${local.pod_certificates_name}",
    "Deployment/${local.namespace}/${local.api_server_name}",
    "Deployment/${local.namespace}/${local.identity_server_name}",
    "Deployment/${local.namespace}/${local.controller_name}",
    "DaemonSet/${local.namespace}/${local.node_agent_name}",
    "Deployment/${local.namespace}/${local.router_name}",
    "Deployment/${local.namespace}/${local.egress_name}",
    "Deployment/${local.namespace}/${local.credential_provider_name}",
    "StatefulSet/${local.namespace}/${local.postgres_name}",
  ]
  foreign_workloads = [
    for key in local.owned_workload_order : {
      kind      = local.owned_workloads[key].kind
      namespace = local.owned_workloads[key].namespace
      name      = local.owned_workloads[key].name
      owner = (
        length(try(data.kubernetes_resources.owned_workload[key].objects[0].metadata.managedFields, [])) > 0
        ? "last written by field manager(s) ${join(", ", [for m in data.kubernetes_resources.owned_workload[key].objects[0].metadata.managedFields : m.manager])}"
        : "it carries no ownership marks"
      )
    }
    if contains(keys(local.owned_workloads), key)
    && length(try(data.kubernetes_resources.owned_workload[key].objects, [])) > 0
    && try(data.kubernetes_resources.owned_workload[key].objects[0].metadata.labels[local.owner_label], "") != local.kind_name
  ]
  foreign_workload = length(local.foreign_workloads) > 0 ? local.foreign_workloads[0] : null

  # The namespace checks, in order: ate-system, then the pod-certificate one.
  namespace_states = [
    for name in [local.namespace, local.pod_certificates_namespace] : {
      name   = name
      create = local.checked_namespaces[name]
      object = try(data.kubernetes_resources.namespace[name].objects[0], null)
      ours   = try(data.kubernetes_resources.namespace[name].objects[0].metadata.labels[local.owner_label], "") == local.kind_name
      owner = (
        length(try(data.kubernetes_resources.namespace[name].objects[0].metadata.managedFields, [])) > 0
        ? "last written by field manager(s) ${join(", ", [for m in data.kubernetes_resources.namespace[name].objects[0].metadata.managedFields : m.manager])}"
        : "it carries no ownership marks"
      )
    }
  ]
  namespace_missing      = [for n in local.namespace_states : n if !n.create && n.object == null]
  namespace_foreign      = [for n in local.namespace_states : n if n.create && n.object != null && !n.ours]
  namespace_created_here = [for n in local.namespace_states : n if !n.create && n.object != null && n.ours]

  # ---- the upgrade rule (upgrade.go) -------------------------------------------
  api_object = try(data.kubernetes_resources.owned_workload["Deployment/${local.namespace}/${local.api_server_name}"].objects[0], null)
  installed_version = (
    try(local.api_object.metadata.labels[local.owner_label], "") == local.kind_name
    ? try(local.api_object.metadata.annotations[local.version_annotation], "") : ""
  )
  volume_minor   = try(data.kubernetes_resources.database_volume.objects[0].metadata.labels[local.minor_label], "")
  volume_version = local.db_kind == "bundled" && local.volume_minor != "" ? "v${local.volume_minor}" : ""

  rule_before_one = "before 1.0 every minor release is a reinstall, because the API server migrates its database schema irreversibly on its first start, and a new minor's components are not built to run on an older minor's data"
  rule_major      = "a major release is a reinstall"
  rule_skip       = "upgrades go one minor release at a time"
  rule_rollback   = "a rollback goes back at most one minor release, because the newer API server has already migrated the database schema forward"

  requested_parts = [for p in slice(split(".", trimprefix(local.version, "v")), 0, 2) : tonumber(p)]
  upgrade_rules = {
    for record, installed in { deployment = local.installed_version, volume = local.volume_version } : record => (
      length(regexall("^v?[0-9]+\\.[0-9]+", installed)) == 0 ? "" : (
        tonumber(split(".", trimprefix(installed, "v"))[0]) == local.requested_parts[0] && tonumber(split(".", trimprefix(installed, "v"))[1]) == local.requested_parts[1] ? "" :
        tonumber(split(".", trimprefix(installed, "v"))[0]) != local.requested_parts[0] ? (
          tonumber(split(".", trimprefix(installed, "v"))[0]) == 0 || local.requested_parts[0] == 0 ? local.rule_before_one : local.rule_major
        ) :
        local.requested_parts[0] == 0 ? local.rule_before_one :
        local.requested_parts[1] > tonumber(split(".", trimprefix(installed, "v"))[1]) + 1 ? local.rule_skip :
        local.requested_parts[1] < tonumber(split(".", trimprefix(installed, "v"))[1]) - 1 ? local.rule_rollback : ""
      )
    )
  }
}

resource "terraform_data" "preflight" {
  lifecycle {
    precondition {
      condition     = local.requested_namespace == local.namespace
      error_message = "observed: spec.namespace resolved to \"${local.requested_namespace}\"\nmeaning: Agent Substrate runs only in ${local.namespace}: upstream's manifests, the API's server name and its token audience name it\nnext step: reference a KubernetesNamespace named ${local.namespace}, or set spec.namespace to ${local.namespace}"
    }
    precondition {
      condition     = local.foreign_workload == null
      error_message = local.foreign_workload == null ? "" : "observed: ${local.foreign_workload.kind} ${local.foreign_workload.namespace}/${local.foreign_workload.name} already runs on the cluster and was not created by this ${local.kind_name} (${local.foreign_workload.owner})\nmeaning: another installation of the same software runs here, such as a managed add-on or a kubectl apply of the upstream release; installing over it would take over its workloads, and destroying this resource later would delete them\nnext step: remove the other installation first, or keep it and do not install ${local.kind_name} on this cluster; to hand over an installation you made yourself, import it, then mark it as this kind's (`kubectl -n ${local.foreign_workload.namespace} label ${lower(local.foreign_workload.kind)} ${local.foreign_workload.name} ${local.owner_label}=${local.kind_name}`) and re-run"
    }
    precondition {
      condition     = length(local.namespace_missing) == 0
      error_message = length(local.namespace_missing) == 0 ? "" : "observed: namespace ${local.namespace_missing[0].name} does not exist on the cluster, and this ${local.kind_name} does not create it (spec.create_namespace is false)\nmeaning: everything installs into ${local.namespace_missing[0].name}, so its owner must create it first; nothing could start without it\nnext step: deploy the KubernetesNamespace named ${local.namespace_missing[0].name} first and reference it from spec.namespace so the two deploy in order, or set spec.create_namespace to true to let this resource create and own it"
    }
    precondition {
      condition     = length(local.namespace_foreign) == 0
      error_message = length(local.namespace_foreign) == 0 ? "" : "observed: namespace ${local.namespace_foreign[0].name} already exists on the cluster and was not created by this ${local.kind_name} (${local.namespace_foreign[0].owner})\nmeaning: with spec.create_namespace true this resource would take the namespace over, and destroying it later would delete the namespace and everything other components put in it\nnext step: set spec.create_namespace to false and reference the namespace's owner from spec.namespace; or, if you created it yourself for this installation, mark it as this kind's (`kubectl label namespace ${local.namespace_foreign[0].name} ${local.owner_label}=${local.kind_name}`) and re-run"
    }
    precondition {
      condition     = length(local.namespace_created_here) == 0
      error_message = length(local.namespace_created_here) == 0 ? "" : "observed: namespace ${local.namespace_created_here[0].name} was created by this ${local.kind_name}, and spec.create_namespace is now false\nmeaning: turning create_namespace off would delete the namespace, and with it this installation's workloads and everything other components put there\nnext step: keep spec.create_namespace true; to move the namespace to a KubernetesNamespace, destroy this resource first (resource definitions it keeps, and everything outside ${local.namespace_created_here[0].name}, stay), deploy the KubernetesNamespace, then deploy this resource again with create_namespace false"
    }
    precondition {
      condition     = local.upgrade_rules.deployment == ""
      error_message = local.upgrade_rules.deployment == "" ? "" : "observed: Deployment ${local.namespace}/${local.api_server_name} records Agent Substrate ${local.installed_version} on this cluster, and this ${local.kind_name} installs ${local.version}\nmeaning: Agent Substrate does not move from ${local.installed_version} to ${local.version} in place: ${local.upgrade_rules.deployment}\nnext step: to install ${local.version}, back up what you need from the database, destroy this resource, then clear the database it used (`kubectl -n ${local.namespace} delete pvc ${local.postgres_volume_claim}` for the bundled one, or drop the substrate schema of an external or Cloud SQL one), and apply again; to keep ${local.installed_version}, set spec.version back to it, or, for a release of another install shape, deploy this resource from the Planton release that installed it"
    }
    precondition {
      condition     = local.upgrade_rules.volume == ""
      error_message = local.upgrade_rules.volume == "" ? "" : "observed: PersistentVolumeClaim ${local.namespace}/${local.postgres_volume_claim}, the bundled database's volume left by an earlier installation, records Agent Substrate ${local.volume_version} on this cluster, and this ${local.kind_name} installs ${local.version}\nmeaning: Agent Substrate does not move from ${local.volume_version} to ${local.version} in place: ${local.upgrade_rules.volume}\nnext step: to install ${local.version}, back up what you need from that database, delete the volume (`kubectl -n ${local.namespace} delete pvc ${local.postgres_volume_claim}`), and apply again; to keep the data, set spec.version to a release of the minor that wrote it, or, for one of another install shape, deploy this resource from the Planton release that wrote it"
    }
    precondition {
      condition = alltrue([
        local.podcert_container != null,
        local.api_container != null,
        local.identity_container != null,
        local.controller_container != null,
        local.node_container != null,
        local.node_sidecar != null,
        local.router_container != null,
        local.egress_extproc != null,
        local.egress_minter != null,
        local.postgres_container != null,
        length([for e in local.podcert_container.env : e if e.name == "WORKERS_PER_SIGNER"]) == 1,
        length(local.postgres_hook) == 3,
        length(local.agw) == 6 && length(local.agw_egress) == 3 && length(local.cordon_patches) == 2,
        local.router_policy_upstream != null,
        alltrue([for role, modes in local.octal_modes : length(modes) <= 4]),
      ])
      error_message = "observed: Agent Substrate ${local.release}'s manifests no longer carry the workloads, containers, start hook and component patches the module patches, where it expects them\nmeaning: the published release changed shape\nnext step: the module's pin needs a catalog fix"
    }
  }
}
