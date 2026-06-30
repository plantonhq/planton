# Azure DNS Record Pulumi Module

This Pulumi module creates DNS records in an existing Azure DNS Zone.

## Prerequisites

- Go 1.21+
- Pulumi CLI
- Azure credentials configured

## Usage

### Environment Variables

The module expects stack input via the `STACK_INPUT` environment variable containing a JSON-serialized `AzureDnsRecordStackInput`.

### Local Testing

1. Set up credentials:
```bash
export AZURE_CLIENT_ID="..."
export AZURE_CLIENT_SECRET="..."
export AZURE_SUBSCRIPTION_ID="..."
export AZURE_TENANT_ID="..."
```

2. Create stack input:
```bash
export STACK_INPUT='{"target":{"apiVersion":"azure.planton.dev/v1","kind":"AzureDnsRecord","metadata":{"name":"test-record"},"spec":{"resource_group":"my-rg","zone_name":{"value":"example.com"},"record_type":"A","name":"www","values":["192.0.2.1"]}},"provider_config":{"client_id":"...","client_secret":"...","subscription_id":"...","tenant_id":"..."}}'
```

3. Run Pulumi:
```bash
pulumi up
```

## Building

```bash
make build
```

## Module Structure

- `main.go` - Pulumi program entry point
- `module/main.go` - Resource creation logic
- `module/locals.go` - Local variable initialization
- `module/outputs.go` - Output constant definitions

## Outputs

| Output | Description |
|--------|-------------|
| `record_id` | Azure Resource Manager ID of the DNS record |
| `fqdn` | Fully qualified domain name |
