# KubernetesPrometheusRule Guide

The judgment this guide carries: a PrometheusRule that applies cleanly can still never be evaluated. Which Prometheus loads it is decided by labels on the object, and one malformed rule makes the operator drop every rule in the object. Get the fence and the rule shape right, and give each alert what its receiver and its responder need.

## Which Prometheus loads it

A PrometheusRule is configuration a Prometheus selects; nothing on the object names a Prometheus. A [KubernetesKubePrometheusStack](../kuberneteskubeprometheusstack/GUIDE.md) on its default discovery (`all_monitors`) loads every rule object in every namespace, so on a cluster with one stack you set no label at all.

A cluster with two stacks is where this goes wrong. The second stack, typically a receiver that only stores remote-written series, runs `release_managed_only` discovery so it doesn't re-scrape what the first already scrapes. That stack loads only objects labelled `release: <its release name>` (its `release_name` output). A rule meant for it carries that label in `labels`; a rule without it is loaded only by the first stack. Rules usually belong next to complete data, which is the agent stack on each cluster, not a receiver that holds only what agents chose to send. So the label is the exception, and it's worth a comment in the manifest when you use it.

`labels` and `annotations` here are the object's own metadata. They are not the labels the rules attach to series or alerts; those are each group's and each rule's `labels`. Confusing the two is the common mistake: a `severity` on the object reaches no alert.

## One bad rule silences the whole object

Prometheus refuses an entire rule file when one rule is malformed, and the operator then drops the PrometheusRule from every Prometheus that selected it. The API server accepts the object, the deploy succeeds, and none of its rules evaluate. Two shapes cause most of it, and the spec refuses both before the object is applied:

- a rule that is both `record` and `alert`, or neither;
- a recording rule carrying `for`, `keep_firing_for` or `annotations`.

What the spec can't check is the PromQL in `expr`. Prove a new expression in Prometheus's query page before committing it. After applying, the operator's admission webhook (when enabled) and the Prometheus's `prometheus_rule_evaluation_failures_total` and `prometheus_rule_group_last_evaluation_timestamp_seconds` say whether the group is evaluating.

Because one bad rule costs the whole object, split rules into objects by owner and blast radius: one object per component or service, never one object for the whole platform.

## Alerts that act

An alert is only as good as its route and its text, and both live in the rule:

- **`labels` are the route.** Alertmanager routes on them: a `severity` that the pager route matches, and the environment and component the receivers group by. A group's `labels` hold what every rule in it shares (the component); a rule's `labels` win where they overlap.
- **`annotations` are what the responder reads.** A `summary` in the language of the symptom ("{{ $labels.job }} is burning its error budget 14x too fast"), and a `runbook_url` whose first line is the first action.
- **`for` and `keep_firing_for` set the noise floor.** A `for` shorter than a few evaluation intervals pages on one noisy sample; a `keep_firing_for` stops a flapping condition from resolving and paging again every few minutes.
- **`limit` is the guard against a label explosion.** A rule that suddenly returns thousands of series fails that evaluation instead of flooding Alertmanager.

For burn-rate alerting, record the error ratios first and alert on the recorded series: the recording rules sit earlier in the same group (a group evaluates in order), the alert reads cheap series, and the same series feed the dashboards. The `01-error-budget-burn-alerts` and `02-recording-rules` presets carry that pair; presets ship in the release's `presets.zip` and the repository, not in this reference pack.

## Rules over remote-written series

A Prometheus that evaluates rules over series arriving by remote write sees the newest minute incomplete. Set the group's `query_offset` to cover the senders' delay ("1m"), or each evaluation reads a gap and the alert flaps. Prometheus 2.53 or later is required; Thanos Ruler ignores it.

## On the diagram

The rule draws one edge, to its namespace. The edge to the Prometheus that evaluates it is a label match, not a reference, so the diagram doesn't draw it. A reviewer checks the fence (which stack, which discovery mode, which `release` label) deliberately. The registry prerequisite orders the rule after the stack that installs its CRDs.

## Design rationale

The spec mirrors the upstream PrometheusRule field for field, under the upstream keys, so a rule written for any prometheus-operator reads the same here. The kind adds three things:

- **The envelope.** `namespace` is a reference so composition can draw it. `labels` and `annotations` are routed to the object's metadata because for this resource the object's labels are configuration.
- **The rule-shape checks.** They are the ones Prometheus applies when it loads a rule file, moved forward to before the apply.
- **`for` is the proto field `for_duration`.** `for` is a reserved word in the validation language and in several generated SDKs. Its manifest key stays `for`.

## Parity accounting

- **Pinned:** prometheus-operator v0.94.1 (`monitoring.coreos.com/v1` PrometheusRule), as shipped in kube-prometheus-stack chart 91.8.2.
- **Coverage:** every leaf of the upstream spec is a spec field under its upstream key. That's `groups[]` with `name`, `interval`, `query_offset`, `limit`, `partial_response_strategy` and `labels`, and `rules[]` with `record`, `alert`, `expr`, `for`, `keep_firing_for`, `labels` and `annotations`. Upstream `expr` accepts an integer or a string, and the string field carries both.
- **Composed:** nothing; the kind applies one object.
- **Excluded:** nothing. The object's `status` is written by the operator, not configured.

## Pairs well with

- [KubernetesKubePrometheusStack](../kuberneteskubeprometheusstack/GUIDE.md): installs the CRDs and the Prometheus that evaluates the rules.
- [The observability stack pattern](../../_patterns/observability-stack.md): where rules sit in an agent-and-hub layout.
