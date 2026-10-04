# OpenSearch Operator

Installs the OpenSearch Kubernetes Operator — the opensearch-project's operator for running OpenSearch (the Apache-2.0 search and analytics engine) on Kubernetes — from the official `opensearch-operator` Helm chart. The operator reconciles `OpenSearchCluster` custom resources (declared with **OpenSearch**) into running search clusters with managed TLS, security bootstrap, safe rolling upgrades, and OpenSearch Dashboards.

This kind installs and configures the **engine**. Search clusters themselves are declared with OpenSearch resources — one per cluster — which this operator reconciles.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** — created only when `createNamespace` is `true`; otherwise installs into an existing namespace
- **Helm release** — the operator controller-manager Deployment, its RBAC (cluster-wide by default, namespace-scoped with `useRoleBindings`), ServiceAccount, and metrics Service
- **The OpenSearch CRDs** — derived from the pinned chart and applied by the module itself outside the release, NOT release-owned: the chart templates them as release resources with no keep-on-uninstall knob, so the module owns the CRD lifecycle. A chart bump moves the CRDs with it, uninstalling the operator never cascade-deletes OpenSearchCluster resources and their data (unless `crds.keepOnUninstall` is false), a reinstall re-adopts them, and a version below what the cluster's CRDs carry is refused before anything changes
- **kube-rbac-proxy sidecar** (default on) — shields the metrics endpoint behind Kubernetes RBAC. The chart's own default sidecar image was deleted upstream and can never be pulled; the module re-points it at the maintainer's quay.io repository at the chart's pinned tag, so the default posture actually installs

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster.

### Kubernetes Cluster

- **Cluster-admin credentials** — the default install creates CRDs and cluster-wide RBAC (ClusterRoles and ClusterRoleBindings), which requires cluster-level permissions; the namespace-scoped posture (`watchNamespace` + `useRoleBindings`) needs only namespace-level rights plus the CRDs already present.
- **A naming budget of 27 characters** on the resource name — the chart derives long child names from its fullname (the longest adds 36 characters) and Kubernetes caps object names at 63; the chart truncates the fullname but NOT the names built from it, so an over-long name fails at the API server mid-install.

## Deploy

### Console

Open the deployment store, find **OpenSearch Operator**, and click **Deploy**. The creation wizard walks you through namespace placement, the chart pin, watch scope and RBAC, operator runtime, manager endpoints, resources and scheduling, image sourcing, and the Helm-values escape hatch. Start from the **Standard preset** in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesOpenSearchOperator
metadata:
  name: opensearch-operator
  org: acme-corp
  env: prod
spec:
  namespace:
    value: opensearch-operator
  createNamespace: true
```

```shell
planton apply -f opensearch-operator.yaml
```

This installs the stable operator with a cluster-wide watch, module-owned CRDs, and the metrics endpoint shielded by kube-rbac-proxy. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to compose the operator with a managed namespace:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: search-platform
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline creates the namespace first, then installs the operator into it — the natural first half of a chart that also declares the OpenSearch clusters it reconciles.

## Key Configuration

These are the most important decisions when configuring an OpenSearch Operator install. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Chart pinning** — `chartVersion` defaults to **2.8.0**, the newest STABLE chart/operator pairing. The chart's default manager image tag is its appVersion, and the served 2.8.3/2.8.4/3.0.x charts all default to a PRERELEASE operator image while bundling next-generation `opensearch.org` API-group CRDs the stable operator does not serve. A version bump is a license and API-group re-check — pin those lines only after upstream cuts a stable 3.x operator release.

**Watch scope** — by default the operator watches ALL namespaces (cluster-wide RBAC): one install serves every OpenSearch cluster on the Kubernetes cluster. Set `watchNamespace` to fence it to one namespace; pair with `useRoleBindings` to swap ClusterRoleBindings for namespace-scoped RoleBindings — the shared-cluster posture needing no cluster-admin sign-off. The pairing is enforced: namespace-scoped RBAC cannot serve a cluster-wide operator. The fence is silent on the outside — a cluster declared beyond it is never reconciled, with no event pointing at the fence.

**CRD lifecycle is module-owned** — the CRDs are derived from the pinned chart, so a version bump moves the schema with the operator; uninstalling the operator never cascade-deletes OpenSearchCluster resources unless `crds.keepOnUninstall` is false; a version that is not published, or one below what the cluster's CRDs carry, is refused with the remedy before anything changes. In Helm values, `installCRDs` is re-pinned by the module after the merge, and `nameOverride`/`fullnameOverride` are off limits (they break the exported `deployment_name` output).

**DNS domain** — `dnsBase` is baked into the TLS certificates the operator generates; set it only on clusters with a non-default DNS domain, or every generated certificate's SANs mismatch the service names nodes advertise.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | The installation namespace | Composition, debugging |
| `release_name` | Helm release name (= metadata.name) | Helm-level operations |
| `deployment_name` | The controller-manager Deployment name (the chart's fixed component name within the release) | Monitoring, log collection |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard** — cluster-wide watch, pinned stable chart, chart defaults everywhere. Start from the **Standard preset**.

**Namespace-scoped** — a fenced watch with RoleBindings: the multi-tenant / no-cluster-admin posture. Start from the **Namespace scoped preset**.

**Private mirror** — the air-gap path: manager image from your own registry with pull secrets and explicit resources. Start from the **Private mirror preset**.

## Works With

- [**OpenSearch**](/infra-catalog/kubernetes-open-search) — the search clusters this operator reconciles; deploy the operator FIRST (it is the registered prerequisite)
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) — reference a managed namespace to compose governance (quotas, pod-security labels) with the installation
- [**kube-prometheus-stack**](/infra-catalog/kubernetes-kube-prometheus-stack) — scrape the shielded metrics endpoint with an RBAC-authorized ServiceMonitor
