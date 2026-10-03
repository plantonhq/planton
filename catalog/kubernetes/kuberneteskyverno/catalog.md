# Kyverno

Deploys Kyverno -- the Kubernetes-native policy engine -- from the official `kyverno` chart. Policies are written as plain Kubernetes YAML (ClusterPolicy / Policy resources), no new policy language: the fastest path from "we need guardrails" to enforced guardrails. This component installs the ENGINE only -- four controllers (admission always; background, cleanup, and reports individually optional) plus the policy CRDs. The policies themselves are declared separately, and one Kyverno serves the whole cluster.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **Helm Release** -- the `kyverno` chart, creating:
  - Deployment for the admission controller (fixed replicas or an HPA -- one sizing story, never both), which registers the validating/mutating webhook configurations at RUNTIME -- the chart templates none
  - Deployments for the background, cleanup, and reports controllers -- each independently enabled, sized, resourced, and scheduled
  - The policy CRDs (~23, across `kyverno.io`, `policies.kyverno.io`, `reports.kyverno.io`, and `wgpolicyk8s.io`) from the crds SUBCHART -- release-owned, so they delete with the release unless `crds.keepOnUninstall` flips the posture (see Key Configuration)
  - The Kyverno ConfigMap carrying resource filters, namespace exclusions, and the default-registry setting -- its name is exported as `config_map_name`
  - ServiceMonitors across all four controllers, only when `metrics.serviceMonitor` is enabled
