##############################################
# main.tf
#
# Main orchestration file for KubernetesDeployment
# deployment using Terraform.
#
# This module creates a production-ready Kubernetes
# deployment with the following capabilities:
#
# Infrastructure Components:
#  1. Kubernetes Namespace (defined here)
#  2. Service Account with RBAC (deployment.tf)
#  3. Kubernetes Deployment with:
#     - Main application container
#     - Optional sidecar containers
#     - Health probes (startup, liveness, readiness)
#     - Resource requests and limits
#     - Environment variables and secrets
#  4. Kubernetes Service for internal networking (service.tf)
#  5. Environment Secrets (if configured) (secret.tf)
#  6. Image Pull Secret for private registries (deployment.tf)
#  7. Istio Gateway & VirtualService for ingress (ingress.tf)
#
# Production Features:
#  - Zero-downtime rolling updates
#  - Horizontal Pod Autoscaling (HPA)
#  - Pod Disruption Budgets (PDB)
#  - Graceful shutdown with preStop hooks
#  - Comprehensive health monitoring
#  - QoS-aware resource management
#  - TLS certificate automation
#
# Module Structure:
#  - main.tf: Namespace creation and documentation (this file)
#  - deployment.tf: Deployment, ServiceAccount, ImagePullSecret
#  - service.tf: Kubernetes Service for pod networking
#  - ingress.tf: Istio Gateway and VirtualService
#  - secret.tf: Application secrets management
#  - locals.tf: Computed values and label management
#  - variables.tf: Input variable definitions
#  - outputs.tf: Module outputs (FQDNs, service names, etc.)
#
# Design Philosophy:
# This module follows Kubernetes best practices by:
#  - Isolating each microservice in its own namespace
#  - Using semantic versioning for deployments (spec.version)
#  - Implementing defense-in-depth with multiple layers
#  - Providing production-ready defaults
#  - Supporting progressive delivery patterns
#
# For detailed examples and usage patterns, see:
#  - examples.md: Terraform configuration examples
#  - ../README.md: Component overview and features
#  - ../../docs/README.md: Research and production best practices
#  - ../pulumi/examples.md: Alternative Pulumi examples
#
# Zero-Downtime Deployment Strategy:
# The module implements zero-downtime deployments through:
#  1. Rolling update with maxUnavailable: 0
#  2. Readiness probes prevent traffic to non-ready pods
#  3. PreStop hooks delay SIGTERM for connection draining
#  4. Pod Disruption Budgets prevent cascading failures
#  5. Graceful shutdown with 60s termination grace period
#
# Scaling Strategy:
# Horizontal scaling is managed through:
#  - spec.availability.min_replicas: Baseline pod count
#  - HPA metrics: CPU and memory utilization targets
#  - Automatic scale-up/down based on observed metrics
#  - Pod anti-affinity for distribution (when configured)
#
# Security Posture:
#  - Dedicated service account per deployment
#  - Secret management via Kubernetes Secrets
#  - Private registry support with imagePullSecrets
#  - Network policies (when cluster configured)
#  - Pod Security Standards compliance
#
# Observability:
#  - Labels for filtering and grouping
#  - Resource tagging for cost allocation
#  - Structured logging through env vars
#  - Metrics exposure via service ports
#  - Health probe endpoints for monitoring
##############################################

##############################################
# Namespace Resource
#
# Creates a dedicated Kubernetes namespace for the
# microservice deployment. Each deployment gets its
# own namespace for:
#
# Benefits:
#  - Resource isolation and quota management
#  - RBAC boundary and security isolation
#  - Network policy enforcement
#  - Easier resource cleanup and lifecycle management
#  - Multi-tenancy support
#  - Environment separation
#
# Naming Convention:
#  - Namespace name is derived from metadata.id or metadata.name
#  - This ensures uniqueness and traceability
#  - Format: <resource_id> (e.g., "todo-api-prod")
#
# Labels:
#  - resource: "true" (marks as Planton-managed)
#  - resource_id: Unique identifier for this resource
#  - resource_kind: "microservice_kubernetes"
#  - organization: Organization name (if specified)
#  - environment: Environment name (dev/staging/prod)
#
# The namespace serves as the foundation for all other
# resources in this deployment. All resources (Deployment,
# Service, Ingress, Secrets) are created within this namespace.
##############################################
resource "kubernetes_namespace" "this" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.final_labels
  }
}

##############################################
# Additional Resources
#
# Other resources are organized in dedicated files
# for better maintainability and separation of concerns:
#
# Core Workload:
#  - deployment.tf: Kubernetes Deployment with app containers
#                   ServiceAccount for pod identity
#                   ImagePullSecret for private registries
#
# Networking:
#  - service.tf: Kubernetes Service for pod discovery
#                ClusterIP service for internal routing
#                Port mappings from container to service
#
#  - ingress.tf: Istio Gateway for external traffic
#                VirtualService for routing rules
#                TLS Certificate automation
#                HTTPRoute configuration
#
# Configuration:
#  - secret.tf: Kubernetes Secret for sensitive data
#               Environment variable secrets
#               Secret name matches spec.version
#
# Module Metadata:
#  - locals.tf: Local variables and computed values
#               Label construction and merging
#               FQDN calculation for services
#               Ingress hostname derivation
#
#  - variables.tf: Input variable definitions
#                  Type constraints and validation
#                  Documentation for each variable
#
#  - outputs.tf: Module outputs for consumers
#                Service FQDNs and endpoints
#                Namespace and resource names
#                Ingress hostnames
#                Port-forward commands
#
# This modular structure provides:
#  - Clear separation of concerns
#  - Easier code navigation and maintenance
#  - Reusable patterns across deployments
#  - Simplified testing and validation
##############################################
