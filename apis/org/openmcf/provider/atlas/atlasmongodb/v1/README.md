# AtlasMongodb

Atlas MongoDB cluster resource for deploying managed MongoDB databases across AWS, GCP, and Azure. Defines cluster configuration including topology (replica set, sharded, or geo-sharded), provider settings, backup policies, and multi-region/multi-cloud deployments.

## Spec Fields (80/20)

### Essential Fields (80% Use Case)
- **project_id**: The 24-character Atlas Project ID where the cluster will be deployed.
- **cluster_type**: Cluster topology - REPLICASET (default), SHARDED, or GEOSHARDED.
- **provider_name**: Cloud provider - AWS, GCP, or AZURE.
- **provider_instance_size_name**: Cluster tier (M10, M20, M30, M50, etc.). M30+ recommended for production.
- **mongo_db_major_version**: MongoDB version (e.g., "7.0", "6.0"). Defaults to latest stable if omitted.
- **electable_nodes**: Number of voting, data-bearing nodes (typically 3 for HA).
- **priority**: Election priority (1-7, with 7 being the preferred primary region).
- **cloud_backup**: Enable cloud provider snapshots (recommended: true).
- **auto_scaling_disk_gb_enabled**: Enable automatic storage scaling to prevent disk space issues.

### Advanced Fields (20% Use Case)
- **read_only_nodes**: Dedicated read-only nodes for read scaling (analytics workloads).

## Stack Outputs

- **id**: The provider-assigned unique cluster ID for the Atlas MongoDB cluster.
- **bootstrap_endpoint**: MongoDB connection string in SRV format (recommended for drivers) - mongodb+srv://...
- **crn**: The cluster identifier (same as id) for resource identification.
- **rest_endpoint**: MongoDB connection string in standard format - mongodb://host:port/...

## How It Works

OpenMCF provisions Atlas MongoDB clusters via Pulumi or Terraform modules defined in this repository. The API contract is protobuf-based (api.proto, spec.proto) and stack execution is orchestrated by the platform using the AtlasMongodbStackInput (includes Atlas credentials and IaC metadata).

### Multi-Cloud Architecture

Atlas is uniquely designed for true multi-cloud deployments. A single cluster can span AWS, GCP, and Azure simultaneously, enabling:
- **Provider-Level Disaster Recovery**: Survive complete cloud provider outages by distributing electable nodes across AWS and GCP.
- **Multi-Region Resilience**: Deploy nodes across multiple regions within the same or different cloud providers.
- **Geo-Sharding**: Distribute data geographically for low-latency access and data sovereignty compliance (GEOSHARDED type).

### Cluster Service Models

- **M0 (Free Tier)**: 512 MB storage, shared resources. For learning and prototypes only.
- **Flex**: Modern shared tier (~$30/month cap) for development and testing.
- **M10-M20**: Entry-level dedicated tiers with burstable performance. Suitable for non-critical workloads.
- **M30+**: Production-grade dedicated tiers with consistent performance (recommended for production).

### Network Security Tiers

1. **Basic (IP Access Lists)**: Simple IP whitelisting. Suitable for development only.
2. **Good (VPC Peering)**: Private connection between your VPC and Atlas. Isolates database traffic from internet.
3. **Best (Private Endpoints)**: Unidirectional private connectivity via cloud provider native services (PrivateLink, Private Link, Private Service Connect). Production best practice.

### Backup Strategies

- **Cloud Backups**: Scheduled snapshots using native cloud provider capabilities. RPO determined by snapshot frequency (e.g., 6 hours).
- **Continuous Backups (PITR)**: Point-in-Time Recovery with near-zero RPO. Captures continuous oplog for precise recovery to any second.

## Multi-Environment Best Practice

Following Atlas's recommended pattern, use separate Atlas Projects for each environment:
- `company-dev` → Development clusters
- `company-staging` → Staging clusters
- `company-prod` → Production clusters

Each environment references a different `project_id`, providing complete isolation for billing, alerts, and security boundaries.

## Common Use Cases

### Single-Region Production Cluster
Standard 3-node replica set in one region with M30+ tier, cloud backups, and auto-scaling enabled.

### Multi-Region Cluster with Failover
Primary region with 3 electable nodes + secondary region with read-only replicas for disaster recovery and read scaling.

### Multi-Cloud High-Availability
Electable nodes distributed across AWS and GCP for provider-level resilience.

### Sharded Cluster for Scale
Horizontal scaling with multiple shards for high-throughput applications or massive datasets.

### Global Cluster for Geographic Distribution
Geo-sharded deployment with location-aware reads/writes for globally distributed users.

## Performance Considerations

- **Indexing**: The single most critical performance factor. Queries without supporting indexes perform collection scans (extremely slow).
- **Connection Pooling**: All MongoDB drivers support connection pooling—enable it in your connection string to avoid exhausting connection limits.
- **Co-location**: Deploy application servers in the same cloud provider and region as the Atlas primary node to minimize network costs and latency.

## References

- Atlas MongoDB Documentation: https://www.mongodb.com/docs/atlas/
- Atlas Admin API: https://www.mongodb.com/docs/atlas/reference/api-resources/
- Atlas Best Practices: https://www.mongodb.com/docs/atlas/best-practices/
- Terraform Provider: https://registry.terraform.io/providers/mongodb/atlasmongodb/latest
- Pulumi Provider: https://www.pulumi.com/registry/packages/atlasmongodb/

