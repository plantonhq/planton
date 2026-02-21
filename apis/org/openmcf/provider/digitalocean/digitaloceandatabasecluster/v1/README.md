# DigitalOcean Database Cluster

Manage fully-managed database clusters on DigitalOcean using a type-safe, protobuf-defined API with OpenMCF.

## Overview

**DigitalOceanDatabaseCluster** enables you to provision and manage highly-available, fully-managed database clusters on DigitalOcean. Supports PostgreSQL, MySQL, Redis, and MongoDB with automated backups, patching, and monitoring.

DigitalOcean Managed Databases delivers simplicity and predictable pricing for startups and SMBs. Unlike AWS RDS or Google Cloud SQL with overwhelming configuration options, DigitalOcean focuses on the essential 80% of features with transparent, flat-rate pricing.

## Why Use This Component?

- **Type-Safe Configuration**: Protobuf-based API with compile-time validation prevents invalid cluster configurations
- **80/20 Focused**: Exposes only essential fields (name, engine, version, size, nodes, region, VPC, storage)
- **Production-Ready Defaults**: Enforces best practices (HA node counts, VPC networking, private connectivity)
- **Cost-Effective**: DigitalOcean's bandwidth is $0.01/GB vs AWS's $0.15/GB—10x cheaper for data-heavy workloads

## Quick Start

### Development PostgreSQL Cluster (Single Node)

For non-production workloads:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDatabaseCluster
metadata:
  name: dev-postgres
spec:
  cluster_name: dev-postgres
  engine: postgres
  engine_version: "16"
  region: nyc3
  size_slug: db-s-1vcpu-1gb
  node_count: 1
  enable_public_connectivity: false
```

### Production PostgreSQL Cluster (High Availability)

For mission-critical workloads:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDatabaseCluster
metadata:
  name: prod-postgres
spec:
  cluster_name: prod-postgres
  engine: postgres
  engine_version: "16"
  region: nyc3
  size_slug: db-s-4vcpu-8gb
  node_count: 3  # Primary + 2 standbys (HA)
  vpc:
    value: "12345678-1234-1234-1234-123456789012"  # VPC UUID
  storage_gib: 100
  enable_public_connectivity: false
```

## Key Features

### Supported Engines
- **PostgreSQL**: Industry-standard relational database with ACID guarantees
- **MySQL**: Popular open-source RDBMS for web applications
- **Redis**: In-memory data store for caching and session management
- **MongoDB**: Document-oriented NoSQL database

### High Availability
- **Single Node (node_count: 1)**: No redundancy, dev/test only
- **Two Nodes (node_count: 2)**: Minimum HA, automatic failover
- **Three Nodes (node_count: 3)**: Maximum HA, production-ready

### VPC Integration
- **Private Networking**: Deploy clusters in DigitalOcean VPCs for security
- **Zero Egress Costs**: VPC-internal traffic is free (saves 10x vs public bandwidth)
- **Network Isolation**: Clusters accessible only from authorized VPC resources

### Automated Operations
- **Daily Backups**: Automatic backups with point-in-time recovery
- **Automatic Patching**: Security updates applied during maintenance windows
- **Monitoring**: Built-in metrics for CPU, memory, disk, connections
- **Connection Pooling**: PgBouncer for PostgreSQL (required for production)

## Configuration Reference

### Core Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `cluster_name` | string | Yes | Unique cluster identifier (max 64 chars) |
| `engine` | enum | Yes | postgres, mysql, redis, or mongodb |
| `engine_version` | string | Yes | Major.minor version (e.g., "16", "8.0") |
| `region` | enum | Yes | DigitalOcean region (nyc3, sfo3, etc.) |
| `size_slug` | string | Yes | Node size (db-s-1vcpu-1gb, db-s-4vcpu-8gb, etc.) |
| `node_count` | uint32 | Yes | Number of nodes (1-3) |
| `vpc` | StringValueOrRef | No | VPC UUID for private networking |
| `storage_gib` | uint32 | No | Custom storage size (if larger than default) |
| `enable_public_connectivity` | bool | No | Allow public internet access (default: false) |

### Size Slugs

**Development:**
- `db-s-1vcpu-1gb` - $15/month, suitable for dev/test
- `db-s-1vcpu-2gb` - $30/month, small production workloads

**Production:**
- `db-s-2vcpu-4gb` - $60/month, standard production
- `db-s-4vcpu-8gb` - $120/month, high-traffic production
- `db-s-8vcpu-16gb` - $240/month, enterprise workloads

