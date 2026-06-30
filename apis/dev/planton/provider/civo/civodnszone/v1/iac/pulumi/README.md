# Civo DNS Zone - Pulumi Module

This directory contains the Pulumi implementation for managing Civo DNS zones. It's designed to be invoked by the Planton CLI, but can also be used standalone for direct Pulumi workflows.

## Overview

The Pulumi module provisions:
- A DNS zone (domain) on Civo
- DNS records (A, AAAA, CNAME, MX, TXT, SRV) within the zone
- Appropriate metadata and labels for resource tracking

## Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/get-started/install/) installed (v3.x)
- [Go](https://golang.org/dl/) 1.21 or later
- Civo account and API token
- Domain registered and ready to use

## Quick Start (Standalone Usage)

### 1. Set up stack input

Create a `stack-input.json` file with your configuration:

```json
{
  "target": {
    "apiVersion": "civo.planton.dev/v1",
    "kind": "CivoDnsZone",
    "metadata": {
      "name": "example-zone",
      "id": "cidns-example-123"
    },
    "spec": {
      "domainName": "example.com",
      "records": [
        {
          "name": "@",
          "type": "A",
          "values": [
            {
              "value": "198.51.100.42"
            }
          ],
          "ttlSeconds": 3600
        },
        {
          "name": "www",
          "type": "CNAME",
          "values": [
            {
              "value": "example.com"
            }
          ],
          "ttlSeconds": 3600
        }
      ]
    }
  },
  "providerConfig": {
    "credential": {
      "credentialType": "API_TOKEN",
      "apiToken": "YOUR_CIVO_API_TOKEN"
    }
  }
}
```

### 2. Initialize Pulumi stack

```bash
# Navigate to this directory
cd iac/pulumi

# Initialize a new Pulumi stack
pulumi stack init dev

# Set Civo region (if needed for provider config)
pulumi config set civo:region LON1
```

### 3. Deploy

```bash
# Set stack input as a config value
pulumi config set --path stackInput --plaintext "$(cat stack-input.json)"

# Preview changes
pulumi preview

# Deploy the zone and records
pulumi up
```

### 4. View outputs

```bash
# Get all outputs
pulumi stack output

# Get specific output
pulumi stack output zone_name
pulumi stack output zone_id
pulumi stack output name_servers
```

### 5. Update records

Modify your `stack-input.json` and run:

```bash
pulumi up
```

Pulumi will detect changes and update only the modified records.

### 6. Destroy

```bash
pulumi destroy
```

**Warning:** This will delete the DNS zone and all records. The domain itself remains registered at your registrar.

## Configuration

### Stack Input Structure

The module expects a `CivoDnsZoneStackInput` protobuf message converted to JSON:

```json
{
  "target": {
    "apiVersion": "civo.planton.dev/v1",
    "kind": "CivoDnsZone",
    "metadata": {
      "name": "resource-name",
      "id": "cidns-unique-id",
      "org": "my-org",
      "env": "production"
    },
    "spec": {
      "domainName": "example.com",
      "records": []
    }
  },
  "providerConfig": {
    "credential": {
      "credentialType": "API_TOKEN",
      "apiToken": "your-civo-api-token"
    }
  }
}
```

### Environment Variables

Alternatively, set Civo credentials via environment:

```bash
export CIVO_TOKEN="your-civo-api-token"
```

Then omit `providerConfig.credential` from stack input.

## Module Structure

```
iac/pulumi/
├── main.go              # Pulumi entrypoint
├── Pulumi.yaml          # Project configuration
├── Makefile             # Build and test commands
├── debug.sh             # Debug helper script
├── README.md            # This file
├── overview.md          # Architecture documentation
└── module/
    ├── main.go          # Module entry point (Resources function)
    ├── locals.go        # Local variables and initialization
    ├── outputs.go       # Output constants
    └── dns_zone.go      # DNS zone and records provisioning
```

## Key Files

### `module/main.go`

The `Resources` function is the single entry point invoked by Planton CLI:

```go
func Resources(
    ctx *pulumi.Context,
    stackInput *civodnszonev1.CivoDnsZoneStackInput,
) error
```

### `module/dns_zone.go`

Contains the core logic:
- Creates Civo DNS domain
- Creates DNS records (one Pulumi resource per record value)
- Exports stack outputs

### `module/locals.go`

Initializes local variables and standard Planton labels:
- `resource: true`
- `resource_name: <metadata.name>`
- `resource_kind: CivoDnsZone`
- `resource_id: <metadata.id>`
- `organization: <metadata.org>`
- `environment: <metadata.env>`

### `module/outputs.go`

Defines output constants:
- `zone_name` - Domain name
- `zone_id` - Civo zone UUID
- `name_servers` - List of Civo nameservers

## Stack Outputs

After deployment, the following outputs are available:

| Output | Type | Description | Example |
|--------|------|-------------|---------|
| `zone_name` | string | The domain name | `"example.com"` |
| `zone_id` | string | Civo's unique zone identifier | `"abc123-def456-ghi789"` |
| `name_servers` | []string | Nameservers to configure at registrar | `["ns0.civo.com", "ns1.civo.com", "ns2.civo.com"]` |

Access in your Pulumi program:

```go
zoneName := ctx.Export("zone_name", pulumi.String("example.com"))
```

Or via CLI:

```bash
pulumi stack output zone_name --json
```

## Integration with Planton

When using the Planton CLI, you don't interact with Pulumi directly. The CLI:

1. Reads your `CivoDnsZone` YAML manifest
2. Converts it to `CivoDnsZoneStackInput` protobuf
3. Invokes this Pulumi module's `Resources` function
4. Manages Pulumi stacks automatically
5. Returns outputs in standardized format

Example CLI workflow:

```bash
# Create DNS zone
planton apply -f dns-zone.yaml

# View outputs
planton outputs civodnszones/example-zone

# Update
planton apply -f dns-zone.yaml

# Delete
planton delete civodnszones/example-zone
```

## Development

### Build and test

```bash
# Build module
make build

# Run tests
make test

# Format code
make fmt

# Lint code
make lint
```

### Debug

Use the provided debug script:

```bash
# Set up debug environment
export CIVO_TOKEN="your-api-token"
export DEBUG_STACK_INPUT="$(cat test-input.json)"

# Run debug script
./debug.sh
```

This creates a temporary Pulumi stack, runs the module, and cleans up.

### Local testing with Pulumi

```bash
# Use local stack
pulumi stack init local-test

# Set test configuration
pulumi config set --path stackInput "$(cat test-input.json)"

# Run with detailed logging
pulumi up --logtostderr -v=9

# Clean up
pulumi destroy
pulumi stack rm local-test
```

## Common Operations

### Adding a new DNS record

1. Update your stack input JSON:

```json
{
  "target": {
    "spec": {
      "records": [
        {
          "name": "new-subdomain",
          "type": "A",
          "values": [{"value": "192.0.2.100"}],
          "ttlSeconds": 3600
        }
      ]
    }
  }
}
```

2. Apply:

```bash
pulumi up
```

### Updating TTL values

Modify `ttlSeconds` in your input and run `pulumi up`. Existing records will be updated in-place.

### Deleting a record

Remove the record from your `records` array and run `pulumi up`. Pulumi will delete the corresponding DNS record.

### Changing domain name

**Warning:** Changing `domainName` will **delete** the existing zone and create a new one. All records will be recreated.

```bash
# Preview the replacement
pulumi preview

# Apply
pulumi up
```

## Troubleshooting

### Error: "Domain already exists"

The domain is already registered in Civo DNS (possibly in another account or region). Check:

```bash
# Via CLI
civo domain list

# Via API
curl -H "Authorization: Bearer $CIVO_TOKEN" https://api.civo.com/v2/dns
```

### Records not resolving

1. Verify zone was created:

```bash
pulumi stack output zone_id
```

2. Check nameservers at your registrar match:

```bash
pulumi stack output name_servers
```

3. Test resolution directly against Civo NS:

```bash
dig @ns0.civo.com example.com
```

4. Wait for propagation (up to 48 hours)

### State drift

If someone manually changed DNS records in Civo dashboard:

```bash
# Refresh Pulumi state
pulumi refresh

# Review differences
pulumi preview

# Reapply desired state
pulumi up
```

### Rate limiting

Civo API has rate limits. If you hit them:

```
Error: 429 Too Many Requests
```

Wait a few minutes and retry. For large zones (100+ records), consider batching changes.

## Best Practices

1. **Version control** - Check Pulumi stack files into Git (exclude secrets)

2. **Use remote state** - Configure Pulumi backend for team collaboration:

```bash
# S3 backend
pulumi login s3://my-pulumi-state-bucket

# Pulumi Cloud (recommended)
pulumi login
```

3. **Separate stacks per environment**:

```bash
pulumi stack init production
pulumi stack init staging
pulumi stack init development
```

4. **Protect production stacks**:

```bash
pulumi stack select production
pulumi stack change-secrets-provider
```

5. **Tag resources** - Use metadata fields (`org`, `env`) for cost tracking and organization

6. **Test changes** - Always run `pulumi preview` before `up`

## Security

- **Never commit API tokens** to Git
- Use Pulumi secrets for sensitive data:

```bash
pulumi config set --secret civo:token YOUR_TOKEN
```

- Restrict Civo API token permissions to DNS only
- Use separate tokens per environment
- Rotate tokens regularly

## Advanced Usage

### Programmatic invocation

Import the module in your Go code:

```go
import (
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
    civodnszone "github.com/plantonhq/planton/apis/dev/planton/provider/civo/civodnszone/v1/iac/pulumi/module"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        stackInput := &civodnszonev1.CivoDnsZoneStackInput{
            // ... your config
        }
        return civodnszone.Resources(ctx, stackInput)
    })
}
```

### Custom Pulumi transforms

Apply custom transformations to all resources:

```go
pulumi.Run(func(ctx *pulumi.Context) error {
    // Add custom tags to all resources
    ctx.RegisterStackTransformation(func(args *pulumi.ResourceTransformationArgs) *pulumi.ResourceTransformationResult {
        // Custom logic
        return &pulumi.ResourceTransformationResult{
            Props: args.Props,
            Opts:  args.Opts,
        }
    })
    
    return civodnszone.Resources(ctx, stackInput)
})
```

## References

- [Pulumi Civo Provider](https://www.pulumi.com/registry/packages/civo/)
- [Civo DNS API Documentation](https://www.civo.com/api/dns)
- [Planton Documentation](https://github.com/plantonhq/planton)
- [Architecture Overview](overview.md)

## Support

- Issues: [GitHub Issues](https://github.com/plantonhq/planton/issues)
- Pulumi Community: [Pulumi Slack](https://slack.pulumi.com/)
- Civo Support: support@civo.com

