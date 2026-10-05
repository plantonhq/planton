# OpenTelemetry Operator

Deploys the OpenTelemetry Operator -- the controller that turns `OpenTelemetryCollector` declarations into running collector fleets -- from the official `opentelemetry-operator` Helm chart. This component installs the ENGINE only: collectors are declared separately as KubernetesOtelCollector resources (one per pipeline shape), and the operator reconciles them into Deployments, DaemonSets, StatefulSets or sidecar injections. One operator per cluster is the grain: it watches every namespace, so every collector on the cluster is served by it. The opentelemetry.io CRDs are module-owned and RETAINED on destroy -- removing the operator never deletes the fleet's declarations. Requires cert-manager on the cluster for the admission-webhook certificate.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **The opentelemetry.io CRDs** -- derived from the pinned chart at apply time (at the default pin: `opentelemetrycollectors`, `instrumentations`, `opampbridges`, `targetallocators`; the feature-gated `clusterobservabilities` CRD appears when its gate is enabled), applied OUTSIDE the Helm release and retained on destroy; chart version bumps re-apply them at the new pin.
- **Helm Release** -- the `opentelemetry-operator` chart with `crds.create` pinned to `false` (the module owns the CRD lifecycle), creating:
  - Deployment for the operator manager with the configured replica count and resource limits; with more than one replica leader election picks ONE active reconciler (extra replicas are warm standbys, not workload shards)
  - The admission/conversion webhook Service and a cert-manager Certificate for its serving certificate -- either against the chart's own self-signed Issuer (the default) or against an Issuer/ClusterIssuer you name
  - A ServiceMonitor for the operator's own metrics, only when `serviceMonitorEnabled` is `true`
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **cert-manager (HARD prerequisite)** -- a working KubernetesCertManager install. cert-manager issues and ROTATES the operator's webhook serving certificate, and its CA injector keeps the retained CRDs' conversion trust current through the `cert-manager.io/inject-ca-from` annotation. Because the CRDs outlive the release, their conversion trust must be maintained by a running reconciler -- a certificate embedded once at install time goes stale on rotation and silently breaks collector-CR conversion later, which is why no self-signed one-shot arm exists. Without cert-manager the install fails its readiness wait (atomic, 600s) by design.
- **Cluster-admin-grade permissions on the first install** -- applying the cluster-scoped opentelemetry.io CRDs requires them.
- **A name within budget** -- keep `metadata.name` at 30 characters or fewer: the chart derives a `<name>-controller-manager-service-cert` Secret (a 33-character suffix) and Kubernetes caps names at 63; the modules fail loudly over the budget.

## Deploy

### Console

Open the deployment store, find **OpenTelemetry Operator**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Default preset** for the standard install, or the **Private-mirror preset** for the air-gapped posture (both image seams mirrored) in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesOtelOperator
metadata:
  name: otel-operator
  org: acme-corp
  env: prod
