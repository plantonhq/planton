# KubernetesFlagd Terraform Module

Runs flagd as module-owned manifests (flagd publishes no Helm chart) through the hashicorp `kubernetes` provider, with the optional ServiceMonitor applied through `kubectl_manifest` (alekc/kubectl), which needs no cluster connection at plan time.

## Module Behavior

- **Arguments and sources are computed in `locals.tf`** -- the same `flagd start` arguments and a byte-identical sources document as the Pulumi twin, so the `checksum/sources` pod annotation matches across engines and a source change rolls the pods.
- **Sources live in the `<name>-sources` Secret** -- HTTP authorization headers and credential headers are part of the SourceConfig array, which reaches flagd as `FLAGD_SOURCES`.
- **ConfigMap sources mount as directories** at `/etc/flagd/sources/<index>`, never `subPath`, so the kubelet's volume sync delivers flag edits with no restart.
- **Readiness waits for every source** -- the readiness probe is `/readyz` on the management port, which flagd answers only after each source has synced once.
- **Preconditions fail the plan loudly** -- `metadata.name` past 63 characters (the Service is named after the resource, and a Service name is a 63-character DNS label), or a header declared both plain and sensitive.
- **Empty topology-spread selectors self-spread** on flagd's own selector labels.

## Resources

| Resource | Condition |
|---|---|
| `kubernetes_namespace_v1.flagd` | `spec.create_namespace` |
| `kubernetes_secret_v1.sources` | always |
| `kubernetes_service_account_v1.flagd` | unless `spec.service_account.existing_name` |
| `kubernetes_role_v1.flag_reader` / `kubernetes_role_binding_v1.flag_reader` | one per namespace a `feature_flag` source reads |
| `kubernetes_deployment_v1.flagd` | always |
| `kubernetes_service_v1.flagd` | always |
| `kubernetes_horizontal_pod_autoscaler_v2.flagd` | `spec.hpa.enabled` |
| `kubernetes_pod_disruption_budget_v1.flagd` | `spec.pdb.enabled` |
| `kubectl_manifest.service_monitor` | `spec.metrics.service_monitor_enabled` |

## Usage

```shell
planton tofu init --manifest ../../e2e/manifest.yaml --module-dir .
planton tofu plan --manifest ../../e2e/manifest.yaml --module-dir .
planton tofu apply --manifest ../../e2e/manifest.yaml --module-dir .
planton tofu destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

State is kept by the default local backend.
