# Grafana Dashboard

This preset ships one Grafana dashboard as code: a ConfigMap labeled `grafana_dashboard: "1"` whose single data key is the dashboard's JSON. A `KubernetesGrafana` installed with its dashboard sidecar on (the default) loads it from any namespace within a minute, and the dashboard stays read-only in Grafana, so the committed file is the only way it changes.

## When to Use

- Dashboards that must survive a Grafana restart or rebuild, and change only through review
- A team shipping its own dashboard next to its workload, without editing the Grafana resource
- An Infra Chart that declares its operator screens beside the components they watch

## Key Configuration Choices

- **`labels.grafana_dashboard`**: the label the sidecar watches. Its value is not read; `"1"` is the convention.
- **The data key ends in `.json`** (`capacity.json`): the sidecar writes each key as a file, and Grafana loads only `.json` files. The key also becomes the dashboard's `provisionedExternalId` in Grafana's API, which is how a checker traces a dashboard back to the ConfigMap that shipped it.
- **`uid` pinned in the JSON**: the dashboard's address (`/d/capacity`) and the key alerts and runbooks link to. Never let Grafana derive it.
- **The datasource named by `uid`** (`prometheus`): pin the same uid on the `KubernetesGrafana` datasource, and the dashboard survives that datasource moving to another Prometheus.
- **`schemaVersion` matching the running Grafana** (42 on Grafana 13.1): a lower version is migrated in the browser on every load, so Grafana's copy would drift from the file.
- **`editable: false`**: hides the edit affordances; Grafana already refuses to save over a provisioned dashboard ("Cannot save provisioned dashboard"), even for an Admin.
- **Pretty-printed JSON in an Infra Chart**: chart templates are rendered, and a Prometheus legend format with double braces collides with the engine. Pretty-printing keeps closing braces apart; name series with a `displayName: "${__field.labels.<label>}"` field override instead of a legend format.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-grafana-namespace>` | Any namespace: the sidecar searches all of them. In an Infra Chart, reference the chart's `KubernetesNamespace` with `valueFrom` instead of a literal | Your namespace management |

Also replace the one panel with your dashboard's panels. Each panel's `description` is the question it answers, and the dashboard's `title` is the question the whole screen answers.

## Related Presets

- **01-app-config**: mutable configuration with a properties file
- **02-immutable-versioned**: versioned, immutable configuration
