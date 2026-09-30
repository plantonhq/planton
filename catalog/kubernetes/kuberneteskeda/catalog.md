# KEDA

Installs KEDA — Kubernetes Event-Driven Autoscaling — from the official Helm chart. KEDA scales workloads on real-world signals (queue depth, stream lag, database rows, cron schedules, cloud metrics — 70+ scalers) instead of only CPU/memory: its operator watches ScaledObject/ScaledJob resources, drives the workload's HPA including scale-to-ZERO, and serves the `external.metrics.k8s.io` API the HPA controller reads. This component installs the ENGINE only — the scaling declarations (ScaledObjects, TriggerAuthentications) live alongside the workloads they scale. One installation per cluster.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Helm Release** -- the `keda-operator` Deployment, the `keda-operator-metrics-apiserver` Deployment (registers the cluster-wide `v1beta1.external.metrics.k8s.io` APIService), the `keda-admission-webhooks` Deployment, and RBAC
- **CRDs** -- ScaledObject, ScaledJob, TriggerAuthentication, and companions — annotated to survive uninstall by default, so removing the release does not cascade-delete every scaling declaration in the cluster
- **ServiceAccount** -- the `keda-operator` identity, optionally annotated for keyless cloud metric access (EKS IRSA, Azure Workload Identity, GCP Workload Identity)
- **Namespace** (optional) -- created with standard governance labels when `createNamespace` is true (`keda` is the upstream convention)

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.

### Kubernetes Cluster

- A cluster WITHOUT an existing external-metrics provider — the `v1beta1.external.metrics.k8s.io` APIService is a singleton and Kubernetes allows only one.
- With cert-manager-issued internal TLS: cert-manager must already run on the cluster (the Cert Manager component).
- With the Prometheus ServiceMonitor: the Prometheus operator CRDs must already be on the cluster — the release fails to install without them.
- For the keyless cloud identity arms (IRSA / Azure Workload Identity / GCP Workload Identity): the cloud-side trust or binding must name the `keda-operator` service account.

## Deploy

### Console

Open the deployment store, find **KEDA**, and click **Deploy**. The creation wizard walks you through placement, the chart pin and CRD lifecycle, cloud identity for keyless scalers, internal TLS, availability, observability, and scheduling. Start from the **Cluster Standard** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesKeda
metadata:
  name: keda
  org: acme-corp
  env: prod
spec:
  namespace:
    value: keda
  createNamespace: true
```

```shell
planton apply -f keda.yaml
```

The engine then watches ScaledObjects in all namespaces — deploy the scaling declarations alongside the workloads they scale. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, hand KEDA's internal TLS to a cert-manager issuer managed by another Cloud Resource:

```yaml
spec:
  namespace:
    value: keda
  createNamespace: true
  certificates:
    type: cert_manager
    certManagerIssuer:
      clusterIssuer:
        name:
          valueFrom:
            name: internal-ca
```

The InfraPipeline deploys the issuer first, then installs KEDA with certificates signed and renewed by it.

## Key Configuration

These are the most important decisions when configuring KEDA. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The namespace is effectively permanent** -- kept CRDs (the default) retain the release's namespace in their ownership metadata, so moving the installation later means deleting the CRDs — which cascades to every ScaledObject and TriggerAuthentication in the cluster. `keda` is the convention.

**CRD lifecycle** -- CRDs install with the release and survive uninstall by default: the scaling declarations are the cluster's, not the release's. Deleting them on uninstall is the one genuinely destructive dial.

**Cloud identity for keyless scalers** -- the `podIdentity` arms bind the `keda-operator` service account to a cloud identity so scalers read cloud metric sources (SQS, Azure Monitor, Stackdriver) with zero stored keys. The arms are INDEPENDENT — a multi-cloud cluster may enable more than one.

**Internal TLS** -- KEDA's components speak TLS internally; cert-manager can own those certificates in place of the self-generated ones once autoscaling is load-bearing.

**One external-metrics provider per cluster** -- the APIService is a singleton. If something else already serves external metrics (a Prometheus adapter), the two cannot coexist.

**The escape hatch** -- `helmValues` carries additional chart values as a YAML document, merged LAST over everything the typed fields render — never the substitute for typed fields, never a place for secrets.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesIssuer** | `certificates.certManagerIssuer.issuer.name` | `status.outputs.issuer_name` |
| **KubernetesClusterIssuer** | `certificates.certManagerIssuer.clusterIssuer.name` | `status.outputs.cluster_issuer_name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Installation namespace | Debugging and composition |
| `release_name` | Helm release name (always `keda`) | Debugging the release (`helm status`) |
| `operator_service_account_name` | Always `keda-operator` | The subject cloud-side keyless bindings are written against (IRSA trust policy, Azure federated credential, GCP Workload Identity member) |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard install** -- conventional `keda` namespace, pinned chart, CRDs managed by the release. Start from the **Cluster Standard** preset.

**EKS with IRSA scalers** -- IRSA binds the operator to an IAM role so AWS scalers (SQS, CloudWatch) read metrics with zero stored keys. Start from the **EKS with IRSA for AWS Scalers** preset.

**Load-bearing autoscaling** -- standby operator replicas, `system-cluster-critical` priority, cert-manager TLS, and Prometheus telemetry. Start from the **HA Production** preset.

## Works With

- [**Kubernetes Deployment**](/cloud-catalog/kubernetes-deployment) -- the workloads ScaledObjects and ScaledJobs scale, including to zero (the same holds for StatefulSets and Jobs)
- [**Cert Manager**](/cloud-catalog/kubernetes-cert-manager) -- can own KEDA's internal TLS certificates in place of the self-generated ones
- [**Cert Manager Issuer**](/cloud-catalog/kubernetes-issuer) -- the issuer KEDA's certificates reference when cert-manager owns them
- [**Metrics Server**](/cloud-catalog/kubernetes-metrics-server) -- the complementary pipeline: metrics-server covers instantaneous CPU/memory, KEDA covers event-driven and external signals
- [**Kubernetes HorizontalPodAutoscaler**](/cloud-catalog/kubernetes-horizontal-pod-autoscaler) -- KEDA drives HPAs under the hood; hand-authored HPAs must not target the same workload a ScaledObject manages
