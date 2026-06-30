# Confluent Cloud Kafka Deployment Guide

## Introduction: The Shift from Self-Managed to Platform-as-a-Service

For years, the conventional wisdom in the Apache Kafka community was clear: if you want to run Kafka in production, you need a dedicated team of distributed systems experts. The reality of managing Kafka clusters—handling broker failures, optimizing partition rebalancing, ensuring zero-downtime upgrades, and maintaining the surrounding ecosystem of Schema Registry and Kafka Connect—was considered a rite of passage for serious engineering organizations.

That paradigm has fundamentally shifted. Confluent Cloud represents not just "managed Kafka," but a complete re-architecture of Apache Kafka as a cloud-native, serverless data streaming platform. The strategic question is no longer "Can we afford to use a managed service?" but rather "Can we afford not to?"

This document explains the deployment landscape for Apache Kafka, compares self-managed versus fully-managed approaches, and details why Planton provides first-class support for Confluent Cloud as the production-ready default for teams building event-driven architectures.

## The Kafka Deployment Maturity Spectrum

### Level 0: The Anti-Pattern – Manual Kafka on VMs

Running Kafka manually on virtual machines or bare metal represents the foundational approach that most organizations have abandoned for production systems. While it provides absolute control over every configuration parameter, this approach requires:

- **Deep operational expertise**: Teams must handle cluster provisioning, ZooKeeper management (or KRaft consensus), rolling upgrades, and failure recovery manually
- **Significant operational overhead**: 24/7 on-call rotations for a critical piece of infrastructure that never sleeps
- **Complex ecosystem management**: Separate deployment and management of Schema Registry, Kafka Connect, and stream processing infrastructure

**Verdict**: This approach is viable only for organizations with dedicated platform teams and specific requirements that mandate on-premises deployment. For 95% of use cases, the operational burden far outweighs the benefits of low-level control.

### Level 1: Kubernetes Operators – Better, But Still Self-Managed

Deploying Kafka on Kubernetes using operators like Strimzi or the Confluent for Kubernetes operator represents a significant maturity upgrade. These operators provide:

- **Declarative configuration**: Define Kafka clusters as Kubernetes custom resources
- **Automated operations**: Handling rolling upgrades, pod failures, and storage management
- **Integration with cloud-native tooling**: Monitoring, logging, and GitOps workflows

However, this approach still leaves the organization responsible for:

- **Kubernetes cluster management**: Operating the underlying orchestration platform
- **Capacity planning**: Sizing brokers, managing storage, and scaling clusters
- **Network and security**: Implementing private networking, certificate management, and access control
- **The full ecosystem**: Deploying and operating Schema Registry, Connect, and ksqlDB as separate workloads

**Verdict**: A solid middle ground for organizations that have already invested heavily in Kubernetes expertise and infrastructure, but want Kafka-specific automation. However, it's still fundamentally a "build" rather than "buy" approach.

### Level 2: Cloud-Provider Managed Kafka – The Hyperscaler Option

Major cloud providers offer their own managed Kafka services:

- **Amazon MSK** (Managed Streaming for Apache Kafka) on AWS
- **Azure Event Hubs for Kafka** on Microsoft Azure  
- **Managed Service for Apache Kafka** on Google Cloud (in preview)

These services provide value through:

- **Simplified infrastructure**: The cloud provider handles broker provisioning and basic operations
- **Native cloud integration**: Direct integration with VPCs, IAM, and cloud monitoring services
- **Competitive pricing**: Often lower base costs than specialized SaaS platforms

The limitations include:

- **Limited ecosystem**: You get managed Kafka brokers, but Schema Registry, Kafka Connect, and stream processing remain your responsibility
- **Feature lag**: These services typically run older Kafka versions and lack advanced features available in Confluent Cloud
- **Basic operations only**: High-level operational concerns like multi-region replication, disaster recovery, and advanced security features require custom solutions

**Verdict**: A reasonable choice for teams deeply embedded in a single cloud provider's ecosystem who only need basic Kafka functionality and are willing to build the surrounding platform themselves.

### Level 3: The Production Solution – Confluent Cloud

Confluent Cloud represents the culmination of this maturity spectrum: a fully-managed, cloud-native, multi-cloud data streaming platform. It provides:

