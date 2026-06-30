# GcpRedisInstance — Research and Design Documentation

## 1. Introduction

### What Is Memorystore for Redis?

Memorystore for Redis is Google Cloud's fully managed, in-memory data store backed by the Redis protocol. It provides sub-millisecond latency for read and write operations, automatic failover, managed patching, and integration with GCP networking and IAM. Unlike self-managed Redis deployments, Memorystore eliminates operational tasks such as provisioning VMs, configuring replication, applying security patches, and monitoring cluster health.

### Why Caching Matters

In-memory caching is foundational to modern application architecture. By keeping frequently accessed data in RAM instead of repeatedly querying databases or external APIs, applications achieve:

- **Reduced latency**: Sub-millisecond responses vs. tens to hundreds of milliseconds for disk-backed stores
- **Lower database load**: Offloading read traffic from primary databases extends their capacity and reduces costs
- **Improved scalability**: Caches absorb traffic spikes that would otherwise overwhelm backends
- **Session persistence**: Stateless application tiers can store user sessions in Redis for horizontal scaling
- **Real-time features**: Leaderboards, rate limiting, pub/sub messaging, and live dashboards rely on fast in-memory operations

Memorystore for Redis supports these use cases with a managed service that handles availability, security, and maintenance.

---

## 2. Historical Context

### From Self-Managed Redis to Managed Services

**Early era (2009–2015)**: Redis emerged as an open-source in-memory data structure store. Teams ran Redis on bare metal or VMs, managing replication, persistence, and failover manually. Operations required deep Redis expertise and 24/7 on-call coverage for incidents.

**Cloud-managed era (2015–present)**: Cloud providers introduced managed Redis offerings:

| Provider | Service | Launch Era |
|----------|---------|------------|
| AWS | ElastiCache for Redis | 2013 |
| Azure | Azure Cache for Redis | 2016 |
| GCP | Memorystore for Redis | 2018 |
| DigitalOcean | Managed Redis | 2019 |

These services abstract away VM management, networking, patching, and backup. Platform engineers declare desired state (tier, memory, region) and the provider provisions and operates the instance.

**Current landscape**: Managed Redis is the default choice for production workloads. Self-managed Redis remains relevant for edge deployments, air-gapped environments, or when strict control over data locality is required. For most cloud-native applications, Memorystore for Redis offers the best trade-off between operational simplicity and feature richness.

---

## 3. Deployment Methods Landscape

### Level 0: Cloud Console (Manual)

**Workflow**: Navigate to GCP Console → Memorystore → Redis → Create Instance. Fill in region, tier, memory size, network, and optional settings via the web UI. Click Create and wait for provisioning.

**Pros**:
- No tooling required; works from any browser
- Visual feedback and validation
- Suitable for one-off experiments or demos

**Cons**:
- No version control; changes are ad hoc
- No audit trail of who changed what
- Cannot be reproduced or automated
- Error-prone for repeated deployments

**Verdict**: Use only for quick exploration. Not suitable for production or team workflows.

---

### Level 1: gcloud CLI

**Workflow**: Use `gcloud redis instances create` with flags for name, region, tier, memory, network, and other options. Scripts can wrap the command for repeatability.

```bash
gcloud redis instances create my-redis \
  --size=1 \
  --region=us-central1 \
  --redis-version=redis_7_0 \
  --tier=basic \
  --network=projects/my-project/global/networks/default
```

**Pros**:
- Scriptable; can be embedded in CI/CD
- No additional tooling beyond gcloud
- Fast iteration for single-instance creation

**Cons**:
- Imperative; no drift detection or desired-state reconciliation
- Limited composition with other resources (VPC, KMS) unless scripted manually
- No built-in state management; updates require explicit commands

**Verdict**: Useful for quick provisioning or one-off automation. Lacks the declarative and compositional benefits of IaC.

---

### Level 2: Terraform

**Workflow**: Define a `google_redis_instance` resource in HCL. Run `terraform plan` and `terraform apply` to create or update. Terraform tracks state and reconciles drift.

**Example HCL**:

