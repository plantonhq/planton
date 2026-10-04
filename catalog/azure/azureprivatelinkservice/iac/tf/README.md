# AzurePrivateLinkService Terraform Module

## Overview

This Terraform module provisions a Private Link Service using the
`azurerm` provider. It creates an `azurerm_private_link_service` -- the
PROVIDER side of Azure Private Link: a service you run (behind a
Standard internal load balancer, or at one fixed destination IP)
exposed to consumers in other virtual networks through private
endpoints, over the Microsoft backbone.

The destination contract (exactly one of the load-balancer frontend
list or a destination IP), the single-primary NAT contract, and the
NAT-name uniqueness rule are spec-validated before the module runs.
Every NAT subnet must have `private_link_service_network_policies_enabled`
set to `false` -- an ARM requirement checked only at deploy time.

## Resources Created

- `azurerm_private_link_service.main` -- the service

## Variables

| Variable | Type | Description |
|----------|------|-------------|
| `metadata` | object | Planton metadata (name, org, env) |
| `spec` | object | Private Link Service specification |

Key spec fields:

| Field | Required | Description |
|-------|----------|-------------|
| `region` / `resource_group` / `name` | yes | The service's ARM identity; all ForceNew |
| `nat_ip_configurations` | yes | 1-8 NAT addresses on policies-disabled subnets; exactly one `primary` |
| `load_balancer_frontend_ip_configuration_ids` | one-of | Standard LB frontends the service fronts (ForceNew) |
| `destination_ip_address` | one-of | NAT straight to one fixed private IP |
| `proxy_protocol_enabled` | no | PROXY v2 headers carry the consumer's source IP |
| `visibility_subscription_ids` | no | Who can discover the service (UUIDs or `"*"`) |
| `auto_approval_subscription_ids` | no | Whose connections skip manual approval |
| `fqdns` | no | Names surfaced to consumers |
| `tags` | no | User tags, merged over metadata-derived tags |

## Outputs

| Output | Description |
|--------|-------------|
| `private_link_service_id` | Full ARM ID |
| `private_link_service_name` | The service's name as deployed |
| `alias` | The globally unique handle consumers request connections with |

## Usage

```hcl
module "orders_pls" {
  source = "./iac/tf"

  metadata = { name = "orders-api", org = "mycompany", env = "production" }

  spec = {
    region         = "eastus"
    resource_group = "network-rg"
    name           = "orders-api"
    nat_ip_configurations = [{
      name      = "nat-1"
      subnet_id = "/subscriptions/.../subnets/pls-nat"
      primary   = true
    }]
    load_balancer_frontend_ip_configuration_ids = [
      "/subscriptions/.../loadBalancers/orders-lb/frontendIPConfigurations/internal"
    ]
  }
}
```

## Behavior Notes

- The LB frontend set is ForceNew; the NAT list updates in place, but
  ARM refuses to clear an assigned static address or move the PRIMARY
  configuration's subnet (destroy/recreate is the only path).
- Creation waits for ARM's Succeeded provisioning state -- the provider
  polls past the initial accept because the service applies its values
  asynchronously.

## Required Permissions

Least-privilege runner permissions for this kind are declared in [`../permissions.yaml`](../permissions.yaml).