**1. Complete Platform Integration**
- **Managed Kafka clusters** with automatic scaling, patching, and zero-downtime upgrades
- **Managed Schema Registry** for data governance and evolution, deployed per-environment
- **Managed Kafka Connect** with 80+ pre-built, fully-managed connectors
- **Managed ksqlDB** for SQL-based stream processing
- **Cluster Linking** for multi-region and multi-cloud data replication

**2. True Multi-Cloud Architecture**
- Deployable across 100+ regions on AWS, GCP, and Azure
- Unified API and tooling regardless of underlying cloud provider
- Built-in support for multi-cloud data replication via Cluster Linking

**3. Enterprise-Grade Reliability**
- **99.99% uptime SLA** for multi-zone clusters (financially backed)
- Automatic fault tolerance with synchronous 3x replication across availability zones
- **Zero RPO** (Recovery Point Objective) for zone failures

**4. Elastic, Serverless Economics**
- Consumption-based billing with elastic capacity units (eCKUs) for Standard and Enterprise clusters
- No manual capacity planning or overprovisioning for variable workloads
- Provisioned CKU model remains available for predictable, high-throughput workloads (Dedicated clusters)

**5. Advanced Security Without Complexity**
- Native support for **AWS PrivateLink**, **Azure Private Link**, and **GCP Private Service Connect**
- Dual authorization model (RBAC for platform control, ACLs for data access)
- Encryption at rest and in transit (TLS 1.2+) by default

**Verdict**: The production-ready choice for organizations prioritizing development velocity, operational excellence, and a complete streaming platform. The higher per-GB cost compared to hyperscaler alternatives is offset by dramatically reduced operational overhead and engineering time-to-market.

## When to Choose Confluent Cloud vs. Self-Managed Kafka

The decision framework boils down to a clear set of trade-offs:

### Choose Confluent Cloud When:

1. **Application development velocity is the strategic priority**: Teams can focus on building data pipelines and applications rather than managing infrastructure
2. **You need a complete streaming platform**: Not just Kafka, but Schema Registry, Connect, and stream processing as a unified, managed service
3. **Enterprise SLAs are non-negotiable**: The 99.99% uptime guarantee is difficult and expensive to achieve with self-managed infrastructure
4. **Total Cost of Ownership (TCO) favors managed services**: When accounting for the fully-loaded cost of hiring, training, and retaining specialized Kafka operations teams
5. **Multi-cloud or hybrid-cloud is part of your strategy**: Confluent's Cluster Linking provides native multi-cloud replication capabilities

### Choose Self-Managed Kafka When:

1. **Absolute infrastructure control is mandatory**: Specific hardware, network topology, or kernel-level tuning requirements that preclude a SaaS model
2. **Extreme latency sensitivity**: Sub-millisecond latency requirements that benefit from co-location of applications and brokers on the same physical infrastructure
3. **Data sovereignty requires on-premises deployment**: Regulatory or compliance requirements mandate that data never leaves a private data center
4. **The workload is truly small-scale**: For toy projects or very limited use cases where the full Confluent Cloud platform would be overkill

For the vast majority of production workloads, the operational complexity and hidden costs of self-managed Kafka make Confluent Cloud the economically and strategically superior choice.

## Understanding Confluent Cloud's Architecture

### The Resource Hierarchy: Organizations, Environments, and Clusters

Confluent Cloud is built on a strict logical hierarchy that governs resource isolation, billing, and security:

1. **Organization**: The root entity representing your Confluent Cloud subscription. Contains user accounts, service accounts, billing details, and all environments.

2. **Environment**: The primary boundary for resource isolation. Environments are used to separate application lifecycles (dev, staging, production) or organizational units (team A, team B). Critically, **Schema Registry is enabled per-environment**, ensuring governance boundaries align with resource boundaries.

3. **Resources**: The functional components deployed within an Environment:
   - Kafka Clusters
   - ksqlDB Clusters
   - Managed Connectors
   - Networks (for private connectivity)

This hierarchy isn't just organizational—it's fundamental to security and access control. A Kafka cluster is always a child of an Environment, and permissions can be scoped at the environment level.

### Kafka Cluster Types: A Critical Fork in the Road

The most important architectural decision you make is selecting the cluster type. This choice dictates cost, performance, tenancy, networking capabilities, and operational model.

