# KubernetesPrometheusRule

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

KubernetesPrometheusRuleSpec defines a prometheus-operator PrometheusRule: a
namespaced set of Prometheus alerting and recording rules that every
Prometheus (and Thanos Ruler) whose rule selector matches the object loads
and evaluates.

100% fidelity with the upstream PrometheusRule custom resource
(monitoring.coreos.com/v1), pinned to prometheus-operator v0.94.1 -- the
operator the kube-prometheus-stack chart 91.8.2 ships, which is what
KubernetesKubePrometheusStack installs. The upstream spec (`groups`) follows
the Planton envelope directly; there is no nested `prometheus_rule`
sub-message.

The envelope: `namespace` places the object, and `labels` and `annotations`
are the object's own metadata. The labels are configuration here, not
decoration: a Prometheus picks rules up through its rule selector, so a rule
object that a fenced Prometheus should load carries the labels that selector
asks for. KubernetesKubePrometheusStack's default discovery (all_monitors)
loads every PrometheusRule in the cluster; its release_managed_only discovery
loads only objects labelled `release: <the stack's release name>` (its
release_name output).

What makes an alert worth its page: give every alerting rule the labels its
alert routing reads (a `severity`, the environment and component the
receivers group by) and the annotations a responder needs (a summary, and a
`runbook_url` whose first line is the first action). Alertmanager routes on
the labels and renders the annotations; a rule with neither fires into a
default route nobody watches.

## Example

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPrometheusRule
metadata:
  name: api-slo
