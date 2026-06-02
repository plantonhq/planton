##############################################
# main.tf
#
# Main orchestration file for deploying Istio
# service mesh on a Kubernetes cluster.
#
# This module installs Istio using the official
# Helm charts with three separate releases:
#  1. istio/base - CRDs and base resources
#  2. istiod - Control plane (pilot)
#  3. istio-gateway - Ingress gateway
#
# The module follows Istio's recommended Helm-based
# installation approach for production deployments.
#
# Resources Created:
#  1. Kubernetes Namespace (istio-system)
#  2. Kubernetes Namespace (istio-ingress)
#  3. Helm Release (istio base)
#  4. Helm Release (istiod control plane)
#  5. Helm Release (istio ingress gateway)
#
# For more information see:
#  - examples.md for usage examples
#  - README.md for component documentation
#  - ../docs/README.md for deployment patterns
##############################################

##############################################
# 1. Create Istio System Namespace
#
# The istio-system namespace hosts the Istio
# control plane components (istiod).
#
# Only created if var.spec.create_namespace is true.
##############################################
resource "kubernetes_namespace" "istio_system" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.system_namespace
    labels = local.final_labels
  }
}

##############################################
# 2. Create Istio Ingress Namespace
#
# The istio-ingress namespace hosts the Istio
# ingress gateway for handling external traffic.
#
# Only created if var.spec.create_namespace is true.
##############################################
resource "kubernetes_namespace" "istio_ingress" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.gateway_namespace
    labels = local.final_labels
  }
}

##############################################
# 3. Deploy Istio Base via Helm
#
# Installs Istio base resources including CRDs
# from the official Istio Helm repository.
#
# This must be installed before istiod.
#
# Helm release name uses {metadata.name}-base to
# avoid conflicts when multiple instances share a namespace.
##############################################
resource "helm_release" "istio_base" {
  name       = local.base_release_name
  namespace  = local.system_namespace
  repository = local.helm_repo
  chart      = local.base_chart_name
  version    = local.chart_version

  create_namespace = false
  atomic           = true
  cleanup_on_fail  = true
  wait_for_jobs    = true
  timeout          = 180
}

##############################################
# 4. Deploy Istiod (Control Plane) via Helm
#
# Installs the Istio control plane (istiod) which
# includes:
#  - Pilot (traffic management)
#  - Citadel (certificate management)
#  - Galley (configuration validation)
#
# Resource limits are configured from the spec.
#
# Helm release name uses {metadata.name}-istiod to
# avoid conflicts when multiple instances share a namespace.
##############################################
resource "helm_release" "istiod" {
  name       = local.istiod_release_name
  namespace  = local.system_namespace
  repository = local.helm_repo
  chart      = local.istiod_chart_name
  version    = local.chart_version

  create_namespace = false
  atomic           = true
  cleanup_on_fail  = true
  wait_for_jobs    = true
  timeout          = 180

  # Configure pilot (istiod) resources from spec.
  # helm provider v3 expects `set` as a list-of-objects attribute (not nested blocks).
  set = [
    {
      name  = "pilot.resources.requests.cpu"
      value = var.spec.container.resources.requests.cpu
    },
    {
      name  = "pilot.resources.requests.memory"
      value = var.spec.container.resources.requests.memory
    },
    {
      name  = "pilot.resources.limits.cpu"
      value = var.spec.container.resources.limits.cpu
    },
    {
      name  = "pilot.resources.limits.memory"
      value = var.spec.container.resources.limits.memory
    },
  ]

  depends_on = [helm_release.istio_base]
}

##############################################
# 5. Deploy Istio Ingress Gateway via Helm
#
# Installs the Istio ingress gateway for handling
# external traffic entering the service mesh.
#
# Configured as ClusterIP by default (can be
# exposed via LoadBalancer or other ingress).
#
# Helm release name uses {metadata.name}-gateway to
# avoid conflicts when multiple instances share a namespace.
##############################################
resource "helm_release" "istio_gateway" {
  name       = local.gateway_release_name
  namespace  = local.gateway_namespace
  repository = local.helm_repo
  chart      = local.gateway_chart_name
  version    = local.chart_version

  create_namespace = false
  atomic           = true
  cleanup_on_fail  = true
  wait_for_jobs    = true
  timeout          = 180

  # Configure gateway service type.
  # helm provider v3 expects `set` as a list-of-objects attribute (not nested blocks).
  set = [
    {
      name  = "service.type"
      value = "ClusterIP"
    },
  ]

  depends_on = [helm_release.istiod]
}