spec:
  namespace:
    value: "otel-operator"
  createNamespace: true
  replicas: 2
  serviceMonitorEnabled: true
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 512Mi
```

```shell
planton apply -f otel-operator.yaml
```

This deploys the operator with a warm standby behind leader election in the `otel-operator` namespace and a ServiceMonitor exposing its own metrics to Prometheus. The operator watches all namespaces. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the operator to a namespace managed by another Infra Component:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: otel-operator-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline deploys the namespace first, then provisions the operator into it.

## Key Configuration

These are the most important decisions when configuring the OpenTelemetry Operator. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The CRD lifecycle and the destroy contract** -- The chart would template the opentelemetry.io CRDs RELEASE-OWNED, so a Helm uninstall would cascade-delete every collector declaration on the cluster. The modules instead DERIVE the CRDs from the pinned chart (rendered with the release's own values and the CRD switch on), apply them OUTSIDE the release as kept resources, and install the release with CRDs skipped: destroying the operator KEEPS the CRDs -- and every declared collector -- on the cluster unless `crds.keepOnUninstall` is false; a reinstall re-adopts them; chart version bumps re-apply them at the new pin; lowering the chart version below what the cluster's CRDs carry is REFUSED with the remedy, because an older schema can strip fields from existing collectors. `crds.install: false` is a bring-your-own-CRDs arm for clusters where a GitOps bundle owns them -- never a lighter install (with the CRDs absent the operator cannot start). Every refusal states what was observed, what it means, and the next step.

**Chart version lockstep** -- `chartVersion` (default `"0.120.0"`) pins the chart, and chart 0.120.0 pairs with operator v0.156.0. Bumping the chart re-applies the module-owned CRDs at the new pin -- the version must exist as a served chart in the upstream repository index, and a version that does not is reported as such at plan time.

**The webhook certificate rides cert-manager** -- `webhook.issuerRef` is optional: ABSENT, the chart creates its own self-signed Issuer, the right choice for almost everyone (the webhook certificate only needs API-server trust, which cert-manager's CA injection handles). Declared, `kind` must be exactly `Issuer` (namespaced -- it must live in the operator's namespace) or `ClusterIssuer`, and `name` is required. Either way cert-manager is the hard prerequisite -- see Before You Deploy.

**The TWO image dials are different seams** -- `imageRegistry` rewrites ONLY the operator manager image's registry (the one image this component's pods pull; the path stays the upstream `open-telemetry/opentelemetry-operator/opentelemetry-operator`). `defaultCollectorImage` is the FLEET-WIDE dial: what the operator injects into `OpenTelemetryCollector` CRs that declare no image -- collector pods pull that one. Air-gapped clusters mirror BOTH; mirroring the operator without the collector default leaves every collector pod reaching for ghcr.io. Keep the tag on `defaultCollectorImage` so the operator's `--collector-image` flag stays explicit.

**Replicas are warm standbys** -- A single replica suits most clusters. With `replicas: 2` leader election picks ONE active reconciler and the standby takes over on failure -- reconcile throughput does not change with replicas.

**Metrics are opt-in** -- `serviceMonitorEnabled: true` renders a ServiceMonitor for the operator's OWN metrics (reconcile latency, webhook activity). It requires the `monitoring.coreos.com` CRDs on the cluster -- install KubernetesKubePrometheusStack first.

**The Helm-values escape hatch is guarded** -- `helmValues` merges LAST over the typed fields (Helm `-f` semantics) for the chart surface beyond them: kube-rbac-proxy tuning, network policy, PDB. Two keys are re-pinned by the module AFTER the merge: `crds.create: false` and the cert-manager webhook arm -- both are load-bearing design. Feature gates set through `helmValues` (the `operator.clusterobservability` alpha gate, for example) bring their CRDs with them, because the CRDs are derived from the chart with the same values the release installs with.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the operator runs in | Locating the control plane for diagnostics |
| `release_name` | Helm release name (equals metadata.name) | Helm management and debugging |
| `webhook_service` | The admission/conversion webhook Service name | Network-policy allowances for API-server-to-webhook traffic |
| `webhook_cert_secret_name` | The webhook serving-certificate Secret name | Certificate diagnostics and rotation checks |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard operator** -- One operator in its own namespace with the chart's self-signed webhook Issuer: it watches all namespaces and serves every collector on the cluster. Start from the **Default preset**.

**Air-gapped operator** -- Both image seams mirrored (the manager registry AND the fleet-wide collector default), pull secrets referenced by name, a warm standby, and the ServiceMonitor on. Start from the **Private-mirror preset**.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the namespace for the operator install
- [**Cert Manager**](/infra-catalog/kubernetes-cert-manager) -- the HARD prerequisite: issues and rotates the webhook certificate and keeps the retained CRDs' conversion trust current
- [**OpenTelemetry Collector**](/infra-catalog/kubernetes-otel-collector) -- the collector fleets this operator reconciles, one per pipeline shape
- [**kube-prometheus-stack**](/infra-catalog/kubernetes-kube-prometheus-stack) -- scrapes the operator's metrics when `serviceMonitorEnabled` is set
