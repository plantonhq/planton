# CivoComputeInstance Pulumi Module

## Overview

This Pulumi module provisions Civo compute instances with networking, security, and storage configuration. It translates the `CivoComputeInstanceSpec` Protobuf specification into Civo infrastructure resources.

## Module Structure

```
iac/pulumi/
├── main.go              # Pulumi entrypoint
├── Pulumi.yaml          # Project configuration
├── Makefile             # Build automation
├── debug.sh             # Local debugging
├── README.md            # This file
└── module/
    ├── main.go          # Module entry point
    ├── locals.go        # Local variables
    ├── outputs.go       # Output constants
    └── instance.go      # Instance resource creation
```

## Prerequisites

- **Go**: 1.21+
- **Pulumi CLI**: 3.x+
- **Civo Account**: With API token

## How It Works

### Input

`CivoComputeInstanceStackInput` containing:
- `CivoComputeInstance` resource (metadata + spec)
- Civo provider configuration (API token)

### Processing

1. Initialize locals (metadata, labels)
2. Create Civo provider
3. Create instance with:
   - Required: hostname, size, region, image, network
   - Optional: SSH keys, firewalls, volumes, reserved IP, tags, user-data

### Output

Exports:
- `instance_id`: UUID
- `public_ipv4`: Public IP address
- `private_ipv4`: Private IP address
- `status`: Instance state
- `created_at`: Creation timestamp

## Local Development

### Setup

```bash
cd apis/dev/planton/provider/civo/civocomputeinstance/v1/iac/pulumi/
export CIVO_TOKEN=your-api-token
./debug.sh
```

### Manual Execution

```bash
pulumi stack init dev
pulumi config set civo:token $CIVO_TOKEN --secret
pulumi up
pulumi stack output
pulumi destroy
```

## Implementation Details

### `instance.go`

Core resource creation:

1. **Build args**: Map spec fields to `civo.InstanceArgs`
2. **Handle optionals**: SSH keys, firewalls, volumes, reserved IP, tags, user-data
3. **Create instance**: `civo.NewInstance()`
4. **Export outputs**: ID, IPs, status, created_at

### Civo Provider Specifics

- **Single SSH key**: Civo allows one SSH key per instance (uses first from list)
- **Single firewall**: Civo allows one firewall per instance (uses first from list)
- **Multiple volumes**: Can attach multiple volumes
- **Tags**: Supported as string array

## Security

1. **API Token**: Store as Pulumi secret
   ```bash
   pulumi config set civo:token $CIVO_TOKEN --secret
   ```

2. **Private Keys**: Never commit SSH private keys
3. **User Data**: Avoid secrets in cloud-init (use secret stores)

## Testing

### Unit Tests

```bash
cd ../../  # Back to v1/
go test -v
```

### Integration Test

```bash
export CIVO_TOKEN=your-token
./debug.sh
# Verify instance in Civo dashboard
```

## Troubleshooting

### Issue: Instance creation fails

**Cause**: Invalid image/size/network for region  
**Solution**: List available options:
```bash
civo size list --region NYC1
civo diskimage list --region NYC1
civo network list --region NYC1
```

### Issue: Firewall not applied

**Cause**: Firewall in different network  
**Solution**: Ensure firewall and instance in same network

## Performance

- **Provisioning**: ~60 seconds
- **With cloud-init**: +30-120 seconds (depends on script)
- **Total**: 1-3 minutes typical

## Related Documentation

- **API**: [../../README.md](../../README.md)
- **Examples**: [../../examples.md](../../examples.md)
- **Research**: [../../docs/README.md](../../docs/README.md)
- **Pulumi Civo**: [pulumi.com/registry/packages/civo](https://www.pulumi.com/registry/packages/civo/)

## Support

- **Issues**: [GitHub Issues](https://github.com/plantonhq/planton/issues)
- **Civo Support**: [support.civo.com](https://support.civo.com)
