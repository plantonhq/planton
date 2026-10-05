# AzureEventHubCluster - Terraform Module

OpenTofu/Terraform implementation for the AzureEventHubCluster
kind, at 100% behavioral parity with the Pulumi module.

## Resources Created

- `azurerm_eventhub_cluster` -- the dedicated Event Hubs cluster
  (single-tenant capacity units)

## Implementation Notes

- The provider block is EMPTY -- credentials arrive as ARM_* environment
  variables (service principal or keyless OIDC).
- The ARM sku is composed as `Dedicated_{capacity_units}` -- Dedicated
  is the only sku family Azure sells for clusters, so the tier name is a
  constant, not configuration. Unset capacity deploys 1 CU.
- Azure forbids deleting a cluster for 4 hours after creation (the
  deletion moratorium); destroys of young clusters retry for hours by
  the service's own rule.
- Clusters bill per capacity unit per hour at dedicated-tier rates.

## Validate

```bash
tofu init -backend=false && tofu validate
```
