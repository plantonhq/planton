# AzureEventHub - Terraform Module

OpenTofu/Terraform implementation for the AzureEventHub
kind, at 100% behavioral parity with the Pulumi module.

## Resources Created

- `azurerm_eventhub` -- the event hub, with its partition layout,
  retention model, gate state, and optional capture-to-storage

## Implementation Notes

- The provider block is EMPTY -- credentials arrive as ARM_* environment
  variables (service principal or keyless OIDC).
- The status/cleanup-policy/encoding enums arrive as full proto value
  names and are mapped to ARM's wire values in `locals.tf`; unset status
  deploys Active.
- The retention XOR and the policy-to-field pairing are spec CELs, so
  each retention variant sends only its own fields (Azure silently
  ignores mismatched ones -- the spec rejects them up front instead).
- The capture destination's `name` is a one-value constant
  (`EventHubArchive.AzureBlockBlob`) the module sends unconditionally;
  unset `storage_authentication_type` keeps Azure's default,
  service-managed SAS.
- The partition tier caps and the never-decrease contract are enforced
  by Azure at apply time -- they depend on the parent namespace's tier,
  which this module cannot see.
- No tags: ARM does not support tags on Event Hubs entities; the
  platform's identity tags live on the parent namespace.

## Validate

```bash
tofu init -backend=false && tofu validate
```
