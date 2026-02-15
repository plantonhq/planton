# AwsElasticIp — Pulumi Module

This directory contains the Pulumi Go module that provisions an AWS Elastic IP.

## Structure

```
pulumi/
├── main.go          # Entrypoint — loads stack input, calls module
├── Pulumi.yaml      # Pulumi project descriptor
├── Makefile          # Build/preview/up/destroy shortcuts
├── debug.sh          # Local development helper
├── module/
│   ├── main.go      # Resources() — provider setup, orchestration, exports
│   ├── locals.go    # Locals struct, tag initialization
│   ├── outputs.go   # Output key constants
│   └── eip.go       # Elastic IP resource creation
```

## Quick Start

```bash
# Build
make build

# Preview (requires AWS credentials and stack input)
make preview

# Deploy
make up

# Destroy
make destroy
```

## Module API

```go
func Resources(ctx *pulumi.Context, stackInput *AwsElasticIpStackInput) error
```

**Inputs:** `AwsElasticIpStackInput` (target resource + optional provider config)

**Exports:**
- `allocation_id` — EIP allocation ID
- `public_ip` — public IPv4 address
- `arn` — EIP ARN
- `public_dns` — public DNS hostname
