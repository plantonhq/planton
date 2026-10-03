# Kubernetes Flink Operator

## When NOT to Use This

**This component installs the ENGINE, not a Flink job.** The Apache
Flink Kubernetes Operator reconciles `FlinkDeployment` custom
resources — declared with KubernetesFlinkDeployment — into running
Flink clusters (and, unmodeled in this catalog,
FlinkSessionJob/FlinkStateSnapshot/FlinkBlueGreenDeployment CRs
authored directly). Install the operator once, then declare Flink
deployments against it.

Also not the right kind when:

- **You want a Flink cluster or job** — that is
  KubernetesFlinkDeployment; this kind is the controller that
  reconciles it.
- **You cannot run cert-manager and want the webhook** — with the
  webhook enabled (the upstream default this spec keeps),
  KubernetesCertManager is a hard prerequisite: the chart renders
  cert-manager Issuer/Certificate resources UNCONDITIONALLY and there
  is NO self-signed fallback. `webhook_enabled: false` is the way out
  (see below for the trade).
- **You want a second operator in the same namespace** — the chart
  hardcodes its webhook Service, certificate, and issuer names
  (`flink-operator-webhook-service`, `flink-operator-serving-cert`),
  so a second release in the same namespace collides by construction.
  One operator per namespace is the grain; one cluster-wide-watching
  operator is the normal posture.

## Overview

