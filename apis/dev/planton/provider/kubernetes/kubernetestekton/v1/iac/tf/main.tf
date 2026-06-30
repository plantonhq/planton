##############################################
# main.tf
#
# Main orchestration file for deploying Tekton
# on Kubernetes using official release manifests.
#
# This module installs Tekton Pipelines and
# optionally Tekton Dashboard using kubectl apply
# style manifest deployment.
#
# Resources Created:
#  1. Tekton Pipelines manifests
#  2. Tekton Dashboard manifests (if enabled)
#  3. ConfigMap patch for cloud events (if configured)
#  4. Gateway API resources for dashboard ingress (if enabled)
#
# IMPORTANT: Namespace Behavior
# Tekton manifests create the 'tekton-pipelines' namespace
# automatically. This cannot be customized.
#
# For more information see:
#  - examples.md for usage examples
#  - README.md for component documentation
##############################################

##############################################
# 1. Deploy Tekton Pipelines
#
# Apply the official Tekton Pipelines release manifests.
# This creates the tekton-pipelines namespace, CRDs,
# and all pipeline components.
#
# Equivalent to:
#   kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/{version}/release.yaml
##############################################
resource "kubernetes_namespace_v1" "tekton_pipelines" {
  metadata {
    name   = local.namespace
    labels = local.final_labels
  }
}

data "http" "tekton_pipeline_manifest" {
  url = local.pipeline_manifest_url
}

resource "kubectl_manifest" "tekton_pipelines" {
  for_each = {
    for idx, doc in split("---", data.http.tekton_pipeline_manifest.response_body) : idx => doc
    if trimspace(doc) != "" && !startswith(trimspace(doc), "#")
  }

  yaml_body = each.value
  wait      = true

  depends_on = [kubernetes_namespace_v1.tekton_pipelines]
}

##############################################
# 2. Deploy Tekton Dashboard (if enabled)
#
# Apply the official Tekton Dashboard release manifests.
# This adds the web UI for viewing pipelines, tasks, and runs.
#
# Equivalent to:
#   kubectl apply --filename https://infra.tekton.dev/tekton-releases/dashboard/{version}/release.yaml
##############################################
data "http" "tekton_dashboard_manifest" {
  count = local.dashboard_enabled ? 1 : 0
  url   = local.dashboard_manifest_url
}

resource "kubectl_manifest" "tekton_dashboard" {
  for_each = local.dashboard_enabled ? {
    for idx, doc in split("---", data.http.tekton_dashboard_manifest[0].response_body) : idx => doc
    if trimspace(doc) != "" && !startswith(trimspace(doc), "#")
  } : {}

  yaml_body = each.value
  wait      = true

  depends_on = [kubectl_manifest.tekton_pipelines]
}

##############################################
# 3. Configure Cloud Events (if specified)
#
# Patch the config-defaults ConfigMap to set the
# cloud events sink URL. This enables Tekton to
# send CloudEvents for TaskRun and PipelineRun
# lifecycle events.
##############################################
resource "kubernetes_config_map_v1_data" "cloud_events_config" {
  count = local.cloud_events_enabled ? 1 : 0

  metadata {
    name      = "config-defaults"
    namespace = local.namespace
  }

  data = {
    "default-cloud-events-sink" = local.cloud_events_sink_url
  }

  force = true

  depends_on = [kubectl_manifest.tekton_pipelines]
}