| Cluster Type | Use Case | Billing Model | Tenancy | Availability | 99.99% SLA | Private Networking |
|:-------------|:---------|:--------------|:--------|:-------------|:-----------|:-------------------|
| **Basic** | Development, Testing | Elastic (eCKU) | Multi-Tenant | Single-Zone Only | ❌ | ❌ (Public Internet Only) |
| **Standard** | Production (General) | Elastic (eCKU) | Multi-Tenant | Single or Multi-Zone | ✅ (Multi-Zone Only) | ❌ (Public Internet Only) |
| **Enterprise** | Production (Secure) | Elastic (eCKU) | Multi-Tenant | Multi-Zone | ✅ | ✅ (PrivateLink/VNet Peering) |
| **Dedicated** | Production (Critical) | Provisioned (CKU) | **Single-Tenant** | Single or Multi-Zone | ✅ (Multi-Zone Only) | ✅ (PrivateLink/VNet Peering) |

**Key Insights:**

- **Basic** and **Standard** are public-internet-only clusters. They're simpler and cheaper, but cannot be integrated into a private network.
- **Enterprise** is the game-changer: it combines the elasticity and consumption-based pricing of Standard with the private networking capabilities previously reserved for Dedicated clusters.
- **Dedicated** remains the choice for single-tenant isolation and provisioned capacity (CKU), ideal for workloads with predictable, high throughput requirements.

### Multi-Zone High Availability

High availability in Confluent Cloud is achieved through **multi-zone deployment**:

- **Single-Zone**: The cluster runs within a single availability zone. Suitable for development, but ineligible for enterprise SLAs.
- **Multi-Zone**: The cluster spans **three** separate availability zones within a region, with synchronous 3x replication. This provides:
  - Automatic fault tolerance for entire zone failures
  - **Zero RPO** (Recovery Point Objective)
  - Eligibility for the 99.99% uptime SLA

**Critical**: This setting is **immutable after cluster creation**. You cannot change a single-zone cluster to multi-zone later. This decision must be made correctly on day zero.

## Deployment Methods: From Manual to Infrastructure-as-Code

### Manual Provisioning: The Confluent Cloud Console

The Confluent Cloud web console provides a guided, visual workflow for creating clusters, managing connectors, and monitoring metrics. While excellent for learning and ad-hoc exploration, it lacks the repeatability, auditability, and version control required for production deployments.

**Use for**: Initial platform exploration, one-off debugging, and operations monitoring.

**Avoid for**: Production provisioning, multi-environment deployments, and team collaboration.

### Imperative Automation: The Confluent CLI

The `confluent` CLI is the official command-line tool for both Confluent Cloud and Confluent Platform. It provides comprehensive coverage of platform operations:

```bash
# Create a Kafka cluster
confluent kafka cluster create my-cluster \
  --cloud aws \
  --region us-east-2 \
  --type standard

# Create a service account
confluent iam service-account create sa-app-prod \
  --description "Production app service account"

# Generate an API key
confluent api-key create --resource lkc-abc123 \
  --service-account sa-12345
```

The CLI is ideal for:
- Ad-hoc administrative tasks
- Simple shell-based automation
- CI/CD pipeline integrations for imperative actions

However, it's fundamentally imperative, not declarative. For production infrastructure management, declarative Infrastructure-as-Code is the industry standard.

### Declarative Automation: Infrastructure-as-Code with Terraform

**Terraform** with the official `confluentinc/confluent` provider is the first-class, production-grade solution for managing Confluent Cloud infrastructure as code.

The provider offers comprehensive resource coverage:
- `confluent_environment`: Logical resource containers
- `confluent_service_account`: Identity for applications and automation
- `confluent_api_key`: Credentials for authentication
- `confluent_kafka_cluster`: The core Kafka cluster resource
- `confluent_kafka_topic`: Topic configuration
- `confluent_kafka_acl`: Data-plane access control lists
- `confluent_role_binding`: Platform-level RBAC
- `confluent_network`: Private networking configuration
- `confluent_private_link_access`: PrivateLink/Private Link configuration
- `confluent_connector`: Managed Kafka Connect connectors
- `confluent_ksql_cluster`: Managed ksqlDB clusters

**Example: Basic Cluster**

