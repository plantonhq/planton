# Azure Front Door Profile Pulumi Module

This Pulumi module deploys Azure Front Door profiles with endpoints, origin groups, origins, and routes for global CDN and application delivery.

## Usage

### Standalone Module Usage

```go
package main

import (
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
    "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azurefrontdoorprofile/v1/iac/pulumi/module"
    azurefrontdoorprofilev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azurefrontdoorprofile/v1"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        stackInput := &azurefrontdoorprofilev1.AzureFrontDoorProfileStackInput{
            // Configure your input here
        }
        return module.Resources(ctx, stackInput)
    })
}
```

### Via OpenMCF CLI

```bash
# Deploy using Pulumi backend
openmcf pulumi update \
  --manifest frontdoor.yaml \
  --stack org/project/env \
  --module-dir apis/org/openmcf/provider/azure/azurefrontdoorprofile/v1/iac/pulumi
```

## Environment Variables

The module expects stack input via the `STACK_INPUT` environment variable, which is automatically set by the OpenMCF CLI.

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

- Azure Front Door Profile (`azurerm_cdn_frontdoor_profile`)
- Front Door Endpoints (`azurerm_cdn_frontdoor_endpoint`)
- Front Door Origin Groups (`azurerm_cdn_frontdoor_origin_group`)
- Front Door Origins (`azurerm_cdn_frontdoor_origin`)
- Front Door Routes (`azurerm_cdn_frontdoor_route`)

## Outputs

| Output | Description |
|--------|-------------|
| `profile_id` | Azure Resource Manager ID of the profile |
| `profile_name` | The profile name |
| `resource_guid` | Front Door service GUID |
| `endpoint_ids` | Map of endpoint names to resource IDs |
| `endpoint_hostnames` | Map of endpoint names to generated hostnames (*.azurefd.net) |

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

1. **Name conflicts**: Profile names must be globally unique (2-46 characters, alphanumeric and hyphens).

2. **Private link errors**: Private link requires `Premium_AzureFrontDoor` SKU and `certificate_name_check_enabled = true` on the origin.

3. **Route conflicts**: Ensure `patterns_to_match` do not overlap across routes for the same endpoint unless intentional.

4. **Provider errors**: Verify Azure credentials are valid and have sufficient permissions (CDN Profile Contributor role recommended).
