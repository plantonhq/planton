# AKS Environment

The **AKS Environment** InfraChart provisions a configurable Azure Kubernetes
Service environment: a custom VNet with an optional NAT Gateway, an Azure DNS
zone, an Azure Container Registry, Key Vault secrets, an autoscaled node pool,
and a set of toggleable Kubernetes add-ons.

Chart manifests live in the [`templates`](templates) directory; every tunable
value is documented in [`values.yaml`](values.yaml).

## Customisation

- Copy `values.yaml` to a higher-priority file (e.g. `values.dev.yaml`) and
  override only what you need per environment.
- Add-ons and the NAT Gateway are feature-flagged in `values.yaml`; enable only
  what each environment requires.
- Cross-resource references are wired with `valueFrom`, so the templates rarely
  need direct edits.
