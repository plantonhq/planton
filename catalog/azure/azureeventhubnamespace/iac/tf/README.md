# AzureEventHubNamespace - Terraform Module

OpenTofu/Terraform implementation for the AzureEventHubNamespace
kind, at 100% behavioral parity with the Pulumi module.

## Resources Created

- `azurerm_eventhub_namespace` -- the namespace, with optional managed
  identity, auto-inflate, dedicated-cluster placement, and the inline
  network rule set

## Implementation Notes

- The provider block is EMPTY -- credentials arrive as ARM_* environment
  variables (service principal or keyless OIDC).
- The SKU/identity/firewall enums arrive as full proto value names and
  are mapped to ARM's wire values in `locals.tf`; unset sku deploys
  Standard.
- `capacity` is sent only when present so Azure's default (1) applies
  otherwise; the tier-dependent value contracts (TUs 1-40 vs PUs
  1/2/4/8/16) are enforced by the spec's CELs.
- Each `ip_rules` entry is emitted as an allow rule -- Azure's per-rule
  action accepts exactly one value (Allow), so the spec models just the
  mask.
- The Premium boundary is ForceNew (the provider's CustomizeDiff);
  moving on or off a dedicated cluster is likewise ForceNew.
- Sensitive outputs: the root SAS rule's four credential faces plus the
  two geo-DR alias faces (populated only under a pairing).

## Validate

```bash
tofu init -backend=false && tofu validate
```