```hcl
resource "confluent_kafka_cluster" "dev" {
  display_name = "dev-cluster"
  availability = "SINGLE_ZONE"
  cloud        = "AWS"
  region       = "us-east-2"
  
  basic {}
  
  environment {
    id = confluent_environment.dev.id
  }
}
```

**Example: Production Dedicated Cluster with PrivateLink**

```hcl
resource "confluent_kafka_cluster" "prod" {
  display_name = "prod-orders-cluster"
  availability = "MULTI_ZONE"
  cloud        = "GCP"
  region       = "us-central1"
  
  dedicated {
    cku = 2
  }
  
  environment {
    id = confluent_environment.prod.id
  }
}

resource "confluent_network" "prod_private" {
  display_name     = "prod-private-network"
  cloud            = "GCP"
  region           = "us-central1"
  connection_types = ["PRIVATELINK"]
  
  environment {
    id = confluent_environment.prod.id
  }
}
```

**Best Practices:**

1. **Separate environments by directory**: Use distinct Terraform state files for dev, staging, and prod to prevent cross-environment accidents
2. **Never commit secrets**: Pass provider credentials via environment variables (`CONFLUENT_CLOUD_API_KEY`, `CONFLUENT_CLOUD_API_SECRET`)
3. **Use remote state**: Store Terraform state in S3, GCS, or Terraform Cloud, never in version control
4. **Manage secrets in a vault**: Store generated API keys in HashiCorp Vault or cloud-native secret managers, not in Terraform state

### Declarative Automation: Infrastructure-as-Code with Pulumi

**Pulumi** provides a native provider package `@pulumi/confluentcloud` that is **bridged from the official Confluent Terraform provider**. This means:
- **100% feature parity** with Terraform (same resources, same capabilities)
- **Different developer experience**: Write infrastructure code in Python, TypeScript, Go, or C# instead of HCL
- **Integrated secret management**: Pulumi automatically encrypts sensitive outputs in its state file

**Example: Standard Production Cluster (TypeScript)**

```typescript
import * as confluentcloud from "@pulumi/confluentcloud";

const prod = new confluentcloud.KafkaCluster("prod-orders", {
    displayName: "prod-orders-cluster",
    availability: "MULTI_ZONE",
    cloud: "AWS",
    region: "us-west-2",
    standard: {},
    environment: {
        id: prodEnv.id,
    },
});

export const bootstrapEndpoint = prod.bootstrapEndpoint;
```

**Choice Criteria**: Terraform vs. Pulumi is a question of team preference and existing infrastructure patterns, not capability. Both are production-ready and officially supported.

## Networking and Security for Production Workloads

### Private Networking: From Public Internet to PrivateLink

For enterprise deployments, moving from public internet access to private networking is a fundamental security requirement.

**Networking Options by Cluster Type:**

| Cluster Type | Public Internet | VPC/VNet Peering | PrivateLink/Private Link/Private Service Connect |
|:-------------|:----------------|:-----------------|:--------------------------------------------------|
| Basic | ✅ | ❌ | ❌ |
| Standard | ✅ | ❌ | ❌ |
| Enterprise | ✅ | ✅ | ✅ |
| Dedicated | ✅ | ✅ | ✅ |

**1. Public Internet (Default)**
- All traffic encrypted with TLS 1.2+
- Accessible from anywhere with credentials
- Suitable for development and non-sensitive workloads

**2. VPC/VNet Peering**
- Direct private network connection between your cloud VPC and Confluent's VPC
- Traffic flows over the cloud provider's backbone, not the public internet
- Requires careful CIDR management to avoid IP address conflicts

**3. PrivateLink / Private Link / Private Service Connect** (Recommended)
- **AWS PrivateLink**, **Azure Private Link**, **GCP Private Service Connect**
- Creates a private endpoint in your VPC that routes securely to Confluent Cloud
- Unidirectional connection prevents data exfiltration
- No CIDR management complexity
- **This is the preferred method for enterprise security**

### Security: Authentication and Authorization

Confluent Cloud employs a sophisticated two-level security model:

#### Authentication: Service Accounts and API Keys

1. **Service Accounts**: These are the **identities** for non-human access (applications, automation, CI/CD pipelines). Each service account represents a distinct principal (e.g., `sa-payment-processor-prod`).

