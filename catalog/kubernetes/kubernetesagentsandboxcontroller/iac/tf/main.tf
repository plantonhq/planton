# KubernetesAgentSandboxController Terraform module.
#
# Installs agent-sandbox from upstream's own release: the release asset the
# extensions setting selects (sandbox.yaml, or sandbox-with-extensions.yaml),
# and the router's deployment files at the same tag when a router is
# configured, each with the typed settings applied onto upstream's own
# objects (locals.tf, documents.tf).
#
# ORDER: every refusal fires at plan time, before anything is written: a
# controller or router another installation created, a namespace that is
# missing or owned by someone else, a namespace other than upstream's, and an
# upstream release whose shape moved (the preflight below), and a definition
# someone else owns or a newer one (helm_crds.tf). Then the Namespace when
# this resource owns it; the kept definitions; and everything else after
# both, with the rollout awaited. The controller holds no
# finalizers and registers no webhook at the pinned release, so a destroy
# needs no live controller to drain anything, and nothing deadlocks waiting
# for the controller to start before its definitions. A finished apply
# therefore means a serving controller (the module adds readiness probes on
# upstream's own health endpoints).
#
# DESTROY: the controller, the router and their RBAC are deleted, and the
# Namespace when this resource created it; the definitions are kept
# (apply_only) unless crds.keep_on_uninstall is false.
#
# NO depends_on ON THE DATA SOURCES: a data source that depends on a resource
# is read at apply time, and the refusals must fire at plan time.

data "http" "release_asset" {
  url                = local.release_asset_url
  request_timeout_ms = 120000

  lifecycle {
    postcondition {
      condition     = self.status_code == 200
      error_message = "observed: ${self.url} answered HTTP ${self.status_code}\nmeaning: upstream no longer serves agent-sandbox ${local.release} at the address this module pins, or the host is failing\nnext step: confirm what upstream publishes (curl -I ${self.url}); if the release was moved or withdrawn, the module's pin needs a catalog fix"
    }
  }
}

data "http" "router_file" {
  for_each = toset(local.router_files)

  url                = "${local.router_files_base_url}${each.key}"
  request_timeout_ms = 120000

  lifecycle {
    postcondition {
      condition     = self.status_code == 200
      error_message = "observed: ${self.url} answered HTTP ${self.status_code}\nmeaning: upstream no longer serves agent-sandbox ${local.release} at the address this module pins, or the host is failing\nnext step: confirm what upstream publishes (curl -I ${self.url}); if the release was moved or withdrawn, the module's pin needs a catalog fix"
    }
  }
}

# Each Deployment the module applies under upstream's fixed names, read by
# name before anything is written: one that exists without this kind's
# ownership label belongs to another installation (a managed add-on, a
# kubectl apply of the upstream release) and must not be taken over.
data "kubernetes_resources" "owned_deployment" {
  for_each = toset(local.owned_deployment_names)

  api_version    = "apps/v1"
  kind           = "Deployment"
  namespace      = local.namespace
  field_selector = "metadata.name=${each.key}"
}

# Upstream's namespace, read by name: whether it exists, and whether this
# resource created it.
data "kubernetes_resources" "namespace" {
  api_version    = "v1"
  kind           = "Namespace"
  field_selector = "metadata.name=${local.namespace}"
}

