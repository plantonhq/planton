##############################################
# main.tf
#
# Main orchestration file for deploying the
# Tekton Operator on Kubernetes.
#
# This module installs the Tekton Operator using
# official release manifests, which provides
# automated lifecycle management for Tekton
# components (Pipelines, Triggers, Dashboard).
#
# Resources Created:
#  1. Tekton Operator (via kubectl_manifest)
#  2. TektonConfig CRD (to configure components)
#
# IMPORTANT: Namespace Behavior
# Tekton Operator manages its own namespaces:
# - 'tekton-operator' for the operator itself
# - 'tekton-pipelines' for Tekton components
# These are automatically created by the operator
# and cannot be customized by the user.
#
# The Tekton Operator extends the Kubernetes API
# with Custom Resource Definitions (CRDs) and
# provides automated operations including:
#  - Component installation and upgrades
#  - Configuration management
#  - Health monitoring
#
# For more information see:
#  - examples.md for usage examples
#  - ../README.md for component documentation
#  - ../../docs/README.md for deployment patterns
##############################################

##############################################
# 1. Deploy Tekton Operator
#
# Apply the Tekton Operator release manifests.
# This installs the operator controller which
# watches for TektonConfig CRDs.
##############################################
data "http" "tekton_operator_manifest" {
  url = local.operator_release_url
}

resource "kubectl_manifest" "tekton_operator" {
  for_each = {
    for idx, doc in split("---", data.http.tekton_operator_manifest.response_body) : idx => doc
    if trimspace(doc) != "" && !startswith(trimspace(doc), "#")
  }

  yaml_body = each.value

  # Wait for the previous manifest to be applied
  wait = true
}

##############################################
# 2. Create TektonConfig
#
# Configure which Tekton components to install.
# The operator will reconcile this and install
# the requested components in 'tekton-pipelines'
# namespace.
#
# Profiles:
#  - all: Pipelines + Triggers + Dashboard
#  - basic: Pipelines + Triggers
#  - lite: Pipelines only
##############################################
resource "kubectl_manifest" "tekton_config" {
  # Note: Do not set fields that the operator manages automatically (e.g., pipeline.enable-api-fields)
  # to avoid Server-Side Apply field conflicts
  yaml_body = <<-YAML
    apiVersion: operator.tekton.dev/v1alpha1
    kind: TektonConfig
    metadata:
      name: ${local.tekton_config_name}
    spec:
      profile: ${local.tekton_profile}
      targetNamespace: ${local.components_namespace}
  YAML

  depends_on = [kubectl_manifest.tekton_operator]
}
