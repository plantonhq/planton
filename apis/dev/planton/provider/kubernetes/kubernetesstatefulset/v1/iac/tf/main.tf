##############################################
# main.tf
#
# Main orchestration file for KubernetesStatefulSet
# deployment using Terraform.
#
# This module creates a production-ready Kubernetes
# StatefulSet with the following capabilities:
#
# Infrastructure Components:
#  1. Kubernetes Namespace (defined here)
#  2. Service Account with RBAC (statefulset.tf)
#  3. Kubernetes StatefulSet with:
#     - Main application container
#     - Optional sidecar containers
#     - Health probes (startup, liveness, readiness)
#     - Resource requests and limits
#     - Environment variables and secrets
#     - Persistent volume claims
#  4. Headless Service for stable network identity (service.tf)
#  5. ClusterIP Service for client access (service.tf)
#  6. Environment Secrets (if configured) (secret.tf)
#  7. Image Pull Secret for private registries (statefulset.tf)
#
# StatefulSet Features:
#  - Stable, unique network identifiers
#  - Stable, persistent storage via PVCs
#  - Ordered, graceful deployment and scaling
#  - Ordered, automated rolling updates
#
# Module Structure:
#  - main.tf: Namespace creation and documentation (this file)
#  - statefulset.tf: StatefulSet, ServiceAccount, ImagePullSecret
#  - service.tf: Headless and ClusterIP Services
#  - secret.tf: Application secrets management
#  - locals.tf: Computed values and label management
#  - variables.tf: Input variable definitions
#  - outputs.tf: Module outputs (FQDNs, service names, etc.)
##############################################

##############################################
# Namespace Resource
#
# Creates a dedicated Kubernetes namespace for the
# statefulset deployment if create_namespace is true.
##############################################
resource "kubernetes_namespace" "this" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.final_labels
  }
}