```hcl
resource "google_redis_instance" "cache" {
  name           = "prod-session-cache"
  project        = var.project_id
  region         = "us-central1"
  tier           = "STANDARD_HA"
  memory_size_gb = 5

  redis_version           = "REDIS_7_0"
  display_name            = "Production session cache"
  authorized_network      = google_compute_network.vpc.id
  connect_mode            = "DIRECT_PEERING"
  auth_enabled            = true
  transit_encryption_mode = "SERVER_AUTHENTICATION"
  deletion_protection     = true

  maintenance_policy {
    weekly_maintenance_window {
      day = "SUNDAY"
      start_time {
        hours = 3
      }
    }
  }

  persistence_config {
    persistence_mode    = "RDB"
    rdb_snapshot_period = "TWENTY_FOUR_HOURS"
  }

  read_replicas_mode = "READ_REPLICAS_ENABLED"
  replica_count      = 2

  labels = {
    env  = "prod"
    team = "platform"
  }
}
```

**Pros**:
- Declarative; desired state is version-controlled
- Drift detection and reconciliation
- Composition with VPC, KMS, and other resources via references
- Large ecosystem and community modules
- Plan-before-apply safety

**Cons**:
- HCL syntax and provider-specific resource schemas
- State management overhead (remote backend, locking)
- Learning curve for teams new to Terraform

**Verdict**: Industry standard for production IaC. Best choice when Terraform is already the team's tool of choice.

---

### Level 3: Pulumi

**Workflow**: Define a `redis.Instance` in Go, TypeScript, Python, or another language. Run `pulumi up` to create or update. Pulumi manages state and supports imperative logic within the program.

**Example Go**:

```go
package main

import (
    "github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/redis"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        instance, err := redis.NewInstance(ctx, "cache", &redis.InstanceArgs{
            Name:         pulumi.String("prod-session-cache"),
            Project:      pulumi.String("my-project"),
            Region:       pulumi.String("us-central1"),
            Tier:         pulumi.String("STANDARD_HA"),
            MemorySizeGb: pulumi.Int(5),

            RedisVersion:           pulumi.String("REDIS_7_0"),
            DisplayName:            pulumi.String("Production session cache"),
            AuthorizedNetwork:      pulumi.String("projects/my-project/global/networks/default"),
            ConnectMode:            pulumi.String("DIRECT_PEERING"),
            AuthEnabled:            pulumi.Bool(true),
            TransitEncryptionMode:  pulumi.String("SERVER_AUTHENTICATION"),
            DeletionProtection:     pulumi.Bool(true),

            MaintenancePolicy: &redis.InstanceMaintenancePolicyArgs{
                WeeklyMaintenanceWindows: redis.InstanceMaintenancePolicyWeeklyMaintenanceWindowArray{
                    &redis.InstanceMaintenancePolicyWeeklyMaintenanceWindowArgs{
                        Day: pulumi.String("SUNDAY"),
                        StartTime: &redis.InstanceMaintenancePolicyWeeklyMaintenanceWindowStartTimeArgs{
                            Hours: pulumi.Int(3),
                        },
                    },
                },
            },

            PersistenceConfig: &redis.InstancePersistenceConfigArgs{
                PersistenceMode:    pulumi.String("RDB"),
                RdbSnapshotPeriod:  pulumi.String("TWENTY_FOUR_HOURS"),
            },

            ReadReplicasMode: pulumi.String("READ_REPLICAS_ENABLED"),
            ReplicaCount:     pulumi.Int(2),
        })
        if err != nil {
            return err
        }

        ctx.Export("host", instance.Host)
        ctx.Export("port", instance.Port)
        ctx.Export("authString", instance.AuthString)
        return nil
    })
}
```

**Pros**:
- Full programming language; loops, conditionals, and abstractions
- Type safety and IDE support
- Same workflow for multiple clouds
- Rich output exports and composition

**Cons**:
- Requires language runtime (Go, Node, Python)
- Smaller ecosystem than Terraform
- State management similar to Terraform

**Verdict**: Strong choice for teams that prefer general-purpose languages over HCL or need complex logic in IaC.

---

### Level 4: Planton

**Workflow**: Define a `GcpRedisInstance` custom resource in YAML. Apply with `planton apply`. Planton translates the spec to Terraform or Pulumi and provisions the instance. Outputs (host, port, auth_string) are available for downstream resources.

