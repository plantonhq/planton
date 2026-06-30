# Cloudflare D1 Database - Pulumi Module

Pulumi module for provisioning Cloudflare D1 databases using Go.

## Overview

This module implements the CloudflareD1Database resource using Pulumi's Go SDK and the official Cloudflare provider (v6+). It translates the protobuf-defined `CloudflareD1DatabaseSpec` into Pulumi resource arguments and manages the database lifecycle.

## Module Structure

```
iac/pulumi/
├── main.go              # Entrypoint - loads stack input and calls module
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile            # Build and deployment targets
├── debug.sh            # Debug helper script
└── module/
    ├── main.go         # Module entry point
    ├── locals.go       # Locals initialization
    ├── database.go     # D1 database resource creation
    └── outputs.go      # Output constants
```

## Inputs

The module accepts a `CloudflareD1DatabaseStackInput` protobuf message containing:

### Provider Configuration

```go
ProviderConfig *CloudflareProviderConfig
```

The Cloudflare provider configuration, including:
- `credential_id`: Reference to Cloudflare API token credential
- (Provider automatically uses `CLOUDFLARE_API_TOKEN` environment variable)

### Spec (CloudflareD1DatabaseSpec)

```go
CloudflareD1Database *CloudflareD1Database
```

The CloudflareD1Database resource containing:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `account_id` | string | Yes | Cloudflare account ID |
| `database_name` | string | Yes | Database name (max 64 chars) |
| `region` | CloudflareD1Region enum | No | Primary location hint (weur, eeur, apac, oc, wnam, enam) |
| `read_replication` | CloudflareD1ReadReplication | No | Read replication configuration (mode: "auto" or "disabled") |

## Outputs

The module exports the following stack outputs:

| Output | Type | Description |
|--------|------|-------------|
| `database-id` | string | The unique identifier of the created D1 database |
| `database-name` | string | The name of the database (same as input) |
| `connection-string` | string | Connection string (currently empty - D1 uses Worker bindings) |

## Resource Mapping

The module maps the protobuf spec to Pulumi's `cloudflare.D1Database` resource:

| Spec Field | Pulumi Argument | Notes |
|------------|-----------------|-------|
| `account_id` | `AccountId` | Required |
| `database_name` | `Name` | Required |
| `region` | `PrimaryLocationHint` | Optional, enum mapped to string |
| `read_replication.mode` | `ReadReplication.Mode` | Optional, nested object |

### Region Mapping

The module converts the CloudflareD1Region enum to the string values expected by the Cloudflare API:

```go
CloudflareD1Region_weur → "weur"
CloudflareD1Region_eeur → "eeur"
CloudflareD1Region_apac → "apac"
CloudflareD1Region_oc   → "oc"
CloudflareD1Region_wnam → "wnam"
CloudflareD1Region_enam → "enam"
```

If `region` is unspecified or `cloudflare_d1_region_unspecified`, the field is omitted, allowing Cloudflare to select a default location.

## Usage

### Via Planton CLI

The typical usage is through the Planton CLI, which handles stack input creation and Pulumi execution:

```bash
planton apply -f database.yaml
```

### Direct Pulumi Execution

For debugging or manual execution:

1. **Set Environment Variables**:
   ```bash
   export CLOUDFLARE_API_TOKEN="your-cloudflare-api-token"
   ```

2. **Create Stack Input File**:
   Create a `stack-input.json` file with the CloudflareD1DatabaseStackInput protobuf structure serialized to JSON.

3. **Run Pulumi**:
   ```bash
   pulumi up
   ```

### Debug Mode

Use the provided `debug.sh` script to run Pulumi with verbose logging:

```bash
./debug.sh
```

This sets `PULUMI_LOG_LEVEL=debug` and runs `pulumi up` with detailed output.

## Implementation Details

### Locals Initialization

The `initializeLocals` function copies the stack input into a `Locals` struct for easy access throughout the module:

```go
type Locals struct {
	CloudflareProviderConfig *pulumicloudflareprovider.CloudflareCredential
	CloudflareD1Database     *cloudflared1databasev1.CloudflareD1Database
}
```

### Provider Setup

The module uses a shared Pulumi Cloudflare provider helper to instantiate the provider with credentials:

