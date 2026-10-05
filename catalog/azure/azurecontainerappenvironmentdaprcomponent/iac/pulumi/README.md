# AzureContainerAppEnvironmentDaprComponent - Pulumi Module

Pulumi implementation for the AzureContainerAppEnvironmentDaprComponent
kind.

## Architecture

```
containerapp.EnvironmentDaprComponent (one pluggable Dapr backend)
```

## Key Design Decisions

- **Secrets never ride plain metadata** -- metadata entries carry a
  literal value XOR a `secret_name` reference into the component's own
  secret list (CEL-enforced at the spec).
- **`init_timeout` is materialized explicitly** (default `"5s"`) so
  both engines send identical request bodies; IaC inputs never carry
  proto defaults.
- **Empty scopes are omitted** -- ARM treats an absent scope list as
  "every Dapr app in the environment"; production components should
  scope deliberately.
- **No tags** -- ARM does not support tags on `daprComponents`.

## Provider

Built via the shared `pulumiazureprovider.Get` builder -- static client
secret, keyless web identity, or ambient chain, resolved from the IaC
input. Never construct the provider inline.
