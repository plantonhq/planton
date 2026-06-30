# Civo Database

Provision and manage Civo managed database instances (MySQL and PostgreSQL) through Planton's declarative infrastructure-as-code approach.

## Overview

`CivoDatabase` provides a Kubernetes-native way to deploy production-grade managed databases on Civo Cloud. It abstracts away the complexity of multi-step API calls and configuration while providing essential features like high availability, network isolation, and firewall rules.

### Key Features

- **Multiple Database Engines**: MySQL and PostgreSQL with configurable versions
- **High Availability**: Support for up to 5 nodes (1 primary + 4 replicas)
- **Network Security**: Private network isolation with firewall rules
- **Transparent Pricing**: Predictable costs with bundled storage and no hidden fees
- **Automatic Backups**: Daily automated backups with 24-hour recovery granularity
- **Flexible Sizing**: From small development instances to large production clusters

## Prerequisites

- Civo account with API access
- Civo API token configured
- Existing Civo network (private network)
- (Optional) Civo firewall for access control

## Quick Start

### Minimal Configuration

Deploy a development PostgreSQL database:

```yaml
apiVersion: code.planton.io/v1
kind: CivoDatabase
metadata:
  name: dev-postgres
spec:
  dbInstanceName: dev-db
  engine: postgres
  engineVersion: "16"
  region: lon1
  sizeSlug: g3.db.small
  networkId:
    value: "net-12345678-abcd-1234-abcd-1234567890ab"
```

### Production Configuration

Deploy a highly-available PostgreSQL database with replicas and firewall:

```yaml
apiVersion: code.planton.io/v1
kind: CivoDatabase
metadata:
  name: prod-postgres
  env: production
spec:
  dbInstanceName: production-db
  engine: postgres
  engineVersion: "16"
  region: lon1
  sizeSlug: g3.db.large
  replicas: 2  # Creates 3 total nodes (1 primary + 2 replicas)
  networkId:
    value: "net-12345678-abcd-1234-abcd-1234567890ab"
  firewallIds:
    - value: "fw-87654321-dcba-4321-dcba-0987654321ba"
  storageGib: 200
  tags:
    - production
    - backend
    - primary
```

## Configuration Reference

### Database Instance Name (`db_instance_name`)

**Type**: `string` (required)  
**Max Length**: 64 characters

A human-readable name for your database instance. Must be unique within the chosen Civo region.

**Example**:
```yaml
db_instance_name: my-app-database
```

### Engine (`engine`)

**Type**: `enum` (required)  
**Values**: `mysql`, `postgres`

The database engine to use.

**Examples**:
```yaml
engine: postgres  # PostgreSQL
engine: mysql     # MySQL
```

### Engine Version (`engine_version`)

**Type**: `string` (required)  
**Pattern**: `^[0-9]+(\.[0-9]+)?$`

Major (and optionally minor) version number for the database engine.

**Examples**:
```yaml
engine_version: "16"     # PostgreSQL 16
engine_version: "8.0"    # MySQL 8.0
engine_version: "14.10"  # PostgreSQL 14.10
```

### Region (`region`)

**Type**: `enum` (required)

The Civo region where the database will be created.

**Available Regions**:
- `lon1` - London
- `nyc1` - New York City
- `fra1` - Frankfurt
- `phx1` - Phoenix

**Example**:
```yaml
region: lon1
```

### Size Slug (`size_slug`)

**Type**: `string` (required)

The instance tier that defines CPU, RAM, and storage. Civo bundles these into fixed tiers.

**Available Tiers**:

| Size Slug | vCPU | RAM | Storage | Price/month |
|-----------|------|-----|---------|-------------|
| `g3.db.small` | 2 | 4 GB | 40 GB | ~$43 |
| `g3.db.medium` | 4 | 8 GB | 80 GB | ~$87 |
| `g3.db.large` | 6 | 16 GB | 160 GB | ~$174 |
| `g3.db.xlarge` | 8 | 32 GB | 320 GB | ~$348 |
| `g3.db.2xlarge` | 10 | 64 GB | 640 GB | ~$695 |

**Example**:
```yaml
size_slug: g3.db.medium
```

### Replicas (`replicas`)

**Type**: `uint32` (optional, default: 0)  
**Max**: 4

Number of replica nodes to add for high availability. The total cluster size will be `replicas + 1` (primary + replicas).

**Examples**:
```yaml
replicas: 0  # 1 node total (master only)
replicas: 1  # 2 nodes total (master + 1 replica)
replicas: 2  # 3 nodes total (master + 2 replicas) - recommended for production
replicas: 4  # 5 nodes total (master + 4 replicas) - maximum allowed
```

### Network ID (`network_id`)

**Type**: `StringValueOrRef` (required)

The private network where the database will be deployed. This ensures the database is isolated from the public internet.

**Literal Value**:
```yaml
network_id:
  value: "net-12345678-abcd-1234-abcd-1234567890ab"
```

**Reference to CivoVpc Resource**:
```yaml
network_id:
  value_from:
    kind: CivoVpc
    name: my-vpc
    field_path: status.outputs.network_id
```

### Firewall IDs (`firewall_ids`)

**Type**: `list[StringValueOrRef]` (optional)

Firewall rules to control access to the database. Civo currently supports one firewall per database, so only the first entry will be used.

**Literal Value**:
```yaml
firewall_ids:
  - value: "fw-87654321-dcba-4321-dcba-0987654321ba"
```