2. **API Keys**: These are the **credentials** (a key and secret pair) owned by a user or service account.

**Two types of API keys exist:**

- **Cloud API Keys**: Grant access to the Confluent Cloud Management APIs (control plane). Used for provisioning resources like clusters, environments, and networks.
- **Resource API Keys**: Grant access to a specific resource's data plane (e.g., Kafka API keys for producing/consuming messages, Schema Registry API keys for managing schemas).

**Best Practice**: Create dedicated service accounts for each application or logical service, generate resource-scoped API keys, and rotate them regularly.

#### Authorization: The Dual RBAC and ACL Model

Confluent Cloud uses **two distinct authorization systems** that are often confused:

**1. RBAC (Role-Based Access Control)**
- **Scope**: The Confluent Cloud Platform (control plane)
- **Question answered**: "Who can create, delete, or manage platform resources?"
- **Example**: Grant the `EnvironmentAdmin` role to `User:alice` on `Environment:prod`
- **IaC Resource**: `confluent_role_binding` (Terraform), `RoleBinding` (Pulumi)

**2. ACLs (Access Control Lists)**
- **Scope**: Inside a Kafka Cluster (data plane)
- **Question answered**: "What can an authenticated principal do with Kafka data?"
- **Example**: Allow `ServiceAccount:sa-app` to `WRITE` to `Topic:orders`
- **IaC Resource**: `confluent_kafka_acl` (Terraform), `KafkaAcl` (Pulumi)

A common pattern requires both: Use RBAC to grant operators permission to manage a cluster, then use ACLs to grant applications permission to produce/consume data.

## Stream Governance and the Broader Ecosystem

### Schema Registry: The Governance Foundation

For any organization where multiple services produce or consume events, Schema Registry is non-negotiable. It provides:

- **Schema validation**: Ensures data written to Kafka conforms to a defined schema
- **Schema evolution**: Manages backward/forward compatibility as schemas change over time
- **Data governance**: Centralized registry for all data contracts

**Key Architectural Points:**
- Schema Registry is **enabled per-environment**, not per-cluster
- It requires a **separate Resource API Key** distinct from Kafka API keys
- Available in "Essentials" and "Advanced" governance packages

### Managed Kafka Connect: Pre-Built Integrations

Confluent Cloud offers **80+ fully managed connectors** for integrating Kafka with databases, cloud storage, SaaS applications, and data warehouses. Unlike self-managed Kafka Connect, you don't provision a Connect cluster—you simply deploy individual connectors as managed resources.

**Example connectors:**
- **Source**: PostgreSQL CDC, MongoDB, Salesforce, Snowflake, S3
- **Sink**: Elasticsearch, BigQuery, Snowflake, S3, JDBC

**IaC Pattern**: Use `confluent_connector` (Terraform) with separate `config_nonsensitive` and `config_sensitive` fields to cleanly handle secrets.

### Managed ksqlDB: SQL for Stream Processing

ksqlDB allows teams to perform stateful stream processing using familiar SQL syntax, without writing Java code. In Confluent Cloud, ksqlDB is deployed as a managed cluster resource provisioned with Confluent Streaming Units (CSUs).

**Use cases:**
- Real-time aggregations and windowing
- Stream-table joins
- Continuous queries for materialized views

## Migration and Operational Excellence

### Migrating from Self-Managed Kafka

For organizations moving from self-managed Kafka to Confluent Cloud, **Cluster Linking** is the primary technology for zero-downtime migration:

**Migration Pattern:**
1. **Link**: Establish a cluster link from the self-managed cluster to the new Confluent Cloud cluster. Data replicates in real-time.
2. **Migrate Consumers**: Reconfigure consumer applications to read from the Confluent Cloud cluster.
3. **Migrate Producers**: Once consumers are stable, reconfigure producers to write to Confluent Cloud.
4. **Cut Over**: Decommission the cluster link and retire the old cluster.

This approach enables live, gradual migration without downtime or data loss.

### Disaster Recovery: Beyond High Availability

**High Availability (HA)** and **Disaster Recovery (DR)** are distinct concepts:

- **HA** (intra-region): Automatic resilience to zone failures via multi-zone clusters. Confluent Cloud provides this out-of-the-box with zero RPO for zone failures.
- **DR** (inter-region): Protection against full regional outages. **This must be architected by the user.**