### Version Support

- **PostgreSQL**: 12, 13, 14, 15, 16
- **MySQL**: 8 (8.0.x)
- **Redis**: 6, 7
- **MongoDB**: 4.4, 5.0, 6.0, 7.0

## Outputs

After successful provisioning:

| Output | Description |
|--------|-------------|
| `cluster_id` | DigitalOcean cluster UUID |
| `connection_uri` | Full connection string (includes credentials) |
| `host` | Cluster hostname |
| `port` | Cluster port |
| `database_name` | Default database name |
| `username` | Admin username |
| `password` | Admin password (sensitive) |

## Common Use Cases

### 1. Development PostgreSQL (Single Node, Public Access)

```yaml
spec:
  cluster_name: dev-pg
  engine: postgres
  engine_version: "16"
  region: nyc3
  size_slug: db-s-1vcpu-1gb
  node_count: 1
  enable_public_connectivity: true  # For dev access
```

### 2. Production PostgreSQL (HA, VPC-Private)

```yaml
spec:
  cluster_name: prod-pg
  engine: postgres
  engine_version: "16"
  region: nyc3
  size_slug: db-s-4vcpu-8gb
  node_count: 3  # HA configuration
  vpc:
    value: "vpc-12345678"
  storage_gib: 200
  enable_public_connectivity: false
```

### 3. MySQL for Web Applications

```yaml
spec:
  cluster_name: prod-mysql
  engine: mysql
  engine_version: "8"
  region: sfo3
  size_slug: db-s-2vcpu-4gb
  node_count: 2  # HA
  vpc:
    value: "vpc-87654321"
```

### 4. Redis Cache Cluster

```yaml
spec:
  cluster_name: prod-redis-cache
  engine: redis
  engine_version: "7"
  region: nyc3
  size_slug: db-s-2vcpu-4gb
  node_count: 2  # HA
  vpc:
    value: "vpc-abcdef12"
```

### 5. MongoDB Document Store

```yaml
spec:
  cluster_name: prod-mongodb
  engine: mongodb
  engine_version: "7.0"
  region: fra1
  size_slug: db-s-4vcpu-8gb
  node_count: 3  # HA
  vpc:
    value: "vpc-xyz789"
  storage_gib: 500
```

## Best Practices

### High Availability
1. **Use node_count ≥ 2 for production** - Automatic failover requires standbys
2. **Implement retry logic** - Applications must handle brief connection drops during failover
3. **Test failover scenarios** - Validate your app gracefully handles primary node failure

### Security
1. **VPC-First**: Always deploy in a VPC, never expose to public internet for production
2. **Use Private Connection Strings**: Faster, more secure, zero bandwidth costs
3. **Tag-Based Firewalls**: Use DigitalOcean tags instead of IP addresses for access control
4. **Rotate Credentials**: Change default admin password immediately after provisioning

### Cost Optimization
1. **Start Small**: Use db-s-1vcpu-1gb for dev, upgrade for production
2. **Leverage Free VPC Bandwidth**: Keep apps and databases in same region/VPC
3. **Monitor Storage Growth**: Databases don't autoscale storage; plan for growth
4. **Use Read Replicas**: Offload read traffic to replicas instead of oversizing primary

### PostgreSQL-Specific
1. **Mandatory PgBouncer**: Connection limits are severely restricted (97 connections for 4GB RAM)
2. **Enable Connection Pooling**: DigitalOcean provides built-in PgBouncer pools
3. **Monitor Connections**: Alert when approaching connection limit

### MySQL-Specific
1. **Backup Limitations**: Native `mysqldump` restricted; use DigitalOcean's backup system
2. **No SUPER Privileges**: Cannot SET GLOBAL variables or install plugins
3. **Vendor Lock-In Warning**: Limited backup portability (see research doc)

### Redis-Specific
1. **Valkey Migration**: Redis 7+ is actually Valkey (SSPL licensing)
2. **No Clustering**: DigitalOcean Redis doesn't support Redis Cluster mode
3. **Use for Caching**: Treat as ephemeral; don't rely on persistence for critical data

## Production Checklist

Before deploying to production:

- ✅ Set `node_count` ≥ 2 for high availability
- ✅ Deploy in VPC (`vpc` field specified)
- ✅ Set `enable_public_connectivity: false` (private only)
- ✅ Choose production-appropriate `size_slug` (≥ db-s-2vcpu-4gb)
- ✅ Plan storage growth (`storage_gib` for databases that will grow)
- ✅ Configure firewall rules (via separate DigitalOceanDatabaseFirewall resource)
- ✅ Enable PgBouncer for PostgreSQL (via connection pool resource)
- ✅ Test application retry/reconnection logic