**Reference to CivoFirewall Resource**:
```yaml
firewall_ids:
  - value_from:
      kind: CivoFirewall
      name: db-firewall
      field_path: status.outputs.firewall_id
```

### Storage GiB (`storage_gib`)

**Type**: `uint32` (optional)

Custom storage size in GiB, overriding the default provided by `size_slug`.

**Example**:
```yaml
storage_gib: 200  # 200 GiB storage
```

### Tags (`tags`)

**Type**: `list[string]` (optional)

Tags for organizing and categorizing the database instance.

**Example**:
```yaml
tags:
  - production
  - backend
  - primary-db
```

## Connection Details

After the database is provisioned, connection details are available in the resource status:

- **DNS Endpoint**: Recommended for HA setups (automatically updates on failover)
- **Host**: Direct hostname (static)
- **Port**: Database port (typically 3306 for MySQL, 5432 for PostgreSQL)
- **Username**: Master database username
- **Password**: Master database password (stored securely)

### Connecting from Applications

For high availability deployments, **always use the DNS endpoint** rather than the static host. This ensures your application automatically reconnects to the new primary after a failover.

```bash
# PostgreSQL connection string (using DNS endpoint)
postgresql://username:password@dns-endpoint:5432/database_name

# MySQL connection string (using DNS endpoint)
mysql://username:password@dns-endpoint:3306/database_name
```

## Best Practices

### Security

1. **Always use private networks**: Deploy databases in a `CivoNetwork` (private network) to prevent public internet access.

2. **Configure firewall rules**: Attach a `CivoFirewall` with least-privilege access control:
   ```yaml
   firewallIds:
     - value_from:
         kind: CivoFirewall
         name: db-firewall
   ```

3. **Never expose to 0.0.0.0/0**: Restrict access to specific CIDR blocks (e.g., Kubernetes node range).

### High Availability

1. **Use replicas for production**: Set `replicas: 2` or higher for production workloads to ensure availability during maintenance or failures.

2. **Connect via DNS endpoint**: Use the `dns_endpoint` in your connection strings, not the static `host`. This enables automatic failover.

3. **Test failover scenarios**: Regularly test your application's behavior during database failovers in staging environments.

### Backup and Recovery

1. **Civo provides daily backups**: Automated backups run daily with 24-hour granularity.

2. **Implement custom backups for production**: For critical production workloads, supplement Civo's native backups with custom solutions:
   - Schedule `pg_dump` (PostgreSQL) or `mysqldump` (MySQL) using Kubernetes CronJobs
   - Upload backups to Civo Object Storage (S3-compatible)
   - Configure retention policies

3. **Test restore procedures**: Regularly test database restore processes to ensure backups are valid.

### Cost Optimization

1. **Right-size your instances**: Start with smaller tiers (`g3.db.small`) and scale up based on actual usage.

2. **Use replicas judiciously**: Each replica doubles (or more) your cost. Use replicas only when high availability is genuinely required.

3. **Leverage transparent pricing**: Civo's all-inclusive pricing (no egress or IOPS fees) makes cost forecasting simple.

## Troubleshooting

### Database Won't Create

**Symptom**: Resource creation fails or times out.

**Common Causes**:
- Invalid network ID (network doesn't exist or is in a different region)
- Invalid firewall ID
- Region mismatch between database, network, and firewall
- API token lacks necessary permissions

**Solution**: Verify network and firewall exist in the same region, and ensure API token has database creation permissions.

### Cannot Connect to Database

**Symptom**: Application cannot reach the database.

**Common Causes**:
- Firewall rules blocking access
- Application not in the same private network
- Using static `host` instead of `dns_endpoint` in HA setups

**Solution**:
1. Verify firewall rules allow traffic from your application's CIDR
2. Ensure application pods/instances are in the same Civo network
3. Use `dns_endpoint` for connections in HA configurations

### Slow Performance

**Symptom**: Database queries are slow or timing out.

**Common Causes**:
- Under-provisioned instance tier (CPU/RAM)
- Missing database indexes
- Heavy concurrent load exceeding instance capacity

**Solution**:
1. Monitor database metrics using Percona PMM (available in Civo Marketplace)
2. Scale up to a larger `size_slug` if CPU/RAM is consistently high
3. Add read replicas (`replicas: 2+`) for read-heavy workloads
4. Optimize slow queries and add indexes

## Limitations

- **No Point-in-Time Recovery (PITR)**: Civo provides daily backups, not minute-level PITR. Implement custom solutions for critical workloads.
- **One Firewall Per Database**: Civo currently supports attaching only one firewall rule to a database.
- **Bundled Storage**: Storage is coupled with instance tiers. Independent storage scaling requires changing `size_slug`.
- **Downtime on Vertical Scaling**: Changing `size_slug` requires instance reprovisioning, causing brief downtime.

## Additional Resources

- **Research Documentation**: See [`docs/README.md`](docs/README.md) for comprehensive landscape analysis and architectural patterns
- **Examples**: See [`examples.md`](examples.md) for more configuration examples
- **Pulumi Module**: See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific documentation
- **Terraform Module**: See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific documentation

## Support

For questions, issues, or feature requests related to Planton, visit the [Planton repository](https://github.com/plantonhq/planton).

For Civo-specific questions, consult the [Civo documentation](https://www.civo.com/docs) or [Civo support](https://www.civo.com/support).

