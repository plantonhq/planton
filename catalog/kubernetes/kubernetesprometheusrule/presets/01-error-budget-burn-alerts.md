# Error-Budget Burn Alerts

Two multi-window burn-rate alerts for one service's availability objective. The fast one pages when the service would exhaust a 30-day error budget in about two days. The slow one opens a ticket when it would do so in about five days. Each fires only when a long window and a short window agree, so an alert means the burn is both significant and still happening.

## When to Use

- A service has an availability objective (for example 99.9%, an error budget of `0.001`) and you want to be woken only when customers are measurably hurt.
- You are replacing threshold alerts ("error rate above 5%") that page on every short blip.

## How It Works

The alerts read the recorded error ratios from the **Error-Ratio Recording Rules** preset. The fast-burn alert fires when both the 1-hour and the 5-minute ratio exceed 14.4 times the budget, held for 2 minutes. The slow-burn alert fires when both the 6-hour and the 30-minute ratio exceed 6 times the budget, held for 15 minutes. The short window makes each alert resolve soon after the burn stops; `keep_firing_for` keeps a flapping burn from paging again every few minutes.

## Key Configuration Choices

- **`severity: page` and `severity: ticket`.** Alertmanager routes on these, so match them to your routes: only the fast burn should reach a pager.
- **`component` on the group.** Every alert in the group carries it, for receivers that group by component.
- **`runbook_url`.** Point it at a page whose first line is the first action.
- **`limit: 20`.** One alert per job is expected; a selector mistake that matches thousands of series fails the evaluation instead of flooding Alertmanager.

## Prerequisites

- The recording rules from the **Error-Ratio Recording Rules** preset, loaded by the same Prometheus.
- A Prometheus that selects this object, with an Alertmanager whose routes match the `severity` values (`KubernetesKubePrometheusStack`).
- The target namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<service>` / `<Service>` | Name of the service, lowercase in names and capitalized in alert names. |
| `<namespace>` | Namespace the object lives in. |
| `<job>` | The `job` label of the service's scrape target. |
| `<error_budget>` | One minus the objective: `0.001` for 99.9%. |
| `<runbook_url>` | The runbook a responder opens first. |
