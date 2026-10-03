# Flink Operator

Deploys the Apache Flink Kubernetes Operator -- the official ASF controller that turns `FlinkDeployment` declarations (declared with KubernetesFlinkDeployment) into running Flink clusters -- from the official chart served per version by Apache (chart version = operator version = image tag; `1.15.0` pinned, and the modules pin the image tag explicitly where the chart's own default is the unpinned `latest`). This component installs the ENGINE only: Flink pipelines are declared separately, one KubernetesFlinkDeployment per cluster. One operator per namespace by construction -- the chart hardcodes its webhook Service, certificate, and issuer names, so a second release in the same namespace collides -- and one cluster-wide-watching operator is the normal posture. With the admission webhook on (the upstream default), KubernetesCertManager is a hard prerequisite: the chart renders cert-manager Issuer/Certificate resources UNCONDITIONALLY and there is no self-signed fallback. The four flink.apache.org CRDs ride the chart's `crds/` directory -- installed once, never upgraded by chart bumps, and KEPT on uninstall along with every Flink declaration. Uses a Kubernetes Provider Connection for cluster access.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **The flink.apache.org CRDs** -- `flinkdeployments`, `flinksessionjobs`, `flinkstatesnapshots`, and `flinkbluegreendeployments`, installed by Helm from the chart's `crds/` directory: installed once, never upgraded on chart bumps (apply the new release's CRD files manually when a bump changes them), and left on the cluster on uninstall -- removing the operator never deletes Flink declarations
- **Webhook keystore Secret** -- the chart's default webhook keystore credential is a HARDCODED PUBLIC PASSWORD; the modules generate a random password per install, materialize it as a module-owned Secret (`<name>-webhook-keystore`), and point the chart at it, re-pinning `useDefaultPassword: false` after the escape-hatch merge so the public default cannot resurface
- **Watched namespaces** -- under the fenced posture (non-empty `watchNamespaces`), each listed namespace is created by the modules BEFORE the Helm release (the chart plants job RBAC into them but does not create them)
- **Helm Release** -- the `flink-kubernetes-operator` chart, creating:
  - Deployment for the operator with the configured replica count and resources; with more than one replica the modules render the operator's leader-election config for you (the chart refuses multi-replica installs without it, by design) and ONE active reconciler leads while the rest stand by warm
  - With the webhook enabled: the admission webhook, its chart-fixed Service (`flink-operator-webhook-service`), and cert-manager Issuer/Certificate resources for its serving certificate
  - The `flink` job service account (or your `jobServiceAccount` name) with reconcile RBAC, marked `helm.sh/resource-policy: keep` so running jobs never lose their identity on uninstall
  - RBAC scoped to the watch fence: an empty `watchNamespaces` yields cluster-wide RBAC; a non-empty list scopes RBAC AND the webhook's namespaceSelector to exactly that list
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **cert-manager, unless you disable the webhook** -- with `webhookEnabled` true (the default), the chart's webhook certificate is issued and rotated by cert-manager with NO self-signed fallback. Install KubernetesCertManager first, or set `webhookEnabled: false` and trade admission-time validation for reconcile-loop validation.
- **Cluster-admin-grade permissions on the first install** -- applying the cluster-scoped flink.apache.org CRDs and (under the cluster-wide posture) the operator's ClusterRole requires them.
- **A name within budget** -- keep `metadata.name` at 45 characters or fewer: the longest derived name is the module-generated `-webhook-keystore` Secret (17 characters) against the Kubernetes 63-character cap. Both engines fail loudly over the budget. The chart-fixed webhook artifact names are excluded from the budget -- they do not derive from the resource name at all, which is also why one operator per namespace is the grain.

## Deploy

### Console

Open the deployment store, find **Flink Operator**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Default preset** for the standard cluster-wide install with the webhook on, or the **Fenced HA preset** for the multi-tenant posture with a warm standby, in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesFlinkOperator
metadata:
  name: flink-operator
  org: acme-corp
  env: prod
spec:
  namespace:
    value: "flink-system"
  createNamespace: true
  watchNamespaces:
    - stream-team-a
    - stream-team-b
  replicas: 2
