# AzureEventHubDisasterRecoveryConfig - Terraform Module

OpenTofu/Terraform implementation for the
AzureEventHubDisasterRecoveryConfig kind, at 100%
behavioral parity with the Pulumi module.

## Resources Created

- `azurerm_eventhub_namespace_disaster_recovery_config` -- the geo-DR
  pairing, under the primary namespace

## Implementation Notes

- The provider block is EMPTY -- credentials arrive as ARM_* environment
  variables (service principal or keyless OIDC).
- azurerm addresses the primary side by discrete names (namespace +
  resource group); `locals.tf` parses them from the spec's single
  `primary_namespace_id` with an anchored regex that fails the plan
  loudly on a malformed id.
- The provider choreographs the lifecycle: create polls to Succeeded;
  partner change break-pairs then re-pairs; destroy break-pairs,
  deletes, and waits for both the config 404 and the alias-name
  release -- destroys take minutes by the service's own design.
- No credential outputs: Azure's Event Hubs DR resource exposes none;
  alias-addressed connection strings surface on the namespace and
  authorization-rule kinds.

## Validate

```bash
tofu init -backend=false && tofu validate
```
