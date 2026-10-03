# AzureEventHubCluster - Pulumi Module

Pulumi (Go) implementation for the AzureEventHubCluster deployment
component, at 100% behavioral parity with the Terraform module.

## Resources Created

- `eventhub.Cluster` -- the dedicated Event Hubs cluster (single-tenant
  capacity units)

## Implementation Notes

- The Azure provider is built through the shared
  `pulumiazureprovider.Get` builder -- never inline `NewProvider`.
- The ARM sku is composed as `Dedicated_{capacity_units}` -- Dedicated
  is the only sku family Azure sells for clusters, so the tier name is a
  constant, not configuration. `capacity_units` is presence-guarded to 1
  CU -- direct iac-input paths do not materialize proto defaults.
- Azure forbids deleting a cluster for 4 hours after creation (the
  deletion moratorium); destroys of young clusters retry for hours by
  the service's own rule.
- Clusters bill per capacity unit per hour at dedicated-tier rates.

## Build

```bash
make build
```
