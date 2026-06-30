# Civo Database Pulumi Module

This directory contains the Pulumi implementation for provisioning Civo managed database instances.

## Overview

The Pulumi module translates the `CivoDatabaseSpec` protobuf definition into Civo Database resources using the official `pulumi-civo` provider. It handles:

- Database instance provisioning
- High availability configuration (replicas)
- Network attachment for private access
- Firewall rule attachment
- Custom storage configuration
- Tag management
- Output generation for downstream dependencies

## Architecture

### Module Structure

```
iac/pulumi/
├── main.go           # Pulumi program entrypoint
├── Pulumi.yaml       # Pulumi project configuration
├── Makefile          # Build and deployment helpers
├── debug.sh          # Local debugging script
├── module/
│   ├── main.go       # Resources() function (primary entrypoint)
│   ├── database.go   # Database resource provisioning logic
│   ├── locals.go     # Local value initialization
│   └── outputs.go    # Output constant definitions
└── README.md         # This file
```

### Key Components

#### 1. Entrypoint (`main.go`)

The top-level `main.go` serves as the Pulumi program entry point. It:
- Receives stack input (contains `CivoDatabaseSpec`)
- Calls the `module.Resources()` function
- Handles errors and exits

#### 2. Module Entry (`module/main.go`)

The `Resources()` function orchestrates resource provisioning:
1. Initializes local values from stack input
2. Configures the Civo provider
3. Creates the database resource
4. Exports outputs

#### 3. Database Resource (`module/database.go`)

The `database()` function provisions the managed database:
- Translates proto enum (`mysql`/`postgres`) to Civo engine strings
- Maps spec fields to `civo.DatabaseArgs`
- Handles optional fields (firewall, storage, tags)
- Calculates total nodes (primary + replicas)
- Exports connection details

#### 4. Locals (`module/locals.go`)

The `Locals` struct consolidates frequently used values:
- Provider configuration
- Database specification
- Derived tags and labels (for consistency)

#### 5. Outputs (`module/outputs.go`)

Defines output constants for cross-resource wiring:
- `OpDatabaseId`: Database instance ID
- `OpHost`: DNS endpoint (recommended for HA)
- `OpPort`: Database port
- `OpUsername`: Master username
- `OpPasswordSecretRef`: Master password (secured)

## Usage

### Prerequisites

- Go 1.21 or later
- Pulumi CLI installed
- Civo API token configured (`CIVO_TOKEN` environment variable)

### Running Locally

```bash
# Set Civo API token
export CIVO_TOKEN="your-civo-api-token"

# Navigate to module directory
cd apis/dev/planton/provider/civo/civodatabase/v1/iac/pulumi

# Initialize Pulumi stack (if first time)
pulumi stack init dev

# Preview changes
pulumi preview

# Deploy
pulumi up

# View outputs
pulumi stack output

# Destroy
pulumi destroy
```

### Debugging

The `debug.sh` script provides a helper for local debugging:

```bash
./debug.sh
```

This script:
- Checks for required environment variables
- Runs `pulumi up` with verbose logging
- Displays outputs after deployment

### Makefile Targets

```bash
# Preview changes
make preview

# Deploy stack
make deploy

# Destroy stack
make destroy

# View outputs
make outputs
```

## Stack Input Structure

The Pulumi module expects a `CivoDatabaseStackInput` with the following structure:

```go
type CivoDatabaseStackInput struct {
    ProviderConfig *civo.CivoProviderConfig  // Civo credentials
    Target         *CivoDatabase              // Database specification
}
```

### Example Stack Input (JSON)

```json
{
  "provider_config": {
    "civo_token": "abc123..."
  },
  "target": {
    "metadata": {
      "name": "my-database",
      "env": "production"
    },
    "spec": {
      "db_instance_name": "prod-db",
      "engine": "postgres",
      "engine_version": "16",
      "region": "lon1",
      "size_slug": "g3.db.large",
      "replicas": 2,
      "network_id": {
        "value": "net-12345678"
      },
      "firewall_ids": [
        {
          "value": "fw-87654321"
        }
      ],
      "storage_gib": 200,
      "tags": ["production", "backend"]
    }
  }
}
```

## Implementation Details

### Engine Translation

The Pulumi module translates protobuf enum values to Civo-compatible strings:

```go
switch locals.CivoDatabase.Spec.Engine {
case civodatabasev1.CivoDatabaseEngine_mysql:
    engineSlug = "mysql"
case civodatabasev1.CivoDatabaseEngine_postgres:
    engineSlug = "postgres"
default:
    return error
}
```

### Node Count Calculation

Civo expects the **total node count** (primary + replicas), not just replicas:

```go
databaseArgs.Nodes = pulumi.Int(int(locals.CivoDatabase.Spec.Replicas) + 1)
```

**Examples**:
- `replicas: 0` → `nodes: 1` (master only)
- `replicas: 2` → `nodes: 3` (master + 2 replicas)

### Firewall Handling

Civo currently supports **one firewall per database**. The module uses the first firewall ID from the list:

