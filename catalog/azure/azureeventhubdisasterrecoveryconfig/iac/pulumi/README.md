# AzureEventHubDisasterRecoveryConfig - Pulumi Module

Pulumi (Go) implementation for the AzureEventHubDisasterRecoveryConfig
kind, at 100% behavioral parity with the Terraform
module.

## Resources Created

- `eventhub.EventhubNamespaceDisasterRecoveryConfig` -- the geo-DR
  pairing, under the primary namespace

## Implementation Notes

- The Azure provider is built through the shared
  `pulumiazureprovider.Get` builder -- never inline `NewProvider`.
- The provider addresses the primary side by discrete names (namespace
  + resource group); the module parses them from the spec's single
  `primary_namespace_id` with the same anchored contract as the
  Terraform module's regex, failing loudly on a malformed id.
- The provider choreographs the lifecycle: create polls to Succeeded;
  partner change break-pairs then re-pairs; destroy break-pairs,
  deletes, and waits for both the config 404 and the alias-name
  release -- destroys take minutes by the service's own design.
- No credential outputs: Azure's Event Hubs DR resource exposes none;
  alias-addressed connection strings surface on the namespace and
  authorization-rule kinds.

## Build

```bash
make build
```
