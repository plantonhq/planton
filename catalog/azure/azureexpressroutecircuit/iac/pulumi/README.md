# AzureExpressRouteCircuit Pulumi Module

## Overview

This Pulumi module provisions an ExpressRoute circuit using the Azure
Classic provider (`pulumi-azure`). It creates a
`network.ExpressRouteCircuit` plus one
`network.ExpressRouteCircuitAuthorization` child per spec
`authorizations` entry, parented to the circuit -- the keys other
subscriptions redeem to connect their gateways.

The Azure provider is built through the shared provider builder, which
resolves the right credential mechanism (static client secret, keyless
web identity, or ambient chain) from the IaC input.

## Design Decisions

- The provisioning-mode contract (service-provider trio XOR
  ExpressRoute Direct pair) is spec-validated; the module selects the
  argument set from `service_provider_name`'s presence, mirroring the
  provider's own create-shape branch.
- The spec's tier/family enums map to ARM wire values through explicit
  switch helpers in `locals.go` -- the same name-to-wire tables the
  Terraform module carries in its locals.
- ARM-generated secrets (`service_key`, each authorization's key) are
  exported through `pulumi.ToSecret`, matching the Terraform module's
  `sensitive = true` outputs.

## Inputs

The module receives an `AzureExpressRouteCircuitIacInput` containing:

- `target.spec.region` / `target.spec.resource_group` / `target.spec.name` -- the circuit's ARM identity (references resolved to literals by the platform)
- `target.spec.sku_tier` / `target.spec.sku_family` -- the SKU pair
- `target.spec.service_provider_name` + `peering_location` + `bandwidth_in_mbps` OR `express_route_port_id` + `bandwidth_in_gbps` -- the provisioning mode
- `target.spec.authorizations` -- keys to issue, by name
- `target.spec.tags` -- user tags, merged over the metadata-derived tags (user wins)
- `provider_config` -- Azure credentials

## Outputs

| Output | Description |
|--------|-------------|
| `express_route_circuit_id` | Full ARM ID |
| `express_route_circuit_name` | The join key peerings reference |
| `service_key` | The provisioning credential for the provider (secret) |
| `service_provider_provisioning_state` | NotProvisioned / Provisioning / Provisioned / Deprovisioning |
| `authorization_keys` | Name-keyed generated keys (secret) |

## Local Development

```bash
go build ./...   # compile the module and entrypoint
```
