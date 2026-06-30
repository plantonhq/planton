# CloudflareDnsRecord Pulumi Module

This Pulumi module provisions a Cloudflare DNS record.

## Prerequisites

- Go 1.21+
- Pulumi CLI
- Cloudflare API token with DNS:Edit permissions

## Installation

Install required Pulumi plugins:

```bash
make install-pulumi-plugins
```

## Usage

### As Part of Planton

This module is typically invoked through the Planton CLI:

```bash
planton apply -f manifest.yaml
```

### Standalone Usage

1. Set up the stack input as a base64-encoded environment variable:

```bash
export STACK_INPUT=$(cat manifest.yaml | base64)
```

2. Run Pulumi:

```bash
pulumi up
```

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `STACK_INPUT` | Base64-encoded CloudflareDnsRecordStackInput | Yes |
| `CLOUDFLARE_API_TOKEN` | Cloudflare API token (alternative to stack input credentials) | No |

## Build

```bash
make build
```

## Test

```bash
make test
```

## Module Structure

```
.
├── main.go           # Pulumi entry point
├── Pulumi.yaml       # Pulumi project configuration
├── Makefile          # Build and test targets
├── README.md         # This file
├── overview.md       # Architecture overview
└── module/
    ├── main.go       # Resource orchestration
    ├── locals.go     # Data transformations
    ├── outputs.go    # Output constants
    └── dns_record.go # DNS record creation logic
```

## Outputs

| Output | Description |
|--------|-------------|
| `record_id` | Cloudflare DNS record ID |
| `record_name` | The record name as stored by Cloudflare |
| `record_type` | DNS record type |
| `proxied` | Whether the record is proxied |

## Debugging

Use the debug script for local testing:

```bash
./debug.sh ../hack/manifest.yaml
```

## Troubleshooting

### "missing required configuration"

Ensure `STACK_INPUT` environment variable is set with base64-encoded manifest.

### "authentication failed"

Verify your Cloudflare API token has the required permissions:
- Zone:DNS:Edit

### "zone not found"

Verify the `zone_id` in your manifest matches an existing Cloudflare zone.
