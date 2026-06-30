##############################################
# main.tf
#
# Main orchestration file for KafkaKubernetes
# deployment using Terraform.
#
# This module creates a production-ready Apache Kafka
# cluster on Kubernetes using the Strimzi Operator with
# the following capabilities:
#
# Infrastructure Components:
#  1. Kubernetes Namespace (defined here)
#  2. Kafka Cluster (kafka_cluster.tf)
#     - Kafka brokers with configurable replicas
#     - ZooKeeper ensemble for cluster coordination
#     - Entity Operator for topic/user management
#  3. Admin User (kafka_admin_user.tf)
#     - SCRAM-SHA-512 authentication
#     - Admin credentials secret
#  4. Kafka Topics (kafka_topics.tf)
#     - Topic creation and configuration
#     - Partition and replication settings
#  5. Schema Registry (schema_registry.tf)
#     - Optional Confluent Schema Registry deployment
#     - Schema versioning and compatibility
#  6. Kafka UI (kowl.tf)
#     - Optional Kowl/Kafka UI for cluster management
#     - Web-based topic browsing and management
#
# Production Features:
#  - High availability with multiple broker replicas
#  - Persistent storage for Kafka and ZooKeeper
#  - SCRAM-SHA-512 authentication for security
#  - Topic-level configuration and ACLs
#  - Optional ingress for external access
#  - Schema Registry for data governance
#  - Kafka UI for operational visibility
#
# Module Structure:
#  - main.tf: Namespace creation and documentation (this file)
#  - kafka_cluster.tf: Strimzi Kafka custom resource
#  - kafka_admin_user.tf: Admin user creation
#  - kafka_topics.tf: Topic definitions
#  - schema_registry.tf: Schema Registry deployment
#  - kowl.tf: Kafka UI deployment
#  - locals.tf: Computed values and label management
#  - variables.tf: Input variable definitions
#  - outputs.tf: Module outputs (endpoints, credentials)
#
# Design Philosophy:
# This module follows the Strimzi Operator approach:
#  - Declarative Kafka cluster management via CRDs
#  - Operator handles complex lifecycle operations
#  - Separation of concerns with modular files
#  - Production-ready defaults with customization options
#  - Support for both internal and external access patterns
#
# Deployment Flow:
# 1. Namespace is created first (this file)
# 2. Kafka cluster (brokers + ZooKeeper) is deployed
# 3. Admin user is created with credentials
# 4. Topics are created based on spec
# 5. Optional Schema Registry is deployed
# 6. Optional Kafka UI is deployed
#
# Dependencies:
# - Strimzi Kafka Operator must be pre-installed in the cluster
# - Kubernetes cluster with sufficient resources
# - Storage class for persistent volumes
#
# For detailed examples and usage patterns, see:
#  - examples.md: Terraform configuration examples
#  - README.md: Module documentation
#  - ../docs/README.md: Comprehensive deployment guide
##############################################

##############################################
# 1. Create Namespace
#
# Kafka and related components run in a dedicated
# namespace for isolation and resource management.
# The namespace is only created if create_namespace is true.
##############################################
resource "kubernetes_namespace_v1" "kafka_namespace" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.final_labels
  }
}

##############################################
# Note: Additional resources are defined in modular files:
#
# - kafka_cluster.tf: Core Kafka cluster (brokers, ZooKeeper, entity operator)
# - kafka_admin_user.tf: Administrative user with SCRAM-SHA-512 credentials
# - kafka_topics.tf: Topic creation with partition/replication configuration
# - schema_registry.tf: Optional Schema Registry for Avro/JSON schema management
# - kowl.tf: Optional Kafka UI (Kowl) for cluster visualization and management
#
# This modular approach:
# - Improves maintainability by separating concerns
# - Makes it easier to understand each component
# - Allows for selective customization
# - Follows Terraform best practices for large modules
##############################################
