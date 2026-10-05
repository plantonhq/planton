# AzureEventHub - Pulumi Module

Pulumi (Go) implementation for the AzureEventHub kind,
at 100% behavioral parity with the Terraform module.

## Resources Created

- `eventhub.EventHub` -- the event hub, with its partition layout,
  retention model, gate state, and optional capture-to-storage

## Implementation Notes

- The Azure provider is built through the shared
  `pulumiazureprovider.Get` builder -- never inline `NewProvider`.
- Enum wire maps carry an unspecified row where the spec allows unset
  (Active status, StorageSAS capture auth) so a zero enum never sends
  the empty string; cleanup policy and encoding are required-explicit,
  so no fallback rows.
- The retention XOR and the policy-to-field pairing are spec CELs; each
  retention variant sends only its own fields.
- The capture destination's `name` is a one-value constant
  (`EventHubArchive.AzureBlockBlob`) the module sends unconditionally;
  the optional capture cadence fields (interval, size, skip-empty) are
  presence-guarded so Azure's defaults apply when unset.
- The partition tier caps and the never-decrease contract are enforced
  by Azure at apply time -- they depend on the parent namespace's tier,
  which this module cannot see.
- No tags: ARM does not support tags on Event Hubs entities; the
  platform's identity tags live on the parent namespace.

## Build

```bash
make build
```
