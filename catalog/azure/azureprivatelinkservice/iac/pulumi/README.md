# AzurePrivateLinkService Pulumi Module

## Overview

This Pulumi module provisions a Private Link Service using the Azure
Classic provider (`pulumi-azure`). It creates a
`privatedns.LinkService` -- the classic SDK registers this resource
under the legacy "privatedns" module (a historical upstream placement,
not a DNS resource); the wire surface is `azurerm_private_link_service`,
the PROVIDER side of Azure Private Link.

The Azure provider is built through the shared provider builder, which
resolves the right credential mechanism (static client secret, keyless
web identity, or ambient chain) from the IaC input.

## Design Decisions

- The destination contract (exactly one of LB frontends or a
  destination IP) and the single-primary NAT contract are spec-validated
  -- the module maps fields without re-checking.
- The classic SDK carries both `ProxyProtocolEnabled` and the
  deprecated `EnableProxyProtocol`; the module uses the current name,
  matching the Terraform engine exactly.
- Empty optional fields are omitted rather than sent as zero values, so
  both engines produce identical ARM payloads.

## Inputs

The module receives an `AzurePrivateLinkServiceIacInput` containing:

- `target.spec.region` / `target.spec.resource_group` / `target.spec.name` -- the service's ARM identity (references resolved to literals by the platform)
- `target.spec.nat_ip_configurations` -- 1-8 NAT addresses on policies-disabled subnets, exactly one primary
- `target.spec.load_balancer_frontend_ip_configuration_ids` / `target.spec.destination_ip_address` -- the traffic destination (exactly one form)
- `target.spec.visibility_subscription_ids` / `target.spec.auto_approval_subscription_ids` -- discoverability and approval
- `target.spec.tags` -- user tags, merged over the metadata-derived tags (user wins)
- `provider_config` -- Azure credentials

## Outputs

| Output | Description |
|--------|-------------|
| `private_link_service_id` | Full ARM ID |
| `private_link_service_name` | The service's name as deployed |
| `alias` | The globally unique handle consumers request connections with |

## Local Development

```bash
go build ./...   # compile the module and entrypoint
```