- **Post-release webhook cleanup** -- module-owned cleanup of the runtime-registered webhook configurations at uninstall (the chart's own helper at this version targets the wrong API group), kept on by `webhooksCleanupEnabled` defaulting true
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Cluster-admin-grade permissions** -- the install creates cluster-scoped CRDs and webhook configurations.
- **cert-manager, only if you choose issuer-signed webhook certificates** -- the default posture needs nothing: Kyverno generates and rotates its own webhook TLS. Naming a cert-manager issuer requires cert-manager (and that issuer) to already be on the cluster.
- **Name budget** -- the resource name feeds Helm resource names; names longer than 47 characters fail the deploy loudly at install time.

## Deploy

### Console

Open the deployment store, find **Kyverno**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Dev cluster preset** for the chart defaults end to end, the **Production HA preset** for an availability-hardened admission path, or the **Air-gapped mirror preset** for clusters that cannot reach public registries, in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKyverno
metadata:
  name: kyverno
  org: acme-corp
  env: prod
spec:
  namespace:
    value: "kyverno"
  createNamespace: true
  admissionController:
    replicas: 3
    resources:
      requests:
        cpu: 100m
        memory: 384Mi
      limits:
        cpu: "1"
        memory: 1Gi
  backgroundController:
    enabled: true
  cleanupController:
    enabled: true
  reportsController:
    enabled: true
  metrics:
    serviceMonitor: true
```

```shell
planton apply -f kyverno.yaml
```

This deploys a three-replica admission path (the count Kyverno's own high-availability guidance names), all four controllers with explicit resource requests, and ServiceMonitors fanning out across every controller. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire Kyverno to a namespace managed by another Infra Component:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: kyverno-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline deploys the namespace first, then provisions Kyverno into it.

## Key Configuration

These are the most important decisions when configuring Kyverno. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The webhook lifecycle** -- The chart templates NO webhook configurations; the admission controller registers them at runtime once it is healthy. That makes uninstall the moment that matters: a webhook configuration left behind with no backing service blocks every matched admission cluster-wide, because Kyverno's runtime-registered rules default fail-closed. `webhooksCleanupEnabled` (default true) keeps both the chart's cleanup hook and the module's own post-release cleanup on -- leave it on. If webhooks are ever stranded anyway, the recovery is one command: `kubectl delete validatingwebhookconfiguration,mutatingwebhookconfiguration -l webhook.kyverno.io/managed-by=kyverno`.

**The CRD destroy contract** -- The crds subchart templates release-owned CRDs: destroying the release deletes them, and deleting the CRDs cascade-deletes EVERY ClusterPolicy, Policy, PolicyException, and policy report on the cluster. `crds.keepOnUninstall` (default false) is the single toggle that flips to the kept posture -- CRDs and policies survive destroy. Kept CRDs still carry this release's Helm ownership metadata, so a LATER Kyverno install must set `crds.install: false` and ride the existing CRDs.

**Declared lists REPLACE chart defaults** -- `config.excludeGroups` carries the chart default `["system:nodes"]`, and `features.omitEventTypes` carries `["PolicyApplied", "PolicySkipped"]`. Declaring either field REPLACES its default: re-include the default entries unless you mean to drop them. `config.resourceFiltersExclude` removes entries from the chart's default resource-filter list, and each entry must match a default-list entry byte-for-byte.

**The fail-open safety valve** -- `features.forceFailurePolicyIgnore` forces EVERY webhook rule to failurePolicy=Ignore: an engine outage can no longer block admissions anywhere -- and enforcement silently stops applying during that outage. An honest trade for clusters where availability outranks policy guarantees; leave it off where enforcement is the point.

**Sizing the admission path** -- `admissionController.replicas` (fixed count) and `admissionController.autoscaling` (an HPA) are one sizing story each -- declare one, never both; with an HPA declared the chart ignores the replica count. Three replicas is the high-availability floor Kyverno's own documentation names. The background, cleanup, and reports controllers each take their own `replicas`, `resources`, and scheduling.

**Background scanning is report-only** -- The background scan re-evaluates EXISTING resources against audit-mode policies on an interval (`backgroundScan.interval`, Go-duration format, default 1h) and writes policy reports -- it never mutates or blocks anything. `workers` (default 2) trades API-server load for scan throughput.

**Air-gap in one field** -- `imageRegistry` reroutes every Kyverno image through your mirror, including the CRD-migration hook's non-obvious `reg.kyverno.io` home. Separately, the ENGINE's default-registry mutation (`config.enableDefaultRegistryMutation`, chart default ON) rewrites bare image references in resources it admits to `config.defaultRegistry` (`docker.io` unless changed) -- a different registry story that affects other workloads' images, not Kyverno's own.

**Native VAP generation** -- `features.generateValidatingAdmissionPolicy` offloads eligible validations to the API server's own ValidatingAdmissionPolicy objects, cutting webhook round-trips for those rules.

**Webhook TLS** -- By default Kyverno generates and rotates its own webhook certificates: zero prerequisites. Declaring `certificates.certManager` switches issuance to cert-manager; naming no issuer lets the chart create its own self-signed ClusterIssuer, and `issuerName` (with `issuerKind`) chains the certificates to trust you already operate.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesClusterIssuer** | `certificates.certManager.issuerName` | `metadata.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace Kyverno runs in | Locating the engine for diagnostics |
| `release_name` | Helm release name (equals metadata.name) | Helm management and debugging |
| `admission_service_name` | The admission controller's Service name | Webhook diagnostics and network policy |
| `config_map_name` | The Kyverno ConfigMap carrying filters and exclusions | Auditing effective engine configuration |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Dev cluster** -- The smallest declarable Kyverno: the chart defaults end to end -- single admission replica, background/cleanup/reports at their chart defaults, self-managed webhook TLS. Start from the **Dev cluster preset**.

**Production HA** -- Three admission replicas with explicit resources, all optional controllers on, and ServiceMonitor fan-out across the four controllers. Start from the **Production HA preset**.

**Airgapped mirror** -- Every Kyverno image rerouted through an internal registry (including the CRD-migration hook image), pull secrets attached, and the engine's default-registry mutation pointed at the mirror. Start from the **Air-gapped mirror preset**.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the namespace for the Kyverno install
- [**Cert Manager Cluster Issuer**](/infra-catalog/kubernetes-cluster-issuer) -- signs the webhook certificates when issuance is switched to cert-manager
- [**kube-prometheus-stack**](/infra-catalog/kubernetes-kube-prometheus-stack) -- scrapes all four controllers when the ServiceMonitor is enabled
