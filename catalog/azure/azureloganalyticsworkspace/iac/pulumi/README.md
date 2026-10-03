# AzureLogAnalyticsWorkspace - Pulumi Module

Pulumi (Go) implementation for the AzureLogAnalyticsWorkspace deployment
component, at 100% behavioral parity with the Terraform module.

## Resources Created

- `operationalinsights.AnalyticsWorkspace` -- the workspace, carrying the
  pricing tier, retention/quota dials, security and network posture, an
  optional managed identity, and merged governance tags

## Implementation Notes

- The Azure provider is built through the shared
  `pulumiazureprovider.Get` builder (static client secret, keyless web
  identity, or ambient chain) -- never inline `NewProvider`.
- Enum wire maps live in `locals.go` (SKU and identity-type vocabularies,
  mechanically identical to the Terraform module's `locals.tf` maps).
- True-default optional booleans are presence-guarded to the proto
  defaults: IaC inputs built from a manifest materialize defaults, but
  direct iac-input paths do not.
- `workspace_customer_id` exports the provider's WorkspaceId attribute
  (the agent-facing GUID); `workspace_id` exports the ARM resource ID --
  the FK seam downstream kinds reference.

## Build

```bash
make build
```
