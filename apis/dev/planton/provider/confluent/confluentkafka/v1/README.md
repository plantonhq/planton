# ConfluentKafka

Confluent Kafka cluster resource for deploying managed Apache Kafka clusters on Confluent Cloud. Supports deployment across AWS, GCP, and Azure with multiple cluster types (BASIC, STANDARD, ENTERPRISE, DEDICATED) and private networking options.

## Spec Fields (80/20)

### Essential Fields (80% Use Case)
- **cloud**: Cloud provider where the cluster will be deployed - AWS, GCP, or AZURE.
- **region**: Cloud-specific region for deployment (e.g., us-east-2, us-central1, eastus).
- **availability**: High availability configuration:
  - `SINGLE_ZONE`: Development/testing, no SLA
  - `MULTI_ZONE`: Production, 99.99% SLA (required for Standard/Dedicated)
  - `LOW`/`HIGH`: Legacy values for Basic clusters (backward compatibility)
- **environment_id**: The ID of the Confluent Cloud environment (parent container for clusters).
- **cluster_type**: Deployment type and capabilities:
  - `BASIC`: Multi-tenant, development/testing, single-zone only, public internet only
  - `STANDARD`: Multi-tenant, production, elastic scaling, public internet only (default)
  - `ENTERPRISE`: Multi-tenant, production, elastic scaling, supports private networking
  - `DEDICATED`: Single-tenant, production, provisioned capacity (CKU), supports private networking

### Advanced Fields (20% Use Case)
- **dedicated_config**: Required when cluster_type is DEDICATED. Configures provisioned capacity:
  - `cku`: Confluent Kafka Units (minimum 1, can be scaled up/down)
- **network_config**: Configures private networking (PrivateLink, VNet Peering, Private Service Connect):
  - `network_id`: ID of the pre-created Confluent Cloud network resource
  - Only available for ENTERPRISE and DEDICATED cluster types
- **display_name**: Human-readable name shown in Confluent Cloud UI (defaults to metadata.name if not specified)

## Stack Outputs

- **id**: The provider-assigned unique ID for the Confluent Kafka cluster.
- **bootstrap_endpoint**: Bootstrap server endpoint used by Kafka clients to connect to the cluster (e.g., SASL_SSL://pkc-00000.us-central1.gcp.confluent.cloud:9092).
- **crn**: The Confluent Resource Name of the Kafka cluster for RBAC (e.g., crn://confluent.cloud/organization=.../environment=.../cloud-cluster=...).
- **rest_endpoint**: REST API endpoint for cluster management (e.g., https://pkc-00000.us-central1.gcp.confluent.cloud:443).

## How It Works

Planton provisions Confluent Kafka clusters via Pulumi or Terraform modules defined in this repository. The API contract is protobuf-based (api.proto, spec.proto) and stack execution is orchestrated by the platform using the ConfluentKafkaStackInput (includes Confluent credentials and IaC metadata).

### Cluster Types Comparison

| Cluster Type | Tenancy | Scaling | Networking | Best For |
|-------------|---------|---------|------------|----------|
| **BASIC** | Multi-tenant | Fixed | Public only | Development, testing, learning |
| **STANDARD** | Multi-tenant | Elastic | Public only | Variable production workloads |
| **ENTERPRISE** | Multi-tenant | Elastic | Public + Private | Production with private networking |
| **DEDICATED** | Single-tenant | Provisioned CKU | Public + Private | Predictable high-throughput workloads |

### Availability Zones

- **SINGLE_ZONE**: Cluster runs in a single availability zone. No SLA. Suitable only for development/testing.
- **MULTI_ZONE**: Cluster automatically distributes across multiple availability zones within the region. Provides 99.99% SLA and automatic fault tolerance. Required for production workloads.

### Private Networking

For secure, private connectivity without traversing the public internet:
- **ENTERPRISE** and **DEDICATED** cluster types support private networking
- Requires a pre-created Confluent Cloud network resource with:
  - **AWS**: PrivateLink configuration
  - **Azure**: Private Link (VNet Peering)
  - **GCP**: Private Service Connect
- Must configure corresponding endpoints/connections in your VPC/VNet

### Provisioned Capacity (CKU)

Confluent Kafka Units (CKU) are used for DEDICATED clusters:
- **1-2 CKU**: Small to medium workloads (up to 50 MB/s ingress)
- **4 CKU**: High-throughput workloads (up to 150 MB/s ingress)
- **8+ CKU**: Very high-throughput or mission-critical workloads
- Can be scaled up or down (but never to zero)
- Each CKU provides dedicated compute, storage, and network capacity

## Multi-Environment Best Practice

Use separate Confluent Cloud environments for each application lifecycle stage:
- `env-dev` → Development clusters (BASIC type for cost savings)
- `env-staging` → Staging clusters (STANDARD type)
- `env-prod` → Production clusters (STANDARD, ENTERPRISE, or DEDICATED)

Each environment provides complete isolation for billing, access control, and resource management.

## Common Use Cases

### Development and Testing
BASIC cluster with SINGLE_ZONE deployment for minimal cost.

### Standard Production Workload
STANDARD cluster with MULTI_ZONE for elastic scaling and 99.99% SLA.

### Secure Production with Private Networking
ENTERPRISE cluster with private networking via PrivateLink/Private Link.

### High-Throughput Mission-Critical
DEDICATED cluster with provisioned CKU capacity and private networking.

### Disaster Recovery Setup
Deploy ENTERPRISE or DEDICATED clusters in multiple regions with Cluster Linking for cross-region replication.

## Performance and Scaling

### Elastic Scaling (STANDARD/ENTERPRISE)
- Automatically scales throughput based on workload
- No manual intervention required
- Pay only for what you use
- Best for variable workloads

### Provisioned Scaling (DEDICATED)
- Manual CKU adjustments via configuration
- Predictable performance and cost
- Best for stable, high-volume workloads
- Single-tenant isolation

## Cost Optimization

### By Cluster Type
- **BASIC**: Lowest cost, pay-as-you-go, suitable for dev/test only
- **STANDARD**: Elastic pricing, cost-effective for variable workloads
- **ENTERPRISE**: Similar to STANDARD pricing + private networking
- **DEDICATED**: Provisioned pricing, economical for sustained high throughput

### Best Practices
- Use BASIC clusters for development and testing environments
- Use STANDARD for most production workloads (auto-optimizes cost)
- Use DEDICATED only when workload is predictable and consistently high
- Delete or pause non-production clusters when not in use

## References

- Confluent Cloud Documentation: https://docs.confluent.io/cloud/current/overview.html
- Kafka Cluster Types: https://docs.confluent.io/cloud/current/clusters/cluster-types.html
- Confluent Cloud Networking: https://docs.confluent.io/cloud/current/networking/overview.html
- Terraform Provider: https://registry.terraform.io/providers/confluentinc/confluent/latest
- Pulumi Provider: https://www.pulumi.com/registry/packages/confluentcloud/
- Cluster Linking (DR): https://docs.confluent.io/cloud/current/multi-cloud/cluster-linking/index.html
