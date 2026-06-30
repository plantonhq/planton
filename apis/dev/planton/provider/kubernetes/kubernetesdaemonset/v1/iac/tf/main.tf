##############################################
# main.tf
#
# Main orchestration file for KubernetesDaemonSet
# deployment using Terraform.
#
# This module creates a production-ready Kubernetes
# DaemonSet with the following capabilities:
#
# Infrastructure Components:
#  1. Kubernetes Namespace (optional, created if create_namespace=true)
#  2. ServiceAccount with RBAC (optional, created if create_service_account=true)
#  3. Kubernetes DaemonSet with:
#     - Main application container
#     - Optional sidecar containers
#     - Security context for privileged operations
#     - Resource requests and limits
#     - Environment variables and secrets
#     - Volume mounts (HostPath, ConfigMap, Secret, EmptyDir, PVC)
#  4. Environment Secrets (if configured)
#  5. ConfigMaps (if configured)
#  6. RBAC (ClusterRole/Role and bindings)
#
# DaemonSet Use Cases:
#  - Log collection daemons (Fluentd, Fluent Bit, Filebeat)
#  - Node monitoring agents (Prometheus Node Exporter, Datadog)
#  - Network plugins (Calico, Cilium)
#  - Storage daemons (Ceph, Longhorn)
#  - Security agents (Falco, Sysdig)
#
# Module Structure:
#  - main.tf: Namespace creation and documentation (this file)
#  - daemonset.tf: DaemonSet resource with containers
#  - service_account.tf: ServiceAccount and RBAC
#  - secret.tf: Environment secrets management
#  - configmap.tf: ConfigMap resources
#  - locals.tf: Computed values and label management
#  - variables.tf: Input variable definitions
#  - outputs.tf: Module outputs
#
# For detailed examples and usage patterns, see:
#  - examples.md: Terraform configuration examples
#  - ../README.md: Component overview and features
#  - ../../docs/README.md: Research and production best practices
##############################################

##############################################
# Namespace Resource
#
# Creates a dedicated Kubernetes namespace for the
# DaemonSet deployment when create_namespace is true.
##############################################
resource "kubernetes_namespace" "this" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.final_labels
  }
}

