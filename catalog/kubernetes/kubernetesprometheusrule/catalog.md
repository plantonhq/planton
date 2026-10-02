# Prometheus Rule

Declares a prometheus-operator PrometheusRule: a namespaced set of Prometheus alerting and recording rules that every Prometheus selecting the object loads and evaluates. Alerting rules turn a PromQL condition into alerts Alertmanager routes; recording rules precompute expensive queries into cheap series for dashboards and alerts. The spec is faithful to the upstream `monitoring.coreos.com/v1` PrometheusRule (pinned to prometheus-operator v0.94.1), after a Planton envelope that places the object and sets its own labels -- this is how you declare what your platform alerts on, next to the workloads it watches.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **A PrometheusRule** -- one namespaced object carrying your rule groups, which the prometheus-operator renders into the rule files of every Prometheus whose rule selector matches it.
- **Kubernetes Labels** -- your own labels on the object (the ones a Prometheus selects by), with resource metadata labels (resource name, kind, organization, environment) applied automatically on top.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **prometheus-operator CRDs installed** -- **Kube Prometheus Stack** installs them, along with the Prometheus that evaluates the rules.
- **A Prometheus that selects the object** -- the stack's default discovery loads every rule object; a fenced stack loads only objects carrying its `release` label.
- **Target namespace exists** -- reference an existing one or create it first.

## Deploy

### Console

Open the deployment store, find **Prometheus Rule**, and click **Deploy**. The creation wizard walks you through the namespace, the object's labels, and each rule group with its recording and alerting rules.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPrometheusRule
metadata:
  name: api-slo-alerts
  org: acme-corp
  env: prod
spec:
  namespace:
    value: monitoring
  groups:
    - name: api-slo-alerts
      interval: 30s
      rules:
        - alert: ApiErrorBudgetFastBurn
          expr: job:slo_errors_per_request:ratio_rate1h{job="api"} > (14.4 * 0.001)
          for: 2m
          labels:
            severity: page
          annotations:
            summary: The API is burning its error budget 14x too fast
            runbook_url: https://runbooks.example.com/api-error-budget-burn
```

```shell
planton apply -f prometheus-rule.yaml
```

The stack's Prometheus loads the rule on its next reload and pages through Alertmanager's `severity: page` route when the API burns its budget. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the namespace to a resource managed by another Cloud Resource:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: monitoring-namespace
      fieldPath: spec.name
  groups:
    - name: api-slo-alerts
      rules: []
```

The InfraPipeline creates the namespace first, then the rule. The Prometheus that loads it is chosen by labels, not by a reference, so keep a fenced stack's `release` label in the rule's `labels`.

## Key Configuration

These are the most important decisions when configuring a rule object. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The object's labels decide who loads it.** `labels` are the object's own metadata, the thing a Prometheus's rule selector matches. Under the stack's default discovery no label is needed; a stack on `release_managed_only` loads only objects labelled `release: <its release name>`. These are not the labels alerts carry -- those are each group's and rule's `labels`.

**One bad rule silences the whole object.** Prometheus refuses a rule file with one malformed rule, and the operator then drops every rule in the object. The spec refuses the common shapes (a rule that is both or neither recording and alerting, a recording rule with alert-only fields) before the apply; the PromQL itself is checked only by Prometheus. Split rules into objects by owner.

**Alert labels are the route, annotations are the page.** Alertmanager routes on a rule's `labels` (a `severity`, the environment and component), and a responder reads its `annotations` (a `summary`, a `runbook_url`). `for` and `keep_firing_for` set how noisy the alert is.

**Groups evaluate in order.** Recording rules that later alerts read belong earlier in the same group, so each evaluation reads fresh series. `query_offset` shifts evaluation into the past for series that arrive late by remote write.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `prometheus_rule_name` | Name of the created PrometheusRule (equals `metadata.name`) | Ordering resources that depend on the rules being in place |
| `namespace` | The namespace the rule object was created in | Confirming where the rules live |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Alert on an error budget** -- page only when a service burns its availability budget fast, and open a ticket when it burns slowly. Start from the **Error-Budget Burn Alerts** preset.

**Precompute what alerts and dashboards read** -- record a service's error ratio over several windows once, and read the cheap series everywhere. Start from the **Error-Ratio Recording Rules** preset.

## Works With

- [**Kube Prometheus Stack**](/cloud-catalog/kubernetes-kube-prometheus-stack) -- installs the CRDs and the Prometheus and Alertmanager that evaluate and route the rules.
- [**Kubernetes Namespace**](/cloud-catalog/kubernetes-namespace) -- the placement target the rule object lives in.
