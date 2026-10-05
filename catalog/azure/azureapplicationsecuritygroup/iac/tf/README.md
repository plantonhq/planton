# AzureApplicationSecurityGroup - Terraform Module

Terraform implementation for the AzureApplicationSecurityGroup
kind.

## Resources Created

- `azurerm_application_security_group.main` -- the named NIC grouping
  that NSG rules reference as source/destination

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.region` | NICs can only join an ASG in their own region |
| `spec.resource_group` | Resolved literal resource group name |
| `spec.name` | Unique within the resource group; renaming replaces the group and breaks every referrer |
| `spec.tags` | The ONLY field that updates in place; user tags merge over the metadata-derived tags (user wins) |

## Provider Version

`azurerm ~> 5.0`.

## Behavior Notes

- The ASG itself is an empty anchor: membership is declared from the
  NIC / scale-set IP-configuration side, and rule usage from the NSG
  side -- never on the ASG resource.
- Everything except tags is ForceNew; plan name and region before
  workloads reference the group.

## Usage

```hcl
module "asg" {
  source = "./path/to/module"

  metadata = { name = "web-tier" }
  spec = {
    region         = "eastus"
    resource_group = "network-rg"
    name           = "web-tier-asg"
  }
}
```
