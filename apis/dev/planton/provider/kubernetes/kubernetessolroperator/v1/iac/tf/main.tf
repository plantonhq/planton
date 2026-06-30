##############################################
# main.tf
#
# Main orchestration file for deploying the
# Apache Solr Operator on Kubernetes.
#
# This module installs the Solr Operator using Helm,
# which provides automated lifecycle management for
# Apache Solr on Kubernetes.
#
# Resources Created:
#  1. Kubernetes Namespace (conditional)
#  2. CRDs via kubectl_manifest (Solr CRDs)
#  3. Helm Release (solr-operator chart)
#
# The Solr Operator extends the Kubernetes API with
# Custom Resource Definitions (CRDs) for Solr components
# and provides automated operations including:
#  - SolrCloud cluster management
#  - Scaling operations
#  - Backup and restore
#
# For more information see:
#  - examples.md for usage examples
#  - ../README.md for component documentation
#  - ../../docs/README.md for deployment patterns
##############################################

##############################################
# 1. Namespace Management
#
# Conditionally create namespace based on
# spec.create_namespace flag. If false, the
# namespace is assumed to already exist.
##############################################
resource "kubernetes_namespace_v1" "solr_operator" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.final_labels
  }
}

##############################################
# 2. Apply Solr Operator CRDs
#
# The Solr Operator requires CRDs to be installed
# before the operator can be deployed. The CRDs
# are fetched from the official Apache Solr repository.
#
# Note: CRDs are cluster-scoped resources, so the
# resource name uses metadata.name to avoid conflicts
# if multiple Solr Operator instances are deployed.
##############################################
data "http" "solr_crds" {
  url = local.crd_manifest_url
}

resource "kubectl_manifest" "solr_crds" {
  yaml_body = data.http.solr_crds.response_body

  # Force new resource if CRD content changes
  force_new = true
}

##############################################
# 3. Deploy Solr Operator via Helm
#
# Installs the Apache Solr Operator from the
# official Apache Helm repository.
#
# The operator will:
#  - Install controllers for Solr resources
#  - Watch for SolrCloud, SolrBackup, etc. CRs
#  - Automatically manage Solr deployments
#
# Uses computed helm_release_name to avoid conflicts
# when multiple instances share a namespace.
##############################################
resource "helm_release" "solr_operator" {
  name       = local.helm_release_name
  repository = local.helm_chart_repo
  chart      = local.helm_chart_name
  version    = local.helm_chart_version
  namespace  = local.namespace

  # Disable namespace creation - we either created it above or it already exists
  create_namespace = false

  # Helm release options
  atomic          = true  # Atomic rollback on failure
  cleanup_on_fail = true  # Clean up resources if deployment fails
  wait            = true  # Wait for resources to be ready
  wait_for_jobs   = true  # Wait for any jobs to complete
  timeout         = 180   # 3 minutes timeout

  # Configure Solr Operator resource limits
  values = [yamlencode({
    resources = {
      limits = {
        cpu    = try(var.spec.container.resources.limits.cpu, "1000m")
        memory = try(var.spec.container.resources.limits.memory, "1Gi")
      }
      requests = {
        cpu    = try(var.spec.container.resources.requests.cpu, "50m")
        memory = try(var.spec.container.resources.requests.memory, "100Mi")
      }
    }
  })]

  # Depend on namespace and CRDs
  depends_on = [
    kubernetes_namespace_v1.solr_operator,
    kubectl_manifest.solr_crds
  ]
}
