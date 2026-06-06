##############################################
# main.tf
#
# Main orchestration file for deploying the
# Rook Ceph Cluster on Kubernetes.
#
# This module installs a Ceph storage cluster
# using Helm, which provides distributed block,
# file, and object storage on Kubernetes.
#
# Resources Created:
#  1. Kubernetes Namespace (conditional)
#  2. Helm Release (rook-ceph-cluster chart)
#     - CephCluster custom resource
#     - CephBlockPool resources
#     - CephFilesystem resources
#     - CephObjectStore resources
#     - StorageClasses for each pool type
#
# Prerequisites:
#  - Rook Ceph Operator must be installed
#  - Raw block devices on storage nodes
#
# For more information see:
#  - ../README.md for component documentation
#  - ../../docs/README.md for research docs
##############################################

##############################################
# 1. Namespace Management
#
# Conditionally create namespace based on
# spec.create_namespace flag. If false, the
# namespace is assumed to already exist.
##############################################
resource "kubernetes_namespace_v1" "rook_ceph_cluster" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.final_labels
  }
}

##############################################
# 2. Deploy Rook Ceph Cluster via Helm
#
# Installs the Rook Ceph Cluster from the
# official Rook Helm repository.
#
# The chart will create:
#  - CephCluster custom resource
#  - CephBlockPool resources with StorageClasses
#  - CephFilesystem resources with StorageClasses
#  - CephObjectStore resources with StorageClasses
##############################################
resource "helm_release" "rook_ceph_cluster" {
  name       = local.helm_release_name
  repository = local.helm_chart_repo
  chart      = local.helm_chart_name
  version    = local.helm_chart_version
  namespace  = local.namespace

  # Disable namespace creation - we either created it above or it already exists
  create_namespace = false

  # Helm release options
  atomic          = true # Atomic rollback on failure
  cleanup_on_fail = true # Clean up resources if deployment fails
  wait            = true # Wait for resources to be ready
  wait_for_jobs   = true # Wait for any jobs to complete
  timeout         = 600  # 10 minutes timeout (Ceph cluster takes longer)

  # Configure Rook Ceph Cluster
  values = [yamlencode({
    # Operator namespace
    operatorNamespace = local.operator_namespace

    # Cluster name
    clusterName = local.ceph_cluster_name

    # Ceph image
    cephImage = {
      repository       = local.ceph_image.repository
      tag              = local.ceph_image.tag
      allowUnsupported = local.ceph_image.allow_unsupported
    }

    # Toolbox for debugging
    toolbox = {
      enabled = var.spec.enable_toolbox
    }

    # Monitoring
    monitoring = {
      enabled = var.spec.enable_monitoring
    }

    # CephCluster spec
    cephClusterSpec = local.ceph_cluster_spec

    # Block pools
    cephBlockPools = local.ceph_block_pools

    # Filesystems
    cephFileSystems = local.ceph_filesystems

    # Object stores
    cephObjectStores = local.ceph_object_stores
  })]

  # Depend on namespace
  depends_on = [
    kubernetes_namespace_v1.rook_ceph_cluster
  ]
}
