# CivoBucket Pulumi Module

## Overview

This directory contains the Pulumi implementation for provisioning Civo Object Storage buckets. The module translates the `CivoBucketSpec` Protobuf specification into Civo infrastructure resources using the Civo Pulumi provider.

## Module Structure

```
iac/pulumi/
├── main.go              # Pulumi entrypoint (stack initialization)
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Build automation
├── debug.sh             # Local debugging script
├── README.md            # This file
└── module/              # Core module implementation
    ├── main.go          # Module entry point (Resources function)
    ├── locals.go        # Local variables and initialization
    ├── outputs.go       # Output constant definitions
    └── bucket.go        # Bucket resource creation logic
```

## Prerequisites

### Required Tools

- **Go**: 1.21 or later
- **Pulumi CLI**: 3.x or later
- **Civo Account**: With API credentials

### Civo Provider

The module uses the [Pulumi Civo Provider](https://www.pulumi.com/registry/packages/civo/) (`pulumi-civo` v2.x), which wraps the Terraform Civo provider.

Install via:
```bash
pulumi plugin install resource civo v2.x.x
```

## How It Works

### 1. Stack Input

The module receives a `CivoBucketStackInput` containing:
- `CivoBucket` resource (metadata + spec)
- Civo provider configuration (API token, region)

### 2. Resource Creation

The module creates:
1. **ObjectStoreCredential**: Civo-managed access key pair for the bucket
2. **ObjectStore**: The bucket itself with specified region and name

### 3. Stack Outputs

After provisioning, the module exports:
- `bucket_id`: Unique identifier (UUID)
- `endpoint_url`: S3-compatible endpoint URL
- `access_key_secret_ref`: Reference to access key ID
- `secret_key_secret_ref`: Reference to secret access key

These outputs are captured in `CivoBucketStackOutputs` for consumption by applications.

## Local Development

### Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/plantonhq/planton.git
   cd planton/apis/dev/planton/provider/civo/civobucket/v1/iac/pulumi/
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Set Civo credentials**:
   ```bash
   export CIVO_TOKEN=your-civo-api-token
   ```

### Running Locally

Use the provided `debug.sh` script for local testing:

```bash
./debug.sh
```

This script:
- Loads the example manifest from `../../hack/manifest.yaml`
- Initializes a local Pulumi stack
- Runs `pulumi up` with the test configuration
- Displays outputs

**Note**: The debug script uses a temporary stack and won't persist state.

### Manual Execution

For full control:

```bash
# Initialize stack
pulumi stack init dev

# Set Civo token
pulumi config set civo:token $CIVO_TOKEN --secret

# Preview changes
pulumi preview

# Apply
pulumi up

# View outputs
pulumi stack output

# Destroy
pulumi destroy
```

## Module Implementation Details

### `module/main.go`

Entry point for the module. Calls:
1. `initializeLocals()` - Prepares local variables
2. `pulumicivoprovider.Get()` - Initializes Civo provider
3. `bucket()` - Creates bucket resources

### `module/locals.go`

Initializes the `Locals` struct containing:
- `CivoBucket`: The Protobuf spec
- `Metadata`: Cloud resource metadata (name, labels)
- `Labels`: Kubernetes-style labels for organization

### `module/outputs.go`

Defines output constants used in `ctx.Export()`:
- `OpBucketId`: "bucket_id"
- `OpEndpointUrl`: "endpoint_url"
- `OpAccessKeySecretRef`: "access_key_secret_ref"
- `OpSecretKeySecretRef`: "secret_key_secret_ref"

### `module/bucket.go`

Core resource creation logic:

1. **Create Credential**: `civo.NewObjectStoreCredential()`
   - Auto-generates access key pair
   - No explicit key specification required

2. **Create Bucket**: `civo.NewObjectStore()`
   - Uses `bucket_name` from spec
   - Uses `region` from spec
   - Attaches credential from step 1

3. **Handle Versioning** (informational):
   - Logs reminder that versioning must be configured via S3 API
   - Civo provider doesn't expose versioning as a direct attribute

4. **Handle Tags** (informational):
   - Logs that tags aren't currently supported by Civo provider
   - Tags are recorded in metadata but not applied to Civo resource

5. **Export Outputs**: `ctx.Export()`
   - Exports bucket ID, endpoint, and credential references

## Configuration

### Pulumi.yaml

```yaml
name: civobucket
runtime: go
description: Pulumi module for Civo Object Storage bucket provisioning
```

This is a minimal project configuration. The actual settings come from the stack input.

### Provider Configuration

The module expects a `ProviderConfig` with:
```protobuf
message ProviderConfig {
  string civo_token = 1;  // Civo API token
}
```

The token is passed securely via Pulumi config or environment variables.

## Testing

### Unit Tests

Run Go unit tests:
```bash
cd module/
go test -v ./...
```

### Integration Tests

Test with the example manifest:
```bash
# From iac/pulumi/
export CIVO_TOKEN=your-token
./debug.sh
```

Verify:
1. Bucket is created in Civo dashboard
2. Credentials are generated
3. S3 API works:
   ```bash
   aws s3 ls s3://test-bucket --endpoint-url https://objectstore.civo.com
   ```

### Cleanup

```bash
pulumi destroy
```

## Debugging

### Enable Verbose Logging

```bash
pulumi up --logtostderr -v=9
```

### Pulumi Stack Traces

If the stack fails:
```bash
pulumi stack export > stack-state.json
cat stack-state.json | jq '.deployment.resources'
```

### Civo API Debugging

Check Civo API calls:
```bash
export CIVO_DEBUG=1
pulumi up
```

## Known Limitations

1. **Versioning**: Must be configured via S3 API after bucket creation. The Civo provider doesn't expose versioning as a resource attribute.

2. **Tags**: Civo's ObjectStore provider doesn't support tags. Tags in the spec are for organizational purposes only.

3. **Lifecycle Policies**: Not configurable via Civo provider. Use AWS CLI with Civo endpoint to set lifecycle rules.

4. **CORS / Bucket Policies**: Must be configured via S3 API post-deployment.

These limitations reflect the Civo provider's current capabilities, not the module's design.

## Advanced Usage

### Custom Provider Configuration

If you need to override provider settings:

```go
import "github.com/pulumi/pulumi-civo/sdk/v2/go/civo"

civoProvider, err := civo.NewProvider(ctx, "custom-provider", &civo.ProviderArgs{
    Token:  pulumi.String(os.Getenv("CIVO_TOKEN")),
    Region: pulumi.String("LON1"),
})
```

### Multiple Buckets

To provision multiple buckets in a single stack, call `bucket()` multiple times with different names:

```go
bucket1, _ := bucket(ctx, locals1, civoProvider)
bucket2, _ := bucket(ctx, locals2, civoProvider)
```

Ensure bucket names are unique (globally across Civo).

### Post-Deployment Configuration

After provisioning, configure S3-level settings:

```bash
# Enable versioning
aws s3api put-bucket-versioning \
  --bucket my-bucket \
  --versioning-configuration Status=Enabled \
  --endpoint-url https://objectstore.civo.com

# Set lifecycle rule
aws s3api put-bucket-lifecycle-configuration \
  --bucket my-bucket \
  --lifecycle-configuration file://lifecycle.json \
  --endpoint-url https://objectstore.civo.com
```

## Troubleshooting

### Issue: "bucket name already exists"

**Cause**: Bucket names are globally unique across all Civo customers.

**Solution**: Choose a more specific name (e.g., prefix with your company name).

### Issue: "credential not found"

**Cause**: Credential creation failed or is in a different region.

**Solution**: Ensure credentials and bucket are in the same region. Check Civo dashboard for credential status.

### Issue: "401 Unauthorized"

**Cause**: Invalid or expired Civo API token.

**Solution**: Verify token:
```bash
curl -H "Authorization: Bearer $CIVO_TOKEN" https://api.civo.com/v2/regions
```

### Issue: Pulumi state corruption

**Cause**: Interrupted `pulumi up` or network issues.

**Solution**: Refresh state:
```bash
pulumi refresh
```

If still broken, export and manually fix state:
```bash
pulumi stack export > state.json
# Edit state.json
pulumi stack import --file state.json
```

## Performance Considerations

- **Bucket Creation**: Typically 5-15 seconds
- **Credential Generation**: 2-5 seconds
- **Total Provisioning**: ~20-30 seconds for a single bucket

Civo's API is generally fast, but network latency and region selection can affect provisioning time.

## Security Best Practices

1. **Secrets Management**: Store Civo token in Pulumi secrets:
   ```bash
   pulumi config set civo:token $CIVO_TOKEN --secret
   ```

2. **Credential Rotation**: Civo allows multiple credentials per bucket. Rotate by:
   - Creating new credential
   - Updating applications to use new credential
   - Deleting old credential

3. **Least Privilege**: Use separate credentials for each application/service.

4. **Audit Logging**: Enable Civo account audit logs to track bucket access.

## Related Documentation

- **API Specification**: [../../README.md](../../README.md)
- **Examples**: [../../examples.md](../../examples.md)
- **Research**: [../../docs/README.md](../../docs/README.md)
- **Pulumi Civo Provider**: [pulumi.com/registry/packages/civo](https://www.pulumi.com/registry/packages/civo/)
- **Civo Object Storage**: [civo.com/docs/object-stores](https://www.civo.com/docs/object-stores)

## Contributing

To contribute to this module:

1. Follow the [Planton contribution guidelines](../../../../../../../../CONTRIBUTING.md)
2. Ensure all tests pass: `go test ./module/...`
3. Run linters: `golangci-lint run`
4. Update this README if adding new features

## Support

- **Issues**: [GitHub Issues](https://github.com/plantonhq/planton/issues)
- **Discussions**: [GitHub Discussions](https://github.com/plantonhq/planton/discussions)
- **Civo Support**: [support.civo.com](https://support.civo.com)

