# KubernetesPrometheusRule

Declares a prometheus-operator `PrometheusRule`: a namespaced set of Prometheus alerting and recording rules that every Prometheus selecting the object loads and evaluates. The spec mirrors the upstream `monitoring.coreos.com/v1` PrometheusRule field for field (pinned to prometheus-operator v0.94.1, the operator `KubernetesKubePrometheusStack` installs), after a small Planton envelope: the namespace, and the object's own labels and annotations.

## What Gets Created

- **One PrometheusRule** in the target namespace, carrying the rule groups exactly as written.
- **Labels on the object**: the spec's `labels`, with Planton's identity labels (`planton.ai/resource`, `resource-name`, `resource-kind`, `resource-id`, `organization`, `environment`) stamped on top.

The prometheus-operator renders the groups into the rule files of every Prometheus (and Thanos Ruler) whose rule selector matches the object, and that Prometheus reloads and starts evaluating them. Nothing else is created.

## Which Prometheus Loads the Rules

A Prometheus picks rules up by the labels on the object and the namespace it lives in:

- **KubernetesKubePrometheusStack with its default discovery (`all_monitors`)** loads every PrometheusRule in every namespace. No label is needed.
- **KubernetesKubePrometheusStack with `release_managed_only` discovery** (a fenced Prometheus, such as a receiver beside an agent) loads only objects labelled `release: <the stack's release_name output>`. Put that label in `spec.labels`.
- **A Prometheus installed any other way** selects by its own `ruleSelector` and `ruleNamespaceSelector`; set the labels it names.

## Recording Rules and Alerting Rules

Each rule is exactly one of:

- **A recording rule** (`record`): precomputes an expression into a new series, so dashboards and alerts read a cheap series instead of re-running an expensive query. It may carry `labels`, never `for`, `keep_firing_for` or `annotations`.
- **An alerting rule** (`alert`): every series its expression returns is an active alert. `for` holds it pending before it fires, `keep_firing_for` keeps it firing past a flap, `labels` are what Alertmanager routes on, and `annotations` are what a responder reads.

The spec refuses a rule that is both or neither, and a recording rule that carries alert fields, because Prometheus refuses the whole rule file for either mistake and the operator then drops the entire object.

## Prerequisites

- The prometheus-operator CRDs on the cluster. **KubernetesKubePrometheusStack** installs them, and it is this kind's registry prerequisite.
- A Prometheus that selects the object (see above).
- The target namespace (**KubernetesNamespace**).

## Quick Start

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPrometheusRule
metadata:
  name: api-slo
spec:
  namespace:
    value: monitoring
  groups:
    - name: api-slo-alerts
      rules:
        - alert: ApiErrorBudgetFastBurn
          expr: job:slo_errors_per_request:ratio_rate1h > (14.4 * 0.001)
          for: 2m
          labels:
            severity: page
          annotations:
            summary: "{{ $labels.job }} is burning its error budget fast"
            runbook_url: https://runbooks.example.com/api-error-budget-burn
```

```shell
planton apply -f prometheus-rule.yaml
```

## Configuration Reference

### Envelope

| Field | Description |
|-------|-------------|
| `namespace` | Namespace the object is created in. A literal or a reference to a KubernetesNamespace's `spec.name`. Required. |
| `labels` | The object's own labels. How a Prometheus with a rule selector decides the rules are its own. |
| `annotations` | The object's own annotations, for people and tools reading it. They do not reach alerts. |

### Groups (`groups[]`)

| Field | Description |
|-------|-------------|
| `name` | Group name, unique within the object. Required. |
| `interval` | Evaluation interval as a Prometheus duration. Unset, the Prometheus's evaluation interval applies. |
| `query_offset` | Shifts each evaluation into the past, for late-arriving (remote-written) series. Prometheus >= 2.53. |
| `limit` | Most alerts or series one rule may produce per evaluation. Zero means no limit. |
| `partial_response_strategy` | Thanos Ruler only: `abort` or `warn` on a partial store response. |
| `labels` | Labels added to every series or alert the group produces. Prometheus >= 3.0. |
| `rules` | The rules, evaluated in order. |

### Rules (`groups[].rules[]`)

| Field | Description |
|-------|-------------|
| `record` | Series name a recording rule writes. Set instead of `alert`. |
| `alert` | Alert name an alerting rule raises. Set instead of `record`. |
| `expr` | The PromQL expression. Required. |
| `for` | How long the expression must hold before the alert fires. Alerting rules only. |
| `keep_firing_for` | How long the alert keeps firing after the expression clears. Alerting rules only. |
| `labels` | Labels added to the alert or series. |
| `annotations` | Text attached to each alert. Alerting rules only. |

## Composing in Infra Charts

Reference the namespace from a KubernetesNamespace in the same chart; the pipeline creates the namespace first:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: monitoring-namespace
      fieldPath: spec.name
```

The registry prerequisite orders the rule after the KubernetesKubePrometheusStack that installs its CRDs. The `release` label a fenced stack selects by is a plain string; keep the stack's `release_name` output and the label in step.

## Stack Outputs

| Output | Description |
|--------|-------------|
| `prometheus_rule_name` | Name of the created PrometheusRule (equals `metadata.name`). |
| `namespace` | Namespace the PrometheusRule was created in. |

## Related Components

- **KubernetesKubePrometheusStack**: installs the CRDs and the Prometheus that evaluates the rules.
- **KubernetesNamespace**: where the object lives.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
