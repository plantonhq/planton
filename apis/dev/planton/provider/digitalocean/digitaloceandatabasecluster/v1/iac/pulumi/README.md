# DigitalOcean Database Cluster - Pulumi Module

This Pulumi module provisions and manages fully-managed database clusters on DigitalOcean using the `pulumi-digitalocean` provider.

## Overview

The module implements Planton's `DigitalOceanDatabaseCluster` protobuf spec, providing a type-safe interface for deploying PostgreSQL, MySQL, Redis, and MongoDB clusters with automated backups, patching, and monitoring.

## Module Structure

```
iac/pulumi/
├── main.go              # Pulumi entrypoint
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Build and deployment helpers
├── debug.sh             # Debug script for local testing
├── module/
│   ├── main.go          # Module entry point (Resources function)
│   ├── locals.go        # Local variables and context initialization
│   ├── database_cluster.go  # Database cluster resource creation
│   └── outputs.go       # Output constant definitions
└── README.md            # This file
```

## Prerequisites

- **Pulumi CLI**: Install from [pulumi.com/docs/install](https://www.pulumi.com/docs/install/)
- **Go**: Version 1.21 or later
- **DigitalOcean API Token**: Set via environment variable or Pulumi config

## Quick Start

### 1. Set Up Environment

```bash
# Navigate to the Pulumi module directory
cd apis/dev/planton/provider/digitalocean/digitaloceandatabasecluster/v1/iac/pulumi

# Set your DigitalOcean API token
export DIGITALOCEAN_TOKEN="your-digitalocean-api-token"
```

### 2. Configure Stack Input

Create a `stack-input.yaml`:

**Development PostgreSQL:**
```yaml
provider_config:
  credential_id: "digitalocean-prod-credential"

target:
  spec:
    cluster_name: dev-postgres
    engine: postgres
    engine_version: "16"
    region: nyc3
    size_slug: db-s-1vcpu-1gb
    node_count: 1
    enable_public_connectivity: true
```

**Production HA PostgreSQL:**
```yaml
provider_config:
  credential_id: "digitalocean-prod-credential"

target:
  spec:
    cluster_name: prod-postgres
    engine: postgres
    engine_version: "16"
    region: nyc3
    size_slug: db-s-4vcpu-8gb
    node_count: 3
    vpc:
      value: "vpc-12345678"
    storage_gib: 200
    enable_public_connectivity: false
```

### 3. Deploy

```bash
# Initialize stack
pulumi stack init production

# Preview changes
pulumi preview --stack-input stack-input.yaml

# Deploy
pulumi up --stack-input stack-input.yaml
```

### 4. Retrieve Outputs

```bash
# Get connection details
pulumi stack output connection_uri
pulumi stack output host
pulumi stack output port
pulumi stack output username
pulumi stack output password
```

## Module API

### Input: DigitalOceanDatabaseClusterStackInput

```go
type DigitalOceanDatabaseClusterStackInput struct {
    ProviderConfig *ProviderConfig                 // DigitalOcean credentials
    Target         *DigitalOceanDatabaseCluster    // Cluster specification
}
```

### Output: Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `cluster_id` | string | DigitalOcean cluster UUID |
| `connection_uri` | string (sensitive) | Full connection string |
| `private_uri` | string (sensitive) | VPC-private connection string |
| `host` | string | Public hostname |
| `private_host` | string | VPC-private hostname |
| `port` | number | Database port |
| `database_name` | string | Default database name |
| `username` | string (sensitive) | Admin username |
| `password` | string (sensitive) | Admin password |

## Engine-Specific Implementation

### PostgreSQL Configuration

```go
// In database_cluster.go
if spec.Engine == postgres {
    clusterArgs.Engine = pulumi.String("pg")  // DigitalOcean uses "pg"
    // PgBouncer connection pooling recommended (separate resource)
}
```

**Production Requirements:**
- ✅ Enable PgBouncer connection pool (connection limits are severe: 97 for 4GB RAM)
- ✅ Use private connection string for VPC-internal apps
- ✅ Configure application retry logic for failover events

### MySQL Configuration

```go
if spec.Engine == mysql {
    clusterArgs.Engine = pulumi.String("mysql")
    clusterArgs.Version = pulumi.String(spec.EngineVersion)
}
```

**Production Warnings:**
- ⚠️ No SUPER privileges (cannot install plugins or SET GLOBAL)
- ⚠️ Limited backup portability (see research doc for migration considerations)

### Redis Configuration

```go
if spec.Engine == redis {
    clusterArgs.Engine = pulumi.String("redis")
    // Note: Redis 7+ is actually Valkey (SSPL licensing)
}
```

**Limitations:**
- No Redis Cluster mode (single-instance replication only)
- Treat as ephemeral (use for caching, not primary storage)

### MongoDB Configuration

```go
if spec.Engine == mongodb {
    clusterArgs.Engine = pulumi.String("mongodb")
    // Replica sets only, no sharding support
}
```

**Limitations:**
- No sharding (replica sets only)
- Max 3 nodes

## VPC Integration

### With Literal VPC UUID

```yaml
spec:
  vpc:
    value: "12345678-1234-1234-1234-123456789012"
```

### With Reference to DigitalOceanVpc Resource

```yaml
spec:
  vpc:
    ref:
      kind: DigitalOceanVpc
      name: production-vpc
      field_path: status.outputs.vpc_id
```

The module resolves both patterns automatically.

## Development Workflow

### Local Testing

```bash
# Use debug.sh for rapid iteration
./debug.sh
```

### Building

```bash
make build
```

### Debugging

Enable Pulumi debug logging:

```bash
export PULUMI_DEBUG_COMMANDS=true
export PULUMI_DEBUG_GRPC=debug.log
pulumi up --logtostderr -v=9
```

## Common Issues

### Cluster Creation Timeout

**Symptom**: Pulumi times out waiting for cluster to become ready.

**Cause**: Database cluster provisioning takes 10-15 minutes.

**Solution**: Normal behavior; wait for completion. DigitalOcean is creating nodes, setting up replication, and running initial backups.

### Invalid Size Slug Error

**Symptom**: "invalid size slug for engine"

**Cause**: Not all size slugs are available for all engines.

**Solution**: Verify compatibility:
```bash
doctl databases options sizes --engine pg
```

### Connection Refused After Deployment

**Symptom**: Applications cannot connect to cluster.

**Cause**: Firewall rules not configured.

**Solution**: Deploy separate `DigitalOceanDatabaseFirewall` resource with VPC CIDR or tag-based rules.

## Further Reading

- **Component Overview**: See [../../README.md](../../README.md)
- **Comprehensive Guide**: See [../../docs/README.md](../../docs/README.md)
- **Examples**: See [../../examples.md](../../examples.md)
- **Module Architecture**: See [overview.md](./overview.md)

## Support

For module-specific issues:
1. Check Pulumi logs: `pulumi logs --stack <stack-name>`
2. Enable debug logging: `pulumi up --logtostderr -v=9`
3. Refer to [DigitalOcean Pulumi Docs](https://www.pulumi.com/registry/packages/digitalocean/)

