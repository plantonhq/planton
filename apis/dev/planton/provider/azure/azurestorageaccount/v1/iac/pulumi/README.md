# Azure Storage Account Pulumi Module

This Pulumi module deploys Azure Storage Accounts with blob containers, network access controls, and data protection features.

## Usage

### Standalone Module Usage

```go
package main

import (
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
    "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azurestorageaccount/v1/iac/pulumi/module"
    azurestorageaccountv1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azurestorageaccount/v1"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        stackInput := &azurestorageaccountv1.AzureStorageAccountStackInput{
            // Configure your input here
        }
        return module.Resources(ctx, stackInput)
    })
}
```

### Via Planton CLI

```bash
# Deploy using Pulumi backend
planton pulumi update \
  --manifest storage.yaml \
  --stack org/project/env \
  --module-dir apis/dev/planton/provider/azure/azurestorageaccount/v1/iac/pulumi
```

## Environment Variables

The module expects stack input via the `STACK_INPUT` environment variable, which is automatically set by the Planton CLI.

For manual testing:

```bash
export STACK_INPUT=$(cat manifest.yaml | base64)
pulumi up
```

## Required Azure Credentials

The module requires Azure Service Principal credentials:

- `client_id` - Azure AD Application (client) ID
- `client_secret` - Azure AD Application secret
- `tenant_id` - Azure AD Tenant ID
- `subscription_id` - Azure Subscription ID

These are provided via the `provider_config` field in the stack input.

## Resources Created

- Azure Storage Account
- Blob Containers (as specified in manifest)
- Network ACLs (firewall rules)
- Blob soft delete policy
- Container soft delete policy

## Outputs

| Output | Description |
|--------|-------------|
| `storage_account_id` | Azure Resource Manager ID |
| `storage_account_name` | Storage account name |
| `primary_blob_endpoint` | Primary blob service URL |
| `primary_queue_endpoint` | Primary queue service URL |
| `primary_table_endpoint` | Primary table service URL |
| `primary_file_endpoint` | Primary file service URL |
| `primary_dfs_endpoint` | Primary DFS (Data Lake) URL |
| `primary_web_endpoint` | Primary static website URL |
| `container_url_map` | Map of container names to URLs |
| `region` | Azure region |
| `resource_group` | Resource group name |

## Development

### Build

```bash
make build
```

### Update Dependencies

```bash
make update-deps
```

## Troubleshooting

### Common Issues

1. **Name conflicts**: Storage account names must be globally unique and 3-24 characters (lowercase alphanumeric only).

2. **Network lockout**: If you enable network rules with default "Deny", ensure you include your source IP in the allowed list.

3. **Provider errors**: Verify Azure credentials are valid and have sufficient permissions (Storage Account Contributor role recommended).
