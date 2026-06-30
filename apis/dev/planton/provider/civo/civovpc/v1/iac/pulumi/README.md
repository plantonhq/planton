# CivoVpc Pulumi Module

This directory contains the Pulumi implementation for provisioning Civo VPCs (private networks) using Go.

## Overview

The Pulumi module creates isolated private networks on Civo cloud with the following capabilities:

- **Network Creation**: Provisions Civo networks with configurable CIDR blocks
- **Auto-Allocation**: Supports CIDR auto-allocation when not explicitly specified
- **Regional Deployment**: Creates networks in specific Civo regions (LON1, NYC1, FRA1, etc.)
- **Stack Outputs**: Exports network ID and CIDR block for use by dependent resources

## Module Structure

```
pulumi/
├── main.go              # Entry point - loads stack input and calls module
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Build and deployment tasks
├── debug.sh             # Helper script for local debugging
└── module/
    ├── main.go          # Module entry point - orchestrates resource creation
    ├── locals.go        # Locals initialization (labels, metadata)
    ├── vpc.go           # VPC resource creation and configuration
    └── outputs.go       # Output constants definition
```

## Key Files

### main.go (Entry Point)

Loads the `CivoVpcStackInput` and delegates to the module:

```go
pulumi.Run(func(ctx *pulumi.Context) error {
    stackInput := &civovpcv1.CivoVpcStackInput{}
    
    if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
        return errors.Wrap(err, "failed to load stack-input")
    }
    
    return module.Resources(ctx, stackInput)
})
```

### module/main.go (Module Entry Point)

Orchestrates the resource creation workflow:

1. Initialize locals (metadata, labels)
2. Create Civo provider from credentials
3. Create VPC network resource
4. Export stack outputs

### module/vpc.go (VPC Resource)

Creates the `civo.Network` resource with:

- **Label**: Network name from spec
- **Region**: Target Civo region
- **CIDR Block**: Optional IPv4 CIDR (auto-allocated if omitted)
- **Exports**: `network_id` and `cidr_block` outputs

**Note:** The Pulumi Civo provider (v2.4.8) has limited support for:
- `is_default_for_region` field (logs warning, not applied)
- `description` field (not exposed by provider, logged only)

### module/locals.go (Locals Initialization)

Builds the `Locals` struct with:

- **CivoVpc**: Target resource specification
- **CivoProviderConfig**: Provider configuration and credentials
- **CivoLabels**: Standard Planton labels for resource tracking

## Stack Inputs

The module expects a `CivoVpcStackInput` protobuf message containing:

```protobuf
message CivoVpcStackInput {
  CivoVpc target = 1;                           // Target resource spec
  CivoProviderConfig provider_config = 2;       // Provider config and credentials
}
```

## Stack Outputs

The module exports the following outputs:

| Output | Type | Description |
|--------|------|-------------|
| `network_id` | string | Unique ID of the created network (used by clusters, instances, firewalls) |
| `cidr_block` | string | IPv4 CIDR block of the network (auto-allocated if not specified) |

**Note:** `created_at_rfc3339` is defined in the protobuf spec but not available from the Civo provider.

## Usage

### Prerequisites

1. Civo API credentials configured in Planton
2. Valid `CivoVpcSpec` protobuf definition
3. Pulumi Go SDK and Civo provider installed

### Local Development

#### 1. Run Locally

```bash
# Set up environment
export PULUMI_CONFIG_PASSPHRASE="your-passphrase"

# Run Pulumi
make run
```

#### 2. Debug Mode

```bash
# Use the debug helper script
./debug.sh
```

The debug script:
- Loads stack input from standard input or a file
- Sets up Pulumi environment variables
- Runs `pulumi up` in preview mode

#### 3. Manual Execution

```bash
# Preview changes
pulumi preview

# Apply changes
pulumi up

# Destroy resources
pulumi destroy
```

## Example Stack Input

```yaml
target:
  apiVersion: civo.planton.dev/v1
  kind: CivoVpc
  metadata:
    name: prod-main-network
    id: civpc-abc123
    org: myorg
    env: production
  spec:
    civo_credential_id: civo-cred-123
    network_name: prod-main-network
    region: NYC1
    ip_range_cidr: "10.20.1.0/24"
    description: "Production network (NYC1)"

provider_config:
  civo_credential_id: civo-cred-123
```

## Provider Limitations

The Pulumi Civo provider (bridged from Terraform) has some limitations:

### 1. Default Network Flag Not Supported (v2.4.8)

```go
if locals.CivoVpc.Spec.IsDefaultForRegion {
    ctx.Log.Warn(fmt.Sprintf(
        "Network '%s' has 'is_default_for_region' set to true, but this is not supported by "+
        "Pulumi Civo SDK v2.4.8. To set as default, use Civo CLI: 'civo network default <network-id>'",
        locals.CivoVpc.Spec.NetworkName,
    ), nil)
}
```

