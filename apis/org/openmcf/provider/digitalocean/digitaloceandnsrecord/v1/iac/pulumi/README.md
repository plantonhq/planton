# DigitalOcean DNS Record - Pulumi Module

This Pulumi module provisions a DigitalOcean DNS record.

## Usage

This module is designed to be used with OpenMCF's CLI. The manifest is passed via environment variable.

### Via OpenMCF CLI

```bash
planton pulumi up -f manifest.yaml
```

### Direct Pulumi Usage

Set the `STACK_INPUT` environment variable with the JSON-encoded stack input:

```bash
export STACK_INPUT='{"target":{"apiVersion":"digital-ocean.openmcf.org/v1","kind":"DigitalOceanDnsRecord","metadata":{"name":"www-record"},"spec":{"domain":"example.com","name":"www","type":"A","value":"192.0.2.1"}},"providerConfig":{"apiToken":"YOUR_TOKEN","defaultRegion":"nyc1"}}'

pulumi up
```

## Required Environment Variables

- `STACK_INPUT`: JSON-encoded DigitalOceanDnsRecordStackInput

## Provider Configuration

The module uses the DigitalOcean provider configured via `providerConfig`:

- `apiToken`: DigitalOcean API token with write access
- `defaultRegion`: Default region for resources

## Outputs

| Output | Description |
|--------|-------------|
| `record_id` | The unique ID of the created DNS record |
| `hostname` | The fully qualified hostname |
| `record_type` | The type of DNS record created |
| `domain` | The domain where the record was created |
| `ttl_seconds` | The TTL applied to the record |

## Development

```bash
# Build the module
make build

# Update dependencies
make update-deps
```

## Troubleshooting

### Common Issues

1. **Authentication Error**: Ensure `apiToken` is valid and has DNS write permissions
2. **Domain Not Found**: The domain must exist in your DigitalOcean account before creating records
3. **Invalid Record Type**: Supported types: A, AAAA, CNAME, MX, TXT, SRV, NS, CAA

### Debug Mode

Set Pulumi logging for detailed output:

```bash
export PULUMI_LOG_TO_STDERR=1
export PULUMI_DEBUG_COMMANDS=1
pulumi up
```
