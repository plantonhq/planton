# AzureServiceBusAuthorizationRule - Pulumi Module

Pulumi (Go) implementation for the AzureServiceBusAuthorizationRule
kind, at 100% behavioral parity with the Terraform
module.

## Resources Created

Exactly ONE of (the spec's scope XOR picks it):

- `servicebus.NamespaceAuthorizationRule`
- `servicebus.QueueAuthorizationRule`
- `servicebus.TopicAuthorizationRule`

## Implementation Notes

- The Azure provider is built through the shared
  `pulumiazureprovider.Get` builder -- never inline `NewProvider`.
- The scope switch selects the resource; the six credential faces are
  identical across all three types, so the export block is shared.
- The rights trio reads through getters (Azure's default is false for
  each; the spec's CELs guarantee a usable combination).

## Build

```bash
make build
```