**KubernetesFlinkOperator** installs the Apache Flink Kubernetes
Operator — the official ASF controller for Flink on Kubernetes — from
the official chart served per version at
https://downloads.apache.org/flink/flink-kubernetes-operator-\<version\>/.
Version 1.15.0 is the pinned default (chart version = operator
version = image tag; the modules pin the image tag explicitly — the
chart's own default is the unpinned `latest`).

**Key design points:**

- **The webhook lifecycle, read this first.** With the webhook
  enabled (the upstream default), the chart renders cert-manager
  Issuer/Certificate resources UNCONDITIONALLY and the webhook trusts
  the API server through cert-manager's CA injection — there is NO
  self-signed fallback, which makes KubernetesCertManager a hard
  prerequisite. Both webhook configurations are FAIL-CLOSED: if the
  webhook cannot be reached (cert-manager absent, operator down),
  EVERY flink.apache.org admission in scope is rejected — a
  policy-engine blast radius, not a soft degradation. It is also what
  makes bad Flink declarations fail at admission with a real message
  instead of a silent reconcile stall. `webhook_enabled: false`
  removes the webhook, the certificate machinery, and the
  cert-manager dependency; the operator still validates in its
  reconcile loop (failures surface on CR status instead of at
  admission).
- **The hardcoded keystore password never ships.** The chart's
  default webhook keystore credential is a HARDCODED PUBLIC
  PASSWORD. The modules generate a random keystore password per
  install, materialize it as a module-owned Secret
  (`<name>-webhook-keystore`), and point the chart's
  `webhook.keystore.passwordSecretRef` at it —
  `useDefaultPassword: false` is additionally RE-PINNED after the
  escape-hatch merge, so the public default cannot resurface through
  `helm_values`.
- **The CRDs are kept on uninstall — by upstream design.** The chart
  ships its four flink.apache.org CRDs from its `crds/` directory:
  Helm installs them once, NEVER upgrades them on chart upgrades, and
  LEAVES them (and every Flink declaration) on uninstall. Apply the
  new release's CRD files manually when a chart bump changes them.
- **Singleton per namespace, by chart-fixed names.** The webhook
  artifact names are chart-fixed, not fullname-derived — the reason a
  second release in the same namespace collides, and the reason those
  names are excluded from the name budget. Keep `metadata.name` at 45
  characters or fewer: the longest derived name is the
  module-generated `-webhook-keystore` Secret (17 chars) against the
  Kubernetes 63-character cap; both engines fail loudly over the
  budget.
- **Operator config is cluster-wide defaults.** `operator_config`
  entries (Flink's own config format, `kubernetes.operator.*` keys
  appended over the chart defaults) become defaults for EVERY
  FlinkDeployment this operator manages — per-pipeline configuration
  belongs on each KubernetesFlinkDeployment, not here.
- **`helm_values` is the escape hatch** — additional chart values
  merged LAST over everything the typed fields render (Helm `-f`
  semantics, identical on both engines): logging framework, operator
  volume mounts, health-probe tuning, JVM args — a safety valve,
  never the primary interface. One key is re-pinned AFTER this merge:
  `webhook.keystore.useDefaultPassword: false` (see above).

## Essential Configuration Fields

### Required

- **`spec.namespace`**: namespace to install the operator into —
  literal or a KubernetesNamespace reference (`create_namespace` to
  own it)

### Common

- **`spec.chart_version`**: version pin (default `1.15.0`; the chart
  is served per version from the Apache downloads directory — the
  version is part of the repository URL itself; bumps never touch the
  `crds/`-directory CRDs)
- **`spec.webhook_enabled`**: default true (see the webhook lifecycle
  above); false removes the webhook, the certificates, and the
  cert-manager dependency
- **`spec.watch_namespaces`**: empty = every namespace; with a list
  set, the chart scopes RBAC AND the admission webhook to exactly
  these namespaces — Flink declarations outside them are ignored
  without an error. The modules create each listed namespace before
  the Helm release (the chart plants job RBAC into them and does not
  create the namespaces)
- **`spec.replicas`**: operator replicas (default 1); more than 1
  requires leader election — the modules render the operator's
  leader-election config for you (the chart REFUSES multi-replica
  installs without it, by design)
- **`spec.operator_config`**: cluster-wide operator defaults (see
  above)
- **`spec.job_service_account`**: service account Flink JOB pods run
  as (default `flink` — the name every FlinkDeployment references by
  default); the chart marks it `helm.sh/resource-policy: keep`, so it
  survives uninstall and running jobs never lose their identity
- **`spec.resources`**: operator container resources — empty = the
  chart defaults; the operator is a JVM, production installs
  typically set requests explicitly
- **`spec.image_registry`**: registry replacing the registry part of
  the operator image (`ghcr.io/apache/flink-kubernetes-operator`) —
  the air-gap path for the operator's own image; does NOT rewrite the
  Flink images deployments run (those ride each
  KubernetesFlinkDeployment's own image field)
- **`spec.image_pull_secrets`**: names of existing image-pull Secrets
  in the namespace
- **`spec.scheduling`**: node selector, tolerations, priority class
  for the operator pods
- **`spec.helm_values`**: the escape hatch (see above for the one
  re-pinned key)

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Namespace the operator runs in |
| `release_name` | Helm release name (= `metadata.name`; the chart's fullname is pinned to it) |
| `job_service_account` | Service account Flink job pods run as — FlinkDeployment declarations reference it |
| `watched_namespaces` | Namespaces the operator watches for Flink CRs (empty = cluster-wide) |
| `webhook_service` | The operator's webhook Service (chart-fixed `flink-operator-webhook-service`); empty when the webhook is disabled |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace, field path `spec.name`).
- **KubernetesCertManager is a prerequisite** whenever the webhook is
  enabled — the chart's webhook certificate is issued and rotated by
  cert-manager, with no self-signed fallback.
- **KubernetesFlinkDeployment resources depend on this component**:
  the operator must be running before their `FlinkDeployment`
  resources reconcile, and they reference the exported
  `job_service_account`. Under the fenced posture, declarations
  outside `watched_namespaces` are ignored without an error.
- **The install is deliberately blocking**: the Helm release waits for
  the operator to become Available (atomic, 600s timeout), so an
  unpullable image, an absent cert-manager, or a broken config fails
  THIS apply with a readiness timeout instead of surfacing later as
  FlinkDeployments that mysteriously never reconcile.

## Examples

### Standard install

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesFlinkOperator
metadata:
  name: flink-operator
spec:
  namespace:
    value: flink-system
  createNamespace: true
```

### Fenced, leader-elected standby

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesFlinkOperator
metadata:
  name: flink-operator
spec:
  namespace:
    value: flink-system
  createNamespace: true
  watchNamespaces:
    - stream-team-a
    - stream-team-b
  replicas: 2
  operatorConfig:
    kubernetes.operator.reconcile.interval: 15 s
```

### Webhook-less (no cert-manager dependency)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesFlinkOperator
metadata:
  name: flink-operator
spec:
  namespace:
    value: flink-system
  createNamespace: true
  webhookEnabled: false
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
