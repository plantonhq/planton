#########################################################################################################
# Apache Solr on Kubernetes - Main Terraform Configuration
#########################################################################################################
#
# This module deploys Apache Solr on Kubernetes using the Solr Operator pattern.
# It creates a production-ready SolrCloud cluster with integrated Zookeeper ensemble.
#
# Architecture:
#   - SolrCloud: Scalable Solr cluster deployed as a StatefulSet
#   - Zookeeper: Coordination service for Solr nodes (deployed via Solr Operator)
#   - Ingress: Optional external access via Kubernetes Gateway API (Istio)
#   - Namespace: Dedicated namespace for isolation
#
# Key Features:
#   - Horizontal scaling via replica configuration
#   - Persistent storage for both Solr and Zookeeper
#   - Customizable JVM settings and garbage collection tuning
#   - TLS-enabled ingress with cert-manager integration
#   - Resource limits and requests for production stability
#
# File Organization:
#   - main.tf: Core namespace and orchestration (this file)
#   - solr_cloud.tf: SolrCloud custom resource definition
#   - ingress.tf: Gateway API resources (Certificate, Gateway, HTTPRoute)
#   - locals.tf: Computed values and naming conventions
#   - variables.tf: Input variables from spec.proto
#   - outputs.tf: Stack outputs for consumers
#   - provider.tf: Kubernetes provider configuration
#
# Dependencies:
#   - Solr Operator: Must be pre-installed in the cluster
#   - cert-manager: Required for TLS certificate generation (if ingress enabled)
#   - Istio: Required for Gateway API support (if ingress enabled)
#
# Usage:
#   terraform init
#   terraform plan -var-file="hack/manifest.yaml"
#   terraform apply -var-file="hack/manifest.yaml"
#
#########################################################################################################

# Create dedicated namespace for Solr deployment
# This isolates Solr resources and provides a security boundary
# The namespace is conditionally created based on the create_namespace flag
resource "kubernetes_namespace" "solr_namespace" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.final_labels
  }
}

# Note: The SolrCloud resource is defined in solr_cloud.tf
# It depends on this namespace and creates:
#   - SolrCloud StatefulSet
#   - Zookeeper StatefulSet (via operator)
#   - Associated Services, ConfigMaps, and PVCs

# Note: Ingress resources are defined in ingress.tf
# They are conditionally created when spec.ingress.is_enabled = true
# and include:
#   - TLS Certificate (cert-manager)
#   - Gateway (Istio Gateway API)
#   - HTTPRoute for HTTP->HTTPS redirect
#   - HTTPRoute for HTTPS traffic routing