```

```shell
planton apply -f flink-operator.yaml
```

This deploys the operator with a warm standby behind leader election (configured for you) in the `flink-system` namespace, fenced to the `stream-team-a` and `stream-team-b` namespaces: the modules create both, the chart plants job RBAC in each, and the operator's RBAC AND admission webhook are scoped to exactly that list. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the operator to a namespace managed by another Infra Component:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: flink-system-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline deploys the namespace first, then provisions the operator into it.

## Key Configuration

These are the most important decisions when configuring the Flink Operator. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The webhook lifecycle, read this first** -- With `webhookEnabled` true (the upstream default this spec keeps), cert-manager is a hard prerequisite and both webhook configurations FAIL CLOSED: if the webhook cannot be reached (cert-manager absent, operator down), EVERY flink.apache.org admission in scope is rejected -- a policy-engine blast radius, not a soft degradation. It is also what makes bad Flink declarations fail at admission with a real message instead of a silent reconcile stall. `webhookEnabled: false` removes the webhook, the certificate machinery, and the cert-manager dependency; the operator still validates in its reconcile loop, with failures surfacing on CR status instead of at admission.

**The watch fence scopes RBAC and the webhook together** -- Empty `watchNamespaces` = cluster-wide: the operator watches every namespace. A non-empty list is the fenced multi-tenant posture: the modules create each listed namespace before the release, the chart plants job RBAC in each, and the operator's reconcile RBAC AND the webhook's namespaceSelector confine to exactly that list. Flink declarations outside the fence are ignored WITHOUT an error -- a missing namespace in the list looks like a deployment that never reconciles.

**The hardcoded password never ships** -- The chart's default webhook keystore credential is a hardcoded public password. The modules generate a random per-install Secret instead and RE-PIN `useDefaultPassword: false` after the escape-hatch merge, so the public default cannot resurface through `helmValues`.

**The CRD lifecycle is upstream's keep-forever posture** -- The chart ships its four CRDs from its `crds/` directory: Helm installs them once, NEVER upgrades them on chart bumps, and LEAVES them (and every Flink declaration) on uninstall. The modules neither re-own nor template them -- when a chart bump changes the CRDs, apply the new release's CRD files manually.

**Chart version lockstep** -- `chartVersion` (default `"1.15.0"`) pins the chart, and the chart is served per version from the Apache downloads directory -- the version is part of the repository URL itself. Chart version = operator version = image tag; bumps never touch the `crds/`-directory CRDs.

**Replicas are warm standbys** -- A single replica suits most clusters. With `replicas: 2` the modules render the operator's leader-election config for you (the chart refuses multi-replica installs without it) and ONE active reconciler leads -- reconcile throughput does not change with replicas.

**Operator config is cluster-wide defaults** -- `operatorConfig` entries (Flink's own config format, `kubernetes.operator.*` keys appended over the chart defaults) become defaults for EVERY FlinkDeployment this operator manages -- per-pipeline configuration belongs on each KubernetesFlinkDeployment, not here.

**The job service account stays the upstream convention** -- `jobServiceAccount` (default `flink`) is the identity Flink job pods run as; every KubernetesFlinkDeployment references it by that name by default. The chart marks it `helm.sh/resource-policy: keep`, so it survives uninstall and running jobs never lose their identity.

**Sizing is deliberate** -- `resources` empty means the chart defaults; the operator is a JVM, and production installs typically set requests explicitly -- an OOM-killed operator strands every reconciling Flink pipeline.

**The image dial covers the operator ONLY** -- `imageRegistry` rewrites the registry part of the operator's own image (`ghcr.io/apache/flink-kubernetes-operator`) -- the air-gap path for the operator. It does NOT rewrite the Flink images deployments run: those ride each KubernetesFlinkDeployment's own image field.

**The Helm-values escape hatch has one re-pinned key** -- `helmValues` merges LAST over the typed fields (Helm `-f` semantics, identical on both engines) for the chart surface beyond them: logging framework, operator volume mounts, health-probe tuning, JVM args. Exactly one key is re-pinned AFTER this merge: `webhook.keystore.useDefaultPassword: false`, so the chart's public default password cannot be resurrected.

**The install is deliberately blocking** -- The Helm release waits for the operator to become Available (atomic, 600-second timeout), so an unpullable image, an absent cert-manager, or a broken config fails THIS apply with a readiness timeout instead of surfacing later as FlinkDeployments that mysteriously never reconcile.

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
| `release_name` | Helm release name (equals metadata.name; the chart's fullname is pinned to it) | Helm management and debugging |
| `job_service_account` | Service account Flink job pods run as | KubernetesFlinkDeployment declarations reference it |
| `watched_namespaces` | Namespaces the operator watches for Flink CRs (empty = cluster-wide) | Verifying a pipeline's namespace is inside the fence |
| `webhook_service` | The operator's webhook Service (chart-fixed `flink-operator-webhook-service`); empty when the webhook is disabled | Diagnosing admission failures |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard operator** -- One operator in its own `flink-system` namespace watching cluster-wide, the webhook on, chart defaults for sizing. Requires cert-manager. Start from the **Default preset**.

**Fenced multi-tenant platform** -- The operator watches an explicit namespace list; the modules create the namespaces and the chart plants job RBAC in each; a warm standby behind leader election; reconcile cadence tuned via `operatorConfig`. Start from the **Fenced HA preset**.

**Webhook-less install** -- `webhookEnabled: false` on clusters without cert-manager: no webhook, no certificates, no dependency -- validation moves from admission time to the reconcile loop.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the namespace for the operator install
- [**Cert Manager**](/infra-catalog/kubernetes-cert-manager) -- a hard prerequisite whenever the webhook is enabled: it issues and rotates the webhook's serving certificate
- [**Flink Deployment**](/infra-catalog/kubernetes-flink-deployment) -- declares the Flink clusters and jobs this operator reconciles