```go
if len(locals.CivoDatabase.Spec.FirewallIds) > 0 {
    databaseArgs.FirewallId = pulumi.String(
        locals.CivoDatabase.Spec.FirewallIds[0].GetValue())
}
```

### Storage Configuration

If `storage_gib` is specified, it overrides the default storage bundled with the `size_slug`:

```go
if locals.CivoDatabase.Spec.StorageGib > 0 {
    databaseArgs.SizeGb = pulumi.Int(int(locals.CivoDatabase.Spec.StorageGib))
}
```

### Tags

Tags from both `metadata.tags` and `spec.tags` are collected and applied:

```go
if len(locals.CivoDatabase.Spec.Tags) > 0 {
    for _, t := range locals.CivoDatabase.Spec.Tags {
        locals.CivoTags = append(locals.CivoTags, pulumi.String(t))
    }
}
```

## Outputs

The module exports the following outputs for downstream resource wiring:

| Output Key | Description | Example Value |
|------------|-------------|---------------|
| `database_id` | Unique database instance ID | `db-abc123...` |
| `host` | DNS endpoint (recommended) | `db-abc123.civo.com` |
| `port` | Database connection port | `5432` (PostgreSQL) or `3306` (MySQL) |
| `username` | Master username | `civo` |
| `password` | Master password (sensitive) | `SecurePassword123!` |

### Accessing Outputs

```bash
# Get DNS endpoint
pulumi stack output host

# Get all outputs as JSON
pulumi stack output --json
```

### Using Outputs in Other Resources

Pulumi outputs can be used to wire dependencies between resources:

```go
// Example: Create Kubernetes secret with database credentials
dbSecret := k8s.NewSecret(ctx, "db-credentials", &k8s.SecretArgs{
    StringData: pulumi.StringMap{
        "hostname": database.DnsEndpoint,
        "port":     database.Port.ApplyT(func(p int) string { return fmt.Sprintf("%d", p) }).(pulumi.StringOutput),
        "username": database.Username,
        "password": database.Password,
    },
})
```

## Error Handling

The module performs validation and returns descriptive errors:

### Unsupported Engine

```
Error: unsupported database engine: civo_database_engine_unspecified
```

**Solution**: Ensure `engine` is set to `mysql` or `postgres`.

### Missing Network ID

```
Error: network_id is required
```

**Solution**: Provide a valid `network_id` in the spec.

### Provider Configuration Error

```
Error: failed to setup Civo provider
```

**Solution**: Verify `CIVO_TOKEN` is set and valid.

## Testing

### Unit Tests

Run Go unit tests for the module:

```bash
cd module
go test -v
```

### Integration Tests

Test the full deployment workflow:

```bash
# Set test environment variables
export CIVO_TOKEN="your-test-token"
export PULUMI_STACK="test"

# Run deployment
make deploy

# Verify outputs
make outputs

# Clean up
make destroy
```

## Best Practices

### 1. Use DNS Endpoint

Always export and use the `dns_endpoint` output for application connections, not the static `host`. This ensures automatic failover in HA configurations.

```go
ctx.Export(OpHost, createdDatabase.DnsEndpoint)
```

### 2. Secure Credentials

Mark password outputs as sensitive:

```go
ctx.Export(OpPasswordSecretRef, createdDatabase.Password)  // Pulumi handles sensitivity
```

### 3. Tag Everything

Apply consistent tags for resource organization:

```go
databaseArgs.Tags = locals.CivoTags
```

### 4. Handle Optional Fields

Check for nil/empty values before setting optional fields:

```go
if locals.CivoDatabase.Spec.StorageGib > 0 {
    databaseArgs.SizeGb = pulumi.Int(int(locals.CivoDatabase.Spec.StorageGib))
}
```

### 5. Validate Inputs

Validate critical inputs early in the function:

```go
if locals.CivoDatabase.Spec.Engine == civodatabasev1.CivoDatabaseEngine_civo_database_engine_unspecified {
    return nil, errors.New("engine must be specified")
}
```

## Troubleshooting

### Issue: "Provider not configured"

**Symptom**: Pulumi fails with provider configuration error.

**Solution**: Ensure `CIVO_TOKEN` environment variable is set:

```bash
export CIVO_TOKEN="your-api-token"
```

### Issue: "Resource already exists"

**Symptom**: Database creation fails because name is taken.

**Solution**: Civo database names must be unique per region. Use a different `db_instance_name` or delete the existing database.

### Issue: "Invalid network ID"

**Symptom**: Database creation fails with network error.

**Solution**: Verify the `network_id` exists and is in the same region as the database.

### Issue: "Firewall not found"

**Symptom**: Database creation fails with firewall error.

**Solution**: Ensure the `firewall_id` exists and is associated with the correct network.

## Additional Resources

- **Pulumi Civo Provider**: [pulumi-civo GitHub](https://github.com/pulumi/pulumi-civo)
- **Civo API Documentation**: [Civo API Docs](https://www.civo.com/api)
- **Planton Overview**: See [`../overview.md`](overview.md)
- **Examples**: See [`examples.md`](examples.md)
- **Parent README**: See [`../../README.md`](../../README.md)

