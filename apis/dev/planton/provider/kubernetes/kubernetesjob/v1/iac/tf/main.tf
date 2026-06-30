##############################################
# main.tf
#
# Main orchestration file for KubernetesJob
# deployment using Terraform.
#
# This module creates the following resources:
#  1. Kubernetes Namespace (defined here)
#  2. Service Account for the Job (job.tf)
#  3. Image Pull Secret (if docker_config_json provided) (job.tf)
#  4. Environment Secrets (if spec.env.secrets provided) (secret.tf)
#  5. Job Resource (job.tf)
#
# The module follows best practices:
#  - Uses namespace isolation for each Job
#  - Creates dedicated service account for RBAC
#  - Supports private container registries
#  - Implements proper resource management
#  - Provides comprehensive labeling
#
# For examples and usage patterns, see:
#  - examples.md for Terraform examples
#  - ../README.md for component documentation
#  - ../../docs/README.md for research and best practices
##############################################

##############################################
# 1. Create or Reference Namespace
#
# Conditionally create namespace based on create_namespace flag:
#  - If true: Create new namespace with labels
#  - If false: Reference existing namespace
#
# Each Job gets its own namespace for:
#  - Resource isolation
#  - Security boundary
#  - Easier cleanup and management
##############################################
resource "kubernetes_namespace" "this" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.final_labels
  }
}

# Data source for existing namespace (when not creating)
data "kubernetes_namespace" "existing" {
  count = var.spec.create_namespace ? 0 : 1

  metadata {
    name = local.namespace
  }
}

##############################################
# Additional Resources
#
# Other resources are defined in dedicated files:
#  - job.tf: ServiceAccount, ImagePullSecret, and Job
#  - secret.tf: Environment secrets (if configured)
#  - configmap.tf: ConfigMaps from spec.config_maps
#  - outputs.tf: Module outputs
#  - locals.tf: Local variables and computed values
#  - variables.tf: Input variable definitions
##############################################