```go
cloudflareProvider, err := pulumicloudflareprovider.Get(ctx, stackInput.ProviderConfig)
```

### Database Resource Creation

The `database` function creates the `cloudflare.D1Database` resource:

1. Builds `D1DatabaseArgs` from spec fields
2. Optionally adds `PrimaryLocationHint` if region is specified
3. Optionally adds `ReadReplication` if read_replication is specified
4. Creates the resource with the Cloudflare provider
5. Exports stack outputs

### Read Replication Handling

If `spec.ReadReplication` is non-nil, the module creates a `D1DatabaseReadReplicationArgs` struct:

```go
if locals.CloudflareD1Database.Spec.ReadReplication != nil {
	d1Args.ReadReplication = &cloudflare.D1DatabaseReadReplicationArgs{
		Mode: pulumi.String(locals.CloudflareD1Database.Spec.ReadReplication.Mode),
	}
}
```

## Error Handling

The module uses `github.com/pkg/errors.Wrap` for descriptive error propagation:

```go
if err != nil {
	return errors.Wrap(err, "failed to create cloudflare d1 database")
}
```

Errors bubble up to the Pulumi CLI, which displays them to the user.

## Testing

### Unit Tests

The module does not include unit tests (Pulumi modules are typically integration-tested). The component-level tests in `v1/spec_test.go` validate the protobuf spec.

### Integration Testing

To test the module end-to-end:

1. Create a test manifest in `hack/manifest.yaml`
2. Run `planton apply -f hack/manifest.yaml`
3. Verify the database is created in the Cloudflare dashboard
4. Verify outputs: `planton output database-id`
5. Clean up: `planton destroy -f hack/manifest.yaml`

## Dependencies

- **Pulumi SDK**: `github.com/pulumi/pulumi/sdk/v3/go/pulumi`
- **Cloudflare Provider**: `github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare`
- **Planton Proto**: `github.com/plantonhq/planton/apis/...`
- **Errors Package**: `github.com/pkg/errors`

## Cloudflare Provider Version

The module uses Pulumi's Cloudflare provider **v6.4.1+**, which includes support for:
- D1 databases
- Primary location hints
- Read replication (Beta)

**Note**: The provider does not expose a `connection_string` output for D1 databases (D1 uses Worker bindings instead). The module exports an empty string to satisfy the protobuf schema.

## Limitations

### What This Module Does

- ✅ Provisions the D1 database resource (the "container")
- ✅ Configures region and read replication
- ✅ Exports database ID and name

### What This Module Does NOT Do

- ❌ Create tables or manage schema (use Wrangler CLI migrations)
- ❌ Configure Worker bindings (use `wrangler.toml` or Terraform `cloudflare_workers_script`)
- ❌ Manage data migrations or seed data

## Troubleshooting

### "failed to setup cloudflare provider"

**Cause**: Missing or invalid `CLOUDFLARE_API_TOKEN` environment variable.

**Fix**: Set the token:
```bash
export CLOUDFLARE_API_TOKEN="your-token"
```

### "failed to create cloudflare d1 database"

**Cause**: Invalid `account_id`, `database_name` already exists, or API permission issue.

**Fix**:
- Verify `account_id` is correct
- Choose a unique `database_name`
- Ensure API token has D1 permissions

### Empty Connection String

**Expected Behavior**: D1 does not use traditional connection strings. Worker bindings are configured separately via `wrangler.toml`.

## Further Reading

- **Cloudflare Provider Docs**: [pulumi.com/registry/packages/cloudflare](https://www.pulumi.com/registry/packages/cloudflare/)
- **D1 API Reference**: [developers.cloudflare.com/api/operations/cloudflare-d1](https://developers.cloudflare.com/api/operations/cloudflare-d1)
- **Pulumi Go Guide**: [pulumi.com/docs/languages-sdks/go](https://www.pulumi.com/docs/languages-sdks/go/)
- **Architecture Overview**: [../../docs/README.md](../../docs/README.md)

## Support

For issues specific to this Pulumi module, check:
1. Component tests pass: `go test ./v1/`
2. Pulumi build succeeds: `make build`
3. Stack input is valid: Validate against protobuf schema

For general Cloudflare D1 questions, see [../../README.md](../../README.md).