# The refusals. Their values are all known at plan time, so a failing
# condition stops the plan before anything is written, and conditions are not
# evaluated on destroy, so no refusal stands between a user and deleting
# their stack.
resource "terraform_data" "preflight" {
  lifecycle {
    precondition {
      condition     = local.requested_namespace == local.namespace
      error_message = "observed: spec.namespace resolved to \"${local.requested_namespace}\"\nmeaning: agent-sandbox runs only in ${local.namespace}: upstream's manifests, RBAC and the default sandbox network policy name it\nnext step: reference a KubernetesNamespace named ${local.namespace}, or set spec.namespace to ${local.namespace}"
    }
    precondition {
      condition     = length(local.foreign_deployments) == 0
      error_message = length(local.foreign_deployments) == 0 ? "" : "observed: Deployment ${local.namespace}/${keys(local.foreign_deployments)[0]} already runs on the cluster and was not created by this ${local.kind_name} (${values(local.foreign_deployments)[0]})\nmeaning: another installation of the same software runs here, such as a managed add-on or a kubectl apply of the upstream release; installing over it would take over its workloads, and destroying this resource later would delete them\nnext step: remove the other installation first, or keep it and do not install ${local.kind_name} on this cluster; to hand over an installation you made yourself, import it, then mark it as this kind's (`kubectl -n ${local.namespace} label deployment ${keys(local.foreign_deployments)[0]} ${local.owner_label}=${local.kind_name}`) and re-run"
    }
    precondition {
      condition     = !local.namespace_missing
      error_message = local.namespace_missing ? "observed: namespace ${local.namespace} does not exist on the cluster, and this ${local.kind_name} does not create it (spec.create_namespace is false)\nmeaning: everything installs into ${local.namespace}, so its owner must create it first; nothing could start without it\nnext step: deploy the KubernetesNamespace named ${local.namespace} first and reference it from spec.namespace so the two deploy in order, or set spec.create_namespace to true to let this resource create and own it" : ""
    }
    precondition {
      condition     = !local.namespace_created_here
      error_message = local.namespace_created_here ? "observed: namespace ${local.namespace} was created by this ${local.kind_name}, and spec.create_namespace is now false\nmeaning: turning create_namespace off would delete the namespace, and with it this installation's workloads and everything other components put there\nnext step: keep spec.create_namespace true; to move the namespace to a KubernetesNamespace, destroy this resource first (resource definitions it keeps, and everything outside ${local.namespace}, stay), deploy the KubernetesNamespace, then deploy this resource again with create_namespace false" : ""
    }
    precondition {
      condition     = !local.namespace_foreign
      error_message = local.namespace_foreign ? "observed: namespace ${local.namespace} already exists on the cluster and was not created by this ${local.kind_name} (${local.namespace_owner})\nmeaning: with spec.create_namespace true this resource would take the namespace over, and destroying it later would delete the namespace and everything other components put in it\nnext step: set spec.create_namespace to false and reference the namespace's owner from spec.namespace; or, if you created it yourself for this installation, mark it as this kind's (`kubectl label namespace ${local.namespace} ${local.owner_label}=${local.kind_name}`) and re-run" : ""
    }
    precondition {
      condition     = local.namespace_upstream != null && local.controller_upstream_container != null && try(local.documents[local.controller_service_id], null) != null
      error_message = "observed: upstream's release asset at ${local.release} no longer carries the ${local.namespace} Namespace, the agent-sandbox-controller Deployment, its container and its Service the module patches\nmeaning: upstream changed the release's shape\nnext step: the module's pin needs a catalog fix"
    }
    precondition {
      condition     = !local.router_on || local.router_upstream_image_repo == local.router_image_repo
      error_message = "observed: upstream's router deployment at ${local.release} runs ${try(local.router_upstream_container.image, "no container named sandbox-router")}, not ${local.router_image_repo}\nmeaning: upstream renamed the router image or the container\nnext step: the module's pin needs a catalog fix"
    }
    precondition {
      condition     = local.router_netpol_shape_as_expected
      error_message = "observed: upstream's router NetworkPolicy at ${local.release} no longer has the proxy, metrics and probe ingress rules and the sandbox egress rule where the module expects them\nmeaning: upstream changed the policy's shape\nnext step: the module's pin needs a catalog fix"
    }
  }
}

# The Namespace, only when this resource owns it: applied first and deleted
# last, and its delete blocks until the namespace is fully gone, because the
# name is fixed and a destroy followed by an apply would otherwise race the
# terminating namespace.
resource "kubectl_manifest" "namespace" {
  for_each = local.namespace_documents

  yaml_body = yamlencode(each.value)

  server_side_apply = true
  force_conflicts   = true
  wait              = true

  depends_on = [terraform_data.preflight]
}

# Everything else: the controller (with the typed settings and probes), its
# RBAC and metrics Service, the label-domain ConfigMap, and the router. Applied
# after the namespace and the kept definitions, so the controller starts with
# its definitions served and the rollout can be awaited.
resource "kubectl_manifest" "workloads" {
  for_each = local.workload_documents

  yaml_body = yamlencode(each.value)

  server_side_apply = true
  force_conflicts   = true
  wait_for_rollout  = true

  depends_on = [terraform_data.preflight, kubectl_manifest.namespace, kubectl_manifest.helm_crds]
}