**Example YAML**:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpRedisInstance
metadata:
  name: prod-session-cache
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpRedisInstance.prod-session-cache
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      field: status.outputs.project_id
  instanceName: prod-session-cache
  region: us-central1
  tier: STANDARD_HA
  memorySizeGb: 5
  redisVersion: REDIS_7_0
  displayName: Production session cache
  authorizedNetwork:
    valueFrom:
      kind: GcpVpc
      name: prod-vpc
      field: status.outputs.network_self_link
  connectMode: DIRECT_PEERING
  authEnabled: true
  transitEncryptionMode: SERVER_AUTHENTICATION
  deletionProtection: true
  maintenanceWindow:
    day: SUNDAY
    hour: 3
  persistenceConfig:
    persistenceMode: RDB
    rdbSnapshotPeriod: TWENTY_FOUR_HOURS
  readReplicasMode: READ_REPLICAS_ENABLED
  replicaCount: 2
```

**Pros**:
- Declarative YAML; familiar to Kubernetes users
- Cloud-agnostic abstraction; same pattern across GCP, AWS, Azure
- Cross-resource references via `valueFrom` (project, VPC, KMS)
- Provisioner-agnostic (Terraform or Pulumi backend)
- Minimal surface area; 80/20 coverage of common use cases

**Cons**:
- Abstraction layer; advanced provider features may require extension
- Additional tool (Planton) to learn and operate
- Newer ecosystem than Terraform/Pulumi

**Verdict**: Ideal for platform teams standardizing on a unified, cloud-agnostic IaC model with strong composition and minimal boilerplate.

---

## 4. Comparative Analysis

| Dimension | Console | gcloud | Terraform | Pulumi | Planton |
|-----------|---------|--------|-----------|--------|---------|
| **Declarative** | No | No | Yes | Yes | Yes |
| **Version control** | No | Partial | Yes | Yes | Yes |
| **Drift detection** | No | No | Yes | Yes | Yes |
| **Composition** | Manual | Scripted | Native | Native | Native |
| **Cross-resource refs** | Manual | Manual | Yes | Yes | Yes (valueFrom) |
| **Learning curve** | Low | Low | Medium | Medium | Low |
| **Automation** | No | Scriptable | Yes | Yes | Yes |
| **Cloud-agnostic** | No | GCP only | Multi-cloud | Multi-cloud | Multi-cloud |
| **State management** | N/A | N/A | Required | Required | Delegated |
| **Production readiness** | No | Limited | Yes | Yes | Yes |

---

## 5. The Planton Approach

### How the Abstraction Works

Planton defines a **GcpRedisInstance** as a custom resource with a spec that maps to the underlying `google_redis_instance` (Terraform) or `redis.Instance` (Pulumi). The spec is designed around the 80/20 principle: cover the majority of production use cases with a minimal, consistent API while avoiding provider-specific quirks.

**Key design choices**:

1. **StringValueOrRef**: Fields like `projectId`, `authorizedNetwork`, and `customerManagedKey` accept either a literal value or a reference to another Planton resource (`valueFrom`). This enables infra chart composition without hardcoding IDs.

2. **Sensible defaults**: Omitted optional fields use provider defaults. Required fields (projectId, instanceName, region, tier, memorySizeGb) are explicitly validated.

3. **Immutable field awareness**: The spec documents which fields are immutable (instanceName, tier, connectMode, transitEncryptionMode, authorizedNetwork, reservedIpRange, customerManagedKey). Changing them triggers replacement, not in-place update.

4. **Output contract**: All provisioners export `host`, `port`, `readEndpoint`, `readEndpointPort`, `currentLocationId`, and `authString` (when auth is enabled). Downstream resources consume these via `valueFrom`.

### 80/20 Principle

**Included (covers ~80% of deployments)**:
- Tier selection (BASIC, STANDARD_HA)
- Memory sizing (1–300 GiB)
- Redis version (REDIS_6_X, REDIS_7_0, REDIS_7_2)
- VPC networking (authorized_network, connect_mode, reserved_ip_range)
- Auth and TLS (auth_enabled, transit_encryption_mode)
- Persistence (RDB snapshots with configurable period)
- Read replicas (1–5 for STANDARD_HA)
- Maintenance windows (day + hour)
- CMEK (customer_managed_key)
- Deletion protection
- Redis configs (key-value overrides)

**Excluded (deferred to v2 or out of scope)**:
- Redis Cluster mode (sharding) — different topology model
- Cross-region replication — niche, complex
- Automated backup/restore workflows — often handled by external tooling
- Metrics/alerting presets — typically configured in monitoring stack

---

## 6. Implementation Landscape

### Pulumi Module Architecture

The Planton GcpRedisInstance Pulumi module lives at `apis/dev/planton/provider/gcp/gcpredisinstance/v1/iac/pulumi/`. The main program:

1. Loads the GcpRedisInstance spec from the Planton manifest (or stack input)
2. Maps spec fields to `redis.InstanceArgs`
3. Creates the instance via `redis.NewInstance`
4. Exports host, port, readEndpoint, readEndpointPort, currentLocationId, authString

**Resource**: `github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/redis.Instance`

**Key mappings**:
| Spec Field | Pulumi Property |
|------------|-----------------|
| instanceName | Name |
| projectId | Project |
| region | Region |
| tier | Tier |
| memorySizeGb | MemorySizeGb |
| authEnabled | AuthEnabled |
| transitEncryptionMode | TransitEncryptionMode |
| maintenanceWindow | MaintenancePolicy |
| persistenceConfig | PersistenceConfig |
| readReplicasMode, replicaCount | ReadReplicasMode, ReplicaCount |
| customerManagedKey | CustomerManagedKey |
| deletionProtection | DeletionProtection |

### Terraform Module Architecture

The Terraform module lives at `apis/dev/planton/provider/gcp/gcpredisinstance/v1/iac/tf/`. It:

1. Accepts `spec` as a variable (object matching the GcpRedisInstance spec)
2. Uses `locals` to derive values (with defaults for optional fields)
3. Creates a `google_redis_instance` resource
4. Outputs host, port, read_endpoint, read_endpoint_port, current_location_id, auth_string

**Resource**: `google_redis_instance` (Hashicorp Google provider)

**Key mappings**:
| Spec Field | Terraform Argument |
|------------|-------------------|
| instanceName | name |
| projectId | project |
| region | region |
| tier | tier |
| memorySizeGb | memory_size_gb |
| authEnabled | auth_enabled |
| transitEncryptionMode | transit_encryption_mode |
| maintenanceWindow | maintenance_policy |
| persistenceConfig | persistence_config |
| readReplicasMode, replicaCount | read_replicas_mode, replica_count |
| customerManagedKey | customer_managed_key |
| deletionProtection | deletion_protection |

---

## 7. Production Best Practices

### Sizing

- **Memory**: Start with 1 GiB for dev; production typically needs 5–50 GiB depending on cache size and eviction policy. Memorystore supports 1–300 GiB.
- **Tier**: Use BASIC for dev/test; use STANDARD_HA for production (99.9% SLA, automatic failover).
- **Read replicas**: Enable `READ_REPLICAS_ENABLED` with 1–5 replicas for read-heavy workloads. Distributes read traffic and improves throughput.

### Networking

- **VPC attachment**: Always specify `authorized_network` to attach the instance to a private VPC. Never expose Redis on the public internet.
- **Connect mode**: Use `DIRECT_PEERING` for simple VPC setups. Use `PRIVATE_SERVICE_ACCESS` for Shared VPC or when DIRECT_PEERING is not available.
- **Reserved IP range**: For DIRECT_PEERING, reserve a /29 block via GcpGlobalAddress (purpose: VPC_PEERING) and reference it in `reserved_ip_range`. Ensures predictable addressing and avoids conflicts.

### Security (AUTH + TLS)

- **Redis AUTH**: Enable `auth_enabled: true` for all production instances. GCP generates and auto-rotates the AUTH string. Store it in Secret Manager or inject via `valueFrom`; never commit to version control.
- **TLS in transit**: Set `transit_encryption_mode: SERVER_AUTHENTICATION` for encrypted client connections. Required when using AUTH in many client libraries.
- **CMEK**: Use `customer_managed_key` for encryption at rest when compliance requires customer-managed keys. Ensure the KMS key has appropriate IAM bindings.

### Persistence

- **RDB snapshots**: For STANDARD_HA, enable `persistence_config` with `persistence_mode: RDB` and `rdb_snapshot_period` (ONE_HOUR, SIX_HOURS, TWELVE_HOURS, TWENTY_FOUR_HOURS). Balances durability with performance impact.
- **Cache vs. durable store**: If Redis is a pure cache (data can be rebuilt), persistence may be unnecessary. For session storage or critical state, enable RDB.

### Maintenance Windows

- **Schedule**: Set `maintenance_window` to a low-traffic period (e.g., Sunday 03:00 UTC). GCP performs patching and upgrades during this 1-hour window.
- **Impact**: Expect brief connectivity interruptions during maintenance. Design clients for reconnection and retry.

### Monitoring

- **Cloud Monitoring**: Memorystore exports metrics (connected clients, memory usage, commands/sec, latency). Create alerts for high memory, connection limits, or elevated error rates.
- **Uptime checks**: Use uptime checks to verify Redis is reachable from your application network.
- **Logging**: Enable Redis slow query logs if supported; tune `maxmemory-policy` and eviction settings based on usage patterns.

---

## 8. Common Pitfalls

1. **Immutable field changes**: Changing `tier`, `connect_mode`, `authorized_network`, `transit_encryption_mode`, or `customer_managed_key` forces instance replacement. Plan for downtime or use blue-green deployment (create new, migrate, delete old).

2. **Reserved IP range conflicts**: The /29 block for DIRECT_PEERING must not overlap with existing subnets or other reserved ranges. Validate before apply.

3. **AUTH string handling**: The AUTH string is a secret. Do not log it, commit it, or expose it in client configs stored in plain text. Use Secret Manager or environment variables injected at runtime.

4. **Read replicas on BASIC**: Read replicas require STANDARD_HA. Setting `read_replicas_mode: READ_REPLICAS_ENABLED` on BASIC tier will fail validation.

5. **Persistence on BASIC**: RDB persistence is only meaningful for STANDARD_HA. BASIC instances have no replica to fail over to; persistence provides limited benefit.

6. **Private Service Access setup**: PRIVATE_SERVICE_ACCESS requires a private connection and allocated IP range. Ensure the necessary APIs and peering are configured before creating the instance.

7. **Region and zone placement**: Instance creation can fail if the region has no capacity. Use `location_id` to pin to a specific zone when needed; otherwise let GCP choose.

8. **Deletion protection**: Enable `deletion_protection: true` for production. Disable it explicitly before destroying the instance, or Terraform/Pulumi will error.

---

## 9. Conclusion

### When to Use Memorystore for Redis

- **Caching**: Database query cache, API response cache, session cache
- **Session storage**: Stateless web apps storing user sessions
- **Rate limiting**: Sliding window or token bucket counters
- **Real-time features**: Leaderboards, live dashboards, counters
- **Pub/sub**: Decoupling services with Redis pub/sub channels

### When to Consider Alternatives

- **Multi-terabyte datasets**: Redis Cluster (sharding) is not yet supported by Memorystore; consider ElastiCache or self-managed Redis Cluster
- **Cross-region replication**: Memorystore is regional; use application-level replication or a different service
- **Edge/low-latency**: Self-managed Redis co-located with compute may offer lower latency for specific topologies

### References

- [Memorystore for Redis Documentation](https://cloud.google.com/memorystore/docs/redis)
- [Redis Instance REST API](https://cloud.google.com/memorystore/docs/redis/reference/rest/v1/projects.locations.instances)
- [Terraform google_redis_instance](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/redis_instance)
- [Pulumi GCP Redis Instance](https://www.pulumi.com/registry/packages/gcp/api-docs/redis/instance/)
- [Memorystore for Redis Best Practices](https://cloud.google.com/memorystore/docs/redis/best-practices)
- [VPC Peering for Memorystore](https://cloud.google.com/memorystore/docs/redis/connect-redis-instance-vpc)