spec:
  namespace:
    value: monitoring
  labels:
    release: kube-prometheus-stack
  groups:
    - name: api-slo-recording
      interval: 30s
      rules:
        - record: job:slo_errors_per_request:ratio_rate5m
          expr: sum by (job) (rate(http_requests_total{code=~"5.."}[5m])) / sum by (job) (rate(http_requests_total[5m]))
        - record: job:slo_errors_per_request:ratio_rate1h
          expr: sum by (job) (rate(http_requests_total{code=~"5.."}[1h])) / sum by (job) (rate(http_requests_total[1h]))
    - name: api-slo-alerts
      interval: 30s
      limit: 50
      labels:
        component: api
      rules:
        - alert: ApiErrorBudgetFastBurn
          expr: job:slo_errors_per_request:ratio_rate1h > (14.4 * 0.001) and job:slo_errors_per_request:ratio_rate5m > (14.4 * 0.001)
          for: 2m
          keep_firing_for: 5m
          labels:
            severity: page
          annotations:
            summary: "{{ $labels.job }} is burning its 30-day error budget 14x too fast"
            runbook_url: https://runbooks.example.com/api-error-budget-burn
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.groups` | `[]KubernetesPrometheusRuleGroup` |  |  |  |
| `spec.groups[].name` | `string` | yes |  |  |
| `spec.groups[].interval` | `string` |  |  |  |
| `spec.groups[].query_offset` | `string` |  |  |  |
| `spec.groups[].limit` | `int32` |  |  |  |
| `spec.groups[].partial_response_strategy` | `string` |  |  |  |
| `spec.groups[].labels` | `map<string, string>` |  |  |  |
| `spec.groups[].rules` | `[]KubernetesPrometheusRuleRule` |  |  |  |
| `spec.groups[].rules[].record` | `string` |  |  |  |
| `spec.groups[].rules[].alert` | `string` | yes |  |  |
| `spec.groups[].rules[].expr` | `string` | yes |  |  |
| `spec.groups[].rules[].for` | `string` |  |  |  |
| `spec.groups[].rules[].keep_firing_for` | `string` | yes |  |  |
| `spec.groups[].rules[].labels` | `map<string, string>` |  |  |  |
| `spec.groups[].rules[].annotations` | `map<string, string>` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Kubernetes namespace the PrometheusRule is created in. Typically a
reference to a KubernetesNamespace resource's `spec.name`.

A Prometheus loads rules from the namespaces its rule namespace selector
matches (KubernetesKubePrometheusStack selects every namespace), so the
namespace is where the rules live, not a limit on what they query: a rule's
expression sees every series the Prometheus holds.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.labels

`map<string, string>`

Labels on the PrometheusRule object itself (its metadata.labels), merged
under Planton's identity labels (a key in both keeps the identity value).

These are how a Prometheus with a rule selector decides the rules are its
own: a KubernetesKubePrometheusStack on release_managed_only discovery loads
only objects labelled `release: <its release_name output>`, and a
Prometheus installed any other way selects by whatever its ruleSelector
names. Under the stack's default all_monitors discovery no label is needed.

These are NOT the labels the rules attach to their series or alerts; those
are each group's and each rule's `labels`.

- rule: {"map":{"keys":{"string":{"minLen":"1","maxLen":"317","pattern":"^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?[A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?$"}},"values":{"string":{"maxLen":"63","pattern":"^([A-Za-z0-9]([-A-Za-z0-9_.]{0,61}[A-Za-z0-9])?)?$"}}}}

### spec.annotations

`map<string, string>`

Annotations on the PrometheusRule object itself (its
metadata.annotations): notes for people and tools reading the object, such
as an owning team or a link to the alert's design. They do not reach the
alerts; each alerting rule's `annotations` do.

### spec.groups

`[]KubernetesPrometheusRuleGroup`

The rule groups: the content of one Prometheus rule file. Each group is
evaluated as a unit, sequentially, at its own interval; groups run in
parallel with each other.

Group names are the list's key upstream, so two groups may not share a
name. Recording rules that later rules read belong earlier in the same
group, because a group evaluates its rules in order within one evaluation.

- rule: Each rule group needs a unique name: groups are keyed by name, and the API server refuses two with the same one

### spec.groups[].name

`string` · required

Name of the rule group, unique within the PrometheusRule. It appears in
Prometheus's rules API and UI, and in the `rule_group` label of the
evaluation metrics, so name it for what the rules watch
("api-availability", "postgres-backups").

- rule: {"string":{"minLen":"1"}}

### spec.groups[].interval

`string` · optional (explicit presence)

How often the group's rules are evaluated, as a Prometheus duration
("30s", "1m", "1h30m"). Unset, the evaluating Prometheus's own evaluation
interval applies (KubernetesKubePrometheusStack's default is 30s).

A fast-burn alert wants a short interval; a slow recording rule over long
windows can afford a longer one. An interval shorter than the scrape
interval of the series it reads evaluates the same data twice.

- rule: {"string":{"pattern":"^(0|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$"}}

### spec.groups[].query_offset

`string` · optional (explicit presence)

Delays the evaluation timestamp of this group into the past by this
duration ("1m"), so each evaluation sees samples that arrived late -- the
setting for rules over remote-written or federated series. Requires
Prometheus >= 2.53; Thanos Ruler does not support it. Unset, the
Prometheus's global rule query offset applies (zero by default).

- rule: {"string":{"pattern":"^(0|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$"}}

### spec.groups[].limit

`int32` · optional (explicit presence)

The most alerts an alerting rule, or series a recording rule, in this
group may produce per evaluation. A rule that exceeds it fails that
evaluation instead of flooding Alertmanager or the database, which makes a
limit the cheap guard against a label explosion. Zero (the default) means
no limit. Requires Prometheus >= 2.31 or Thanos Ruler >= 0.24.

- rule: {"int32":{"gte":0}}

### spec.groups[].partial_response_strategy

`string` · optional (explicit presence)

Only Thanos Ruler reads this; Prometheus ignores it. What Thanos Ruler does
when a store returns a partial response during evaluation: "abort" fails
the evaluation, "warn" evaluates with the partial data and logs a warning.
Case-insensitive. Unset, Thanos Ruler's own default applies.

- rule: {"string":{"pattern":"^(?i)(abort|warn)?$"}}

### spec.groups[].labels

`map<string, string>`

Labels added to, or overwriting, every series or alert the group's rules
produce, before they are stored. A rule's own `labels` take precedence
over these. Use them for what every rule in the group shares (the
component the group watches). Requires Prometheus >= 3.0; Thanos Ruler
ignores them.

### spec.groups[].rules

`[]KubernetesPrometheusRuleRule`

The group's rules, evaluated in order. Each is either a recording rule
(`record`) or an alerting rule (`alert`), never both.

- rule: A rule is either a recording rule (record) or an alerting rule (alert): set exactly one of the two
- rule: A recording rule (record) cannot carry for, keep_firing_for or annotations; those belong to alerting rules, and Prometheus refuses the whole rule file when a recording rule has them

### spec.groups[].rules[].record

`string` · optional (explicit presence)

Name of the time series a recording rule writes, set instead of `alert`.
Must be a valid metric name; by convention it reads
level:metric:operations ("job:http_requests:rate5m"), so a reader can tell
a recorded series from a scraped one.

- rule: {"string":{"pattern":"^[a-zA-Z_:][a-zA-Z0-9_:]*$"}}

### spec.groups[].rules[].alert

`string` · required · optional (explicit presence)

Name of the alert an alerting rule raises, set instead of `record`. It
becomes the alert's `alertname` label, so it is what a receiver shows
first: name the symptom a person would act on ("ApiErrorBudgetBurn"), not
the query.

- rule: {"string":{"minLen":"1"}}

### spec.groups[].rules[].expr

`string` · required

The PromQL expression to evaluate. For an alerting rule, every series the
expression returns is an active alert, one per distinct label set; for a
recording rule, its result is written as the `record` series.

Upstream accepts an integer or a string here; a string carries both,
since an integer is itself a valid PromQL expression ("1").

- rule: {"string":{"minLen":"1"}}

### spec.groups[].rules[].for

`string` · optional (explicit presence)

Upstream `for`: how long an alerting rule's expression must keep returning
a series before the alert moves from pending to firing, as a Prometheus
duration ("5m"). Unset or "0", the alert fires on the first evaluation
that returns it. A duration shorter than a few evaluation intervals pages
on a single noisy sample. Alerting rules only. The manifest key is `for`,
as upstream spells it; the proto field is named for_duration because `for`
is a reserved word in the validation language and in several of the
generated SDKs' languages.

- rule: {"string":{"pattern":"^(0|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$"}}

### spec.groups[].rules[].keep_firing_for

`string` · required · optional (explicit presence)

How long an alert keeps firing after its expression stops returning it,
as a Prometheus duration ("10m"). It stops an alert on a flapping
condition from resolving and re-paging every few minutes. Alerting rules
only.

- rule: {"string":{"minLen":"1","pattern":"^(0|(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?)$"}}

### spec.groups[].rules[].labels

`map<string, string>`

Labels added to, or overwriting, the alert's or the recorded series'
labels. On an alerting rule these are what Alertmanager routes on: a
`severity`, and the environment and component the receivers group by.
They take precedence over the group's `labels`. Values may use Go
templating over the alert's labels and value ("{{ $labels.namespace }}").

### spec.groups[].rules[].annotations

`map<string, string>`

Annotations attached to each alert an alerting rule raises: the text a
responder reads. Conventionally a `summary`, a `description` and a
`runbook_url` whose first line is the first action. Values may use Go
templating ("{{ $value | humanizePercentage }} of requests failed").
Alerting rules only.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesPrometheusRule, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.prometheus_rule_name` | `string` | Name of the created PrometheusRule (equals metadata.name). |
| `status.outputs.namespace` | `string` | Namespace the PrometheusRule was created in (the resolved spec.namespace). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |

## See Also

- [Overview](../README.md)
