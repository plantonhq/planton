##############################################
# main.tf
#
# Main orchestration file for deploying the
# Elastic Cloud on Kubernetes (ECK) operator.
#
# This module installs the ECK operator using Helm,
# which provides automated lifecycle management for
# Elasticsearch, Kibana, APM Server, Enterprise Search,
# Beats, Elastic Agent, and Logstash on Kubernetes.
#
# Resources Created:
#  1. Kubernetes Namespace (elastic-system)
#  2. Helm Release (eck-operator chart)
#
# The ECK operator extends the Kubernetes API with
# Custom Resource Definitions (CRDs) for Elastic Stack
# components and provides automated operations including:
#  - Certificate management and rotation
#  - Rolling upgrades
#  - Scaling operations
#  - Cross-cluster replication
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
resource "kubernetes_namespace" "elastic_system" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.final_labels
  }
}

##############################################
# 2. Deploy ECK Operator via Helm
#
# Installs the Elastic Cloud on Kubernetes operator
# from the official Elastic Helm repository.
#
# The operator will:
#  - Install CRDs for Elastic Stack components
#  - Run as a controller that watches for custom resources
#  - Automatically manage Elastic Stack deployments
#  - Handle certificate management and rotation
#  - Orchestrate rolling upgrades
#
# The namespace reference depends on whether we created
# it or are using an existing one.
#
# Uses computed helm_release_name to avoid conflicts
# when multiple instances share a namespace.
##############################################
resource "helm_release" "eck_operator" {
  name       = local.helm_release_name
  repository = local.helm_chart_repo
  chart      = local.helm_chart_name
  version    = local.helm_chart_version
  namespace  = local.namespace

  # Disable namespace creation - we either created it above or it already exists
  create_namespace = false

  # Helm release options
  atomic          = false # Allow partial deployment for debugging
  cleanup_on_fail = true  # Clean up resources if deployment fails
  wait            = true  # Wait for resources to be ready
  wait_for_jobs   = true  # Wait for any jobs to complete
  timeout         = 180   # 3 minutes timeout

  # Configure ECK to inherit Planton labels
  values = [yamlencode({
    # Propagate Planton labels to all ECK-managed resources
    configKubernetes = {
      inherited_labels = local.inherited_labels
    }

    # Resource limits and requests for the ECK operator pod
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

  # Ignore changes to these fields to prevent drift
  lifecycle {
    ignore_changes = [
      metadata,
    ]
  }

  # Only depend on namespace if we created it
  depends_on = [
    kubernetes_namespace.elastic_system
  ]
}

