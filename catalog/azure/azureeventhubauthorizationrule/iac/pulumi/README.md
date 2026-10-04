# AzureEventHubAuthorizationRule - Pulumi Module

Pulumi (Go) implementation for the AzureEventHubAuthorizationRule
kind, at 100% behavioral parity with the Terraform
module.

## Resources Created

Exactly ONE of (the spec's scope XOR picks it):

- `eventhub.EventHubNamespaceAuthorizationRule`
- `eventhub.EventHubAuthorizationRule`

## Implementation Notes

- The Azure provider is built through the shared
  `pulumiazureprovider.Get` builder -- never inline `NewProvider`.
- Both provider resources still use legacy name addressing, so the
  locals parse the parent names from the resolved ARM id with the same
  anchored regexes as the Terraform module -- a malformed id fails
  loudly on both engines.
- The scope switch selects the resource; the six credential faces are
  identical across both types, so the export block is shared.
- The rights trio is presence-guarded to false (Azure's default for
  each; the spec's CELs guarantee a usable combination).
- The alias connection strings are only populated when the namespace
  carries a geo-DR pairing.

## Build

```bash
make build
```
