##############################################
# main.tf
#
# Main orchestration file for KubernetesGhaRunnerScaleSet
# deployment using Terraform.
#
# This module deploys a GitHub Actions Runner Scale Set
# using the official Helm chart.
#
# Infrastructure Components:
#  1. Kubernetes Namespace (if create_namespace is true)
#  2. PVCs for persistent volumes
#  3. Helm Release for the runner scale set
#
# After deployment, runners will register with GitHub
# and start processing jobs from matching workflows.
##############################################

# Create namespace if requested
resource "kubernetes_namespace" "this" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.labels
  }
}

# Create PVCs for persistent volumes
resource "kubernetes_persistent_volume_claim" "this" {
  for_each = { for pv in var.spec.persistent_volumes : pv.name => pv }

  metadata {
    name      = "${local.release_name}-${each.value.name}"
    namespace = local.namespace
    labels    = local.labels
  }

  spec {
    access_modes       = each.value.access_modes
    storage_class_name = each.value.storage_class != "" ? each.value.storage_class : null

    resources {
      requests = {
        storage = each.value.size
      }
    }
  }

  depends_on = [kubernetes_namespace.this]
}

# Deploy the runner scale set via Helm
# For OCI charts, the full URL must be passed as the chart parameter
# (repository doesn't work with OCI registries in Terraform helm_release)
resource "helm_release" "this" {
  name             = local.release_name
  namespace        = local.namespace
  create_namespace = false # We handle namespace creation ourselves

  chart   = local.chart_oci
  version = local.chart_version

  values = [yamlencode(local.helm_values_final)]

  depends_on = [
    kubernetes_namespace.this,
    kubernetes_persistent_volume_claim.this
  ]
}

