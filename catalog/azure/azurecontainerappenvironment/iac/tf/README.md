# AzureContainerAppEnvironment Terraform Module

The Terraform/OpenTofu implementation of the `AzureContainerAppEnvironment` component.

## Structure

```
tf/
├── main.tf          # Environment + optional custom-domain association
├── variables.tf     # Input variables (metadata + spec)
├── outputs.tf       # Outputs
├── locals.tf        # Tag merge + enum wire-value maps
└── provider.tf      # Empty azurerm provider (credentials injected as ARM_* env)
```

## Resources Created

| Resource | Type | Condition |
|----------|------|-----------|
| Container App Environment | `azurerm_container_app_environment` | Always |
| Custom DNS suffix | `azurerm_container_app_environment_custom_domain` | `spec.custom_domain` set |

## Behavior Notes

- Enum vocabularies (logs destination, public network access, workload profile SKUs, identity types) arrive as the spec enum's name strings and are mapped to ARM wire values in `locals.tf` -- a vocabulary drift fails at plan time rather than deploying something wrong.
- `logs_destination` unset with a workspace deploys `log-analytics` (azurerm's own legacy inference), unset without one omits the property (streaming-only).
- User `spec.tags` merge over the metadata-derived tags; user tags win on key collision.

## Validation

```bash
tofu init -backend=false && tofu validate
```

The full offline proof runs through the CLI against `../../e2e/manifest.yaml`:

```bash
planton tofu plan --manifest ../../e2e/manifest.yaml --module-dir .
```
