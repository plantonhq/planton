# AzureContainerApp Terraform Module

The Terraform/OpenTofu implementation of the `AzureContainerApp` kind.

## Structure

```
tf/
├── main.tf          # The container app (template, ingress, secrets, registries, dapr, identity)
├── variables.tf     # Input variables (metadata + spec)
├── outputs.tf       # Outputs
├── locals.tf        # Tag merge + enum wire-value maps
└── provider.tf      # Empty azurerm provider (credentials injected as ARM_* env)
```

## Resources Created

| Resource | Type | Condition |
|----------|------|-----------|
| Container App | `azurerm_container_app` | Always |

## Behavior Notes

- Enum vocabularies (revision mode, probe/ingress transports, mTLS modes, restriction actions, volume storage types, Dapr protocol, identity types) map to ARM wire values in `locals.tf`; the mTLS vocabulary is lowercase on the wire (`accept`/`require`/`ignore`).
- Per-probe-type initial-delay defaults (1 liveness, 0 readiness/startup) are applied with `coalesce`, matching azurerm's own schema defaults; the other scaler/replica dials resolve in `variables.tf`.
- `custom_domain_verification_id` is wrapped in `nonsensitive()` -- it is published in a DNS TXT record, but the provider marks the attribute Sensitive.
- User `spec.tags` merge over the metadata-derived tags; user tags win on key collision.

## Validation

```bash
tofu init -backend=false && tofu validate
planton tofu plan --manifest ../../e2e/manifest.yaml --module-dir .
```
