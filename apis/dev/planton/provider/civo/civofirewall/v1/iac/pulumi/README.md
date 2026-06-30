# Civo Firewall - Pulumi Module

This directory contains the Pulumi implementation for managing Civo firewalls. It's designed to be invoked by the Planton CLI, but can also be used standalone for direct Pulumi workflows.

## Overview

The Pulumi module provisions:
- A firewall on Civo (scoped to a network/VPC)
- Inbound (ingress) rules for controlling traffic **to** instances
- Outbound (egress) rules for controlling traffic **from** instances
- Appropriate metadata and labels for resource tracking

## Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/get-started/install/) installed (v3.x)
- [Go](https://golang.org/dl/) 1.21 or later
- Civo account and API token
- Existing Civo network (VPC)

## Quick Start (Standalone Usage)

### 1. Set up stack input

Create a `stack-input.json` file with your configuration:

```json
{
  "target": {
    "apiVersion": "civo.planton.dev/v1",
    "kind": "CivoFirewall",
    "metadata": {
      "name": "web-server-firewall",
      "id": "cifw-web-123"
    },
    "spec": {
      "name": "web-server-fw",
      "networkId": {
        "value": "your-network-id-here"
      },
      "inboundRules": [
        {
          "protocol": "tcp",
          "portRange": "22",
          "cidrs": ["203.0.113.10/32"],
          "action": "allow",
          "label": "SSH from office"
        },
        {
          "protocol": "tcp",
          "portRange": "80",
          "cidrs": ["0.0.0.0/0"],
          "action": "allow",
          "label": "HTTP"
        },
        {
          "protocol": "tcp",
          "portRange": "443",
          "cidrs": ["0.0.0.0/0"],
          "action": "allow",
          "label": "HTTPS"
        }
      ],
      "tags": ["web-server"]
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

# Set Civo region (if needed)
pulumi config set civo:region LON1
```

### 3. Deploy

```bash
# Set stack input as a config value
pulumi config set --path stackInput --plaintext "$(cat stack-input.json)"

# Preview changes
pulumi preview

# Deploy the firewall
pulumi up
```

### 4. View outputs

```bash
# Get all outputs
pulumi stack output

# Get firewall ID
pulumi stack output firewall_id
```

### 5. Update rules

Modify your `stack-input.json` and run:

```bash
pulumi up
```

Pulumi will detect changes and update only the modified rules.

### 6. Destroy

```bash
pulumi destroy
```

**Warning:** This will delete the firewall and all its rules. Instances using this firewall will revert to default firewall (if available).

## Configuration

### Stack Input Structure

The module expects a `CivoFirewallStackInput` protobuf message converted to JSON:

```json
{
  "target": {
    "apiVersion": "civo.planton.dev/v1",
    "kind": "CivoFirewall",
    "metadata": {
      "name": "resource-name",
      "id": "cifw-unique-id",
      "org": "my-org",
      "env": "production"
    },
    "spec": {
      "name": "firewall-name",
      "networkId": {
        "value": "network-uuid"
      },
      "inboundRules": [],
      "outboundRules": [],
      "tags": []
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
    └── firewall.go      # Firewall and rules provisioning
```

## Key Files

### `module/main.go`

The `Resources` function is the single entry point invoked by Planton CLI:

```go
func Resources(
    ctx *pulumi.Context,
    stackInput *civofirewallv1.CivoFirewallStackInput,
) error
```

### `module/firewall.go`

Contains the core logic:
- Translates protobuf rules to Pulumi Civo provider format
- Creates Civo firewall with inbound and outbound rules
- Exports stack outputs

### `module/locals.go`

Initializes local variables and standard Planton labels:
- `resource: true`
- `resource_name: <metadata.name>`
- `resource_kind: CivoFirewall`
- `resource_id: <metadata.id>`
- `organization: <metadata.org>`
- `environment: <metadata.env>`

### `module/outputs.go`

Defines output constants:
- `firewall_id` - Civo firewall UUID

## Stack Outputs

After deployment, the following outputs are available:

| Output | Type | Description | Example |
|--------|------|-------------|---------|
| `firewall_id` | string | Civo's unique firewall identifier | `"abc123-def456-ghi789"` |

Access in your Pulumi program:

```go
firewallId := ctx.Export("firewall_id", createdFirewall.ID())
```

Or via CLI:

```bash
pulumi stack output firewall_id --json
```

## Integration with Planton

When using the Planton CLI, you don't interact with Pulumi directly. The CLI:

1. Reads your `CivoFirewall` YAML manifest
2. Converts it to `CivoFirewallStackInput` protobuf
3. Invokes this Pulumi module's `Resources` function
4. Manages Pulumi stacks automatically
5. Returns outputs in standardized format

Example CLI workflow:

```bash
# Create firewall
planton apply -f firewall.yaml

# View outputs
planton outputs civofirewalls/web-server-firewall

# Update
planton apply -f firewall.yaml

# Delete
planton delete civofirewalls/web-server-firewall
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

### Adding a new rule

1. Update your stack input JSON:

```json
{
  "target": {
    "spec": {
      "inboundRules": [
        {
          "protocol": "tcp",
          "portRange": "8080",
          "cidrs": ["10.0.0.0/8"],
          "action": "allow",
          "label": "Custom app port"
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

### Updating a rule

Modify the rule in your input and run `pulumi up`. The firewall will be updated in-place.

### Deleting a rule

Remove the rule from your `inboundRules` or `outboundRules` array and run `pulumi up`. Pulumi will delete the corresponding rule.

### Changing firewall name

**Warning:** Changing `spec.name` will **replace** the entire firewall (delete old, create new). All rules will be recreated.

```bash
# Preview the replacement
pulumi preview

# Apply
pulumi up
```

## Troubleshooting

### Error: "Firewall already exists"

A firewall with this name already exists in the network. Either:
- Use a different name
- Import the existing firewall:

```bash
pulumi import civo:index/firewall:Firewall firewall <firewall-id>
```

### Rules not taking effect

1. Verify firewall was created:

```bash
pulumi stack output firewall_id
```

2. Check rule syntax:
   - Protocol must be lowercase: `tcp`, `udp`, or `icmp`
   - Port range format: `80` or `8000-9000`
   - CIDR notation: `203.0.113.10/32`

3. Wait for propagation (10-30 seconds)

4. Test directly against Civo:

```bash
civo firewall show <firewall-id>
```

### State drift

If someone manually changed firewall rules in Civo dashboard:

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

Wait a few minutes and retry. For firewalls with many rules (50+), consider splitting into multiple firewalls.

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

7. **Document rules** - Use descriptive labels for every rule

## Security

- **Never commit API tokens** to Git
- Use Pulumi secrets for sensitive data:

```bash
pulumi config set --secret civo:token YOUR_TOKEN
```

- Restrict Civo API token permissions to network/firewall only
- Use separate tokens per environment
- Rotate tokens regularly
- Review firewall rules monthly for unnecessary access

## Rule Management Patterns

### Programmatic Rule Generation

If you need many similar rules, generate them programmatically:

```go
// Example: Allow multiple office IPs
officeIPs := []string{"203.0.113.10/32", "198.51.100.20/32", "192.0.2.30/32"}
rules := []*civofirewallv1.CivoFirewallInboundRule{}

for i, ip := range officeIPs {
    rules = append(rules, &civofirewallv1.CivoFirewallInboundRule{
        Protocol:  "tcp",
        PortRange: "22",
        Cidrs:     []string{ip},
        Action:    "allow",
        Label:     fmt.Sprintf("SSH from office IP %d", i+1),
    })
}
```

### Rule Templates

Create reusable rule templates:

```json
{
  "webServerRules": [
    {"protocol": "tcp", "portRange": "80", "cidrs": ["0.0.0.0/0"], "label": "HTTP"},
    {"protocol": "tcp", "portRange": "443", "cidrs": ["0.0.0.0/0"], "label": "HTTPS"}
  ],
  "databaseRules": [
    {"protocol": "tcp", "portRange": "5432", "cidrs": ["10.0.1.0/24"], "label": "PostgreSQL"}
  ]
}
```

## Advanced Usage

### Programmatic invocation

Import the module in your Go code:

```go
import (
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
    civofirewall "github.com/plantonhq/planton/apis/dev/planton/provider/civo/civofirewall/v1/iac/pulumi/module"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        stackInput := &civofirewallv1.CivoFirewallStackInput{
            // ... your config
        }
        return civofirewall.Resources(ctx, stackInput)
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
    
    return civofirewall.Resources(ctx, stackInput)
})
```

## References

- [Pulumi Civo Provider](https://www.pulumi.com/registry/packages/civo/)
- [Civo Firewall API Documentation](https://www.civo.com/api/firewalls)
- [Planton Documentation](https://github.com/plantonhq/planton)
- [Architecture Overview](overview.md)

## Support

- Issues: [GitHub Issues](https://github.com/plantonhq/planton/issues)
- Pulumi Community: [Pulumi Slack](https://slack.pulumi.com/)
- Civo Support: support@civo.com