## Integration

### With DigitalOcean VPC

```yaml
# First, create VPC
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanVpc
metadata:
  name: prod-vpc
spec:
  name: prod-vpc
  region: nyc3
  ip_range: 10.10.0.0/16

---
# Then reference VPC in database cluster
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanDatabaseCluster
metadata:
  name: prod-postgres
spec:
  cluster_name: prod-postgres
  engine: postgres
  engine_version: "16"
  region: nyc3
  size_slug: db-s-4vcpu-8gb
  node_count: 3
  vpc:
    ref:
      kind: DigitalOceanVpc
      name: prod-vpc
      field_path: status.outputs.vpc_id
```

### With Application Workloads

Applications retrieve connection details from outputs:

```yaml
# Application deployment references database outputs
env:
  - name: DATABASE_URL
    value: ${digitalocean_database_cluster.prod_postgres.connection_uri}
  - name: DB_HOST
    value: ${digitalocean_database_cluster.prod_postgres.host}
  - name: DB_PORT
    value: ${digitalocean_database_cluster.prod_postgres.port}
```

## Troubleshooting

### Cluster Creation Fails

**Cause**: Invalid size_slug for selected engine/version.

**Solution**: Verify size_slug compatibility:
```bash
doctl databases options sizes --engine postgres
```

### Connection Timeouts

**Cause**: Firewall rules not configured or VPC misconfigured.

**Solution**:
1. Verify cluster is in same VPC as application
2. Check firewall rules allow VPC CIDR or application tags
3. Use private connection string, not public

### PostgreSQL: "Too Many Connections"

**Cause**: Exceeded connection limit (97 for 4GB RAM cluster).

**Solution**:
1. **Mandatory**: Enable PgBouncer connection pooling
2. Configure application to use connection pool instead of direct connections
3. Monitor connection usage in DigitalOcean dashboard

### MySQL: Cannot Restore mysqldump Backup

**Cause**: DigitalOcean restricts native MySQL utilities.

**Solution**: Use DigitalOcean's backup system. For migrations, export data as SQL and import via supported methods.

## Validation Rules

The protobuf spec enforces these constraints:

- `cluster_name`: Max 64 characters
- `engine`: Must be postgres, mysql, redis, or mongodb
- `engine_version`: Must match pattern `^[0-9]+(\.[0-9]+)?$` (e.g., "16", "8.0")
- `node_count`: 1-3 nodes (enforced by buf.validate)
- `region`: Must be valid DigitalOcean region enum
- `size_slug`: Required, no pattern validation (validated by DigitalOcean API)

## Limitations

### What DigitalOcean Manages
✅ Automatic daily backups (retained 7 days)  
✅ Automatic security patching  
✅ Automatic failover (multi-node clusters)  
✅ Monitoring and alerting  
✅ PgBouncer connection pooling (PostgreSQL)

### What You Must Manage
❌ Application connection retry logic  
❌ Query optimization and indexing  
❌ Database schema migrations  
❌ Firewall rules (separate resource)  
❌ User management (beyond default admin)  
❌ Read replicas (separate resource)

### Platform Limitations
- **PostgreSQL**: Severely limited connection counts (requires PgBouncer)
- **MySQL**: No SUPER privileges, restricted mysqldump
- **Redis**: No Redis Cluster mode (single-instance only)
- **MongoDB**: No sharding support (replica sets only)
- **Storage**: No autoscaling (plan for growth upfront)

## Further Reading

- **Comprehensive Guide**: See [docs/README.md](./docs/README.md) for deep-dive coverage of deployment methods, production essentials, and platform-specific gotchas
- **Examples**: See [examples.md](./examples.md) for copy-paste ready manifests
- **Pulumi Module**: See [iac/pulumi/README.md](./iac/pulumi/README.md) for standalone Pulumi usage
- **Terraform Module**: See [iac/tf/README.md](./iac/tf/README.md) for standalone Terraform usage

## Support

For issues, questions, or contributions, refer to the [OpenMCF documentation](https://openmcf.org) or file an issue in the repository.

---

**TL;DR**: Use node_count ≥ 2 for production HA. Always deploy in VPC. Enable PgBouncer for PostgreSQL. Plan storage growth upfront (no autoscaling).
