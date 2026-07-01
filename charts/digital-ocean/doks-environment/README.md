# DOKS Environment

The **DOKS Environment** InfraChart provisions a configurable DigitalOcean
Kubernetes environment: a custom VPC, an optional DNS zone and container
registry, an autoscaled node pool, and toggleable Kubernetes add-ons
(Cert-Manager, Istio, and more).

Chart manifests live in the [`templates`](templates) directory; every tunable
value is documented in [`values.yaml`](values.yaml).

## Customisation

- Copy `values.yaml` to a higher-priority file (e.g. `values.dev.yaml`) and
  override only what you need per environment.
- Add-ons, the DNS zone, and the container registry are feature-flagged in
  `values.yaml`; enable only what each environment requires.
- Cross-resource references are wired with `valueFrom`, so the templates rarely
  need direct edits.