**Workaround:** Use the Civo CLI to set a network as default after creation:
```bash
civo network default <network-id>
```

### 2. Description Field Not Exposed

The Civo Network resource doesn't support a `description` field. Any description specified in the spec is logged but not applied to the Civo resource.

**Workaround:** Use Planton metadata or resource labels for documentation.

### 3. No Created Timestamp

The provider doesn't expose the network's `created_at` timestamp. This field is defined in `CivoVpcStackOutputs` for API consistency but will be empty.

## Build and Test

### Build

```bash
# Build the Pulumi program
make build
```

### Test Locally

```bash
# 1. Create a test stack input file
cat > test-input.yaml <<EOF
target:
  apiVersion: civo.planton.dev/v1
  kind: CivoVpc
  metadata:
    name: test-network
  spec:
    civo_credential_id: civo-cred-test
    network_name: test-network
    region: LON1
provider_config:
  civo_credential_id: civo-cred-test
EOF

# 2. Run Pulumi in preview mode
pulumi preview --stack test
```

### Clean Up

```bash
# Remove all resources
pulumi destroy --yes

# Remove stack
pulumi stack rm test
```

## Integration with Planton

This module is invoked by Planton's orchestration layer:

1. **API Request**: User creates/updates a `CivoVpc` resource via API
2. **Stack Input**: Planton generates `CivoVpcStackInput` protobuf
3. **Pulumi Execution**: This module is invoked with the stack input
4. **Outputs**: `network_id` and `cidr_block` are captured and stored in resource status
5. **Dependent Resources**: Other resources (clusters, firewalls) reference the `network_id`

## Labels and Metadata

The module applies standard Planton labels for resource tracking:

```go
locals.CivoLabels = map[string]string{
    civolabelkeys.Resource:     strconv.FormatBool(true),
    civolabelkeys.ResourceName: locals.CivoVpc.Metadata.Name,
    civolabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_CivoVpc.String(),
    civolabelkeys.Organization: locals.CivoVpc.Metadata.Org,
    civolabelkeys.Environment:  locals.CivoVpc.Metadata.Env,
    civolabelkeys.ResourceId:   locals.CivoVpc.Metadata.Id,
}
```

**Note:** Civo Network resources don't currently support labels/tags via the provider. These labels are stored in Planton metadata only.

## Error Handling

The module uses error wrapping for clear error messages:

```go
if err != nil {
    return errors.Wrap(err, "failed to create civo network")
}
```

Common errors:

- **Invalid credentials**: Check `civo_credential_id` is correct
- **CIDR conflicts**: Verify CIDR block doesn't overlap with existing networks
- **Region not found**: Ensure region code is valid (LON1, NYC1, FRA1, etc.)
- **Quota exceeded**: Check Civo account network quota

## Dependencies

### Go Dependencies

```
github.com/pulumi/pulumi-civo/sdk/v2/go/civo  # Civo provider
github.com/pulumi/pulumi/sdk/v3/go/pulumi     # Pulumi Go SDK
github.com/pkg/errors                         # Error wrapping
```

### Protobuf Dependencies

```
github.com/plantonhq/planton/apis/dev/planton/provider/civo/civovpc/v1
github.com/plantonhq/planton/apis/dev/planton/provider/civo
```

## Further Reading

- **Overview**: [overview.md](./overview.md) - Architecture and design decisions
- **User Guide**: [../../README.md](../../README.md) - CivoVpc resource documentation
- **Examples**: [../../examples.md](../../examples.md) - Configuration examples
- **Research**: [../../docs/README.md](../../docs/README.md) - Comprehensive deployment guide

## Troubleshooting

### Issue: "Failed to setup civo provider"

**Cause:** Invalid or missing Civo credentials

**Solution:**
1. Verify `civo_credential_id` is correct
2. Ensure credentials exist in Planton
3. Check API token has required permissions

### Issue: "CIDR block already in use"

**Cause:** Specified CIDR overlaps with existing network

**Solution:**
1. Use `civo network list` to check existing networks
2. Choose a different CIDR block
3. Consider using auto-allocation (omit `ip_range_cidr`)

### Issue: "Region not found"

**Cause:** Invalid region code

**Solution:**
1. Use `civo region list` to see available regions
2. Common regions: LON1, NYC1, FRA1, PHX1, SYD1
3. Region codes are case-sensitive (use uppercase)

### Issue: Network created but output is empty

**Cause:** Provider limitation (created_at timestamp not exposed)

**Solution:**
- This is expected behavior
- `network_id` and `cidr_block` will be populated
- `created_at_rfc3339` will be empty (provider limitation)

---

**Maintained by:** Planton  
**Last Updated:** 2025-11-21

