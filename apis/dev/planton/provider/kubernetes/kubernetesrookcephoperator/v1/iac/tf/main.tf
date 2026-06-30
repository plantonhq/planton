##############################################
# main.tf
#
# Main orchestration file for deploying the
# Rook Ceph Operator on Kubernetes.
#
# This module installs the Rook Ceph Operator
# using Helm, which provides automated lifecycle
# management for Ceph storage on Kubernetes.
#
# Resources Created:
#  1. Kubernetes Namespace (conditional)
#  2. Helm Release (rook-ceph chart)
#
# The Rook Ceph Operator extends the Kubernetes
# API with Custom Resource Definitions (CRDs) for
# Ceph storage components and provides automated
# operations including:
#  - CephCluster management
#  - Block storage (RBD)
#  - File storage (CephFS)
#  - Object storage (RGW)
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
resource "kubernetes_namespace_v1" "rook_ceph_operator" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.final_labels
  }
}

##############################################
# 2. Deploy Rook Ceph Operator via Helm
#
# Installs the Rook Ceph Operator from the
# official Rook Helm repository.
#
# The operator will:
#  - Install controllers for Ceph resources
#  - Watch for CephCluster, CephBlockPool, etc.
#  - Automatically manage Ceph daemons
#  - Deploy CSI drivers for PV provisioning
##############################################
resource "helm_release" "rook_ceph_operator" {
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
  timeout         = 300   # 5 minutes timeout (Rook may take longer)

  # Configure Rook Ceph Operator
  values = [yamlencode({
    # CRD management
    crds = {
      enabled = var.spec.crds_enabled
    }

    # Resource limits
    resources = {
      limits = {
        cpu    = try(var.spec.container.resources.limits.cpu, "500m")
        memory = try(var.spec.container.resources.limits.memory, "512Mi")
      }
      requests = {
        cpu    = try(var.spec.container.resources.requests.cpu, "200m")
        memory = try(var.spec.container.resources.requests.memory, "128Mi")
      }
    }

    # CSI configuration
    csi = {
      enableRbdDriver      = local.csi_config.enable_rbd_driver
      enableCephfsDriver   = local.csi_config.enable_cephfs_driver
      disableCsiDriver     = local.csi_config.disable_csi_driver ? "true" : "false"
      enableCSIHostNetwork = local.csi_config.enable_csi_host_network
      provisionerReplicas  = local.csi_config.provisioner_replicas
      csiAddons = {
        enabled = local.csi_config.enable_csi_addons
      }
      nfs = {
        enabled = local.csi_config.enable_nfs_driver
      }
    }
  })]

  # Depend on namespace
  depends_on = [
    kubernetes_namespace_v1.rook_ceph_operator
  ]
}
