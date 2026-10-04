# AzureServiceBusQueue - Pulumi Module

Pulumi (Go) implementation for the AzureServiceBusQueue
kind, at 100% behavioral parity with the Terraform module.

## Resources Created

- `servicebus.Queue` -- the queue, addressed by the parent namespace's
  ARM id

## Implementation Notes

- The Azure provider is built through the shared
  `pulumiazureprovider.Get` builder -- never inline `NewProvider`.
- The gate-state wire map carries an unspecified row (Active) so a zero
  enum never sends the empty string.
- `batched_operations_enabled` is presence-guarded to Azure's default
  (true); every other optional dial is sent only when set so Azure's
  defaults apply.
- `namespace_name` is parsed from the resolved parent id with the same
  anchored contract as the Terraform module's regex.

## Build

```bash
make build
```