**DR Patterns with Cluster Linking:**
1. **Active/Passive**: Primary cluster in Region A replicates to standby cluster in Region B. Failover during disaster.
2. **Active/Active**: Multiple clusters in different regions serve live traffic with bi-directional replication.

## Planton's Approach: Simplified, Secure, Production-Ready

Planton provides first-class support for Confluent Cloud through the `ConfluentKafka` resource (`confluent.planton.dev/v1`). The API is designed according to the **80/20 principle**: expose the 20% of configuration that 80% of users actually need.

### The 80/20 Configuration Philosophy

Based on analysis of real-world Terraform and Pulumi deployments, most users configure only these essential fields:

**Essential (Required):**
- **Environment**: The parent container for the cluster (via `metadata.org` in Planton)
- **Cloud Provider**: AWS, GCP, or Azure
- **Region**: Cloud-specific region (e.g., `us-east-2`)
- **Availability**: `SINGLE_ZONE` or `MULTI_ZONE`
- **Cluster Type**: Basic, Standard, Enterprise, or Dedicated

**Advanced (Optional):**
- **Network Configuration**: For private networking (VPC Peering, PrivateLink)
- **Dedicated CKU**: Provisioned capacity for Dedicated clusters

### Current API Design

The `ConfluentKafkaSpec` API focuses on the most critical decisions:

```protobuf
message ConfluentKafkaSpec {
  // Cloud provider: AWS, AZURE, or GCP
  string cloud = 1;
  
  // Availability: SINGLE_ZONE (dev), MULTI_ZONE (prod)
  // LOW and HIGH are legacy values maintained for compatibility
  string availability = 2;
  
  // Environment ID: The Confluent Cloud environment parent
  string environment = 3;
}
```

**Simplified Abstraction:**
- The **cluster type** (Basic, Standard, Enterprise, Dedicated) is inferred or defaulted based on the environment and availability
- The **region** is often derived from cloud resource metadata or defaults
- **Display name** comes from `metadata.name`

This design optimizes for the common case: teams want "a production Kafka cluster in AWS us-east-2 in our prod environment" without needing to specify every parameter.

### What Planton Handles for You

1. **Environment Management**: Integrates with Confluent Cloud's environment hierarchy
2. **Credential Management**: Securely handles Cloud API Keys and Resource API Keys
3. **IaC Abstraction**: Generates Pulumi/Terraform under the hood from the protobuf spec
4. **Output Management**: Exposes critical outputs (bootstrap endpoint, cluster ID, REST endpoint, CRN) via `ConfluentKafkaStackOutputs`

### Recommended Usage Patterns

**Development Cluster:**

```yaml
apiVersion: confluent.planton.dev/v1
kind: ConfluentKafka
metadata:
  name: dev-kafka
spec:
  cloud: AWS
  availability: SINGLE_ZONE
  environment: env-dev-abcde
```

**Production Cluster:**

```yaml
apiVersion: confluent.planton.dev/v1
kind: ConfluentKafka
metadata:
  name: prod-orders-kafka
spec:
  cloud: GCP
  availability: MULTI_ZONE  # Required for 99.99% SLA
  environment: env-prod-xyz
```

For advanced configurations (private networking, Dedicated clusters with specific CKU counts), the API can be extended or users can leverage the underlying IaC modules directly.

## Conclusion: Platform Thinking Over Infrastructure Thinking

The strategic insight behind Confluent Cloud is a shift from infrastructure thinking to platform thinking. Teams no longer ask "How do I deploy Kafka brokers?" but rather "How do I build real-time data pipelines?"

This shift is why Planton provides Confluent Cloud as a first-class resource. By abstracting the operational complexity of Kafka infrastructure, development teams can focus on the 20% of decisions that matter (cloud, region, availability, environment) and leave the other 80% to the platform.

For teams building event-driven architectures, microservices, or data streaming applications, Confluent Cloud combined with Planton's declarative API provides the fastest path from idea to production-grade infrastructure.

**Next Steps:**
- Explore [Planton's ConfluentKafka examples](../examples.md)
- Review the [Pulumi implementation](../iac/pulumi/README.md) for advanced customization
- Learn about integrating Schema Registry and Kafka Connect in your deployment

